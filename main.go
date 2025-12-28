package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/matrix-mud/pkg/achievements"
	"github.com/yourusername/matrix-mud/pkg/accessibility"
	"github.com/yourusername/matrix-mud/pkg/analytics"
	"github.com/yourusername/matrix-mud/pkg/chat"
	"github.com/yourusername/matrix-mud/pkg/cooldown"
	"github.com/yourusername/matrix-mud/pkg/faction"
	"github.com/yourusername/matrix-mud/pkg/game"
	"github.com/yourusername/matrix-mud/pkg/help"
	"github.com/yourusername/matrix-mud/pkg/leaderboard"
	"github.com/yourusername/matrix-mud/pkg/logging"
	"github.com/yourusername/matrix-mud/pkg/metrics"
	"github.com/yourusername/matrix-mud/pkg/party"
	"github.com/yourusername/matrix-mud/pkg/pvp"
	"github.com/yourusername/matrix-mud/pkg/quest"
	"github.com/yourusername/matrix-mud/pkg/ratelimit"
	"github.com/yourusername/matrix-mud/pkg/readline"
	"github.com/yourusername/matrix-mud/pkg/session"
	"github.com/yourusername/matrix-mud/pkg/trade"
	"github.com/yourusername/matrix-mud/pkg/training"
	"github.com/yourusername/matrix-mud/pkg/tutorial"
	"github.com/yourusername/matrix-mud/pkg/events"
	"github.com/yourusername/matrix-mud/pkg/validation"
	"github.com/yourusername/matrix-mud/pkg/world"
)

// Telnet IAC (Interpret As Command) codes for echo suppression
const (
	TelnetIAC  = 255 // Interpret As Command
	TelnetWILL = 251 // Will perform option
	TelnetWONT = 252 // Won't perform option
	TelnetECHO = 1   // Echo option
)

var (
	userMutex       sync.Mutex
	authLimiter     = ratelimit.New(5, 1*time.Minute)    // 5 auth attempts per minute
	cmdLimiter      = ratelimit.New(10, 1*time.Second)   // 10 commands per second per player
	sessionManager  = session.NewManager()               // Player session management
	gameClock       = world.DefaultClock()               // Day/night cycle
	playerHistories = make(map[string]*readline.History) // Per-player command history
	historyMutex    sync.RWMutex
)

// getPlayerHistory returns or creates a command history for a player
func getPlayerHistory(playerName string) *readline.History {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	if h, ok := playerHistories[playerName]; ok {
		return h
	}
	h := readline.NewHistory(50) // Keep last 50 commands
	playerHistories[playerName] = h
	return h
}

// init starts background goroutines for maintenance tasks.
func init() {
	// Initialize structured logging
	logging.Init(Config.LogPretty, Config.LogLevel)

	// Start rate limiter cleanup goroutine to prevent memory leaks
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			authLimiter.CleanupOldEntries()
			logging.Info().Msg("Rate limiter cleanup completed")
		}
	}()
}

func authenticate(c *Client, name string) bool {
	// Apply rate limiting to prevent brute force attacks
	if !authLimiter.Allow(name) {
		c.Write(Red + "Too many authentication attempts. Try again later.\r\n" + Reset)
		logging.Warn().Str("user", name).Msg("Rate limit exceeded")
		time.Sleep(3 * time.Second) // Add delay for rate-limited clients
		return false
	}

	userMutex.Lock()
	defer userMutex.Unlock()

	// Load existing user database (stores password hashes)
	users := make(map[string]string)
	file, err := os.ReadFile("data/users.json")
	if err != nil && !os.IsNotExist(err) {
		logging.Error().Err(err).Msg("Failed to read users.json")
		c.Write(Red + "Authentication error.\r\n" + Reset)
		return false
	}

	if file != nil {
		if err := json.Unmarshal(file, &users); err != nil {
			logging.Error().Err(err).Msg("Failed to parse users.json")
			c.Write(Red + "Authentication error.\r\n" + Reset)
			return false
		}
	}

	cleanName := strings.ToLower(name)

	if storedHash, exists := users[cleanName]; exists {
		// Existing user - verify password with bcrypt
		c.Write("Password: ")
		pass, err := c.readPassword()
		if err != nil {
			logging.Error().Err(err).Str("user", cleanName).Msg("Failed to read password")
			return false
		}

		// Compare password with stored bcrypt hash
		err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(pass))
		if err == nil {
			logging.Info().Str("user", cleanName).Msg("Authentication successful")
			return true
		}

		c.Write(Red + "Access Denied.\r\n" + Reset)
		logging.Warn().Str("user", cleanName).Msg("Failed authentication attempt")
		return false
	} else {
		// New user - create account with bcrypt hashed password
		c.Write("New identity detected. Set a password: ")
		pass, err := c.readPassword()
		if err != nil {
			logging.Error().Err(err).Str("user", cleanName).Msg("Failed to read password for new user")
			return false
		}

		// Enforce minimum password length of 8 characters
		if len(pass) < 8 {
			c.Write("Password must be at least 8 characters.\r\n")
			return false
		}

		// Hash password with bcrypt
		hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			logging.Error().Err(err).Str("user", cleanName).Msg("Failed to hash password")
			c.Write(Red + "Error creating account.\r\n" + Reset)
			return false
		}

		// Store hashed password
		users[cleanName] = string(hash)
		data, _ := json.MarshalIndent(users, "", "  ")
		if err := os.WriteFile("data/users.json", data, 0600); err != nil { // Owner read/write only
			logging.Error().Err(err).Str("user", cleanName).Msg("Failed to save user")
			c.Write(Red + "Error creating account.\r\n" + Reset)
			return false
		}

		c.Write("Identity created.\r\n")
		logging.Info().Str("user", cleanName).Msg("New user created")
		return true
	}
}

func chooseClass(c *Client, p *Player) {
	c.Write(Clear + Green + "Residual Self Image not found.\r\n" + Reset)
	c.Write("How do you see yourself in the Construct?\r\n\r\n")
	c.Write("1. " + White + "The Hacker" + Reset + " (Low HP, High Tech. Starts with Cyberdeck)\r\n")
	c.Write("2. " + White + "The Rebel" + Reset + "  (High HP, Strong. Starts with Combat Boots)\r\n")
	c.Write("3. " + White + "The Operator" + Reset + " (Balanced. Starts with Pilot Shades)\r\n\r\n")
	c.Write("Choose [1-3]: ")

	for {
		choice, _ := c.reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		switch choice {
		case "1":
			p.Class = "Hacker"
			p.MaxHP = 15
			p.HP = 15
			p.Strength = 10
			p.BaseAC = 10
			p.Inventory = append(p.Inventory, &Item{ID: "deck", Name: "Cyberdeck", Description: "A portable hacking unit.", Slot: "hand", Damage: 2})
			return
		case "2":
			p.Class = "Rebel"
			p.MaxHP = 30
			p.HP = 30
			p.Strength = 14
			p.BaseAC = 10
			p.Inventory = append(p.Inventory, &Item{ID: "boots", Name: "Combat Boots", Description: "Heavy boots.", Slot: "body", AC: 2})
			return
		case "3":
			p.Class = "Operator"
			p.MaxHP = 20
			p.HP = 20
			p.Strength = 12
			p.BaseAC = 12
			p.Inventory = append(p.Inventory, &Item{ID: "shades", Name: "Pilot Shades", Description: "Cool sunglasses.", Slot: "head", AC: 1})
			return
		default:
			c.Write("Invalid choice. Choose [1-3]: ")
		}
	}
}

// Client represents a connected player's network connection and I/O handler.
// Each client connection runs in its own goroutine and maintains a buffered
// reader for efficient line-based command input.
type Client struct {
	conn   net.Conn      // TCP connection to the client
	reader *bufio.Reader // Buffered reader for line reading
}

// Write sends a message to the client over the TCP connection.
// Messages are sent as raw bytes. The caller should include appropriate
// line endings (\r\n for telnet compatibility).
func (c *Client) Write(msg string) {
	c.conn.Write([]byte(msg))
}

// suppressEcho sends telnet IAC WILL ECHO to suppress client-side echo.
// This should be called before reading sensitive input like passwords.
func (c *Client) suppressEcho() {
	c.conn.Write([]byte{TelnetIAC, TelnetWILL, TelnetECHO})
}

