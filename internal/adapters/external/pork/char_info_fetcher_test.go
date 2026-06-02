package pork

import (
	"errors"
	"testing"
)

// mockProvider is a test double for the charInfoProvider interface.
type mockProvider struct {
	data *PlayerData
	err  error
}

func (m *mockProvider) FetchCharInfo(name string, dimension int) (*PlayerData, error) {
	return m.data, m.err
}

// TestAdapterFetchByNameAsPlayer_success verifies correct mapping from
// PlayerData to domain.Player when all fields are present.
// It reuses the fixture that is already tested by parsePlayerData tests.
func TestAdapterFetchByNameAsPlayer_success(t *testing.T) {
	pd, err := parsePlayerData(validCharacterInOrg)
	if err != nil {
		t.Fatalf("failed to parse fixture: %v", err)
	}
	if pd == nil {
		t.Fatal("expected fixture to produce PlayerData, got nil")
	}

	mp := &mockProvider{data: pd}
	adapt := NewAdapter(mp)
	player, err := adapt.FetchByNameAsPlayer("Pigtail", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player == nil {
		t.Fatal("expected player, got nil")
	}

	// Character fields
	if player.Nickname != "Pigtail" {
		t.Errorf("Nickname = %q, want %q", player.Nickname, "Pigtail")
	}
	if player.FirstName != "" {
		t.Errorf("FirstName = %q, want %q", player.FirstName, "")
	}
	if player.LastName != "" {
		t.Errorf("LastName = %q, want %q", player.LastName, "")
	}
	if player.Gender != "Neuter" {
		t.Errorf("Gender = %q, want %q", player.Gender, "Neuter")
	}
	if player.Breed != "Atrox" {
		t.Errorf("Breed = %q, want %q", player.Breed, "Atrox")
	}
	if player.Profession != "Enforcer" {
		t.Errorf("Profession = %q, want %q", player.Profession, "Enforcer")
	}
	if player.ProfessionTitle != "Don" {
		t.Errorf("ProfessionTitle = %q, want %q", player.ProfessionTitle, "Don")
	}
	if player.Level != 200 {
		t.Errorf("Level = %d, want %d", player.Level, 200)
	}
	if player.DefenderRank != 0 {
		t.Errorf("DefenderRank = %d, want %d", player.DefenderRank, 0)
	}
	if player.Faction != "Neutral" {
		t.Errorf("Faction = %q, want %q", player.Faction, "Neutral")
	}
	if player.DefenderRankName != "None" {
		t.Errorf("DefenderRankName = %q, want %q", player.DefenderRankName, "None")
	}
	if player.CharID == nil || *player.CharID != 1178213 {
		t.Errorf("CharID = %v, want pointer to 1178213", player.CharID)
	}
	if player.Server != 5 {
		t.Errorf("Server = %d, want %d", player.Server, 5)
	}

	// Guild fields
	if player.MyGuild == nil {
		t.Fatal("expected MyGuild, got nil")
	}
	if player.MyGuild.ID != 725003 {
		t.Errorf("MyGuild.ID = %d, want %d", player.MyGuild.ID, 725003)
	}
	if player.MyGuild.Name != "Troet" {
		t.Errorf("MyGuild.Name = %q, want %q", player.MyGuild.Name, "Troet")
	}
	if player.MyGuild.Rank != 1 {
		t.Errorf("MyGuild.Rank = %d, want %d", player.MyGuild.Rank, 1)
	}
	if player.MyGuild.RankName != "Advisor" {
		t.Errorf("MyGuild.RankName = %q, want %q", player.MyGuild.RankName, "Advisor")
	}

	// Timestamps
	if player.LastChanged.IsZero() {
		t.Error("expected LastChanged to be set")
	}
	if player.LastChecked.IsZero() {
		t.Error("expected LastChecked to be set")
	}
}

// TestAdapterFetchByNameAsPlayer_noOrg verifies that a PlayerData without an
// Org results in a Player with MyGuild == nil.
func TestAdapterFetchByNameAsPlayer_noOrg(t *testing.T) {
	pd, err := parsePlayerData(validCharacterWithoutOrg)
	if err != nil {
		t.Fatalf("failed to parse fixture: %v", err)
	}
	if pd == nil {
		t.Fatal("expected fixture to produce PlayerData, got nil")
	}

	mp := &mockProvider{data: pd}
	adapt := NewAdapter(mp)
	player, err := adapt.FetchByNameAsPlayer("SoloChar", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player == nil {
		t.Fatal("expected player, got nil")
	}
	if player.MyGuild != nil {
		t.Errorf("expected MyGuild=nil, got %+v", player.MyGuild)
	}
}

// TestAdapterFetchByNameAsPlayer_nilData verifies that a nil PlayerData
// (character not found) returns nil without error.
func TestAdapterFetchByNameAsPlayer_nilData(t *testing.T) {
	mp := &mockProvider{data: nil}

	adapt := NewAdapter(mp)
	player, err := adapt.FetchByNameAsPlayer("Unknown", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player != nil {
		t.Errorf("expected nil, got %+v", player)
	}
}

// TestAdapterFetchByNameAsPlayer_providerError verifies that errors from the
// underlying provider are propagated unchanged.
func TestAdapterFetchByNameAsPlayer_providerError(t *testing.T) {
	mp := &mockProvider{err: errors.New("network timeout")}

	adapt := NewAdapter(mp)
	_, err := adapt.FetchByNameAsPlayer("ErrorChar", 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "network timeout" {
		t.Errorf("error = %q, want %q", err.Error(), "network timeout")
	}
}
