// Package trade implements player trading and auction house for Matrix MUD.
// Supports direct trades, auction listings, and market pricing.
package trade

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// TradeState represents the state of a trade
type TradeState string

const (
	StatePending   TradeState = "pending"
	StateActive    TradeState = "active"
	StateConfirmed TradeState = "confirmed"
	StateCompleted TradeState = "completed"
	StateCanceled  TradeState = "canceled"
)

// TradeItem represents an item in a trade
type TradeItem struct {
	ItemID   string
	Name     string
	Quantity int
}

// TradeOffer represents one side of a trade
type TradeOffer struct {
	Items     []TradeItem
	Money     int
	Confirmed bool
}

// Trade represents an active trade between two players
type Trade struct {
	ID             string
	Initiator      string
	Target         string
	InitiatorOffer TradeOffer
	TargetOffer    TradeOffer
	State          TradeState
	CreatedAt      time.Time
	UpdatedAt      time.Time
	mu             sync.RWMutex
}

// AuctionListing represents an item listed for auction
type AuctionListing struct {
	ID            string
	SellerName    string
	ItemID        string
	ItemName      string
	Quantity      int
	StartPrice    int
	BuyoutPrice   int
	CurrentBid    int
	CurrentBidder string
	ExpiresAt     time.Time
	CreatedAt     time.Time
	Category      string
	Sold          bool
}

// PriceHistory tracks historical prices
type PriceHistory struct {
	ItemID       string
	AveragePrice int
	MinPrice     int
	MaxPrice     int
	TotalSold    int
	LastSoldAt   time.Time
	PricePoints  []PricePoint
}

// PricePoint is a single price data point
type PricePoint struct {
	Price     int
	Quantity  int
	Timestamp time.Time
}

// ExpiredAuction represents an auction that has expired
type ExpiredAuction struct {
	Listing *AuctionListing
	HasBid  bool
}

// Manager handles all trade operations
type Manager struct {
	mu             sync.RWMutex
	Trades         map[string]*Trade
	PlayerTrades   map[string]string
	Auctions       map[string]*AuctionListing
	PlayerAuctions map[string][]string
	PriceHistory   map[string]*PriceHistory
	nextTradeID    int
	nextAuctionID  int
}

// NewManager creates a new trade manager
func NewManager() *Manager {
	return &Manager{
		Trades:         make(map[string]*Trade),
		PlayerTrades:   make(map[string]string),
		Auctions:       make(map[string]*AuctionListing),
		PlayerAuctions: make(map[string][]string),
		PriceHistory:   make(map[string]*PriceHistory),
		nextTradeID:    1,
		nextAuctionID:  1,
	}
}

// InitiateTrade starts a trade between two players
func (m *Manager) InitiateTrade(initiatorName, targetName string) (*Trade, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	initiator := strings.ToLower(initiatorName)
	target := strings.ToLower(targetName)

	if initiator == target {
		return nil, fmt.Errorf("you cannot trade with yourself")
	}

	if _, ok := m.PlayerTrades[initiator]; ok {
		return nil, fmt.Errorf("you are already in a trade")
	}
	if _, ok := m.PlayerTrades[target]; ok {
		return nil, fmt.Errorf("%s is already in a trade", targetName)
	}

	tradeID := fmt.Sprintf("trade_%d", m.nextTradeID)
	m.nextTradeID++

	trade := &Trade{
		ID:        tradeID,
		Initiator: initiatorName,
		Target:    targetName,
		InitiatorOffer: TradeOffer{
			Items: make([]TradeItem, 0),
		},
		TargetOffer: TradeOffer{
			Items: make([]TradeItem, 0),
		},
		State:     StatePending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.Trades[tradeID] = trade
	m.PlayerTrades[initiator] = tradeID

	return trade, nil
}

// AcceptTrade accepts a pending trade request
func (m *Manager) AcceptTrade(playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)

	var trade *Trade
	for _, t := range m.Trades {
		if strings.ToLower(t.Target) == name && t.State == StatePending {
			trade = t
			break
		}
	}

	if trade == nil {
		return fmt.Errorf("no pending trade requests")
	}

	trade.mu.Lock()
	defer trade.mu.Unlock()

	trade.State = StateActive
	trade.UpdatedAt = time.Now()
	m.PlayerTrades[name] = trade.ID

	return nil
}