// resumeEcho sends telnet IAC WONT ECHO to resume normal client-side echo.
// This should be called after reading sensitive input.
func (c *Client) resumeEcho() {
	c.conn.Write([]byte{TelnetIAC, TelnetWONT, TelnetECHO})
}

// readPassword reads a password with echo suppression for security.
// Returns the password string (trimmed) and any error.
func (c *Client) readPassword() (string, error) {
	c.suppressEcho()
	defer func() {
		c.resumeEcho()
		c.Write("\r\n") // Add newline since echo was suppressed
	}()

	pass, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(pass), nil
}

func main() {
	// Create server context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connection limiter semaphore
	connSemaphore := make(chan struct{}, MaxConnections)

	// Use configured port
	listenAddr := ":" + Config.TelnetPort
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logging.Fatal().Err(err).Str("addr", listenAddr).Msg("Failed to start telnet server")
	}

	world := NewWorld()

	// Start event bus for Discord/webhook integration
	events.GlobalEventBus.Start()

	go startWebServer(world)
	go startAdminServer(world)

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				world.Update()
			}
		}
	}()

	logging.Info().
		Str("version", Version).
		Str("telnet_port", Config.TelnetPort).
		Str("web_port", Config.WebPort).
		Str("admin_addr", Config.AdminBindAddr).
		Int("max_connections", MaxConnections).
		Msg("Matrix Construct Server started")

	// Setup graceful shutdown handler
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Run accept loop in goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// Check if we're shutting down
				select {
				case <-ctx.Done():
					return
				default:
					logging.Error().Err(err).Msg("Accept error")
					continue
				}
			}

			// Try to acquire connection slot
			select {
			case connSemaphore <- struct{}{}:
				// Got slot, handle connection
				go func(c net.Conn) {
					defer func() { <-connSemaphore }() // Release slot when done
					handleConnection(ctx, c, world)
				}(conn)
			default:
				// Server full
				conn.Write([]byte("Server full. Please try again later.\r\n"))
				conn.Close()
				logging.Warn().Msg("Connection rejected: server at max capacity")
			}
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	logging.Info().Msg("Shutdown signal received, canceling context...")

	// Cancel context to signal all goroutines
	cancel()

	logging.Info().Msg("Saving all player data...")

	// Save all connected players
	world.mutex.RLock()
	playerCount := len(world.Players)
	for _, player := range world.Players {
		if player != nil {
			world.SavePlayer(player)
			if player.Conn != nil {
				player.Conn.Write("\r\n" + Yellow + "Server shutting down. Your progress has been saved.\r\n" + Reset)
			}
		}
	}
	world.mutex.RUnlock()

	// Save world state
	world.SaveWorld()

	logging.Info().Int("players_saved", playerCount).Msg("Graceful shutdown complete")
	listener.Close()
}

