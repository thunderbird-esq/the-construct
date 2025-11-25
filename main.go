package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/matrix-mud/pkg/ratelimit"
	"github.com/yourusername/matrix-mud/pkg/validation"
)

var (
	userMutex   sync.Mutex
	authLimiter = ratelimit.New(5, 1*time.Minute) // 5 auth attempts per minute
)

func authenticate(c *Client, name string) bool {
	// Apply rate limiting to prevent brute force attacks
	if !authLimiter.Allow(name) {
		c.Write(Red + "Too many authentication attempts. Try again later.\r\n" + Reset)
		log.Printf("Rate limit exceeded for user: %s", name)
		time.Sleep(3 * time.Second) // Add delay for rate-limited clients
		return false
	}

	userMutex.Lock()
	defer userMutex.Unlock()

	// Load existing user database (stores password hashes)
	users := make(map[string]string)
	file, err := os.ReadFile("data/users.json")
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error reading users.json: %v", err)
		c.Write(Red + "Authentication error.\r\n" + Reset)
		return false
	}

	if file != nil {
		if err := json.Unmarshal(file, &users); err != nil {
			log.Printf("Error parsing users.json: %v", err)
			c.Write(Red + "Authentication error.\r\n" + Reset)
			return false
		}
	}

	cleanName := strings.ToLower(name)

	if storedHash, exists := users[cleanName]; exists {
		// Existing user - verify password with bcrypt
		c.Write("Password: ")
		pass, err := c.reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading password: %v", err)
			return false
		}
		pass = strings.TrimSpace(pass)

		// Compare password with stored bcrypt hash
		err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(pass))
		if err == nil {
			log.Printf("User %s authenticated successfully", cleanName)
			return true
		}

		c.Write(Red + "Access Denied.\r\n" + Reset)
		log.Printf("Failed auth attempt for user %s", cleanName)
		return false
	} else {
		// New user - create account with bcrypt hashed password
		c.Write("New identity detected. Set a password: ")
		pass, err := c.reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading password: %v", err)
			return false
		}
		pass = strings.TrimSpace(pass)

		// Enforce minimum password length of 8 characters
		if len(pass) < 8 {
			c.Write("Password must be at least 8 characters.\r\n")
			return false
		}

		// Hash password with bcrypt
		hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			c.Write(Red + "Error creating account.\r\n" + Reset)
			return false
		}

		// Store hashed password
		users[cleanName] = string(hash)
		data, _ := json.MarshalIndent(users, "", "  ")
		if err := os.WriteFile("data/users.json", data, 0600); err != nil { // Owner read/write only
			log.Printf("Error saving user: %v", err)
			c.Write(Red + "Error creating account.\r\n" + Reset)
			return false
		}

		c.Write("Identity created.\r\n")
		log.Printf("New user created: %s", cleanName)
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

func main() {
	listener, err := net.Listen("tcp", ":2323")
	if err != nil {
		log.Fatal(err)
	}

	world := NewWorld()
	go startWebServer(world)
	go startAdminServer(world)

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		for range ticker.C {
			world.Update()
		}
	}()

	fmt.Println("Matrix Construct Server v1.28 (Phase 28) started on port 2323...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, world)
	}
}

func handleConnection(conn net.Conn, world *World) {
	client := &Client{conn: conn, reader: bufio.NewReader(conn)}
	defer conn.Close()

	client.Write(Clear + Green + "Wake up...\r\n" + Reset)
	client.Write("Identify yourself: ")
	name, err := client.reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading name: %v", err)
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
		log.Printf("Invalid username attempt: %s", name)
		return
	}

	if !authenticate(client, name) {
		return
	}

	player := world.LoadPlayer(name, client)
	if player.Class == "" {
		chooseClass(client, player)
		world.SavePlayer(player)
	}

	world.mutex.Lock()
	world.Players[client] = player
	world.mutex.Unlock()

	defer func() {
		world.SavePlayer(player)
		world.mutex.Lock()
		delete(world.Players, client)
		world.mutex.Unlock()
	}()

	client.Write(Matrixify(world.Look(player, "")))
	client.Write("> ")

	for {
		input, err := client.reader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)
		if input == "" {
			client.Write("> ")
			continue
		}

		parts := strings.Fields(input)
		cmd := strings.ToLower(parts[0])
		arg := ""
		if len(parts) > 1 {
			arg = strings.ToLower(strings.Join(parts[1:], " "))
		}

		var response string
		switch cmd {
		case "look", "l":
			response = Matrixify(world.Look(player, arg))
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
		case "down":
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
			response = Matrixify(world.StartCombat(player, arg))
		case "cast", "c":
			skillParts := strings.Fields(arg)
			if len(skillParts) > 0 {
				skill := skillParts[0]
				target := ""
				if len(skillParts) > 1 {
					target = strings.Join(skillParts[1:], " ")
				}
				response = Matrixify(world.CastSkill(player, skill, target))
			} else {
				response = "Cast what?\r\n"
			}
		case "flee", "stop":
			response = Matrixify(world.StopCombat(player))
		case "wear", "wield", "equip":
			response = Matrixify(world.WearItem(player, arg))
		case "remove", "unequip":
			response = Matrixify(world.RemoveItem(player, arg))
		case "use", "eat", "take":
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

		case "who":
			response = Matrixify(world.ListPlayers())
		case "help":
			response = Matrixify("Commands: look, n/s/e/w, get, drop, inv, score, wear, remove, use, cast, give, deposit, withdraw, storage, say, gossip, tell, list, buy, sell, kill, flee, who, quit\r\nBuilder: generate city [r] [c], dig, create, delete, edit desc, save world\r\n")
		case "teleport":
			response = Matrixify(world.Teleport(player, arg))

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

		case "quit":
			return
		default:
			response = "Unknown.\r\n"
		}
		if response != "" {
			client.Write(response + "> ")
		} else {
			client.Write("> ")
		}
	}
}

func broadcast(w *World, sender *Player, msg string) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	formatted := fmt.Sprintf("\r\n%s%s says: \"%s\"%s\r\n> ", White, sender.Name, msg, Green)
	for _, p := range w.Players {
		if p.RoomID == sender.RoomID && p != sender {
			p.Conn.Write(formatted)
		}
	}
}
