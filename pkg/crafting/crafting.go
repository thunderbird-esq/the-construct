// Package crafting provides the item crafting system for Matrix MUD.
package crafting

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/yourusername/matrix-mud/pkg/logging"
)

// Ingredient represents a required crafting material
type Ingredient struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// RecipeResult represents the output of a recipe
type RecipeResult struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// Recipe represents a crafting recipe
type Recipe struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Ingredients   []Ingredient `json:"ingredients"`
	Result        RecipeResult `json:"result"`
	SkillRequired int          `json:"skill_required"`
	XPReward      int          `json:"xp_reward"`
}

// RecipesData is the JSON structure for recipes.json
type RecipesData struct {
	Recipes map[string]*Recipe `json:"recipes"`
}

// Manager handles crafting operations
type Manager struct {
	Recipes map[string]*Recipe
}

// NewManager creates a new crafting manager
func NewManager() *Manager {
	m := &Manager{
		Recipes: make(map[string]*Recipe),
	}
	m.loadRecipes()
	return m
}

// loadRecipes loads recipes from data/recipes.json
func (m *Manager) loadRecipes() {
	file, err := os.ReadFile("data/recipes.json")
	if err != nil {
		logging.Warn().Err(err).Msg("Could not read recipes.json")
		return
	}

	var data RecipesData
	if err := json.Unmarshal(file, &data); err != nil {
		logging.Warn().Err(err).Msg("Could not parse recipes.json")
		return
	}

	m.Recipes = data.Recipes
	logging.Info().Int("count", len(m.Recipes)).Msg("Loaded crafting recipes")
}

// GetRecipe returns a recipe by ID (case-insensitive partial match)
func (m *Manager) GetRecipe(name string) *Recipe {
	name = strings.ToLower(name)

	// Exact match first
	if r, ok := m.Recipes[name]; ok {
		return r
	}

	// Partial match
	for id, recipe := range m.Recipes {
		if strings.HasPrefix(strings.ToLower(id), name) ||
			strings.HasPrefix(strings.ToLower(recipe.Name), name) {
			return recipe
		}
	}

	return nil
}

// ListRecipes returns all available recipes
func (m *Manager) ListRecipes() []*Recipe {
	recipes := make([]*Recipe, 0, len(m.Recipes))
	for _, r := range m.Recipes {
		recipes = append(recipes, r)
	}
	return recipes
}

// CanCraft checks if crafting is possible with given inventory
// Returns: canCraft bool, missingItems map[string]int
func (m *Manager) CanCraft(recipe *Recipe, inventory map[string]int, craftingSkill int) (bool, map[string]int) {
	if craftingSkill < recipe.SkillRequired {
		return false, nil
	}

	missing := make(map[string]int)
	for _, ing := range recipe.Ingredients {
		have := inventory[ing.ItemID]
		if have < ing.Quantity {
			missing[ing.ItemID] = ing.Quantity - have
		}
	}

	return len(missing) == 0, missing
}

// GetRecipeList returns a formatted list of recipes for display
func (m *Manager) GetRecipeList() string {
	var sb strings.Builder
	sb.WriteString("=== CRAFTING RECIPES ===\r\n\r\n")

	for _, recipe := range m.Recipes {
		sb.WriteString(recipe.Name)
		if recipe.SkillRequired > 0 {
			sb.WriteString(" (Skill: ")
			sb.WriteString(string(rune('0' + recipe.SkillRequired)))
			sb.WriteString(")")
		}
		sb.WriteString("\r\n  ")
		sb.WriteString(recipe.Description)
		sb.WriteString("\r\n  Ingredients: ")

		for i, ing := range recipe.Ingredients {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(ing.ItemID)
			sb.WriteString(" x")
			sb.WriteString(string(rune('0' + ing.Quantity)))
		}
		sb.WriteString("\r\n\r\n")
	}

	return sb.String()
}