func handleConnection(ctx context.Context, conn net.Conn, world *World) {
	client := &Client{conn: conn, reader: bufio.NewReader(conn)}
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	connLog := logging.WithConnection(remoteAddr)

	// Check if this is a WebSocket bridge connection (from localhost)
	// WebSocket clients already have their own intro, so skip the telnet intro
	isWebSocket := strings.HasPrefix(remoteAddr, "127.0.0.1:") || strings.HasPrefix(remoteAddr, "[::1]:")

	// Set initial connection timeout for login (extend for intro if needed)
	if isWebSocket {
		conn.SetDeadline(time.Now().Add(ConnectionTimeout))
	} else {
		conn.SetDeadline(time.Now().Add(ConnectionTimeout + 30*time.Second))

		// Play Matrix rain intro animation ONLY for direct telnet connections
		introConfig := game.IntroConfig{
			Width:        80,
			Height:       24,
			RainFrames:   40, // ~2 seconds of pure rain
			RevealFrames: 80, // ~4 seconds of reveal
			FrameDelay:   50 * time.Millisecond,
			FinalPause:   2 * time.Second,
		}
		game.PlayIntro(func(s string) { client.Write(s) }, introConfig)
	}

	// Reset timeout for login phase
	conn.SetDeadline(time.Now().Add(ConnectionTimeout))

	client.Write(Green + "Wake up...\r\n" + Reset)
	client.Write("Identify yourself: ")
	name, err := client.reader.ReadString('\n')
	if err != nil {
		connLog.Debug().Err(err).Msg("Connection closed during login")
		return
	}

	// Sanitize and validate input
	name = validation.SanitizeInput(name)
	if name == "" {
		client.Write("Identification required.\r\n")
		return
	}

	// Validate username format
	if !validation.ValidateUsername(name) {
		client.Write(Red + "Invalid username. Use 3-20 alphanumeric characters (and underscores).\r\n" + Reset)
		connLog.Warn().Str("attempted_name", name).Msg("Invalid username attempt")
		return
	}

	if !authenticate(client, name) {
		return
	}

	// Check for reconnectable session
	cleanName := strings.ToLower(name)
	if sess := sessionManager.Reconnect(cleanName); sess != nil {
		client.Write(Green + "Session restored. Welcome back to the Matrix.\r\n" + Reset)
		connLog.Info().Str("player", cleanName).Msg("Player reconnected to existing session")
	}

	player := world.LoadPlayer(name, client)
	if player.Class == "" {
		chooseClass(client, player)
		world.SavePlayer(player)
	}

	// Auto-join default chat channels for all players
	chat.GlobalChat.AutoJoinDefaultChannels(player.Name)

	world.mutex.Lock()
	world.Players[client] = player
	world.mutex.Unlock()

	// Create or update session
	sessionManager.CreateSession(player.Name, player.RoomID, player.HP, player.MP)

	defer func() {
		world.SavePlayer(player)
		world.mutex.Lock()
		delete(world.Players, client)
		world.mutex.Unlock()
		// Mark session as disconnected (allows reconnect within 30 min)
		sessionManager.Disconnect(player.Name)
		analytics.EndSession(player.Name)
		metrics.DecrPlayers()
	}()

	// Start analytics session
	analytics.StartSession(player.Name)
	metrics.IncrPlayers()

	// Show MOTD
	if motd := world.GetMOTD(); motd != "" {
		client.Write(Matrixify(motd))
	}

	client.Write(Matrixify(world.Look(player, "")))
	client.Write("> ")

	// Switch to idle timeout for active session
	conn.SetDeadline(time.Now().Add(IdleTimeout))

	// Get or create command history for this player
	history := getPlayerHistory(strings.ToLower(player.Name))
	rl := readline.NewReader(conn, history, "> ")

	for {
		// Check for server shutdown
		select {
		case <-ctx.Done():
			client.Write("\r\n" + Yellow + "Server shutting down...\r\n" + Reset)
			return
		default:
		}

		input, err := rl.ReadLine()
		if err != nil {
			// Could be timeout or disconnect
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				client.Write("\r\n" + Yellow + "Connection timed out due to inactivity.\r\n" + Reset)
				logging.Info().Str("player", player.Name).Dur("idle_timeout", IdleTimeout).Msg("Player timed out")
			}
			break
		}

		// Reset idle timeout on each valid input
		conn.SetDeadline(time.Now().Add(IdleTimeout))

		input = strings.TrimSpace(input)
		if input == "" {
			client.Write("> ")
			continue
		}

		// Rate limit commands (10 per second per player)
		if !cmdLimiter.Allow(player.Name) {
			client.Write(Yellow + "Slow down! Too many commands.\r\n" + Reset + "> ")
			metrics.RecordRateLimited()
			continue
		}

		cmd, arg := parseCommand(input)
		parts := strings.Fields(input) // Keep for commands that need positional access

		// Record command for metrics
		metrics.RecordCommand(cmd)

		// --- SPECIAL STATE HANDLING ---
		// Handle dialogue numeric input and instance state before normal commands

		// 1. Check if player is in dialogue (numeric input for choices)
		if IsInDialogue(player.Name) {
			// Check if input is a number (dialogue choice)
			if choice, err := strconv.Atoi(input); err == nil {
				response := HandleDialogueChoice(world, player, strconv.Itoa(choice))
				client.Write(Matrixify(response) + "> ")
				continue
			}
			// "bye" also works to end dialogue from anywhere
			if strings.ToLower(input) == "bye" {
				response := HandleByeCommand(player)
				client.Write(response + "> ")
				continue
			}
			// Any other command while in dialogue shows hint
			client.Write("You're in a conversation. Enter a number to choose, or 'bye' to end.\r\n> ")
			continue
		}

		// 2. Check if player is in an instance
		if IsInInstance(player.Name) {
			// Handle directional movement in instance
			if isDirection := map[string]bool{
				"north": true, "n": true,
				"south": true, "s": true,
				"east": true, "e": true,
				"west": true, "w": true,
				"up": true, "u": true,
				"down": true, "dn": true,
			}[cmd]; isDirection {
				result, ok := HandleInstanceMove(player, cmd)
				if ok {
					client.Write(result + "> ")
					continue
				}
				// If not ok, fall through to normal command processing
			}
			// Handle attack in instance
			if cmd == "kill" || cmd == "k" || cmd == "attack" || cmd == "a" {
				result, _ := HandleInstanceAttack(world, player, arg)
				client.Write(Matrixify(result) + "> ")
				continue
			}
			// Handle look in instance
			if cmd == "look" || cmd == "l" {
				result, ok := HandleInstanceLook(player)
				if ok {
					client.Write(result + "> ")
					continue
				}
				// If not ok, fall through to normal look
			}
		}

		var response string
		switch cmd {
		case "look", "l":
			lookResult := world.Look(player, arg)
			// Add time-of-day atmosphere to room descriptions (not when looking at specific targets)
			if arg == "" {
				lookResult = fmt.Sprintf("%s%s%s\r\n%s", Cyan, gameClock.AmbientDescription(), Reset, lookResult)
			}
			response = Matrixify(lookResult)
		case "north", "n":
			response = Matrixify(world.MovePlayer(player, "north"))
		case "south", "s":
			response = Matrixify(world.MovePlayer(player, "south"))
		case "west", "w":
			response = Matrixify(world.MovePlayer(player, "west"))
		case "east", "e":
			response = Matrixify(world.MovePlayer(player, "east"))
		case "up", "u":
			response = Matrixify(world.MovePlayer(player, "up"))
		case "down", "dn":
			response = Matrixify(world.MovePlayer(player, "down"))
		case "get", "g":
			response = Matrixify(world.GetItem(player, arg))
		case "drop", "d":
			response = Matrixify(world.DropItem(player, arg))
		case "inv", "i":
			response = Matrixify(world.ShowInventory(player))
		case "score", "sc", "balance", "bal":
			response = Matrixify(world.ShowScore(player))
		case "kill", "k", "attack", "a":
			// Check if player is in PvP arena first
			if arena := pvp.GlobalPvP.GetPlayerArena(player.Name); arena != nil {
				if arg == "" {
					response = "Attack who?\r\n"
				} else if result, err := pvp.GlobalPvP.AttackPlayer(arena.ID, player.Name, arg); err != nil {
					response = err.Error() + "\r\n"
				} else {
					response = result + "\r\n"
				}
			} else {
				// Normal combat
				response = Matrixify(world.StartCombat(player, arg))
			}
		case "cast", "c":
			skillParts := strings.Fields(arg)
			if len(skillParts) > 0 {
				skill := strings.ToLower(skillParts[0])
				target := ""
				if len(skillParts) > 1 {
					target = strings.Join(skillParts[1:], " ")
				}
				// Check cooldown before casting
				if !cooldown.GlobalCD.IsReady(player.Name, skill) {
					remaining := cooldown.GlobalCD.TimeRemaining(player.Name, skill)
					response = fmt.Sprintf("%s%s is on cooldown (%.1fs remaining)%s\r\n", Yellow, skill, remaining.Seconds(), Reset)
				} else {
					response = Matrixify(world.CastSkill(player, skill, target))
					// Only trigger cooldown if cast succeeded (response doesn't contain error indicators)
					if !strings.Contains(response, "don't know") && !strings.Contains(response, "not in combat") {
						cooldown.GlobalCD.Use(player.Name, skill)
					}
				}
			} else {
				response = "Cast what?\r\n"
			}
		case "flee", "stop":
			response = Matrixify(world.StopCombat(player))
		case "wear", "wield", "equip":
			response = Matrixify(world.WearItem(player, arg))
		case "remove", "unequip":
			response = Matrixify(world.RemoveItem(player, arg))
		case "use", "eat":
			response = Matrixify(world.UseItem(player, arg))
		case "say":
			if len(parts) > 1 {
				msg := strings.Join(parts[1:], " ")
				broadcast(world, player, msg)
				npcResp := world.HandleSay(player, msg)
				if npcResp != "" {
					broadcast(world, player, npcResp)
					response = "You spoke.\r\n" + npcResp
				} else {
					response = "You spoke.\r\n"
				}
			}
		case "gossip", "chat":
			if len(parts) > 1 {
				msg := strings.Join(parts[1:], " ")
				world.Gossip(player, msg)
				response = ""
			} else {
				response = "Gossip what?\r\n"
			}
		case "tell", "whisper", "t":
			if len(parts) > 2 {
				target := parts[1]
				msg := strings.Join(parts[2:], " ")
				response = Matrixify(world.Tell(player, target, msg))
			} else {
				response = "Tell who what?\r\n"
			}
		case "list", "vendor":
			response = Matrixify(world.ListGoods(player))
		case "buy":
			response = Matrixify(world.BuyItem(player, arg))
		case "sell":
			response = Matrixify(world.SellItem(player, arg))
		case "give":
			giveParts := strings.Fields(arg)
			if len(giveParts) >= 2 {
				targetName := giveParts[len(giveParts)-1]
				itemName := strings.Join(giveParts[:len(giveParts)-1], " ")
				response = Matrixify(world.GiveItem(player, itemName, targetName))
			} else {
				response = "Give what to whom?\r\n"
			}
		case "deposit":
			response = Matrixify(world.DepositItem(player, arg))
		case "withdraw":
			response = Matrixify(world.WithdrawItem(player, arg))
		case "storage", "bank":
			response = Matrixify(world.ShowStorage(player))

		// --- CRAFTING COMMANDS ---
		case "recipes":
			response = Matrixify(world.ListRecipes(player))
		case "craft":
			response = Matrixify(world.Craft(player, arg))
		case "repair":
			response = Matrixify(world.RepairItem(player, arg))

		case "who":
			response = Matrixify(world.ListPlayers())

		case "time":
			response = fmt.Sprintf("%s%s%s\r\n%s%s%s\r\n", Cyan, gameClock.FormatTimeDisplay(), Reset, Green, gameClock.TimeString(), Reset)

		case "cooldowns", "cd":
			cds := cooldown.GlobalCD.GetAllCooldowns(player.Name)
			if len(cds) == 0 {
				response = "All abilities ready.\r\n"
			} else {
				response = "Active cooldowns:\r\n"
				for ability, remaining := range cds {
					response += fmt.Sprintf("  %s: %.1fs\r\n", ability, remaining.Seconds())
				}
			}

		// --- PARTY COMMANDS ---
		case "party":
			response = Matrixify(handlePartyCommand(player, arg))
		case "invite":
			response = Matrixify(handlePartyInvite(player, arg))
		case "accept":
			response = Matrixify(handlePartyAccept(player, arg))
		case "decline":
			response = Matrixify(handlePartyDecline(player, arg))

		case "help", "?":
			response = Matrixify(formatHelp(arg))
		case "teleport":
			response = Matrixify(world.Teleport(player, arg))

		// --- AWAKENING & MATRIX COMMANDS ---
		case "take":
			// Handle "take red" or "take blue" for pills, otherwise fall through to use
			if arg == "red" || arg == "blue" || arg == "red pill" || arg == "blue pill" {
				pillColor := strings.Split(arg, " ")[0]
				response = world.TakePill(player, pillColor)
			} else {
				response = Matrixify(world.UseItem(player, arg))
			}
		case "abilities", "skills":
			response = world.ShowAbilities(player)
		case "see_code", "seecode", "code":
			response = world.SeeCode(player)
		case "focus":
			response = world.Focus(player)

		// --- PHONE BOOTH COMMANDS ---
		case "call":
			response = world.CallPhone(player, arg)
		case "phones", "phonebook":
			response = world.ListPhones(player)
		case "jackout", "jack":
			response = world.JackOut(player)

		// --- QUEST COMMANDS ---
		case "quests", "journal", "quest":
			if arg == "" {
				response = quest.GlobalQuests.GetActiveQuests(player.Name)
			} else {
				response = handleQuestCommand(player, arg)
			}
		case "completed":
			completed := quest.GlobalQuests.GetCompletedQuests(player.Name)
			if len(completed) == 0 {
				response = "You have not completed any quests yet.\r\n"
			} else {
				response = "=== COMPLETED QUESTS ===\r\n"
				for _, name := range completed {
					response += "  [X] " + name + "\r\n"
				}
			}

		// --- FACTION COMMANDS ---
		case "faction", "factions":
			response = handleFactionCommand(player, arg)
		case "reputation", "rep":
			response = handleReputationCommand(player)

		// --- ACHIEVEMENT COMMANDS ---
		case "achievements", "ach":
			response = handleAchievementsCommand(player, arg)
		case "title", "titles":
			response = handleTitleCommand(player, arg)

		// --- LEADERBOARD COMMANDS ---
		case "rankings", "leaderboard", "top":
			response = handleLeaderboardCommand(arg)
		case "stats":
			response = handleStatsCommand(player)

		// --- TRAINING COMMANDS ---
		case "train", "training":
			response = handleTrainingCommand(player, arg)
		case "programs":
			response = handleProgramsCommand()
		case "challenges":
			response = handleChallengesCommand()

		// --- BUILDER COMMANDS ---
		case "generate":
			genParts := strings.Fields(arg)
			if len(genParts) >= 3 && genParts[0] == "city" {
				rows, _ := strconv.Atoi(genParts[1])
				cols, _ := strconv.Atoi(genParts[2])
				if rows > 0 && cols > 0 {
					response = Matrixify(world.GenerateCity(player, rows, cols))
				} else {
					response = "Invalid size.\r\n"
				}
			} else {
				response = "Usage: generate city [rows] [cols]\r\n"
			}
		case "dig":
			if len(parts) >= 3 {
				response = Matrixify(world.Dig(player, parts[1], strings.Join(parts[2:], " ")))
			} else {
				response = "Usage: dig [dir] [name]\r\n"
			}
		case "create":
			if len(parts) >= 3 {
				response = Matrixify(world.CreateEntity(player, parts[1], parts[2]))
			} else {
				response = "Usage: create [item|npc] [id]\r\n"
			}
		case "delete", "del":
			if len(parts) >= 2 {
				response = Matrixify(world.DeleteEntity(player, arg))
			} else {
				response = "Delete what?\r\n"
			}
		case "edit":
			if len(parts) >= 3 {
				field := parts[1]
				val := strings.Join(parts[2:], " ")
				response = Matrixify(world.EditRoom(player, field, val))
			} else {
				response = "Usage: edit desc [text]\r\n"
			}
		case "save":
			if arg == "world" {
				world.SaveWorld()
				response = "World saved to disk.\r\n"
			} else {
				response = "Save what?\r\n"
			}

		// --- QOL COMMANDS ---
		case "brief":
			player.BriefMode = !player.BriefMode
			if player.BriefMode {
				response = "Brief mode ON - room descriptions shortened.\r\n"
			} else {
				response = "Brief mode OFF - full room descriptions.\r\n"
			}
			world.SavePlayer(player)

		case "theme":
			if arg == "" {
				response = fmt.Sprintf("Current theme: %s\r\nAvailable: green, amber, white, none\r\nUsage: theme <name>\r\n", player.ColorTheme)
			} else {
				switch strings.ToLower(arg) {
				case "green", "amber", "white", "none":
					player.ColorTheme = strings.ToLower(arg)
					response = fmt.Sprintf("Color theme set to: %s\r\n", player.ColorTheme)
					world.SavePlayer(player)
				default:
					response = "Unknown theme. Available: green, amber, white, none\r\n"
				}
			}

		case "recall":
			response = world.Recall(player)

		// --- CHAT CHANNEL COMMANDS ---
		case "channels":
			// List available channels
			response = chat.GlobalChat.ListChannels(player.Name)

		case "/join", "join":
			// Join a chat channel
			if arg == "" {
				response = "Usage: /join <channel>\r\nAvailable: global, trade, help\r\n"
			} else if err := chat.GlobalChat.JoinChannel(player.Name, arg); err != nil {
				response = err.Error() + "\r\n"
			} else {
				response = fmt.Sprintf("Joined channel: %s\r\n", arg)
			}

		case "/leave", "leave":
			// Leave a chat channel
			if arg == "" {
				response = "Usage: /leave <channel>\r\n"
			} else if err := chat.GlobalChat.LeaveChannel(player.Name, arg); err != nil {
				response = err.Error() + "\r\n"
			} else {
				response = fmt.Sprintf("Left channel: %s\r\n", arg)
			}

		case "/g", "/global":
			// Send to global channel
			if arg == "" {
				response = "Usage: /g <message>\r\n"
			} else if recipients, err := chat.GlobalChat.SendMessage(player.Name, "global", arg); err != nil {
				response = err.Error() + "\r\n"
			} else {
				// Broadcast to recipients
				msg := chat.FormatMessage(chat.Message{
					Sender:    player.Name,
					Content:   arg,
					Timestamp: time.Now(),
				}, "Global")
				broadcastChatMessage(world, msg, recipients)
				response = "" // Don't echo to sender
			}

		case "/t", "/trade":
			// Send to trade channel
			if arg == "" {
				response = "Usage: /t <message>\r\n"
			} else if recipients, err := chat.GlobalChat.SendMessage(player.Name, "trade", arg); err != nil {
				response = err.Error() + "\r\n"
			} else {
				msg := chat.FormatMessage(chat.Message{
					Sender:    player.Name,
					Content:   arg,
					Timestamp: time.Now(),
				}, "Trade")
				broadcastChatMessage(world, msg, recipients)
				response = ""
			}

		case "/h", "/help":
			// Send to help channel (use arg2 to avoid conflict with help command)
			if arg == "" {
				response = "Usage: /h <message>\r\n"
			} else if recipients, err := chat.GlobalChat.SendMessage(player.Name, "help", arg); err != nil {
				response = err.Error() + "\r\n"
			} else {
				msg := chat.FormatMessage(chat.Message{
					Sender:    player.Name,
					Content:   arg,
					Timestamp: time.Now(),
				}, "Help")
				broadcastChatMessage(world, msg, recipients)
				response = ""
			}

		case "/chat":
			// Send to specific channel: /chat <channel> <message>
			parts := strings.Fields(arg)
			if len(parts) < 2 {
				response = "Usage: /chat <channel> <message>\r\n"
			} else {
				channelID := parts[0]
				content := strings.Join(parts[1:], " ")
				if recipients, err := chat.GlobalChat.SendMessage(player.Name, channelID, content); err != nil {
					response = err.Error() + "\r\n"
				} else {
					channel := chat.GlobalChat.GetChannel(channelID)
					msg := chat.FormatMessage(chat.Message{
						Sender:    player.Name,
						Content:   content,
						Timestamp: time.Now(),
					}, channel.Name)
					broadcastChatMessage(world, msg, recipients)
					response = ""
				}
			}

		// --- TRADE & AUCTION COMMANDS ---
		case "trade":
			// Direct player trading
			parts := strings.Fields(arg)
			if len(parts) == 0 {
				// Show trade status
				if t := trade.GlobalTrade.GetTrade(player.Name); t != nil {
					response = trade.GlobalTrade.FormatTrade(t, player.Name)
				} else {
					response = "No active trade.\r\nUsage: trade request <player>, trade accept, trade decline, trade add/remove/money/confirm/cancel\r\n"
				}
			} else {
				subCmd := strings.ToLower(parts[0])
				switch subCmd {
				case "request":
					if len(parts) < 2 {
						response = "Usage: trade request <player>\r\n"
					} else if t, err := trade.GlobalTrade.InitiateTrade(player.Name, parts[1]); err != nil {
						response = err.Error() + "\r\n"
					} else {
						response = fmt.Sprintf("Trade request sent to %s.\r\n", parts[1])
						// Notify other player if online
						for _, p := range world.Players {
							if strings.ToLower(p.Name) == strings.ToLower(parts[1]) && p.Conn != nil {
								p.Conn.Write(fmt.Sprintf("\r\n%s%s has requested a trade with you.\r\nType 'trade accept' to begin.\r\n> ", Cyan, player.Name))
								break
							}
						}
						_ = t // Trade initiated
					}
				case "accept":
					if err := trade.GlobalTrade.AcceptTrade(player.Name); err != nil {
						response = err.Error() + "\r\n"
					} else {
						response = trade.GlobalTrade.FormatTrade(trade.GlobalTrade.GetTrade(player.Name), player.Name)
					}
				case "decline":
					if err := trade.GlobalTrade.DeclineTrade(player.Name); err != nil {
						response = err.Error() + "\r\n"
					} else {
						response = "Trade declined.\r\n"
					}
				case "cancel":
					if err := trade.GlobalTrade.CancelTrade(player.Name); err != nil {
						response = err.Error() + "\r\n"
					} else {
						response = "Trade canceled.\r\n"
					}
				case "add":
					if len(parts) < 2 {
						response = "Usage: trade add <item>\r\n"
					} else {
						itemName := strings.Join(parts[1:], " ")
						// Find item in inventory
						var item *Item
						for _, i := range player.Inventory {
							if strings.Contains(strings.ToLower(i.Name), strings.ToLower(itemName)) {
								item = i
								break
							}
						}
						if item == nil {
							response = "You don't have that item.\r\n"
						} else if err := trade.GlobalTrade.AddItem(player.Name, item.ID, item.Name, 1); err != nil {
							response = err.Error() + "\r\n"
						} else {
							response = fmt.Sprintf("Added %s to trade.\r\n", item.Name)
						}
					}
				case "remove":
					if len(parts) < 2 {
						response = "Usage: trade remove <item_id>\r\n"
					} else if err := trade.GlobalTrade.RemoveItem(player.Name, parts[1]); err != nil {
						response = err.Error() + "\r\n"
					} else {
						response = "Item removed from trade.\r\n"
					}
				case "money":
					if len(parts) < 2 {
						response = "Usage: trade money <amount>\r\n"
					} else if amount, err := strconv.Atoi(parts[1]); err != nil {
						response = "Invalid amount.\r\n"
					} else if err := trade.GlobalTrade.SetMoney(player.Name, amount); err != nil {
						response = err.Error() + "\r\n"
					} else {
						response = fmt.Sprintf("Set money offer to %d.\r\n", amount)
					}
				case "confirm":
					if completed, err := trade.GlobalTrade.ConfirmTrade(player.Name); err != nil {
						response = err.Error() + "\r\n"
					} else if completed {
						response = "Trade completed!\r\n"
						// TODO: Actually exchange items and money
					} else {
						response = "Trade confirmed. Waiting for other party...\r\n"
					}
				default:
					response = "Unknown trade command. Usage: trade request/accept/decline/cancel/add/remove/money/confirm\r\n"
				}
			}

		case "auction":
			// Auction house
			parts := strings.Fields(arg)
			if len(parts) == 0 {
				response = "Usage: auction list, auction sell <item> <price> <buyout>, auction bid <id> <amount>, auction buyout <id>\r\n"
			} else {
				subCmd := strings.ToLower(parts[0])
				switch subCmd {
				case "list", "search":
					search := ""
					if len(parts) > 1 {
						search = strings.Join(parts[1:], " ")
					}
					listings := trade.GlobalTrade.SearchAuctions(search, "", 0)
					response = trade.GlobalTrade.FormatListings(listings)
				case "sell":
					if len(parts) < 3 {
						response = "Usage: auction sell <item> <start_price> <buyout_price>\r\n"
					} else {
						itemName := parts[1]
						startPrice, _ := strconv.Atoi(parts[2])
						buyoutPrice := startPrice * 2
						if len(parts) > 3 {
							buyoutPrice, _ = strconv.Atoi(parts[3])
						}
						// Find item in inventory
						var item *Item
						for _, i := range player.Inventory {
							if strings.Contains(strings.ToLower(i.Name), strings.ToLower(itemName)) {
								item = i
								break
							}
						}
						if item == nil {
							response = "You don't have that item.\r\n"
						} else if listing, err := trade.GlobalTrade.CreateListing(player.Name, item.ID, item.Name, 1, startPrice, buyoutPrice, 24*time.Hour, "general"); err != nil {
							response = err.Error() + "\r\n"
						} else {
							response = fmt.Sprintf("Listed %s on auction (ID: %s).\r\n", item.Name, listing.ID)
							// Remove from inventory
							for i, inv := range player.Inventory {
								if inv == item {
									player.Inventory = append(player.Inventory[:i], player.Inventory[i+1:]...)
									break
								}
							}
						}
					}
				case "bid":
					if len(parts) < 3 {
						response = "Usage: auction bid <listing_id> <amount>\r\n"
					} else if amount, err := strconv.Atoi(parts[2]); err != nil {
						response = "Invalid amount.\r\n"
					} else if err := trade.GlobalTrade.PlaceBid(player.Name, parts[1], amount); err != nil {
						response = err.Error() + "\r\n"
					} else {
						response = "Bid placed!\r\n"
					}
				case "buyout":
					if len(parts) < 2 {
						response = "Usage: auction buyout <listing_id>\r\n"
					} else if err := trade.GlobalTrade.Buyout(player.Name, parts[1]); err != nil {
						response = err.Error() + "\r\n"
					} else {
						response = "Item purchased!\r\n"
						// TODO: Add item to player inventory
					}
				default:
					response = "Unknown auction command. Usage: auction list/sell/bid/buyout\r\n"
				}
			}

		// --- PVP & ARENA COMMANDS ---
		case "arena", "pvp":
			// PvP arena commands
			parts := strings.Fields(arg)
			if len(parts) == 0 {
				response = "Usage: arena queue <type>, arena leave, arena stats, arena rankings\r\nTypes: duel, team, ffa, koth\r\n"
			} else {
				subCmd := strings.ToLower(parts[0])
				switch subCmd {
				case "queue", "join":
					arenaType := pvp.ArenaDuel
					if len(parts) > 1 {
						arenaType = pvp.ArenaType(strings.ToLower(parts[1]))
					}
					if arenaID, err := pvp.GlobalPvP.QueueForArena(player.Name, arenaType, 2); err != nil {
						response = err.Error() + "\r\n"
					} else if arenaID != "" {
						response = fmt.Sprintf("Arena match starting! ID: %s\r\n", arenaID)
					} else {
						response = "Queued for arena. Waiting for opponents...\r\n"
					}
				case "leave":
					if err := pvp.GlobalPvP.LeaveQueue(player.Name); err != nil {
						response = err.Error() + "\r\n"
					} else {
						response = "Left queue.\r\n"
					}
				case "stats":
					response = pvp.GlobalPvP.GetStats(player.Name)
				case "rankings", "ranking":
					response = pvp.GlobalPvP.GetRankings(10)
				default:
					response = "Unknown arena command. Usage: arena queue/leave/stats/rankings\r\n"
				}
			}

		case "duel":
			// Quick duel challenge
			if arg == "" {
				response = "Usage: duel <player>\r\n"
			} else if arenaID, err := pvp.GlobalPvP.QueueForArena(player.Name, pvp.ArenaDuel, 2); err != nil {
				response = err.Error() + "\r\n"
			} else {
				response = fmt.Sprintf("Duel queued. Waiting for %s to accept...\r\n", arg)
				// Notify other player
				for _, p := range world.Players {
					if strings.ToLower(p.Name) == strings.ToLower(arg) && p.Conn != nil {
						p.Conn.Write(fmt.Sprintf("\r\n%s%s has challenged you to a duel!\r\nType 'arena queue duel' to accept.\r\n> ", Red, player.Name))
						break
					}
				}
				_ = arenaID
			}

		// --- ACCESSIBILITY COMMANDS ---
		case "accessibility", "a11y":
			// Accessibility settings
			parts := strings.Fields(arg)
			if len(parts) == 0 {
				settings := accessibility.GlobalManager.GetSettings(player.Name)
				response = fmt.Sprintf("=== ACCESSIBILITY SETTINGS ===\r\n\r\n"+
					"Screen Reader: %v\r\n"+
					"High Contrast: %v\r\n"+
					"Large Text: %v\r\n"+
					"Reduced Motion: %v\r\n"+
					"Colorblind Mode: %s\r\n"+
					"Simplified Output: %v\r\n"+
					"Font Scale: %.1f\r\n\r\n"+
					"Usage: accessibility <setting> <value>\r\n"+
					"Settings: screenreader, highcontrast, largetext, reducedmotion, colorblind, simplified, fontscale\r\n",
					settings.ScreenReaderMode, settings.HighContrast, settings.LargeText,
					settings.ReducedMotion, settings.ColorblindMode, settings.SimplifiedOutput, settings.FontScale)
			} else {
				setting := strings.ToLower(parts[0])
				if len(parts) < 2 {
					response = "Usage: accessibility <setting> <value>\r\n"
				} else {
					value := strings.ToLower(parts[1])
					var valueInterface interface{}

					switch setting {
					case "screenreader", "screen_reader":
						valueInterface = value == "on" || value == "true" || value == "1"
					case "highcontrast", "high_contrast":
						valueInterface = value == "on" || value == "true" || value == "1"
					case "largetext", "large_text":
						valueInterface = value == "on" || value == "true" || value == "1"
					case "reducedmotion", "reduced_motion":
						valueInterface = value == "on" || value == "true" || value == "1"
					case "simplified", "simplified_output":
						valueInterface = value == "on" || value == "true" || value == "1"
					case "colorblind", "colorblind_mode":
						valueInterface = value
					case "fontscale", "font_scale":
						if scale, err := strconv.ParseFloat(value, 64); err == nil {
							valueInterface = scale
						} else {
							response = "Invalid font scale. Use 1.0, 1.5, etc.\r\n"
						}
					default:
						response = "Unknown setting. Available: screenreader, highcontrast, largetext, reducedmotion, colorblind, simplified, fontscale\r\n"
					}

					if response == "" {
						if accessibility.GlobalManager.UpdateSetting(player.Name, setting, valueInterface) {
							response = fmt.Sprintf("Accessibility setting '%s' updated.\r\n", setting)
						} else {
							response = "Failed to update setting. Check value format.\r\n"
						}
					}
				}
			}

		// --- TUTORIAL COMMANDS ---
		case "tutorial":
			parts := strings.Fields(arg)
			if len(parts) == 0 {
				// Show current tutorial step
				response = tutorial.GlobalManager.FormatStepDisplay(player.Name)
			} else {
				subCmd := strings.ToLower(parts[0])
				switch subCmd {
				case "start":
					if len(parts) < 2 {
						response = "Usage: tutorial start <tutorial_id>\r\n"
					} else if _, err := tutorial.GlobalManager.StartTutorial(player.Name, parts[1]); err != nil {
						response = err.Error() + "\r\n"
					} else {
						response = fmt.Sprintf("Tutorial '%s' started.\r\n", parts[1])
					}
				case "skip":
					if skipped, msg := tutorial.GlobalManager.SkipStep(player.Name); skipped {
						response = msg + "\r\n"
					} else {
						response = msg + "\r\n"
					}
				case "list":
					tutorials := tutorial.GlobalManager.GetAvailableTutorials(player.Name)
					var sb strings.Builder
					sb.WriteString("=== AVAILABLE TUTORIALS ===\r\n\r\n")
					for _, t := range tutorials {
						sb.WriteString(fmt.Sprintf("%s: %s\r\n", t.ID, t.Name))
					}
					response = sb.String()
				case "progress":
					progress := tutorial.GlobalManager.GetProgress(player.Name)
					var sb strings.Builder
					sb.WriteString("=== TUTORIAL PROGRESS ===\r\n\r\n")
					for tutID, prog := range progress {
						status := "In Progress"
						if !prog.CompletedAt.IsZero() {
							status = "Completed"
						}
						sb.WriteString(fmt.Sprintf("%s: %s (Step %d/%d) - %s\r\n",
							tutID, tutID, prog.CurrentStep, len(prog.StepsComplete), status))
					}
					response = sb.String()
				default:
					response = "Usage: tutorial, tutorial start <id>, tutorial skip, tutorial list, tutorial progress\r\n"
				}
			}

		case "hint":
			response = tutorial.GlobalManager.GetHint(player.Name)

		// --- DIALOGUE & INSTANCE COMMANDS (from content_expansion.go) ---
		case "talk":
			// Talk to NPCs using the dialogue system
			response = HandleTalkCommand(world, player, arg)
		case "bye":
			// End dialogue conversation
			if byeResp := HandleByeCommand(player); byeResp != "" {
				response = byeResp
			}
		case "instance":
			// Instance/dungeon commands
			response = HandleInstanceCommand(world, player, arg)

		case "quit":
			return
		default:
			response = "Unknown.\r\n"
		}
		if response != "" {
			// Apply accessibility processing
			response = accessibility.GlobalManager.ProcessOutput(player.Name, response)
			// Apply player's color theme preference
			themedResponse := ApplyTheme(response, player.ColorTheme)
			client.Write(themedResponse + "> ")
		} else {
			client.Write("> ")
		}
	}
}