// DeclineTrade declines a pending trade request
func (m *Manager) DeclineTrade(playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)

	for tradeID, t := range m.Trades {
		if strings.ToLower(t.Target) == name && t.State == StatePending {
			t.State = StateCanceled
			delete(m.PlayerTrades, strings.ToLower(t.Initiator))
			delete(m.Trades, tradeID)
			return nil
		}
	}

	return fmt.Errorf("no pending trade requests")
}

// CancelTrade cancels an active trade
func (m *Manager) CancelTrade(playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	tradeID, ok := m.PlayerTrades[name]
	if !ok {
		return fmt.Errorf("you are not in a trade")
	}

	trade := m.Trades[tradeID]
	trade.mu.Lock()
	trade.State = StateCanceled
	trade.mu.Unlock()

	delete(m.PlayerTrades, strings.ToLower(trade.Initiator))
	delete(m.PlayerTrades, strings.ToLower(trade.Target))
	delete(m.Trades, tradeID)

	return nil
}

// AddItem adds an item to a player's trade offer
func (m *Manager) AddItem(playerName, itemID, itemName string, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	tradeID, ok := m.PlayerTrades[name]
	if !ok {
		return fmt.Errorf("you are not in a trade")
	}

	trade := m.Trades[tradeID]
	trade.mu.Lock()
	defer trade.mu.Unlock()

	if trade.State != StateActive {
		return fmt.Errorf("trade is not active")
	}

	trade.InitiatorOffer.Confirmed = false
	trade.TargetOffer.Confirmed = false

	item := TradeItem{
		ItemID:   itemID,
		Name:     itemName,
		Quantity: quantity,
	}

	if strings.ToLower(trade.Initiator) == name {
		trade.InitiatorOffer.Items = append(trade.InitiatorOffer.Items, item)
	} else {
		trade.TargetOffer.Items = append(trade.TargetOffer.Items, item)
	}

	trade.UpdatedAt = time.Now()
	return nil
}

