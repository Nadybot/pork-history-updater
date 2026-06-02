package domain

import (
	"testing"
	"time"
)

func TestGuildMembership_Equals_equal(t *testing.T) {
	a := &GuildMembership{ID: 1, Name: "Alpha", Rank: 5, RankName: "Officer"}
	b := &GuildMembership{ID: 1, Name: "Alpha", Rank: 5, RankName: "Officer"}
	if !a.Equals(b) {
		t.Error("expected equal memberships to be equal")
	}
}

func TestGuildMembership_Equals_differentFields(t *testing.T) {
	tests := []struct {
		name string
		a    *GuildMembership
		b    *GuildMembership
	}{
		{
			name: "different ID",
			a:    &GuildMembership{ID: 1, Name: "A", Rank: 1, RankName: "R1"},
			b:    &GuildMembership{ID: 2, Name: "A", Rank: 1, RankName: "R1"},
		},
		{
			name: "different Name",
			a:    &GuildMembership{ID: 1, Name: "A", Rank: 1, RankName: "R1"},
			b:    &GuildMembership{ID: 1, Name: "B", Rank: 1, RankName: "R1"},
		},
		{
			name: "different Rank",
			a:    &GuildMembership{ID: 1, Name: "A", Rank: 1, RankName: "R1"},
			b:    &GuildMembership{ID: 1, Name: "A", Rank: 2, RankName: "R1"},
		},
		{
			name: "different RankName",
			a:    &GuildMembership{ID: 1, Name: "A", Rank: 1, RankName: "R1"},
			b:    &GuildMembership{ID: 1, Name: "A", Rank: 1, RankName: "R2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.a.Equals(tt.b) {
				t.Error("expected different memberships to be unequal")
			}
		})
	}
}

func TestGuildMembership_Equals_nilCases(t *testing.T) {
	g := &GuildMembership{ID: 1, Name: "Test"}

	if !(*GuildMembership)(nil).Equals((*GuildMembership)(nil)) {
		t.Error("expected two nil pointers to be equal")
	}
	if g.Equals(nil) {
		t.Error("expected non-nil vs nil to be unequal")
	}
	if (*GuildMembership)(nil).Equals(g) {
		t.Error("expected nil vs non-nil to be unequal")
	}
}

