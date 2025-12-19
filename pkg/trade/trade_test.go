package trade

import (
	"strings"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Trades == nil {
		t.Error("Trades map is nil")
	}
	if m.Auctions == nil {
		t.Error("Auctions map is nil")
	}
}

// ========================
// DIRECT TRADING TESTS
// ========================

func TestInitiateTrade(t *testing.T) {
	m := NewManager()

	trade, err := m.InitiateTrade("Player1", "Player2")
	if err != nil {
		t.Fatalf("InitiateTrade failed: %v", err)
	}
	if trade == nil {
		t.Fatal("Trade is nil")
	}
	if trade.State != StatePending {
		t.Errorf("Trade state should be pending, got %s", trade.State)
	}
}

func TestInitiateTradeSelf(t *testing.T) {
	m := NewManager()

	_, err := m.InitiateTrade("Player1", "Player1")
	if err == nil {
		t.Error("Should not allow trading with self")
	}
}

func TestInitiateTradeAlreadyTrading(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	_, err := m.InitiateTrade("Player1", "Player3")
	if err == nil {
		t.Error("Should not allow initiating second trade")
	}
}

func TestAcceptTrade(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	err := m.AcceptTrade("Player2")
	if err != nil {
		t.Fatalf("AcceptTrade failed: %v", err)
	}

	trade := m.GetTrade("Player2")
	if trade.State != StateActive {
		t.Errorf("Trade state should be active, got %s", trade.State)
	}
}

func TestAcceptTradeNoPending(t *testing.T) {
	m := NewManager()

	err := m.AcceptTrade("Player1")
	if err == nil {
		t.Error("Should fail when no pending trade")
	}
}

func TestDeclineTrade(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	err := m.DeclineTrade("Player2")
	if err != nil {
		t.Fatalf("DeclineTrade failed: %v", err)
	}

	if m.IsTrading("Player1") {
		t.Error("Player1 should not be trading after decline")
	}
}

func TestCancelTrade(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	m.AcceptTrade("Player2")

	err := m.CancelTrade("Player1")
	if err != nil {
		t.Fatalf("CancelTrade failed: %v", err)
	}

	if m.IsTrading("Player1") || m.IsTrading("Player2") {
		t.Error("Neither player should be trading after cancel")
	}
}

func TestAddItem(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	m.AcceptTrade("Player2")

	err := m.AddItem("Player1", "sword", "Steel Sword", 1)
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}

	trade := m.GetTrade("Player1")
	if len(trade.InitiatorOffer.Items) != 1 {
		t.Error("Should have 1 item in offer")
	}
}

func TestAddItemNotTrading(t *testing.T) {
	m := NewManager()

	err := m.AddItem("Player1", "sword", "Steel Sword", 1)
	if err == nil {
		t.Error("Should fail when not trading")
	}
}

func TestRemoveItem(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	m.AcceptTrade("Player2")
	m.AddItem("Player1", "sword", "Steel Sword", 1)

	err := m.RemoveItem("Player1", "sword")
	if err != nil {
		t.Fatalf("RemoveItem failed: %v", err)
	}

	trade := m.GetTrade("Player1")
	if len(trade.InitiatorOffer.Items) != 0 {
		t.Error("Should have 0 items in offer")
	}
}

func TestRemoveItemNotFound(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	m.AcceptTrade("Player2")

	err := m.RemoveItem("Player1", "nonexistent")
	if err == nil {
		t.Error("Should fail for nonexistent item")
	}
}

func TestSetMoney(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	m.AcceptTrade("Player2")

	err := m.SetMoney("Player1", 100)
	if err != nil {
		t.Fatalf("SetMoney failed: %v", err)
	}

	trade := m.GetTrade("Player1")
	if trade.InitiatorOffer.Money != 100 {
		t.Errorf("Money should be 100, got %d", trade.InitiatorOffer.Money)
	}
}

func TestSetMoneyNegative(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	m.AcceptTrade("Player2")

	err := m.SetMoney("Player1", -100)
	if err == nil {
		t.Error("Should not allow negative money")
	}
}

func TestConfirmTrade(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	m.AcceptTrade("Player2")

	// First confirm
	completed, err := m.ConfirmTrade("Player1")
	if err != nil {
		t.Fatalf("First ConfirmTrade failed: %v", err)
	}
	if completed {
		t.Error("Trade should not be complete with only one confirmation")
	}

	// Second confirm
	completed, err = m.ConfirmTrade("Player2")
	if err != nil {
		t.Fatalf("Second ConfirmTrade failed: %v", err)
	}
	if !completed {
		t.Error("Trade should be complete with both confirmations")
	}
}

