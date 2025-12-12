package party

import (
	"fmt"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.parties == nil {
		t.Error("parties map not initialized")
	}
	if m.byPlayer == nil {
		t.Error("byPlayer map not initialized")
	}
}

func TestCreate(t *testing.T) {
	m := NewManager()

	party, err := m.Create("leader")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if party == nil {
		t.Fatal("Create returned nil party")
	}
	if party.Leader != "leader" {
		t.Errorf("Leader = %q, want leader", party.Leader)
	}
	if len(party.Members) != 1 {
		t.Errorf("Members count = %d, want 1", len(party.Members))
	}
	if party.Members[0] != "leader" {
		t.Errorf("First member = %q, want leader", party.Members[0])
	}
}

func TestCreateAlreadyInParty(t *testing.T) {
	m := NewManager()

	m.Create("leader")
	_, err := m.Create("leader")

	if err == nil {
		t.Error("Should error when already in party")
	}
}

func TestInvite(t *testing.T) {
	m := NewManager()
	m.Create("leader")

	err := m.Invite("leader", "player2")
	if err != nil {
		t.Fatalf("Invite failed: %v", err)
	}

	party := m.GetParty("leader")
	if _, ok := party.Invites["player2"]; !ok {
		t.Error("player2 should have pending invite")
	}
}

func TestInviteNotLeader(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")

	err := m.Invite("player2", "player3")
	if err == nil {
		t.Error("Non-leader should not be able to invite")
	}
}

func TestInvitePartyFull(t *testing.T) {
	m := NewManager()
	m.Create("leader")

	// Fill party to max
	for i := 2; i <= MaxPartySize; i++ {
		name := "player" + string(rune('0'+i))
		m.Invite("leader", name)
		m.Accept(name, "leader")
	}

	err := m.Invite("leader", "extraplayer")
	if err == nil {
		t.Error("Should error when party is full")
	}
}

func TestAccept(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")

	err := m.Accept("player2", "leader")
	if err != nil {
		t.Fatalf("Accept failed: %v", err)
	}

	party := m.GetParty("player2")
	if party == nil {
		t.Fatal("player2 should be in party")
	}
	if len(party.Members) != 2 {
		t.Errorf("Party should have 2 members, got %d", len(party.Members))
	}
}

func TestAcceptNoInvite(t *testing.T) {
	m := NewManager()
	m.Create("leader")

	err := m.Accept("player2", "leader")
	if err == nil {
		t.Error("Should error without invite")
	}
}

func TestDecline(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")

	err := m.Decline("player2", "leader")
	if err != nil {
		t.Fatalf("Decline failed: %v", err)
	}

	party := m.GetParty("leader")
	if _, ok := party.Invites["player2"]; ok {
		t.Error("Invite should be removed after decline")
	}
}

func TestLeave(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")

	err := m.Leave("player2")
	if err != nil {
		t.Fatalf("Leave failed: %v", err)
	}

	if m.IsInParty("player2") {
		t.Error("player2 should not be in party after leaving")
	}

	party := m.GetParty("leader")
	if len(party.Members) != 1 {
		t.Errorf("Party should have 1 member, got %d", len(party.Members))
	}
}

func TestLeaveLeaderPromotes(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")

	m.Leave("leader")

	party := m.GetParty("player2")
	if party == nil {
		t.Fatal("Party should still exist")
	}
	if party.Leader != "player2" {
		t.Errorf("player2 should be promoted to leader, got %s", party.Leader)
	}
}

func TestLeaveLastMemberDisbands(t *testing.T) {
	m := NewManager()
	m.Create("leader")

	m.Leave("leader")

	if m.IsInParty("leader") {
		t.Error("leader should not be in party")
	}
	// Party should be disbanded
	if len(m.parties) != 0 {
		t.Errorf("No parties should exist, got %d", len(m.parties))
	}
}

func TestKick(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")

	err := m.Kick("leader", "player2")
	if err != nil {
		t.Fatalf("Kick failed: %v", err)
	}

	if m.IsInParty("player2") {
		t.Error("player2 should be kicked")
	}
}

func TestKickNotLeader(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")
	m.Invite("leader", "player3")
	m.Accept("player3", "leader")

	err := m.Kick("player2", "player3")
	if err == nil {
		t.Error("Non-leader should not be able to kick")
	}
}

func TestKickSelf(t *testing.T) {
	m := NewManager()
	m.Create("leader")

	err := m.Kick("leader", "leader")
	if err == nil {
		t.Error("Should not be able to kick yourself")
	}
}

func TestPromote(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")

	err := m.Promote("leader", "player2")
	if err != nil {
		t.Fatalf("Promote failed: %v", err)
	}

	party := m.GetParty("leader")
	if party.Leader != "player2" {
		t.Errorf("Leader should be player2, got %s", party.Leader)
	}
}

func TestDisband(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")

	err := m.Disband("leader")
	if err != nil {
		t.Fatalf("Disband failed: %v", err)
	}

	if m.IsInParty("leader") {
		t.Error("leader should not be in party")
	}
	if m.IsInParty("player2") {
		t.Error("player2 should not be in party")
	}
}

