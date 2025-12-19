// Package api provides a REST API for Matrix MUD.
// Supports external tools, bots, and integrations.
package api

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config holds API configuration
type Config struct {
	BindAddr      string
	APIKeys       map[string]*APIKey
	RateLimitRPS  int
	CORSOrigins   []string
	EnableSwagger bool
}

// APIKey represents an API key with permissions
type APIKey struct {
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used"`
	Enabled     bool      `json:"enabled"`
}

// Response is a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta contains response metadata
type Meta struct {
	Total      int    `json:"total,omitempty"`
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	TotalPages int    `json:"total_pages,omitempty"`
	Version    string `json:"version,omitempty"`
}

// ErrorResponse creates an error response
func ErrorResponse(err string, code int) *Response {
	return &Response{
		Success: false,
		Error:   err,
	}
}

// SuccessResponse creates a success response
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

// SuccessResponseWithMeta creates a success response with metadata
func SuccessResponseWithMeta(data interface{}, meta *Meta) *Response {
	return &Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

// RateLimiter limits API requests per key
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    requestsPerSecond,
		window:   time.Second,
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old requests
	reqs := rl.requests[key]
	var valid []time.Time
	for _, t := range reqs {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	rl.requests[key] = append(valid, now)
	return true
}

// Server is the API server
type Server struct {
	config      *Config
	mux         *http.ServeMux
	rateLimiter *RateLimiter
	server      *http.Server
	version     string

	// Data providers (set by main application)
	GetOnlinePlayers    func() []PlayerInfo
	GetPlayerByName     func(name string) *PlayerInfo
	GetServerStatus     func() *ServerStatus
	GetRooms            func() []RoomInfo
	GetRoom             func(id string) *RoomInfo
	GetNPCs             func() []NPCInfo
	GetItems            func() []ItemInfo
	GetLeaderboard      func(category string, limit int) []LeaderboardEntry
	SendMessageToPlayer func(name, message string) error
}

// PlayerInfo represents player data for API responses
type PlayerInfo struct {
	Name     string    `json:"name"`
	Class    string    `json:"class"`
	Level    int       `json:"level"`
	XP       int       `json:"xp"`
	HP       int       `json:"hp"`
	MaxHP    int       `json:"max_hp"`
	MP       int       `json:"mp"`
	MaxMP    int       `json:"max_mp"`
	Money    int       `json:"money"`
	RoomID   string    `json:"room_id"`
	Title    string    `json:"title,omitempty"`
	Faction  string    `json:"faction,omitempty"`
	Online   bool      `json:"online"`
	LastSeen time.Time `json:"last_seen,omitempty"`
}

// ServerStatus represents server status
type ServerStatus struct {
	Status        string    `json:"status"`
	Version       string    `json:"version"`
	Uptime        string    `json:"uptime"`
	PlayersOnline int       `json:"players_online"`
	TotalPlayers  int       `json:"total_players"`
	TotalRooms    int       `json:"total_rooms"`
	TotalNPCs     int       `json:"total_npcs"`
	StartedAt     time.Time `json:"started_at"`
}

// RoomInfo represents room data
type RoomInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Exits       []string `json:"exits"`
	NPCs        []string `json:"npcs,omitempty"`
	Items       []string `json:"items,omitempty"`
	PlayerCount int      `json:"player_count"`
}

// NPCInfo represents NPC data
type NPCInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Level    int    `json:"level"`
	HP       int    `json:"hp"`
	MaxHP    int    `json:"max_hp"`
	RoomID   string `json:"room_id"`
	Hostile  bool   `json:"hostile"`
	Merchant bool   `json:"merchant"`
}

