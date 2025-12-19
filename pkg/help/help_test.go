package help

import (
	"strings"
	"testing"
)

func TestGetHelp(t *testing.T) {
	entry := GetHelp("look")
	if entry == nil {
		t.Fatal("GetHelp(look) returned nil")
	}
	if entry.Command != "look" {
		t.Errorf("Command = %s, want look", entry.Command)
	}
}

func TestGetHelpAlias(t *testing.T) {
	entry := GetHelp("l")
	if entry == nil {
		t.Fatal("GetHelp(l) returned nil")
	}
	if entry.Command != "look" {
		t.Errorf("Command = %s, want look", entry.Command)
	}
}

func TestGetHelpCaseInsensitive(t *testing.T) {
	entry := GetHelp("LOOK")
	if entry == nil {
		t.Fatal("GetHelp(LOOK) returned nil")
	}
	if entry.Command != "look" {
		t.Errorf("Command = %s, want look", entry.Command)
	}
}

func TestGetHelpNotFound(t *testing.T) {
	entry := GetHelp("nonexistent")
	if entry != nil {
		t.Error("GetHelp should return nil for nonexistent command")
	}
}

func TestGetAllByCategory(t *testing.T) {
	categories := GetAllByCategory()
	if len(categories) == 0 {
		t.Error("Should have categories")
	}
	
	if _, ok := categories[CatMovement]; !ok {
		t.Error("Should have Movement category")
	}
	if _, ok := categories[CatCombat]; !ok {
		t.Error("Should have Combat category")
	}
}

func TestGetCategories(t *testing.T) {
	categories := GetCategories()
	if len(categories) == 0 {
		t.Error("Should have categories")
	}
	
	found := false
	for _, cat := range categories {
		if cat == CatMovement {
			found = true
			break
		}
	}
	if !found {
		t.Error("Categories should include Movement")
	}
}

func TestSearch(t *testing.T) {
	// Search for "move"
	results := Search("move")
	if len(results) == 0 {
		t.Error("Should find results for 'move'")
	}
	
	// Check that movement commands are included
	foundNorth := false
	for _, r := range results {
		if r.Command == "north" {
			foundNorth = true
			break
		}
	}
	if !foundNorth {
		t.Error("Search for 'move' should find 'north'")
	}
}

func TestSearchDescription(t *testing.T) {
	// Search for "teleport"
	results := Search("teleport")
	if len(results) == 0 {
		t.Error("Should find results for 'teleport'")
	}
}

func TestSearchCategory(t *testing.T) {
	// Search for category name
	results := Search("combat")
	if len(results) == 0 {
		t.Error("Should find results for 'combat'")
	}
}

func TestSearchEmpty(t *testing.T) {
	results := Search("zzzznonexistentzzzz")
	if len(results) != 0 {
		t.Error("Should find no results for nonsense")
	}
}

func TestSuggestCommand(t *testing.T) {
	// Typo in "look"
	suggestions := SuggestCommand("lok")
	if len(suggestions) == 0 {
		t.Error("Should suggest corrections for 'lok'")
	}
	
	found := false
	for _, s := range suggestions {
		if s == "look" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should suggest 'look' for 'lok'")
	}
}

func TestSuggestCommandKill(t *testing.T) {
	suggestions := SuggestCommand("kil")
	found := false
	for _, s := range suggestions {
		if s == "kill" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should suggest 'kill' for 'kil'")
	}
}

func TestSuggestCommandNoMatch(t *testing.T) {
	suggestions := SuggestCommand("zzzzzzzzzzzz")
	if len(suggestions) != 0 {
		t.Error("Should not suggest for very different input")
	}
}

func TestSuggestCommandLimit(t *testing.T) {
	suggestions := SuggestCommand("a")
	if len(suggestions) > 3 {
		t.Errorf("Should limit to 3 suggestions, got %d", len(suggestions))
	}
}