func TestGuildMembership_String(t *testing.T) {
	if s := (*GuildMembership)(nil).String(); s != "<none>" {
		t.Errorf("nil.String() = %q, want %q", s, "<none>")
	}

	g := &GuildMembership{ID: 42, Name: "Dragons", Rank: 7, RankName: "General"}
	want := `{ID=42, Name="Dragons", Rank=7, RankName="General"}`
	if got := g.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// ============================================================================
// Player.Equals
// ============================================================================

func TestPlayer_Equals_equal(t *testing.T) {
	now := time.Now()
	charID := uint32(1)
	a := &Player{
		Nickname:         "Alice",
		CharID:           &charID,
		FirstName:        "A",
		LastName:         "B",
		Level:            10,
		Faction:          "Omni",
		Profession:       "MA",
		ProfessionTitle:  "Martial Artist",
		Gender:           "Female",
		Breed:            "Solitus",
		DefenderRank:     5,
		DefenderRankName: "Soldier",
		MyGuild:          &GuildMembership{ID: 1, Name: "Guild", Rank: 1, RankName: "Leader"},
		Server:           5,
		LastChecked:      now,
		LastChanged:      now,
		Deleted:          false,
	}
	b := *a // shallow copy of the value

	if !a.Equals(&b) {
		t.Error("expected identical players to be equal")
	}
}

func TestPlayer_Equals_nil(t *testing.T) {
	p := &Player{Nickname: "Test"}

	if p.Equals(nil) {
		t.Error("expected player vs nil to be unequal")
	}
	if !(*Player)(nil).Equals(nil) {
		t.Error("expected two nil players to be equal")
	}
	if (*Player)(nil).Equals(p) {
		t.Error("expected nil vs player to be unequal")
	}
}

func TestPlayer_Equals_differentFields(t *testing.T) {
	charID := uint32(1)
	newTestPlayer := func() *Player {
		return &Player{
			Nickname:         "Alice",
			CharID:           &charID,
			FirstName:        "A",
			LastName:         "B",
			Level:            10,
			Faction:          "Omni",
			Profession:       "MA",
			ProfessionTitle:  "Martial Artist",
			Gender:           "Female",
			Breed:            "Solitus",
			DefenderRank:     5,
			DefenderRankName: "Soldier",
			MyGuild:          &GuildMembership{ID: 1, Name: "Guild", Rank: 1, RankName: "Leader"},
			Server:           5,
			Deleted:          false,
		}
	}

	otherCharID := uint32(2)
	tests := []struct {
		name   string
		modify func(*Player)
	}{
		{
			name:   "different Nickname",
			modify: func(p *Player) { p.Nickname = "Bob" },
		},
		{
			name:   "different CharID",
			modify: func(p *Player) { p.CharID = &otherCharID },
		},
		{
			name:   "nil CharID",
			modify: func(p *Player) { p.CharID = nil },
		},
		{
			name:   "different FirstName",
			modify: func(p *Player) { p.FirstName = "Z" },
		},
		{
			name:   "different LastName",
			modify: func(p *Player) { p.LastName = "Z" },
		},
		{
			name:   "different Level",
			modify: func(p *Player) { p.Level = 99 },
		},
		{
			name:   "different Faction",
			modify: func(p *Player) { p.Faction = "Clan" },
		},
		{
			name:   "different Profession",
			modify: func(p *Player) { p.Profession = "ENF" },
		},
		{
			name:   "different ProfessionTitle",
			modify: func(p *Player) { p.ProfessionTitle = "Enforcer" },
		},
		{
			name:   "different Gender",
			modify: func(p *Player) { p.Gender = "Male" },
		},
		{
			name:   "different Breed",
			modify: func(p *Player) { p.Breed = "Atrox" },
		},
		{
			name:   "different DefenderRank",
			modify: func(p *Player) { p.DefenderRank = 99 },
		},
		{
			name:   "different DefenderRankName",
			modify: func(p *Player) { p.DefenderRankName = "General" },
		},
		{
			name:   "different Server",
			modify: func(p *Player) { p.Server = 99 },
		},
		{
			name:   "different Deleted",
			modify: func(p *Player) { p.Deleted = true },
		},
		{
			name:   "different MyGuild",
			modify: func(p *Player) { p.MyGuild = &GuildMembership{ID: 99, Name: "Other", Rank: 2, RankName: "Officer"} },
		},
		{
			name:   "nil MyGuild",
			modify: func(p *Player) { p.MyGuild = nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestPlayer()
			b := newTestPlayer()
			tt.modify(b)

			if a.Equals(b) {
				t.Errorf("expected players to be unequal when %s differs", tt.name)
			}
		})
	}
}

func TestPlayer_ChangesTo_nil(t *testing.T) {
	p := &Player{Nickname: "Test"}
	if p.ChangesTo(nil) != nil {
		t.Error("expected ChangesTo(nil) to return nil")
	}
}

func TestPlayer_ChangesTo_noChanges(t *testing.T) {
	p := &Player{Nickname: "Alice", Level: 10, Server: 5}
	changes := p.ChangesTo(&Player{Nickname: "Alice", Level: 10, Server: 5})
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestPlayer_ChangesTo_singleChange(t *testing.T) {
	p := &Player{Nickname: "Alice", Level: 10, Server: 5}
	changes := p.ChangesTo(&Player{Nickname: "Alice", Level: 99, Server: 5})

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Name != "Level" {
		t.Errorf("change.Name = %q, want %q", changes[0].Name, "Level")
	}
	if changes[0].Old != 10 {
		t.Errorf("change.Old = %v, want %v", changes[0].Old, 10)
	}
	if changes[0].New != 99 {
		t.Errorf("change.New = %v, want %v", changes[0].New, 99)
	}
}

func TestPlayer_ChangesTo_MyGuildChanges(t *testing.T) {
	p := &Player{
		Nickname: "Alice",
		MyGuild:  &GuildMembership{ID: 1, Name: "OldGuild", Rank: 1, RankName: "Leader"},
	}
	other := &Player{
		Nickname: "Alice",
		MyGuild:  &GuildMembership{ID: 1, Name: "NewGuild", Rank: 1, RankName: "Leader"},
	}

	changes := p.ChangesTo(other)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Name != "MyGuild.Name" {
		t.Errorf("change.Name = %q, want %q", changes[0].Name, "MyGuild.Name")
	}
}

func TestPlayer_ChangesTo_MyGuildToNil(t *testing.T) {
	old := &Player{
		Nickname: "Alice",
		MyGuild:  &GuildMembership{ID: 1, Name: "Guild"},
	}
	new := &Player{
		Nickname: "Alice",
		MyGuild:  nil,
	}

	changes := old.ChangesTo(new)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Name != "MyGuild" {
		t.Errorf("change.Name = %q, want %q", changes[0].Name, "MyGuild")
	}
}

func TestPlayer_ChangesTo_ignoresLastCheckedAndLastChanged(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	p := &Player{
		Nickname:    "Alice",
		LastChecked: now,
		LastChanged: now,
	}
	other := &Player{
		Nickname:    "Alice",
		LastChecked: later,
		LastChanged: later,
	}

	changes := p.ChangesTo(other)
	for _, c := range changes {
		if c.Name == "LastChecked" || c.Name == "LastChanged" {
			t.Errorf("expected %s to be ignored, but it was reported", c.Name)
		}
	}
}
