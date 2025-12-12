package game

import (
	"testing"
)

func TestAddToInventory(t *testing.T) {
	player := &Player{
		Name:      "TestPlayer",
		Inventory: make([]*Item, 0),
	}

	item := &Item{
		ID:   "test_item",
		Name: "Test Item",
	}

	result := AddToInventory(player, item)
	if !result.Success {
		t.Errorf("Should successfully add item: %s", result.Message)
	}
	if len(player.Inventory) != 1 {
		t.Errorf("Inventory should have 1 item, has %d", len(player.Inventory))
	}
}

func TestInventoryFull(t *testing.T) {
	player := &Player{
		Name:      "TestPlayer",
		Inventory: make([]*Item, MaxInventorySize),
	}

	// Fill inventory
	for i := 0; i < MaxInventorySize; i++ {
		player.Inventory[i] = &Item{ID: "filler"}
	}

	newItem := &Item{
		ID:   "overflow",
		Name: "Overflow Item",
	}

	result := AddToInventory(player, newItem)
	if result.Success {
		t.Error("Should fail when inventory is full")
	}
	if result.Message != "Your inventory is full!" {
		t.Errorf("Wrong error message: %s", result.Message)
	}
}

func TestRemoveFromInventory(t *testing.T) {
	item := &Item{
		ID:   "test_item",
		Name: "Test Item",
	}

	player := &Player{
		Name:      "TestPlayer",
		Inventory: []*Item{item},
	}

	result := RemoveFromInventory(player, "test_item")
	if !result.Success {
		t.Errorf("Should successfully remove item: %s", result.Message)
	}
	if len(player.Inventory) != 0 {
		t.Errorf("Inventory should be empty, has %d items", len(player.Inventory))
	}
}

func TestRemoveNonexistentItem(t *testing.T) {
	player := &Player{
		Name:      "TestPlayer",
		Inventory: make([]*Item, 0),
	}

	result := RemoveFromInventory(player, "nonexistent")
	if result.Success {
		t.Error("Should fail to remove nonexistent item")
	}
}

func TestFindInInventory(t *testing.T) {
	item := &Item{
		ID:   "katana",
		Name: "Training Katana",
	}

	player := &Player{
		Name:      "TestPlayer",
		Inventory: []*Item{item},
	}

	// Find by ID
	found := FindInInventory(player, "katana")
	if found == nil {
		t.Error("Should find item by ID")
	}

	// Find by name prefix
	found = FindInInventory(player, "training")
	if found == nil {
		t.Error("Should find item by name prefix")
	}

	// Not found
	found = FindInInventory(player, "nonexistent")
	if found != nil {
		t.Error("Should not find nonexistent item")
	}
}

func TestEquipItem(t *testing.T) {
	weapon := &Item{
		ID:     "sword",
		Name:   "Sword",
		Slot:   "hand",
		Damage: 5,
	}

	player := &Player{
		Name:      "TestPlayer",
		Inventory: []*Item{weapon},
		Equipment: make(map[string]*Item),
	}

	result := EquipItem(player, weapon)
	if !result.Success {
		t.Errorf("Should successfully equip item: %s", result.Message)
	}
	if player.Equipment["hand"] != weapon {
		t.Error("Weapon should be in hand slot")
	}
	if len(player.Inventory) != 0 {
		t.Error("Weapon should be removed from inventory")
	}
}

func TestEquipNonEquippable(t *testing.T) {
	item := &Item{
		ID:   "trash",
		Name: "Digital Trash",
		Slot: "", // Not equippable
	}

	player := &Player{
		Name:      "TestPlayer",
		Inventory: []*Item{item},
		Equipment: make(map[string]*Item),
	}

	result := EquipItem(player, item)
	if result.Success {
		t.Error("Should not equip non-equippable item")
	}
}

func TestUnequipItem(t *testing.T) {
	weapon := &Item{
		ID:     "sword",
		Name:   "Sword",
		Slot:   "hand",
		Damage: 5,
	}

	player := &Player{
		Name:      "TestPlayer",
		Inventory: make([]*Item, 0),
		Equipment: map[string]*Item{"hand": weapon},
	}

	result := UnequipItem(player, "hand")
	if !result.Success {
		t.Errorf("Should successfully unequip item: %s", result.Message)
	}
	if player.Equipment["hand"] != nil {
		t.Error("Hand slot should be empty")
	}
	if len(player.Inventory) != 1 {
		t.Error("Weapon should be in inventory")
	}
}

func TestUnequipEmptySlot(t *testing.T) {
	player := &Player{
		Name:      "TestPlayer",
		Inventory: make([]*Item, 0),
		Equipment: make(map[string]*Item),
	}

	result := UnequipItem(player, "hand")
	if result.Success {
		t.Error("Should fail to unequip empty slot")
	}
}

func TestGetEquippedWeaponDamage(t *testing.T) {
	weapon := &Item{
		ID:     "sword",
		Name:   "Sword",
		Slot:   "hand",
		Damage: 10,
	}

	// No equipment
	player := &Player{Name: "Test"}
	damage := GetEquippedWeaponDamage(player)
	if damage != 0 {
		t.Errorf("Should return 0 with no equipment, got %d", damage)
	}

	// With weapon
	player.Equipment = map[string]*Item{"hand": weapon}
	damage = GetEquippedWeaponDamage(player)
	if damage != 10 {
		t.Errorf("Should return weapon damage 10, got %d", damage)
	}
}

func TestGetTotalAC(t *testing.T) {
	player := &Player{
		Name:   "Test",
		BaseAC: 10,
		Equipment: map[string]*Item{
			"body": {Name: "Armor", AC: 3},
			"head": {Name: "Helmet", AC: 1},
		},
	}

	ac := GetTotalAC(player)
	expected := 10 + 3 + 1
	if ac != expected {
		t.Errorf("Total AC should be %d, got %d", expected, ac)
	}
}

func TestMatchesItemName(t *testing.T) {
	item := &Item{Name: "Training Katana"}

	tests := []struct {
		search string
		match  bool
	}{
		{"training", true},
		{"Training", true},
		{"TRAINING", true},
		{"tra", true},
		{"katana", false}, // Not a prefix
		{"", false},
		{"training katana long", false},
	}

	for _, tt := range tests {
		result := matchesItemName(item, tt.search)
		if result != tt.match {
			t.Errorf("matchesItemName(%q) = %v, want %v", tt.search, result, tt.match)
		}
	}
}

func TestEquipReplacesExisting(t *testing.T) {
	oldWeapon := &Item{ID: "old", Name: "Old Sword", Slot: "hand", Damage: 3}
	newWeapon := &Item{ID: "new", Name: "New Sword", Slot: "hand", Damage: 5}

	player := &Player{
		Name:      "Test",
		Inventory: []*Item{newWeapon},
		Equipment: map[string]*Item{"hand": oldWeapon},
	}

	result := EquipItem(player, newWeapon)
	if !result.Success {
		t.Errorf("Should successfully equip: %s", result.Message)
	}
	if player.Equipment["hand"] != newWeapon {
		t.Error("New weapon should be equipped")
	}
	if len(player.Inventory) != 1 || player.Inventory[0] != oldWeapon {
		t.Error("Old weapon should be in inventory")
	}
}