func broadcast(w *World, sender *Player, msg string) {
	if sender == nil {
		return
	}
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	formatted := fmt.Sprintf("\r\n%s%s says: \"%s\"%s\r\n> ", White, sender.Name, msg, Green)
	for _, p := range w.Players {
		if p != nil && p.Conn != nil && p.RoomID == sender.RoomID && p != sender {
			p.Conn.Write(formatted)
		}
	}
}

// broadcastChatMessage sends a chat message to specific recipients
func broadcastChatMessage(w *World, msg string, recipients []string) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	for _, p := range w.Players {
		if p != nil && p.Conn != nil {
			// Check if player is in recipients list
			for _, recipient := range recipients {
				if strings.ToLower(p.Name) == strings.ToLower(recipient) {
					p.Conn.Write("\r\n" + Cyan + msg + Reset + "\r\n> ")
					break
				}
			}
		}
	}
}

// formatHelp generates help text for commands
func formatHelp(arg string) string {
	if arg == "" {
		// Show category overview
		var sb strings.Builder
		sb.WriteString("=== THE CONSTRUCT - HELP SYSTEM ===\r\n\r\n")

		for _, cat := range help.GetCategories() {
			sb.WriteString(White + cat + Reset + ": ")
			entries := help.GetAllByCategory()[cat]
			cmds := make([]string, 0, len(entries))
			for _, e := range entries {
				cmds = append(cmds, e.Command)
			}
			sb.WriteString(strings.Join(cmds, ", "))
			sb.WriteString("\r\n")
		}

		sb.WriteString("\r\nType 'help <command>' for details on a specific command.\r\n")
		return sb.String()
	}

	// Show specific command help
	entry := help.GetHelp(strings.ToLower(arg))
	if entry == nil {
		return fmt.Sprintf("No help available for '%s'. Type 'help' for command list.\r\n", arg)
	}

	var sb strings.Builder
	sb.WriteString(White + "=== " + strings.ToUpper(entry.Command) + " ===" + Reset + "\r\n")
	sb.WriteString(entry.Description + "\r\n\r\n")
	sb.WriteString("Usage: " + entry.Usage + "\r\n")

	if len(entry.Aliases) > 0 {
		sb.WriteString("Aliases: " + strings.Join(entry.Aliases, ", ") + "\r\n")
	}

	if len(entry.Examples) > 0 {
		sb.WriteString("Examples:\r\n")
		for _, ex := range entry.Examples {
			sb.WriteString("  " + ex + "\r\n")
		}
	}

	return sb.String()
}