// RemoveItem removes an item from a player's trade offer
func (m *Manager) RemoveItem(playerName, itemID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	tradeID, ok := m.PlayerTrades[name]
	if !ok {
		return fmt.Errorf("you are not in a trade")
	}

	trade := m.Trades[tradeID]
	trade.mu.Lock()
	defer trade.mu.Unlock()

	if trade.State != StateActive {
		return fmt.Errorf("trade is not active")
	}

	trade.InitiatorOffer.Confirmed = false
	trade.TargetOffer.Confirmed = false

	var offer *TradeOffer
	if strings.ToLower(trade.Initiator) == name {
		offer = &trade.InitiatorOffer
	} else {
		offer = &trade.TargetOffer
	}

	found := false
	for i, item := range offer.Items {
		if item.ItemID == itemID {
			offer.Items = append(offer.Items[:i], offer.Items[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("item not in your offer")
	}

	trade.UpdatedAt = time.Now()
	return nil
}

// SetMoney sets the money amount in a player's trade offer
func (m *Manager) SetMoney(playerName string, amount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	name := strings.ToLower(playerName)
	tradeID, ok := m.PlayerTrades[name]
	if !ok {
		return fmt.Errorf("you are not in a trade")
	}

	trade := m.Trades[tradeID]
	trade.mu.Lock()
	defer trade.mu.Unlock()

	if trade.State != StateActive {
		return fmt.Errorf("trade is not active")
	}

	trade.InitiatorOffer.Confirmed = false
	trade.TargetOffer.Confirmed = false

	if strings.ToLower(trade.Initiator) == name {
		trade.InitiatorOffer.Money = amount
	} else {
		trade.TargetOffer.Money = amount
	}

	trade.UpdatedAt = time.Now()
	return nil
}

// ConfirmTrade confirms a player's side of the trade
func (m *Manager) ConfirmTrade(playerName string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	tradeID, ok := m.PlayerTrades[name]
	if !ok {
		return false, fmt.Errorf("you are not in a trade")
	}

	trade := m.Trades[tradeID]
	trade.mu.Lock()
	defer trade.mu.Unlock()

	if trade.State != StateActive {
		return false, fmt.Errorf("trade is not active")
	}

	if strings.ToLower(trade.Initiator) == name {
		trade.InitiatorOffer.Confirmed = true
	} else {
		trade.TargetOffer.Confirmed = true
	}

	if trade.InitiatorOffer.Confirmed && trade.TargetOffer.Confirmed {
		trade.State = StateCompleted
		trade.UpdatedAt = time.Now()
		delete(m.PlayerTrades, strings.ToLower(trade.Initiator))
		delete(m.PlayerTrades, strings.ToLower(trade.Target))
		return true, nil
	}

	trade.UpdatedAt = time.Now()
	return false, nil
}

// GetTrade returns a player's current trade
func (m *Manager) GetTrade(playerName string) *Trade {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	tradeID, ok := m.PlayerTrades[name]
	if !ok {
		return nil
	}
	return m.Trades[tradeID]
}

// FormatTrade formats a trade for display
func (m *Manager) FormatTrade(trade *Trade, viewerName string) string {
	if trade == nil {
		return "No active trade."
	}

	trade.mu.RLock()
	defer trade.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("=== TRADE ===\r\n\r\n")

	sb.WriteString(fmt.Sprintf("%s's Offer:", trade.Initiator))
	if trade.InitiatorOffer.Confirmed {
		sb.WriteString(" [CONFIRMED]")
	}
	sb.WriteString("\r\n")
	for _, item := range trade.InitiatorOffer.Items {
		sb.WriteString(fmt.Sprintf("  - %s x%d\r\n", item.Name, item.Quantity))
	}
	if trade.InitiatorOffer.Money > 0 {
		sb.WriteString(fmt.Sprintf("  - $%d credits\r\n", trade.InitiatorOffer.Money))
	}
	if len(trade.InitiatorOffer.Items) == 0 && trade.InitiatorOffer.Money == 0 {
		sb.WriteString("  (nothing)\r\n")
	}

	sb.WriteString("\r\n")

	sb.WriteString(fmt.Sprintf("%s's Offer:", trade.Target))
	if trade.TargetOffer.Confirmed {
		sb.WriteString(" [CONFIRMED]")
	}
	sb.WriteString("\r\n")
	for _, item := range trade.TargetOffer.Items {
		sb.WriteString(fmt.Sprintf("  - %s x%d\r\n", item.Name, item.Quantity))
	}
	if trade.TargetOffer.Money > 0 {
		sb.WriteString(fmt.Sprintf("  - $%d credits\r\n", trade.TargetOffer.Money))
	}
	if len(trade.TargetOffer.Items) == 0 && trade.TargetOffer.Money == 0 {
		sb.WriteString("  (nothing)\r\n")
	}

	sb.WriteString("\r\nCommands: trade add <item>, trade remove <item>, trade money <amount>, trade confirm, trade cancel\r\n")

	return sb.String()
}

// HasPendingTrade checks if a player has a pending trade request
func (m *Manager) HasPendingTrade(playerName string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	for _, trade := range m.Trades {
		if strings.ToLower(trade.Target) == name && trade.State == StatePending {
			return trade.Initiator, true
		}
	}
	return "", false
}

// IsTrading checks if a player is currently in a trade
func (m *Manager) IsTrading(playerName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.PlayerTrades[strings.ToLower(playerName)]
	return ok
}

// CreateListing creates a new auction listing
func (m *Manager) CreateListing(sellerName, itemID, itemName string, quantity, startPrice, buyoutPrice int, duration time.Duration, category string) (*AuctionListing, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if startPrice <= 0 {
		return nil, fmt.Errorf("start price must be positive")
	}
	if buyoutPrice > 0 && buyoutPrice < startPrice {
		return nil, fmt.Errorf("buyout price must be >= start price")
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	listingID := fmt.Sprintf("auction_%d", m.nextAuctionID)
	m.nextAuctionID++

	listing := &AuctionListing{
		ID:          listingID,
		SellerName:  sellerName,
		ItemID:      itemID,
		ItemName:    itemName,
		Quantity:    quantity,
		StartPrice:  startPrice,
		BuyoutPrice: buyoutPrice,
		CurrentBid:  0,
		ExpiresAt:   time.Now().Add(duration),
		CreatedAt:   time.Now(),
		Category:    category,
	}

	m.Auctions[listingID] = listing

	seller := strings.ToLower(sellerName)
	m.PlayerAuctions[seller] = append(m.PlayerAuctions[seller], listingID)

	return listing, nil
}

// PlaceBid places a bid on an auction
func (m *Manager) PlaceBid(bidderName, listingID string, amount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bidder := strings.ToLower(bidderName)

	listing, ok := m.Auctions[listingID]
	if !ok {
		return fmt.Errorf("listing not found")
	}

	if listing.Sold {
		return fmt.Errorf("this auction has ended")
	}

	if time.Now().After(listing.ExpiresAt) {
		return fmt.Errorf("this auction has expired")
	}

	if strings.ToLower(listing.SellerName) == bidder {
		return fmt.Errorf("you cannot bid on your own auction")
	}

	minBid := listing.StartPrice
	if listing.CurrentBid > 0 {
		minBid = listing.CurrentBid + 1
	}

	if amount < minBid {
		return fmt.Errorf("bid must be at least %d", minBid)
	}

	if listing.BuyoutPrice > 0 && amount >= listing.BuyoutPrice {
		listing.CurrentBid = listing.BuyoutPrice
		listing.CurrentBidder = bidderName
		listing.Sold = true
		m.recordSale(listing.ItemID, listing.BuyoutPrice, listing.Quantity)
		return nil
	}

	listing.CurrentBid = amount
	listing.CurrentBidder = bidderName

	return nil
}

// Buyout purchases an item at the buyout price
func (m *Manager) Buyout(buyerName, listingID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	buyer := strings.ToLower(buyerName)

	listing, ok := m.Auctions[listingID]
	if !ok {
		return fmt.Errorf("listing not found")
	}

	if listing.Sold {
		return fmt.Errorf("this auction has ended")
	}

	if listing.BuyoutPrice <= 0 {
		return fmt.Errorf("this auction has no buyout price")
	}

	if strings.ToLower(listing.SellerName) == buyer {
		return fmt.Errorf("you cannot buy your own listing")
	}

	listing.CurrentBid = listing.BuyoutPrice
	listing.CurrentBidder = buyerName
	listing.Sold = true

	m.recordSale(listing.ItemID, listing.BuyoutPrice, listing.Quantity)

	return nil
}

// CancelListing cancels an auction listing
func (m *Manager) CancelListing(playerName, listingID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)

	listing, ok := m.Auctions[listingID]
	if !ok {
		return fmt.Errorf("listing not found")
	}

	if strings.ToLower(listing.SellerName) != name {
		return fmt.Errorf("you do not own this listing")
	}

	if listing.CurrentBid > 0 {
		return fmt.Errorf("cannot cancel listing with active bids")
	}

	if listing.Sold {
		return fmt.Errorf("this auction has already sold")
	}

	delete(m.Auctions, listingID)

	for i, id := range m.PlayerAuctions[name] {
		if id == listingID {
			m.PlayerAuctions[name] = append(m.PlayerAuctions[name][:i], m.PlayerAuctions[name][i+1:]...)
			break
		}
	}

	return nil
}

// SearchAuctions searches for auction listings
func (m *Manager) SearchAuctions(query, category string, maxPrice int) []*AuctionListing {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]*AuctionListing, 0)
	query = strings.ToLower(query)

	for _, listing := range m.Auctions {
		if listing.Sold {
			continue
		}
		if time.Now().After(listing.ExpiresAt) {
			continue
		}

		if query != "" && !strings.Contains(strings.ToLower(listing.ItemName), query) {
			continue
		}

		if category != "" && listing.Category != category {
			continue
		}

		price := listing.StartPrice
		if listing.CurrentBid > 0 {
			price = listing.CurrentBid
		}
		if maxPrice > 0 && price > maxPrice {
			continue
		}

		results = append(results, listing)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ExpiresAt.Before(results[j].ExpiresAt)
	})

	return results
}

