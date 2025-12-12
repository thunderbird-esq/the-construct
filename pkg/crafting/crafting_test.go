package crafting

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Change to project root so data/recipes.json can be found
	os.Chdir("../../")
	os.Exit(m.Run())
}

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Recipes == nil {
		t.Error("Recipes map not initialized")
	}
}

func TestGetRecipeExact(t *testing.T) {
	m := NewManager()
	if len(m.Recipes) == 0 {
		t.Skip("No recipes loaded")
	}
	
	recipe := m.GetRecipe("health_vial")
	if recipe == nil {
		t.Fatal("Should find health_vial recipe")
	}
	if recipe.ID != "health_vial" {
		t.Errorf("Recipe ID = %q, want health_vial", recipe.ID)
	}
}

func TestGetRecipePartial(t *testing.T) {
	m := NewManager()
	if len(m.Recipes) == 0 {
		t.Skip("No recipes loaded")
	}
	
	recipe := m.GetRecipe("health")
	if recipe == nil {
		t.Fatal("Should find recipe by partial match 'health'")
	}
}

func TestGetRecipeCaseInsensitive(t *testing.T) {
	m := NewManager()
	if len(m.Recipes) == 0 {
		t.Skip("No recipes loaded")
	}
	
	recipe := m.GetRecipe("HEALTH_VIAL")
	if recipe == nil {
		t.Fatal("Should find recipe case-insensitively")
	}
}

func TestGetRecipeNotFound(t *testing.T) {
	m := NewManager()
	
	recipe := m.GetRecipe("nonexistent_recipe_xyz")
	if recipe != nil {
		t.Error("Should return nil for nonexistent recipe")
	}
}

func TestListRecipes(t *testing.T) {
	m := NewManager()
	if len(m.Recipes) == 0 {
		t.Skip("No recipes loaded")
	}
	
	recipes := m.ListRecipes()
	if len(recipes) == 0 {
		t.Error("Should have at least one recipe")
	}
	
	for _, r := range recipes {
		if r.ID == "" {
			t.Error("Recipe missing ID")
		}
		if r.Name == "" {
			t.Error("Recipe missing Name")
		}
		if len(r.Ingredients) == 0 {
			t.Errorf("Recipe %s has no ingredients", r.ID)
		}
		if r.Result.ItemID == "" {
			t.Errorf("Recipe %s has no result item", r.ID)
		}
	}
}

func TestRecipeIngredients(t *testing.T) {
	m := NewManager()
	recipe := m.GetRecipe("health_vial")
	if recipe == nil {
		t.Skip("health_vial recipe not found")
	}
	
	if len(recipe.Ingredients) == 0 {
		t.Error("health_vial should have ingredients")
	}
	
	for _, ing := range recipe.Ingredients {
		if ing.ItemID == "" {
			t.Error("Ingredient missing ItemID")
		}
		if ing.Quantity <= 0 {
			t.Errorf("Ingredient %s has invalid quantity %d", ing.ItemID, ing.Quantity)
		}
	}
}

func TestCanCraft(t *testing.T) {
	m := NewManager()
	recipe := m.GetRecipe("health_vial")
	if recipe == nil {
		t.Skip("health_vial recipe not found")
	}
	
	// With sufficient materials
	inventory := map[string]int{"trash": 10}
	canCraft, missing := m.CanCraft(recipe, inventory, 0)
	if !canCraft {
		t.Errorf("Should be able to craft with sufficient materials, missing: %v", missing)
	}
	
	// Without sufficient materials
	inventory = map[string]int{"trash": 1}
	canCraft, missing = m.CanCraft(recipe, inventory, 0)
	if canCraft {
		t.Error("Should not be able to craft without sufficient materials")
	}
	if missing["trash"] != 2 {
		t.Errorf("Missing trash = %d, want 2", missing["trash"])
	}
}

func TestCanCraftSkillRequired(t *testing.T) {
	m := NewManager()
	recipe := m.GetRecipe("code_blade")
	if recipe == nil {
		t.Skip("code_blade recipe not found")
	}
	
	inventory := map[string]int{"katana": 1, "red_pill": 1, "trash": 10}
	
	// Without skill
	canCraft, _ := m.CanCraft(recipe, inventory, 0)
	if canCraft {
		t.Error("Should not craft without required skill")
	}
	
	// With skill
	canCraft, _ = m.CanCraft(recipe, inventory, 5)
	if !canCraft {
		t.Error("Should be able to craft with skill")
	}
}

func TestGetRecipeList(t *testing.T) {
	m := NewManager()
	if len(m.Recipes) == 0 {
		t.Skip("No recipes loaded")
	}
	
	list := m.GetRecipeList()
	if list == "" {
		t.Error("Recipe list should not be empty")
	}
	if len(list) < 50 {
		t.Error("Recipe list seems too short")
	}
}

func TestAllRecipesValid(t *testing.T) {
	m := NewManager()
	if len(m.Recipes) == 0 {
		t.Skip("No recipes loaded")
	}
	
	for id, recipe := range m.Recipes {
		if recipe.ID != id {
			t.Errorf("Recipe key %q doesn't match ID %q", id, recipe.ID)
		}
		if recipe.Name == "" {
			t.Errorf("Recipe %s missing name", id)
		}
		if recipe.Description == "" {
			t.Errorf("Recipe %s missing description", id)
		}
	}
}