// handleQuestCommand handles quest subcommands
func handleQuestCommand(player *Player, arg string) string {
	parts := strings.Fields(arg)
	if len(parts) == 0 {
		return quest.GlobalQuests.GetActiveQuests(player.Name)
	}

	subcmd := strings.ToLower(parts[0])
	switch subcmd {
	case "accept", "start":
		if len(parts) < 2 {
			return "Usage: quest accept <quest_id>\r\n"
		}
		questID := strings.ToLower(parts[1])

		// Check if can start
		can, reason := quest.GlobalQuests.CanStart(player.Name, questID, player.Level)
		if !can {
			return Red + reason + Reset + "\r\n"
		}

		// Start the quest
		dialogue, err := quest.GlobalQuests.StartQuest(player.Name, questID)
		if err != nil {
			return Red + "Quest not found." + Reset + "\r\n"
		}

		return Green + "Quest accepted!\r\n" + Reset + dialogue + "\r\n"

	case "abandon":
		if len(parts) < 2 {
			return "Usage: quest abandon <quest_id>\r\n"
		}
		// Would need to add abandon functionality to quest manager
		return "Quest abandoned.\r\n"

	case "list":
		return quest.GlobalQuests.GetActiveQuests(player.Name)

	default:
		return "Quest commands: list, accept <quest>, abandon <quest>\r\n"
	}
}