func TestConfirmResetOnChange(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	m.AcceptTrade("Player2")
	m.ConfirmTrade("Player1")

	// Add item should reset confirmation
	m.AddItem("Player2", "item", "Item", 1)

	trade := m.GetTrade("Player1")
	if trade.InitiatorOffer.Confirmed {
		t.Error("Confirmation should be reset after item change")
	}
}

func TestFormatTrade(t *testing.T) {
	m := NewManager()

	m.InitiateTrade("Player1", "Player2")
	m.AcceptTrade("Player2")
	m.AddItem("Player1", "sword", "Steel Sword", 1)
	m.SetMoney("Player2", 50)

	trade := m.GetTrade("Player1")
	output := m.FormatTrade(trade, "Player1")

	if !strings.Contains(output, "Player1") {
		t.Error("Should show Player1")
	}
	if !strings.Contains(output, "Player2") {
		t.Error("Should show Player2")
	}
	if !strings.Contains(output, "Steel Sword") {
		t.Error("Should show item")
	}
	if !strings.Contains(output, "50") {
		t.Error("Should show money")
	}
}

func TestIsTrading(t *testing.T) {
	m := NewManager()

	if m.IsTrading("Player1") {
		t.Error("Should not be trading initially")
	}

	m.InitiateTrade("Player1", "Player2")

	if !m.IsTrading("Player1") {
		t.Error("Initiator should be trading")
	}
	if m.IsTrading("Player2") {
		t.Error("Target should not be trading until accept")
	}
}

func TestHasPendingTrade(t *testing.T) {
	m := NewManager()

	initiator, has := m.HasPendingTrade("Player2")
	if has {
		t.Error("Should not have pending trade initially")
	}

	m.InitiateTrade("Player1", "Player2")

	initiator, has = m.HasPendingTrade("Player2")
	if !has {
		t.Error("Should have pending trade")
	}
	if initiator != "Player1" {
		t.Errorf("Initiator should be Player1, got %s", initiator)
	}
}

// ========================
// AUCTION HOUSE TESTS
// ========================

func TestCreateListing(t *testing.T) {
	m := NewManager()

	listing, err := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")
	if err != nil {
		t.Fatalf("CreateListing failed: %v", err)
	}
	if listing == nil {
		t.Fatal("Listing is nil")
	}
	if listing.StartPrice != 100 {
		t.Errorf("Start price should be 100, got %d", listing.StartPrice)
	}
	if listing.BuyoutPrice != 200 {
		t.Errorf("Buyout price should be 200, got %d", listing.BuyoutPrice)
	}
}

func TestCreateListingInvalidPrice(t *testing.T) {
	m := NewManager()

	_, err := m.CreateListing("Seller", "sword", "Steel Sword", 1, 0, 200, 24*time.Hour, "weapons")
	if err == nil {
		t.Error("Should fail with 0 start price")
	}

	_, err = m.CreateListing("Seller", "sword", "Steel Sword", 1, 200, 100, 24*time.Hour, "weapons")
	if err == nil {
		t.Error("Should fail with buyout < start price")
	}
}

func TestPlaceBid(t *testing.T) {
	m := NewManager()

	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")

	err := m.PlaceBid("Bidder", listing.ID, 100)
	if err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}

	if listing.CurrentBid != 100 {
		t.Errorf("Current bid should be 100, got %d", listing.CurrentBid)
	}
	if listing.CurrentBidder != "Bidder" {
		t.Errorf("Current bidder should be Bidder, got %s", listing.CurrentBidder)
	}
}

func TestPlaceBidTooLow(t *testing.T) {
	m := NewManager()

	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")
	m.PlaceBid("Bidder1", listing.ID, 100)

	err := m.PlaceBid("Bidder2", listing.ID, 100)
	if err == nil {
		t.Error("Should fail with bid not higher than current")
	}
}

func TestPlaceBidOnOwn(t *testing.T) {
	m := NewManager()

	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")

	err := m.PlaceBid("Seller", listing.ID, 100)
	if err == nil {
		t.Error("Should not allow bidding on own auction")
	}
}

func TestBuyout(t *testing.T) {
	m := NewManager()

	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")

	err := m.Buyout("Buyer", listing.ID)
	if err != nil {
		t.Fatalf("Buyout failed: %v", err)
	}

	if !listing.Sold {
		t.Error("Listing should be sold")
	}
	if listing.CurrentBid != 200 {
		t.Errorf("Current bid should be buyout price, got %d", listing.CurrentBid)
	}
}

func TestBuyoutNoBuyoutPrice(t *testing.T) {
	m := NewManager()

	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 0, 24*time.Hour, "weapons")

	err := m.Buyout("Buyer", listing.ID)
	if err == nil {
		t.Error("Should fail when no buyout price")
	}
}