// ItemInfo represents item data
type ItemInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Value       int    `json:"value"`
	Damage      int    `json:"damage,omitempty"`
	Armor       int    `json:"armor,omitempty"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	Rank  int    `json:"rank"`
	Name  string `json:"name"`
	Value int    `json:"value"`
	Class string `json:"class,omitempty"`
	Title string `json:"title,omitempty"`
}

// NewServer creates a new API server
func NewServer(config *Config, version string) *Server {
	if config.RateLimitRPS == 0 {
		config.RateLimitRPS = 100
	}

	s := &Server{
		config:      config,
		mux:         http.NewServeMux(),
		rateLimiter: NewRateLimiter(config.RateLimitRPS),
		version:     version,
	}

	s.registerRoutes()
	return s
}

// registerRoutes sets up API routes
func (s *Server) registerRoutes() {
	// Health (no auth)
	s.mux.HandleFunc("/api/health", s.handleHealth)

	// Status (no auth)
	s.mux.HandleFunc("/api/status", s.handleStatus)

	// Players
	s.mux.HandleFunc("/api/players", s.withAuth(s.handlePlayers))
	s.mux.HandleFunc("/api/players/", s.withAuth(s.handlePlayerByName))

	// World
	s.mux.HandleFunc("/api/world/rooms", s.withAuth(s.handleRooms))
	s.mux.HandleFunc("/api/world/rooms/", s.withAuth(s.handleRoomByID))
	s.mux.HandleFunc("/api/world/npcs", s.withAuth(s.handleNPCs))
	s.mux.HandleFunc("/api/world/items", s.withAuth(s.handleItems))

	// Leaderboards
	s.mux.HandleFunc("/api/leaderboards", s.withAuth(s.handleLeaderboardCategories))
	s.mux.HandleFunc("/api/leaderboards/", s.withAuth(s.handleLeaderboard))

	// Messages
	s.mux.HandleFunc("/api/messages", s.withAuth(s.handleSendMessage))
}

// withAuth wraps a handler with authentication
func (s *Server) withAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS
		s.setCORSHeaders(w, r)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Get API key
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			apiKey = r.URL.Query().Get("api_key")
		}

		if apiKey == "" {
			s.writeError(w, "API key required", http.StatusUnauthorized)
			return
		}

		// Validate key
		key := s.validateAPIKey(apiKey)
		if key == nil {
			s.writeError(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// Rate limit
		if !s.rateLimiter.Allow(apiKey) {
			s.writeError(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Update last used
		key.LastUsed = time.Now()

		handler(w, r)
	}
}

// validateAPIKey validates an API key
func (s *Server) validateAPIKey(key string) *APIKey {
	for _, k := range s.config.APIKeys {
		if k.Enabled && subtle.ConstantTimeCompare([]byte(k.Key), []byte(key)) == 1 {
			return k
		}
	}
	return nil
}

// setCORSHeaders sets CORS headers
func (s *Server) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	allowed := false
	for _, o := range s.config.CORSOrigins {
		if o == "*" || o == origin {
			allowed = true
			break
		}
	}

	if allowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		w.Header().Set("Access-Control-Max-Age", "86400")
	}
}

// writeJSON writes a JSON response
func (s *Server) writeJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (s *Server) writeError(w http.ResponseWriter, message string, code int) {
	s.writeJSON(w, ErrorResponse(message, code), code)
}

// writeSuccess writes a success response
func (s *Server) writeSuccess(w http.ResponseWriter, data interface{}) {
	s.writeJSON(w, SuccessResponse(data), http.StatusOK)
}

// Handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w, r)
	s.writeSuccess(w, map[string]interface{}{
		"status":  "healthy",
		"version": s.version,
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w, r)
	if s.GetServerStatus != nil {
		s.writeSuccess(w, s.GetServerStatus())
	} else {
		s.writeSuccess(w, map[string]string{"status": "ok"})
	}
}

func (s *Server) handlePlayers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.GetOnlinePlayers == nil {
		s.writeSuccess(w, []PlayerInfo{})
		return
	}

	players := s.GetOnlinePlayers()
	s.writeJSON(w, SuccessResponseWithMeta(players, &Meta{
		Total:   len(players),
		Version: s.version,
	}), http.StatusOK)
}

func (s *Server) handlePlayerByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract name from path
	path := strings.TrimPrefix(r.URL.Path, "/api/players/")
	parts := strings.Split(path, "/")
	name := parts[0]

	if name == "" {
		s.writeError(w, "Player name required", http.StatusBadRequest)
		return
	}

	if s.GetPlayerByName == nil {
		s.writeError(w, "Player not found", http.StatusNotFound)
		return
	}

	player := s.GetPlayerByName(name)
	if player == nil {
		s.writeError(w, "Player not found", http.StatusNotFound)
		return
	}

	// Check for sub-resources
	if len(parts) > 1 {
		switch parts[1] {
		case "inventory":
			s.writeSuccess(w, map[string]string{"message": "Inventory endpoint"})
		case "stats":
			s.writeSuccess(w, map[string]interface{}{
				"name":  player.Name,
				"level": player.Level,
				"class": player.Class,
			})
		default:
			s.writeError(w, "Unknown resource", http.StatusNotFound)
		}
		return
	}

	s.writeSuccess(w, player)
}

func (s *Server) handleRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.GetRooms == nil {
		s.writeSuccess(w, []RoomInfo{})
		return
	}

	rooms := s.GetRooms()
	s.writeJSON(w, SuccessResponseWithMeta(rooms, &Meta{
		Total: len(rooms),
	}), http.StatusOK)
}

func (s *Server) handleRoomByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/world/rooms/")
	if id == "" {
		s.writeError(w, "Room ID required", http.StatusBadRequest)
		return
	}

	if s.GetRoom == nil {
		s.writeError(w, "Room not found", http.StatusNotFound)
		return
	}

	room := s.GetRoom(id)
	if room == nil {
		s.writeError(w, "Room not found", http.StatusNotFound)
		return
	}

	s.writeSuccess(w, room)
}

func (s *Server) handleNPCs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.GetNPCs == nil {
		s.writeSuccess(w, []NPCInfo{})
		return
	}

	npcs := s.GetNPCs()
	s.writeJSON(w, SuccessResponseWithMeta(npcs, &Meta{
		Total: len(npcs),
	}), http.StatusOK)
}

func (s *Server) handleItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.GetItems == nil {
		s.writeSuccess(w, []ItemInfo{})
		return
	}

	items := s.GetItems()
	s.writeJSON(w, SuccessResponseWithMeta(items, &Meta{
		Total: len(items),
	}), http.StatusOK)
}

func (s *Server) handleLeaderboardCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories := []string{"level", "money", "kills", "achievements", "pvp"}
	s.writeSuccess(w, map[string]interface{}{
		"categories": categories,
	})
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	category := strings.TrimPrefix(r.URL.Path, "/api/leaderboards/")
	if category == "" {
		s.writeError(w, "Category required", http.StatusBadRequest)
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if s.GetLeaderboard == nil {
		s.writeSuccess(w, []LeaderboardEntry{})
		return
	}

	entries := s.GetLeaderboard(category, limit)
	s.writeJSON(w, SuccessResponseWithMeta(entries, &Meta{
		Total: len(entries),
	}), http.StatusOK)
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Player  string `json:"player"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Player == "" || req.Message == "" {
		s.writeError(w, "Player and message required", http.StatusBadRequest)
		return
	}

	if s.SendMessageToPlayer == nil {
		s.writeError(w, "Messaging not available", http.StatusServiceUnavailable)
		return
	}

	if err := s.SendMessageToPlayer(req.Player, req.Message); err != nil {
		s.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.writeSuccess(w, map[string]string{"status": "sent"})
}

// Start starts the API server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         s.config.BindAddr,
		Handler:      s.mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.server.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Handler returns the HTTP handler for embedding
func (s *Server) Handler() http.Handler {
	return s.mux
}

// AddAPIKey adds an API key
func (s *Server) AddAPIKey(key *APIKey) {
	if s.config.APIKeys == nil {
		s.config.APIKeys = make(map[string]*APIKey)
	}
	s.config.APIKeys[key.Key] = key
}

// RemoveAPIKey removes an API key
func (s *Server) RemoveAPIKey(key string) {
	delete(s.config.APIKeys, key)
}

// ListAPIKeys returns all API keys (without the actual key values)
func (s *Server) ListAPIKeys() []*APIKey {
	keys := make([]*APIKey, 0, len(s.config.APIKeys))
	for _, k := range s.config.APIKeys {
		keys = append(keys, &APIKey{
			Key:         k.Key[:8] + "...", // Truncate for security
			Name:        k.Name,
			Permissions: k.Permissions,
			CreatedAt:   k.CreatedAt,
			LastUsed:    k.LastUsed,
			Enabled:     k.Enabled,
		})
	}
	return keys
}