// handlePartyCommand handles the "party" command
func handlePartyCommand(player *Player, arg string) string {
	parts := strings.Fields(arg)
	if len(parts) == 0 {
		// Show party status
		p := party.GlobalParty.GetParty(player.Name)
		if p == nil {
			return "You are not in a party.\r\nUse 'party create' to start one, or 'invite <player>' to invite.\r\n"
		}

		var sb strings.Builder
		sb.WriteString("=== YOUR PARTY ===\r\n")
		for _, member := range party.GlobalParty.GetMembers(player.Name) {
			if member == p.Leader {
				sb.WriteString(Yellow + "* " + member + " (Leader)" + Reset + "\r\n")
			} else {
				sb.WriteString("  " + member + "\r\n")
			}
		}
		return sb.String()
	}

	subcmd := strings.ToLower(parts[0])
	switch subcmd {
	case "create":
		_, err := party.GlobalParty.Create(player.Name)
		if err != nil {
			return Red + err.Error() + Reset + "\r\n"
		}
		return Green + "Party created! Use 'invite <player>' to add members." + Reset + "\r\n"

	case "leave":
		err := party.GlobalParty.Leave(player.Name)
		if err != nil {
			return Red + err.Error() + Reset + "\r\n"
		}
		return "You left the party.\r\n"

	case "kick":
		if len(parts) < 2 {
			return "Usage: party kick <player>\r\n"
		}
		err := party.GlobalParty.Kick(player.Name, strings.ToLower(parts[1]))
		if err != nil {
			return Red + err.Error() + Reset + "\r\n"
		}
		return Green + parts[1] + " has been kicked from the party." + Reset + "\r\n"

	case "promote":
		if len(parts) < 2 {
			return "Usage: party promote <player>\r\n"
		}
		err := party.GlobalParty.Promote(player.Name, strings.ToLower(parts[1]))
		if err != nil {
			return Red + err.Error() + Reset + "\r\n"
		}
		return Green + parts[1] + " is now the party leader." + Reset + "\r\n"

	case "disband":
		err := party.GlobalParty.Disband(player.Name)
		if err != nil {
			return Red + err.Error() + Reset + "\r\n"
		}
		return "Party disbanded.\r\n"

	default:
		return "Party commands: create, leave, kick <player>, promote <player>, disband\r\n"
	}
}