func TestDisbandNotLeader(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")

	err := m.Disband("player2")
	if err == nil {
		t.Error("Non-leader should not be able to disband")
	}
}

func TestGetMembers(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")

	members := m.GetMembers("leader")
	if len(members) != 2 {
		t.Errorf("GetMembers returned %d, want 2", len(members))
	}
}

func TestGetMembersNotInParty(t *testing.T) {
	m := NewManager()

	members := m.GetMembers("nobody")
	if members != nil {
		t.Error("GetMembers should return nil for non-party member")
	}
}

func TestIsLeader(t *testing.T) {
	m := NewManager()
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")

	if !m.IsLeader("leader") {
		t.Error("leader should be leader")
	}
	if m.IsLeader("player2") {
		t.Error("player2 should not be leader")
	}
}

func TestAreInSameParty(t *testing.T) {
	m := NewManager()

	// Use unique names with microsecond precision + random component
	ts := time.Now().UnixMicro()
	leader := fmt.Sprintf("asp_lead_%d", ts)
	player2 := fmt.Sprintf("asp_p2_%d", ts+1)
	other := fmt.Sprintf("asp_oth_%d", ts+2)

	party, err := m.Create(leader)
	if err != nil {
		t.Fatalf("Failed to create party: %v", err)
	}
	t.Logf("Created party %s with leader %s", party.ID, leader)

	if err := m.Invite(leader, player2); err != nil {
		t.Fatalf("Invite failed: %v", err)
	}
	if err := m.Accept(player2, leader); err != nil {
		t.Fatalf("Accept failed: %v", err)
	}

	if !m.AreInSameParty(leader, player2) {
		t.Errorf("leader (%s) and player2 (%s) should be in same party", leader, player2)
	}

	otherParty, err := m.Create(other)
	if err != nil {
		t.Fatalf("Failed to create other party: %v", err)
	}
	t.Logf("Created other party %s with leader %s", otherParty.ID, other)

	if m.AreInSameParty(leader, other) {
		// Debug: print party IDs
		leaderParty := m.GetParty(leader)
		otherPartyCheck := m.GetParty(other)
		t.Errorf("leader (%s, party:%v) and other (%s, party:%v) should NOT be in same party",
			leader, leaderParty, other, otherPartyCheck)
	}
}

func TestGetPendingInvites(t *testing.T) {
	m := NewManager()

	// Use unique names for this test to avoid conflicts with other test runs
	leader1 := fmt.Sprintf("invites_leader1_%d", time.Now().UnixNano())
	leader2 := fmt.Sprintf("invites_leader2_%d", time.Now().UnixNano())
	player := fmt.Sprintf("invites_player_%d", time.Now().UnixNano())

	party1, err := m.Create(leader1)
	if err != nil {
		t.Fatalf("Failed to create party1: %v", err)
	}
	t.Logf("Created party1 with leader %s, ID: %s", leader1, party1.ID)

	party2, err := m.Create(leader2)
	if err != nil {
		t.Fatalf("Failed to create party2: %v", err)
	}
	t.Logf("Created party2 with leader %s, ID: %s", leader2, party2.ID)

	err1 := m.Invite(leader1, player)
	if err1 != nil {
		t.Errorf("First invite failed: %v (leader1=%s, party1.Leader=%s)", err1, leader1, party1.Leader)
	}

	err2 := m.Invite(leader2, player)
	if err2 != nil {
		t.Errorf("Second invite failed: %v", err2)
	}

	invites := m.GetPendingInvites(player)
	if len(invites) != 2 {
		t.Errorf("Should have 2 invites, got %d: %v", len(invites), invites)
	}
}

func TestCalculateXPShare(t *testing.T) {
	m := NewManager()

	// Solo player gets full XP
	shares := m.CalculateXPShare("solo", 100)
	if shares["solo"] != 100 {
		t.Errorf("Solo XP = %d, want 100", shares["solo"])
	}

	// Party of 2: 110% total, 55 each
	m.Create("leader")
	m.Invite("leader", "player2")
	m.Accept("player2", "leader")

	shares = m.CalculateXPShare("leader", 100)
	expected := 55 // (100 * 1.1) / 2
	if shares["leader"] != expected {
		t.Errorf("Party XP = %d, want %d", shares["leader"], expected)
	}
	if shares["player2"] != expected {
		t.Errorf("Party XP = %d, want %d", shares["player2"], expected)
	}
}

func TestGlobalParty(t *testing.T) {
	if GlobalParty == nil {
		t.Error("GlobalParty should be initialized")
	}

	// Test that GlobalParty works (use unique names to avoid pollution)
	testPlayer := "globaltest_unique_player"
	GlobalParty.Create(testPlayer)
	if !GlobalParty.IsInParty(testPlayer) {
		t.Error("GlobalParty.Create should work")
	}
	GlobalParty.Leave(testPlayer)
}
