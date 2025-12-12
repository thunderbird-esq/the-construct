package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIncrConnections(t *testing.T) {
	initial := M.ActiveConnections
	IncrConnections()
	if M.ActiveConnections != initial+1 {
		t.Error("ActiveConnections should increment")
	}
	DecrConnections()
	if M.ActiveConnections != initial {
		t.Error("ActiveConnections should decrement")
	}
}

func TestIncrPlayers(t *testing.T) {
	initial := M.PlayersOnline
	IncrPlayers()
	if M.PlayersOnline != initial+1 {
		t.Error("PlayersOnline should increment")
	}
	DecrPlayers()
	if M.PlayersOnline != initial {
		t.Error("PlayersOnline should decrement")
	}
}

func TestRecordCommand(t *testing.T) {
	initial := M.CommandsExecuted
	RecordCommand("look")
	if M.CommandsExecuted != initial+1 {
		t.Error("CommandsExecuted should increment")
	}
	if M.CommandsByType["look"] == 0 {
		t.Error("CommandsByType should track command")
	}
}

func TestRecordCombat(t *testing.T) {
	initialStart := M.CombatsStarted
	initialEnd := M.CombatsEnded
	
	RecordCombatStart()
	if M.CombatsStarted != initialStart+1 {
		t.Error("CombatsStarted should increment")
	}
	
	RecordCombatEnd()
	if M.CombatsEnded != initialEnd+1 {
		t.Error("CombatsEnded should increment")
	}
}

func TestRecordNPCKill(t *testing.T) {
	initial := M.NPCsKilled
	RecordNPCKill()
	if M.NPCsKilled != initial+1 {
		t.Error("NPCsKilled should increment")
	}
}

func TestRecordDamage(t *testing.T) {
	initialDealt := M.DamageDealt
	initialReceived := M.DamageReceived
	
	RecordDamage(10, 5)
	
	if M.DamageDealt != initialDealt+10 {
		t.Error("DamageDealt should increase by 10")
	}
	if M.DamageReceived != initialReceived+5 {
		t.Error("DamageReceived should increase by 5")
	}
}

func TestRecordPurchase(t *testing.T) {
	initialBought := M.ItemsBought
	initialMoney := M.MoneyCirculated
	
	RecordPurchase(100)
	
	if M.ItemsBought != initialBought+1 {
		t.Error("ItemsBought should increment")
	}
	if M.MoneyCirculated != initialMoney+100 {
		t.Error("MoneyCirculated should increase")
	}
}

func TestRecordSale(t *testing.T) {
	initialSold := M.ItemsSold
	initialMoney := M.MoneyCirculated
	
	RecordSale(50)
	
	if M.ItemsSold != initialSold+1 {
		t.Error("ItemsSold should increment")
	}
	if M.MoneyCirculated != initialMoney+50 {
		t.Error("MoneyCirculated should increase")
	}
}

func TestRecordFailedLogin(t *testing.T) {
	initial := M.FailedLogins
	RecordFailedLogin()
	if M.FailedLogins != initial+1 {
		t.Error("FailedLogins should increment")
	}
}

func TestRecordNewUser(t *testing.T) {
	initial := M.NewRegistrations
	RecordNewUser()
	if M.NewRegistrations != initial+1 {
		t.Error("NewRegistrations should increment")
	}
}

func TestRecordError(t *testing.T) {
	initial := M.ErrorCount
	RecordError()
	if M.ErrorCount != initial+1 {
		t.Error("ErrorCount should increment")
	}
}

func TestSetWorldCounts(t *testing.T) {
	SetWorldCounts(100, 50, 200)
	if M.RoomsCount != 100 {
		t.Errorf("RoomsCount = %d, want 100", M.RoomsCount)
	}
	if M.NPCsCount != 50 {
		t.Errorf("NPCsCount = %d, want 50", M.NPCsCount)
	}
	if M.ItemsCount != 200 {
		t.Errorf("ItemsCount = %d, want 200", M.ItemsCount)
	}
}

func TestHandler(t *testing.T) {
	handler := Handler()
	
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Status = %d, want 200", resp.StatusCode)
	}
	
	body := w.Body.String()
	
	// Check for expected metrics
	expectedMetrics := []string{
		"matrix_connections_active",
		"matrix_players_online",
		"matrix_commands_total",
		"matrix_uptime_seconds",
	}
	
	for _, metric := range expectedMetrics {
		if !strings.Contains(body, metric) {
			t.Errorf("Missing metric: %s", metric)
		}
	}
	
	// Check content type
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
}

func TestRecordRateLimited(t *testing.T) {
	initial := M.RateLimitedCmds
	RecordRateLimited()
	if M.RateLimitedCmds != initial+1 {
		t.Error("RateLimitedCmds should increment")
	}
}

func TestRecordUpdateCycle(t *testing.T) {
	initial := M.UpdateCycles
	RecordUpdateCycle()
	if M.UpdateCycles != initial+1 {
		t.Error("UpdateCycles should increment")
	}
	if M.LastUpdate.IsZero() {
		t.Error("LastUpdate should be set")
	}
}

func TestRecordPlayerDeath(t *testing.T) {
	initial := M.PlayerDeaths
	RecordPlayerDeath()
	if M.PlayerDeaths != initial+1 {
		t.Error("PlayerDeaths should increment")
	}
}