func TestGetAutocompleteSuggestions(t *testing.T) {
	// Complete "lo"
	suggestions := GetAutocompleteSuggestions("lo")
	if len(suggestions) == 0 {
		t.Error("Should have suggestions for 'lo'")
	}
	
	found := false
	for _, s := range suggestions {
		if s == "look" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should suggest 'look' for 'lo'")
	}
}

func TestGetAutocompleteSuggestionsSorted(t *testing.T) {
	suggestions := GetAutocompleteSuggestions("s")
	if len(suggestions) < 2 {
		t.Skip("Not enough suggestions to check sorting")
	}
	
	for i := 1; i < len(suggestions); i++ {
		if suggestions[i-1] > suggestions[i] {
			t.Error("Suggestions should be sorted")
		}
	}
}

func TestGetAutocompleteSuggestionsNoMatch(t *testing.T) {
	suggestions := GetAutocompleteSuggestions("zzzzzz")
	if len(suggestions) != 0 {
		t.Error("Should have no suggestions for 'zzzzzz'")
	}
}

func TestGetContextHelp(t *testing.T) {
	// Combat state
	help := GetContextHelp("combat", "", "")
	if !strings.Contains(help, "flee") {
		t.Error("Combat context should mention flee")
	}
	
	// Dialogue state
	help = GetContextHelp("dialogue", "", "")
	if !strings.Contains(help, "number") {
		t.Error("Dialogue context should mention number selection")
	}
}

func TestGetContextHelpError(t *testing.T) {
	help := GetContextHelp("", "", "item not found")
	if !strings.Contains(help, "look") {
		t.Error("Should suggest look when item not found")
	}
}

func TestGetContextHelpDefault(t *testing.T) {
	help := GetContextHelp("", "", "")
	if help == "" {
		t.Error("Should have default help")
	}
}

func TestGetManualTopic(t *testing.T) {
	content := GetManualTopic("basics")
	if content == "" {
		t.Error("Should have 'basics' topic")
	}
	if !strings.Contains(content, "MOVING") {
		t.Error("Basics should mention movement")
	}
}

func TestGetManualTopicCombat(t *testing.T) {
	content := GetManualTopic("combat")
	if content == "" {
		t.Error("Should have 'combat' topic")
	}
}

func TestGetManualTopicNotFound(t *testing.T) {
	content := GetManualTopic("nonexistent")
	if content != "" {
		t.Error("Should return empty for nonexistent topic")
	}
}

func TestListManualTopics(t *testing.T) {
	topics := ListManualTopics()
	if len(topics) == 0 {
		t.Error("Should have manual topics")
	}
	
	found := false
	for _, topic := range topics {
		if topic == "basics" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should include 'basics' topic")
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "ab", 1},
		{"abc", "abcd", 1},
		{"kitten", "sitting", 3},
	}
	
	for _, tt := range tests {
		got := levenshteinDistance(tt.a, tt.b)
		if got != tt.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestEntryRelatedCommands(t *testing.T) {
	entry := GetHelp("look")
	if len(entry.Related) == 0 {
		t.Error("look should have related commands")
	}
}

func TestAllEntriesHaveCategory(t *testing.T) {
	for cmd, entry := range Entries {
		if entry.Category == "" {
			t.Errorf("Entry %s has no category", cmd)
		}
	}
}

func TestAllEntriesHaveDescription(t *testing.T) {
	for cmd, entry := range Entries {
		if entry.Description == "" {
			t.Errorf("Entry %s has no description", cmd)
		}
	}
}

func TestMinFunction(t *testing.T) {
	if min(1, 2, 3) != 1 {
		t.Error("min(1,2,3) should be 1")
	}
	if min(3, 1, 2) != 1 {
		t.Error("min(3,1,2) should be 1")
	}
	if min(2, 3, 1) != 1 {
		t.Error("min(2,3,1) should be 1")
	}
}