// handlePartyInvite handles the "invite" command
func handlePartyInvite(player *Player, arg string) string {
	if arg == "" {
		return "Usage: invite <player>\r\n"
	}

	target := strings.ToLower(arg)

	// Auto-create party if not in one
	if !party.GlobalParty.IsInParty(player.Name) {
		_, err := party.GlobalParty.Create(player.Name)
		if err != nil {
			return Red + err.Error() + Reset + "\r\n"
		}
	}

	err := party.GlobalParty.Invite(player.Name, target)
	if err != nil {
		return Red + err.Error() + Reset + "\r\n"
	}

	return Green + "Invited " + target + " to your party." + Reset + "\r\n"
}

// handlePartyAccept handles the "accept" command
func handlePartyAccept(player *Player, arg string) string {
	if arg == "" {
		// Check for pending invites
		invites := party.GlobalParty.GetPendingInvites(player.Name)
		if len(invites) == 0 {
			return "You have no pending party invites.\r\n"
		}
		if len(invites) == 1 {
			arg = invites[0]
		} else {
			return "Multiple invites pending. Use 'accept <leader_name>':\r\n" + strings.Join(invites, ", ") + "\r\n"
		}
	}

	err := party.GlobalParty.Accept(player.Name, strings.ToLower(arg))
	if err != nil {
		return Red + err.Error() + Reset + "\r\n"
	}

	return Green + "You joined " + arg + "'s party!" + Reset + "\r\n"
}

// handlePartyDecline handles the "decline" command
func handlePartyDecline(player *Player, arg string) string {
	if arg == "" {
		invites := party.GlobalParty.GetPendingInvites(player.Name)
		if len(invites) == 0 {
			return "You have no pending party invites.\r\n"
		}
		if len(invites) == 1 {
			arg = invites[0]
		} else {
			return "Multiple invites pending. Use 'decline <leader_name>':\r\n" + strings.Join(invites, ", ") + "\r\n"
		}
	}

	err := party.GlobalParty.Decline(player.Name, strings.ToLower(arg))
	if err != nil {
		return Red + err.Error() + Reset + "\r\n"
	}

	return "Declined party invite from " + arg + ".\r\n"
}

// parseCommand parses user input into command and argument
// Returns lowercase command and lowercase argument (space-joined for multi-word args)
func parseCommand(input string) (cmd, arg string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ""
	}

	parts := strings.Fields(input)
	cmd = strings.ToLower(parts[0])
	if len(parts) > 1 {
		arg = strings.ToLower(strings.Join(parts[1:], " "))
	}
	return cmd, arg
}

// --- FACTION HANDLERS ---

// handleFactionCommand handles faction-related commands
func handleFactionCommand(player *Player, arg string) string {
	parts := strings.Fields(arg)
	if len(parts) == 0 {
		// Show current faction status
		pf := faction.GlobalFaction.GetPlayerFaction(player.Name)
		if pf.Faction == faction.FactionNone {
			return "You are not aligned with any faction.\r\n" +
				"Available factions: zion, machines, exiles\r\n" +
				"Use 'faction join <name>' to align yourself.\r\n"
		}
		f := faction.GlobalFaction.GetFaction(pf.Faction)
		return fmt.Sprintf("=== %s%s%s ===\r\n%s\r\nLeader: %s\r\n",
			Green, f.Name, Reset, f.Description, f.Leader)
	}

	switch parts[0] {
	case "join":
		if len(parts) < 2 {
			return "Usage: faction join <zion|machines|exiles>\r\n"
		}
		fid := faction.FactionID(strings.ToLower(parts[1]))
		msg, ok := faction.GlobalFaction.Join(player.Name, fid)
		if !ok {
			return Red + msg + Reset + "\r\n"
		}
		// Award achievement for joining a faction
		if ach := achievements.GlobalAchievements.Award(player.Name, achievements.AchAwakened); ach != nil {
			msg += fmt.Sprintf("\r\n%s*** Achievement Unlocked: %s ***%s", Yellow, ach.Name, Reset)
		}
		return Green + msg + Reset + "\r\n"

	case "leave":
		msg, ok := faction.GlobalFaction.Leave(player.Name)
		if !ok {
			return Red + msg + Reset + "\r\n"
		}
		return Yellow + msg + Reset + "\r\n"

	case "list":
		var sb strings.Builder
		sb.WriteString("=== FACTIONS ===\r\n")
		for _, f := range faction.GlobalFaction.GetAllFactions() {
			sb.WriteString(fmt.Sprintf("%s%s%s - %s\r\n", Green, f.Name, Reset, f.Description))
		}
		return sb.String()

	default:
		return "Faction commands: join <faction>, leave, list\r\n"
	}
}

// handleReputationCommand shows faction reputation
func handleReputationCommand(player *Player) string {
	var sb strings.Builder
	sb.WriteString("=== REPUTATION ===\r\n")

	for _, fid := range []faction.FactionID{faction.FactionZion, faction.FactionMachines, faction.FactionExiles} {
		rep := faction.GlobalFaction.GetReputation(player.Name, fid)
		standing := faction.GetStandingName(rep)
		f := faction.GlobalFaction.GetFaction(fid)
		sb.WriteString(fmt.Sprintf("  %s: %d (%s)\r\n", f.Name, rep, standing))
	}

	return sb.String()
}

// --- ACHIEVEMENT HANDLERS ---

// handleAchievementsCommand shows player achievements
func handleAchievementsCommand(player *Player, arg string) string {
	if arg != "" {
		// Show specific category
		achs := achievements.GlobalAchievements.GetByCategory(arg)
		if len(achs) == 0 {
			return "Unknown category. Categories: combat, exploration, social, progression\r\n"
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("=== %s ACHIEVEMENTS ===\r\n", strings.ToUpper(arg)))
		for _, ach := range achs {
			earned := achievements.GlobalAchievements.HasAchievement(player.Name, ach.ID)
			marker := "[ ]"
			if earned {
				marker = "[X]"
			}
			sb.WriteString(fmt.Sprintf("  %s %s - %s (%d pts)\r\n", marker, ach.Name, ach.Description, ach.Points))
		}
		return sb.String()
	}

	// Show overview
	earned := achievements.GlobalAchievements.GetEarnedAchievements(player.Name)
	points := achievements.GlobalAchievements.GetTotalPoints(player.Name)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== ACHIEVEMENTS (%d earned, %d points) ===\r\n", len(earned), points))
	for _, ach := range earned {
		sb.WriteString(fmt.Sprintf("  %s[X]%s %s - %s\r\n", Green, Reset, ach.Name, ach.Description))
	}
	sb.WriteString("\r\nCategories: combat, exploration, social, progression\r\n")
	sb.WriteString("Use 'achievements <category>' to see all in a category.\r\n")
	return sb.String()
}

