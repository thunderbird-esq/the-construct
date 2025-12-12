package help

import (
	"testing"
)

func TestGetHelp(t *testing.T) {
	tests := []struct {
		cmd    string
		expect bool
	}{
		{"look", true},
		{"l", true},         // alias
		{"north", true},
		{"n", true},         // alias
		{"kill", true},
		{"k", true},         // alias
		{"nonexistent", false},
	}

	for _, tt := range tests {
		entry := GetHelp(tt.cmd)
		found := entry != nil
		if found != tt.expect {
			t.Errorf("GetHelp(%q) = %v, want %v", tt.cmd, found, tt.expect)
		}
	}
}

func TestGetHelpContent(t *testing.T) {
	entry := GetHelp("look")
	if entry == nil {
		t.Fatal("look entry should exist")
	}

	if entry.Command != "look" {
		t.Errorf("Command = %q, want look", entry.Command)
	}
	if entry.Description == "" {
		t.Error("Description should not be empty")
	}
	if entry.Usage == "" {
		t.Error("Usage should not be empty")
	}
	if len(entry.Aliases) == 0 {
		t.Error("look should have aliases")
	}
}

func TestGetAllByCategory(t *testing.T) {
	byCategory := GetAllByCategory()

	if len(byCategory) == 0 {
		t.Error("Should have categories")
	}

	// Check that movement has entries
	movement := byCategory[CatMovement]
	if len(movement) == 0 {
		t.Error("Movement category should have entries")
	}

	// Check combat has entries
	combat := byCategory[CatCombat]
	if len(combat) == 0 {
		t.Error("Combat category should have entries")
	}
}

func TestGetCategories(t *testing.T) {
	categories := GetCategories()

	if len(categories) == 0 {
		t.Error("Should have categories")
	}

	// Check expected categories exist
	expected := []string{CatMovement, CatCombat, CatItems, CatSocial, CatInfo, CatEconomy, CatSystem}
	for _, exp := range expected {
		found := false
		for _, cat := range categories {
			if cat == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing category: %s", exp)
		}
	}
}

func TestAllEntriesHaveRequiredFields(t *testing.T) {
	for cmd, entry := range Entries {
		if entry.Command == "" {
			t.Errorf("Entry %s has no Command", cmd)
		}
		if entry.Description == "" {
			t.Errorf("Entry %s has no Description", cmd)
		}
		if entry.Usage == "" {
			t.Errorf("Entry %s has no Usage", cmd)
		}
		if entry.Category == "" {
			t.Errorf("Entry %s has no Category", cmd)
		}
	}
}

func TestAliasesResolve(t *testing.T) {
	// Test each alias resolves to its parent
	for _, entry := range Entries {
		for _, alias := range entry.Aliases {
			resolved := GetHelp(alias)
			if resolved == nil {
				t.Errorf("Alias %q for %s does not resolve", alias, entry.Command)
			}
			if resolved != entry {
				t.Errorf("Alias %q resolved to wrong entry", alias)
			}
		}
	}
}