// GetPlayerListings returns a player's auction listings
func (m *Manager) GetPlayerListings(playerName string) []*AuctionListing {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	listingIDs := m.PlayerAuctions[name]

	results := make([]*AuctionListing, 0)
	for _, id := range listingIDs {
		if listing, ok := m.Auctions[id]; ok {
			results = append(results, listing)
		}
	}
	return results
}

// FormatListings formats auction listings for display
func (m *Manager) FormatListings(listings []*AuctionListing) string {
	if len(listings) == 0 {
		return "No listings found.\r\n"
	}

	var sb strings.Builder
	sb.WriteString("=== AUCTION HOUSE ===\r\n\r\n")
	sb.WriteString("ID          Item                    Price     Buyout    Time Left\r\n")
	sb.WriteString("---------   --------------------    ------    ------    ---------\r\n")

	for _, l := range listings {
		price := l.StartPrice
		if l.CurrentBid > 0 {
			price = l.CurrentBid
		}

		buyout := "-"
		if l.BuyoutPrice > 0 {
			buyout = fmt.Sprintf("$%d", l.BuyoutPrice)
		}

		timeLeft := time.Until(l.ExpiresAt).Round(time.Minute)
		timeStr := fmt.Sprintf("%dh %dm", int(timeLeft.Hours()), int(timeLeft.Minutes())%60)
		if timeLeft < time.Hour {
			timeStr = fmt.Sprintf("%dm", int(timeLeft.Minutes()))
		}

		sb.WriteString(fmt.Sprintf("%-11s %-23s $%-7d %-9s %s\r\n",
			l.ID, truncate(l.ItemName, 20)+" x"+fmt.Sprint(l.Quantity),
			price, buyout, timeStr))
	}

	sb.WriteString("\r\nCommands: auction bid <id> <amount>, auction buyout <id>\r\n")

	return sb.String()
}