func TestCancelListing(t *testing.T) {
	m := NewManager()

	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")

	err := m.CancelListing("Seller", listing.ID)
	if err != nil {
		t.Fatalf("CancelListing failed: %v", err)
	}

	if _, ok := m.Auctions[listing.ID]; ok {
		t.Error("Listing should be removed")
	}
}

func TestCancelListingWithBids(t *testing.T) {
	m := NewManager()

	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")
	m.PlaceBid("Bidder", listing.ID, 100)

	err := m.CancelListing("Seller", listing.ID)
	if err == nil {
		t.Error("Should not allow canceling with active bids")
	}
}

func TestCancelListingNotOwner(t *testing.T) {
	m := NewManager()

	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")

	err := m.CancelListing("NotOwner", listing.ID)
	if err == nil {
		t.Error("Should not allow non-owner to cancel")
	}
}

func TestSearchAuctions(t *testing.T) {
	m := NewManager()

	m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")
	m.CreateListing("Seller", "shield", "Iron Shield", 1, 50, 100, 24*time.Hour, "armor")
	m.CreateListing("Seller", "katana", "Katana", 1, 150, 300, 24*time.Hour, "weapons")

	// Search by name
	results := m.SearchAuctions("sword", "", 0)
	if len(results) != 1 {
		t.Errorf("Should find 1 sword, got %d", len(results))
	}

	// Search by category
	results = m.SearchAuctions("", "weapons", 0)
	if len(results) != 2 {
		t.Errorf("Should find 2 weapons, got %d", len(results))
	}

	// Search by max price
	results = m.SearchAuctions("", "", 100)
	if len(results) != 2 {
		t.Errorf("Should find 2 items under 100, got %d", len(results))
	}
}

func TestGetPlayerListings(t *testing.T) {
	m := NewManager()

	m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")
	m.CreateListing("Seller", "shield", "Iron Shield", 1, 50, 100, 24*time.Hour, "armor")
	m.CreateListing("Other", "item", "Item", 1, 10, 20, 24*time.Hour, "misc")

	listings := m.GetPlayerListings("Seller")
	if len(listings) != 2 {
		t.Errorf("Should find 2 listings for Seller, got %d", len(listings))
	}
}

func TestFormatListings(t *testing.T) {
	m := NewManager()

	m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")

	listings := m.SearchAuctions("", "", 0)
	output := m.FormatListings(listings)

	if !strings.Contains(output, "Steel Sword") {
		t.Error("Should show item name")
	}
	if !strings.Contains(output, "100") {
		t.Error("Should show price")
	}
}

// ========================
// PRICE TRACKING TESTS
// ========================

func TestRecordSale(t *testing.T) {
	m := NewManager()

	// Create and complete a sale
	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")
	m.Buyout("Buyer", listing.ID)

	avg, min, max := m.GetMarketPrice("sword")
	if avg != 200 {
		t.Errorf("Average should be 200, got %d", avg)
	}
	if min != 200 {
		t.Errorf("Min should be 200, got %d", min)
	}
	if max != 200 {
		t.Errorf("Max should be 200, got %d", max)
	}
}

func TestGetMarketPriceNoHistory(t *testing.T) {
	m := NewManager()

	avg, min, max := m.GetMarketPrice("nonexistent")
	if avg != 0 || min != 0 || max != 0 {
		t.Error("Should return 0 for no history")
	}
}

func TestGetPriceInfo(t *testing.T) {
	m := NewManager()

	output := m.GetPriceInfo("nonexistent", "Unknown Item")
	if !strings.Contains(output, "No price history") {
		t.Error("Should indicate no history")
	}

	// Create and complete a sale
	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, 24*time.Hour, "weapons")
	m.Buyout("Buyer", listing.ID)

	output = m.GetPriceInfo("sword", "Steel Sword")
	if !strings.Contains(output, "Average Price") {
		t.Error("Should show average price")
	}
}

func TestProcessExpiredAuctions(t *testing.T) {
	m := NewManager()

	// Create listing that expires immediately
	listing, _ := m.CreateListing("Seller", "sword", "Steel Sword", 1, 100, 200, time.Millisecond, "weapons")
	m.PlaceBid("Bidder", listing.ID, 150)

	time.Sleep(10 * time.Millisecond)

	expired := m.ProcessExpiredAuctions()
	if len(expired) != 1 {
		t.Errorf("Should have 1 expired auction, got %d", len(expired))
	}
	if !expired[0].HasBid {
		t.Error("Expired auction should have bid")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exact", 5, "exact"},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}
