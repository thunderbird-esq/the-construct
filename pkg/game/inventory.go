// Package game implements core game mechanics for Matrix MUD.
package game

import "fmt"

// Inventory constants
const (
	MaxInventorySize = 20
	MaxBankSize      = 100
)

// InventoryResult represents the outcome of an inventory operation
type InventoryResult struct {
	Success bool
	Message string
	Item    *Item
}

// AddToInventory attempts to add an item to a player's inventory
func AddToInventory(player *Player, item *Item) InventoryResult {
	if len(player.Inventory) >= MaxInventorySize {
		return InventoryResult{
			Success: false,
			Message: "Your inventory is full!",
		}
	}

	player.Inventory = append(player.Inventory, item)
	return InventoryResult{
		Success: true,
		Message: fmt.Sprintf("Got %s.", item.Name),
		Item:    item,
	}
}

// RemoveFromInventory removes an item from a player's inventory by ID
func RemoveFromInventory(player *Player, itemID string) InventoryResult {
	for i, item := range player.Inventory {
		if item.ID == itemID || matchesItemName(item, itemID) {
			player.Inventory = append(player.Inventory[:i], player.Inventory[i+1:]...)
			return InventoryResult{
				Success: true,
				Message: fmt.Sprintf("Dropped %s.", item.Name),
				Item:    item,
			}
		}
	}
	return InventoryResult{
		Success: false,
		Message: "You don't have that.",
	}
}

// FindInInventory finds an item in a player's inventory by ID or name
func FindInInventory(player *Player, search string) *Item {
	for _, item := range player.Inventory {
		if item.ID == search || matchesItemName(item, search) {
			return item
		}
	}
	return nil
}

// EquipItem equips an item from inventory to the appropriate slot
func EquipItem(player *Player, item *Item) InventoryResult {
	if item.Slot == "" {
		return InventoryResult{
			Success: false,
			Message: fmt.Sprintf("%s cannot be equipped.", item.Name),
		}
	}

	// Initialize equipment map if needed
	if player.Equipment == nil {
		player.Equipment = make(map[string]*Item)
	}

	// Unequip existing item in slot
	if existing := player.Equipment[item.Slot]; existing != nil {
		player.Inventory = append(player.Inventory, existing)
	}

	// Remove from inventory and equip
	for i, inv := range player.Inventory {
		if inv == item {
			player.Inventory = append(player.Inventory[:i], player.Inventory[i+1:]...)
			break
		}
	}

	player.Equipment[item.Slot] = item
	return InventoryResult{
		Success: true,
		Message: fmt.Sprintf("Equipped %s.", item.Name),
		Item:    item,
	}
}

// UnequipItem removes an item from equipment and returns it to inventory
func UnequipItem(player *Player, slot string) InventoryResult {
	item := player.Equipment[slot]
	if item == nil {
		return InventoryResult{
			Success: false,
			Message: "Nothing equipped there.",
		}
	}

	if len(player.Inventory) >= MaxInventorySize {
		return InventoryResult{
			Success: false,
			Message: "Your inventory is full!",
		}
	}

	delete(player.Equipment, slot)
	player.Inventory = append(player.Inventory, item)
	return InventoryResult{
		Success: true,
		Message: fmt.Sprintf("Unequipped %s.", item.Name),
		Item:    item,
	}
}

// GetEquippedWeaponDamage returns the damage of the equipped weapon, or 0 if none
func GetEquippedWeaponDamage(player *Player) int {
	if player.Equipment == nil {
		return 0
	}
	if weapon := player.Equipment["hand"]; weapon != nil {
		return weapon.Damage
	}
	return 0
}

// GetTotalAC calculates total armor class from base + equipment
func GetTotalAC(player *Player) int {
	total := player.BaseAC
	if player.Equipment != nil {
		for _, item := range player.Equipment {
			if item != nil {
				total += item.AC
			}
		}
	}
	return total
}

// matchesItemName checks if a search string matches an item's name (case-insensitive prefix)
func matchesItemName(item *Item, search string) bool {
	if len(search) == 0 {
		return false
	}
	name := item.Name
	if len(search) > len(name) {
		return false
	}
	// Simple case-insensitive prefix match
	for i := 0; i < len(search); i++ {
		c1, c2 := search[i], name[i]
		// Convert to lowercase for comparison
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 32
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 32
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}