// recordSale records a sale for price tracking
func (m *Manager) recordSale(itemID string, price, quantity int) {
	history, ok := m.PriceHistory[itemID]
	if !ok {
		history = &PriceHistory{
			ItemID:      itemID,
			PricePoints: make([]PricePoint, 0),
		}
		m.PriceHistory[itemID] = history
	}

	history.PricePoints = append(history.PricePoints, PricePoint{
		Price:     price,
		Quantity:  quantity,
		Timestamp: time.Now(),
	})

	if len(history.PricePoints) > 100 {
		history.PricePoints = history.PricePoints[len(history.PricePoints)-100:]
	}

	history.TotalSold += quantity
	history.LastSoldAt = time.Now()

	var total, min, max int
	min = price
	for _, pp := range history.PricePoints {
		total += pp.Price
		if pp.Price < min {
			min = pp.Price
		}
		if pp.Price > max {
			max = pp.Price
		}
	}
	history.AveragePrice = total / len(history.PricePoints)
	history.MinPrice = min
	history.MaxPrice = max
}

// GetMarketPrice returns the market price for an item
func (m *Manager) GetMarketPrice(itemID string) (avg, min, maxP int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, ok := m.PriceHistory[itemID]
	if !ok {
		return 0, 0, 0
	}
	return history.AveragePrice, history.MinPrice, history.MaxPrice
}

// GetPriceInfo returns formatted price info for an item
func (m *Manager) GetPriceInfo(itemID, itemName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, ok := m.PriceHistory[itemID]
	if !ok {
		return fmt.Sprintf("No price history for %s.\r\n", itemName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== Price History: %s ===\r\n\r\n", itemName))
	sb.WriteString(fmt.Sprintf("Average Price: $%d\r\n", history.AveragePrice))
	sb.WriteString(fmt.Sprintf("Price Range: $%d - $%d\r\n", history.MinPrice, history.MaxPrice))
	sb.WriteString(fmt.Sprintf("Total Sold: %d\r\n", history.TotalSold))
	if !history.LastSoldAt.IsZero() {
		sb.WriteString(fmt.Sprintf("Last Sold: %s ago\r\n", time.Since(history.LastSoldAt).Round(time.Minute)))
	}

	return sb.String()
}

// ProcessExpiredAuctions processes expired auctions
func (m *Manager) ProcessExpiredAuctions() []ExpiredAuction {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	results := make([]ExpiredAuction, 0)

	for id, listing := range m.Auctions {
		if listing.Sold {
			continue
		}
		if now.After(listing.ExpiresAt) {
			expired := ExpiredAuction{
				Listing: listing,
				HasBid:  listing.CurrentBid > 0,
			}
			results = append(results, expired)

			if listing.CurrentBid > 0 {
				m.recordSale(listing.ItemID, listing.CurrentBid, listing.Quantity)
				listing.Sold = true
			} else {
				delete(m.Auctions, id)
			}
		}
	}

	return results
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// Global trade manager
var GlobalTrade = NewManager()