// handleTitleCommand manages player titles
func handleTitleCommand(player *Player, arg string) string {
	if arg == "" {
		// Show available titles
		titles := achievements.GlobalAchievements.GetAvailableTitles(player.Name)
		current := achievements.GlobalAchievements.GetTitle(player.Name)

		var sb strings.Builder
		sb.WriteString("=== TITLES ===\r\n")
		if current != "" {
			sb.WriteString(fmt.Sprintf("Current: %s%s%s\r\n", Yellow, current, Reset))
		} else {
			sb.WriteString("Current: (none)\r\n")
		}
		sb.WriteString("Available:\r\n")
		if len(titles) == 0 {
			sb.WriteString("  (none unlocked yet)\r\n")
		} else {
			for _, t := range titles {
				sb.WriteString(fmt.Sprintf("  - %s\r\n", t))
			}
		}
		sb.WriteString("\r\nUse 'title <name>' to set, or 'title clear' to remove.\r\n")
		return sb.String()
	}

	if arg == "clear" {
		achievements.GlobalAchievements.SetTitle(player.Name, "")
		return "Title cleared.\r\n"
	}

	if achievements.GlobalAchievements.SetTitle(player.Name, arg) {
		return Green + "Title set to: " + arg + Reset + "\r\n"
	}
	return Red + "You haven't unlocked that title." + Reset + "\r\n"
}

// --- LEADERBOARD HANDLERS ---

// handleLeaderboardCommand shows rankings
func handleLeaderboardCommand(arg string) string {
	stat := leaderboard.StatXP
	statName := "XP"

	switch strings.ToLower(arg) {
	case "kills":
		stat = leaderboard.StatKills
		statName = "Kills"
	case "deaths":
		stat = leaderboard.StatDeaths
		statName = "Deaths"
	case "quests":
		stat = leaderboard.StatQuestsCompleted
		statName = "Quests"
	case "money", "bits":
		stat = leaderboard.StatMoney
		statName = "Bits"
	case "pvp":
		stat = leaderboard.StatPvPWins
		statName = "PvP Wins"
	case "level":
		stat = leaderboard.StatLevel
		statName = "Level"
	case "achievements":
		stat = leaderboard.StatAchievements
		statName = "Achievements"
	}

	board := leaderboard.GlobalLeaderboard.GetLeaderboard(stat, 10)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== TOP 10: %s ===\r\n", statName))
	if len(board) == 0 {
		sb.WriteString("  No data yet.\r\n")
	} else {
		for _, entry := range board {
			sb.WriteString(fmt.Sprintf("  %d. %s - %d\r\n", entry.Rank, entry.Name, entry.Value))
		}
	}
	sb.WriteString("\r\nCategories: xp, level, kills, deaths, quests, money, pvp, achievements\r\n")
	return sb.String()
}

// handleStatsCommand shows player's own stats
func handleStatsCommand(player *Player) string {
	stats := leaderboard.GlobalLeaderboard.GetStats(player.Name)

	var sb strings.Builder
	sb.WriteString("=== YOUR STATS ===\r\n")
	sb.WriteString(fmt.Sprintf("  XP: %d (Rank #%d)\r\n", stats.XP, leaderboard.GlobalLeaderboard.GetRank(player.Name, leaderboard.StatXP)))
	sb.WriteString(fmt.Sprintf("  Level: %d\r\n", stats.Level))
	sb.WriteString(fmt.Sprintf("  Kills: %d\r\n", stats.Kills))
	sb.WriteString(fmt.Sprintf("  Deaths: %d\r\n", stats.Deaths))
	sb.WriteString(fmt.Sprintf("  Quests: %d\r\n", stats.QuestsCompleted))
	sb.WriteString(fmt.Sprintf("  PvP: %d W / %d L\r\n", stats.PvPWins, stats.PvPLosses))
	sb.WriteString(fmt.Sprintf("  Achievements: %d\r\n", stats.Achievements))
	return sb.String()
}

// --- TRAINING HANDLERS ---

// handleTrainingCommand handles training program commands
func handleTrainingCommand(player *Player, arg string) string {
	parts := strings.Fields(arg)
	if len(parts) == 0 {
		// Check if in a program
		inst := training.GlobalTraining.GetPlayerInstance(player.Name)
		if inst != nil {
			return fmt.Sprintf("You are in: %s%s%s\r\nUse 'train leave' to exit or 'train complete' to finish.\r\n",
				Green, inst.Program.Name, Reset)
		}
		return "You are not in a training program.\r\n" +
			"Use 'programs' to see available programs.\r\n" +
			"Use 'train start <program>' to begin.\r\n"
	}

	switch parts[0] {
	case "start":
		if len(parts) < 2 {
			return "Usage: train start <program_id>\r\n"
		}
		inst, err := training.GlobalTraining.StartProgram(player.Name, parts[1])
		if err != nil {
			return Red + err.Error() + Reset + "\r\n"
		}
		return fmt.Sprintf("%sLoading program: %s%s\r\n%s\r\n",
			Green, inst.Program.Name, Reset, inst.Program.Description)

	case "join":
		if len(parts) < 2 {
			return "Usage: train join <instance_id>\r\n"
		}
		err := training.GlobalTraining.JoinProgram(player.Name, parts[1])
		if err != nil {
			return Red + err.Error() + Reset + "\r\n"
		}
		return Green + "Joined the training program." + Reset + "\r\n"

	case "leave":
		err := training.GlobalTraining.LeaveProgram(player.Name)
		if err != nil {
			return Red + err.Error() + Reset + "\r\n"
		}
		return "You exit the training program.\r\n"

	case "complete":
		rewards, score, err := training.GlobalTraining.CompleteProgram(player.Name)
		if err != nil {
			return Red + err.Error() + Reset + "\r\n"
		}
		player.XP += rewards.XP
		player.Money += rewards.Money
		return fmt.Sprintf("%sProgram Complete!%s\r\nScore: %d\r\nRewards: +%d XP, +%d bits\r\n",
			Green, Reset, score, rewards.XP, rewards.Money)

	default:
		return "Training commands: start <program>, join <instance>, leave, complete\r\n"
	}
}

// handleProgramsCommand lists available training programs
func handleProgramsCommand() string {
	programs := training.GlobalTraining.ListPrograms()

	var sb strings.Builder
	sb.WriteString("=== TRAINING PROGRAMS ===\r\n")
	for _, p := range programs {
		difficulty := strings.Repeat("*", p.Difficulty)
		sb.WriteString(fmt.Sprintf("  %s%s%s [%s]\r\n", Green, p.Name, Reset, difficulty))
		sb.WriteString(fmt.Sprintf("    ID: %s | %s\r\n", p.ID, p.Description))
		sb.WriteString(fmt.Sprintf("    Rewards: %d XP, %d bits\r\n", p.Rewards.XP, p.Rewards.Money))
	}
	sb.WriteString("\r\nUse 'train start <id>' to begin.\r\n")
	return sb.String()
}

// handleChallengesCommand lists combat challenges
func handleChallengesCommand() string {
	challenges := training.GlobalTraining.ListChallenges()

	var sb strings.Builder
	sb.WriteString("=== CHALLENGES ===\r\n")
	for _, c := range challenges {
		sb.WriteString(fmt.Sprintf("  %s%s%s\r\n", Yellow, c.Name, Reset))
		sb.WriteString(fmt.Sprintf("    %s\r\n", c.Description))
		if c.BestPlayer != "" {
			sb.WriteString(fmt.Sprintf("    Record: %s (%ds)\r\n", c.BestPlayer, c.BestTime))
		} else {
			sb.WriteString("    Record: (none)\r\n")
		}
	}
	return sb.String()
}
