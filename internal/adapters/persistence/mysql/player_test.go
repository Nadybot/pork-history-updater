package mysql

import (
	"database/sql"
	"testing"
	"time"

	"pork-history-updater/internal/domain"
)

// ============================================================================
// dbPlayer.ToDomain
// ============================================================================

func TestDbPlayer_ToDomain_full(t *testing.T) {
	refTime := time.Unix(1705321800, 0)

	dto := dbPlayer{
		Nickname:         "Pigtail",
		CharID:           sql.NullInt64{Int64: 1178213, Valid: true},
		FirstName:        "Pigtail",
		LastName:         "",
		GuildRank:        1,
		GuildRankName:    "Advisor",
		Level:            200,
		Faction:          "Neutral",
		Profession:       "Enforcer",
		ProfessionTitle:  "Don",
		Gender:           "Neuter",
		Breed:            "Atrox",
		DefenderRank:     30,
		DefenderRankName: "General",
		GuildID:          725003,
		GuildName:        "Troet",
		Server:           5,
		LastChecked:      int(refTime.Unix()),
		LastChanged:      int(refTime.Unix()),
		Deleted:          false,
	}

	player := dto.ToDomain()

	if player.Nickname != "Pigtail" {
		t.Errorf("Nickname = %q, want %q", player.Nickname, "Pigtail")
	}
	if player.CharID == nil || *player.CharID != 1178213 {
		t.Errorf("CharID = %v, want pointer to 1178213", player.CharID)
	}
	if player.FirstName != "Pigtail" {
		t.Errorf("FirstName = %q, want %q", player.FirstName, "Pigtail")
	}
	if player.LastName != "" {
		t.Errorf("LastName = %q, want %q", player.LastName, "")
	}
	if player.Level != 200 {
		t.Errorf("Level = %d, want %d", player.Level, 200)
	}
	if player.Faction != "Neutral" {
		t.Errorf("Faction = %q, want %q", player.Faction, "Neutral")
	}
	if player.Profession != "Enforcer" {
		t.Errorf("Profession = %q, want %q", player.Profession, "Enforcer")
	}
	if player.ProfessionTitle != "Don" {
		t.Errorf("ProfessionTitle = %q, want %q", player.ProfessionTitle, "Don")
	}
	if player.Gender != "Neuter" {
		t.Errorf("Gender = %q, want %q", player.Gender, "Neuter")
	}
	if player.Breed != "Atrox" {
		t.Errorf("Breed = %q, want %q", player.Breed, "Atrox")
	}
	if player.DefenderRank != 30 {
		t.Errorf("DefenderRank = %d, want %d", player.DefenderRank, 30)
	}
	if player.DefenderRankName != "General" {
		t.Errorf("DefenderRankName = %q, want %q", player.DefenderRankName, "General")
	}
	if player.Server != 5 {
		t.Errorf("Server = %d, want %d", player.Server, 5)
	}
	if player.Deleted != false {
		t.Errorf("Deleted = %v, want %v", player.Deleted, false)
	}
	if !player.LastChecked.Equal(refTime) {
		t.Errorf("LastChecked = %v, want %v", player.LastChecked, refTime)
	}
	if !player.LastChanged.Equal(refTime) {
		t.Errorf("LastChanged = %v, want %v", player.LastChanged, refTime)
	}

	// Guild
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
}

func TestDbPlayer_ToDomain_noGuild(t *testing.T) {
	dto := dbPlayer{
		Nickname: "SoloChar",
		CharID:   sql.NullInt64{Int64: 999, Valid: true},
		Level:    1,
		Server:   5,
		// Guild fields intentionally left at zero values
	}

	player := dto.ToDomain()

	if player.MyGuild != nil {
		t.Errorf("expected MyGuild=nil when GuildID==0, got %+v", player.MyGuild)
	}
}

func TestDbPlayer_ToDomain_deletedChar(t *testing.T) {
	dto := dbPlayer{
		Nickname:    "DeletedChar",
		CharID:      sql.NullInt64{Valid: false},
		Deleted:     true,
		LastChecked: 1705321800,
		LastChanged: 1705321800,
	}

	player := dto.ToDomain()

	if player.CharID != nil {
		t.Errorf("expected CharID = nil, got %v", player.CharID)
	}
	if !player.Deleted {
		t.Error("expected Deleted = true")
	}
}

// ============================================================================
// dbPlayerFromDomain
// ============================================================================

func TestDbPlayerFromDomain_full(t *testing.T) {
	refTime := time.Unix(1705321800, 0)
	charID := uint32(1178213)

	player := &domain.Player{
		Nickname:         "Pigtail",
		CharID:           &charID,
		FirstName:        "Pigtail",
		LastName:         "",
		Level:            200,
		Faction:          "Neutral",
		Profession:       "Enforcer",
		ProfessionTitle:  "Don",
		Gender:           "Neuter",
		Breed:            "Atrox",
		DefenderRank:     30,
		DefenderRankName: "General",
		MyGuild: &domain.GuildMembership{
			ID:       725003,
			Name:     "Troet",
			Rank:     1,
			RankName: "Advisor",
		},
		Server:      5,
		LastChecked: refTime,
		LastChanged: refTime,
		Deleted:     false,
	}

	dto := dbPlayerFromDomain(player)

	if dto.Nickname != "Pigtail" {
		t.Errorf("Nickname = %q, want %q", dto.Nickname, "Pigtail")
	}
	if !dto.CharID.Valid || dto.CharID.Int64 != 1178213 {
		t.Errorf("CharID = %v, want {1178213 true}", dto.CharID)
	}
	if dto.FirstName != "Pigtail" {
		t.Errorf("FirstName = %q, want %q", dto.FirstName, "Pigtail")
	}
	if dto.LastName != "" {
		t.Errorf("LastName = %q, want %q", dto.LastName, "")
	}
	if dto.Level != 200 {
		t.Errorf("Level = %d, want %d", dto.Level, 200)
	}
	if dto.Faction != "Neutral" {
		t.Errorf("Faction = %q, want %q", dto.Faction, "Neutral")
	}
	if dto.Profession != "Enforcer" {
		t.Errorf("Profession = %q, want %q", dto.Profession, "Enforcer")
	}
	if dto.ProfessionTitle != "Don" {
		t.Errorf("ProfessionTitle = %q, want %q", dto.ProfessionTitle, "Don")
	}
	if dto.Gender != "Neuter" {
		t.Errorf("Gender = %q, want %q", dto.Gender, "Neuter")
	}
	if dto.Breed != "Atrox" {
		t.Errorf("Breed = %q, want %q", dto.Breed, "Atrox")
	}
	if dto.DefenderRank != 30 {
		t.Errorf("DefenderRank = %d, want %d", dto.DefenderRank, 30)
	}
	if dto.DefenderRankName != "General" {
		t.Errorf("DefenderRankName = %q, want %q", dto.DefenderRankName, "General")
	}
	if dto.Server != 5 {
		t.Errorf("Server = %d, want %d", dto.Server, 5)
	}
	if dto.Deleted != false {
		t.Errorf("Deleted = %v, want %v", dto.Deleted, false)
	}
	if dto.LastChecked != int(refTime.Unix()) {
		t.Errorf("LastChecked = %d, want %d", dto.LastChecked, int(refTime.Unix()))
	}
	if dto.LastChanged != int(refTime.Unix()) {
		t.Errorf("LastChanged = %d, want %d", dto.LastChanged, int(refTime.Unix()))
	}

	// Guild
	if dto.GuildID != 725003 {
		t.Errorf("GuildID = %d, want %d", dto.GuildID, 725003)
	}
	if dto.GuildName != "Troet" {
		t.Errorf("GuildName = %q, want %q", dto.GuildName, "Troet")
	}
	if dto.GuildRank != 1 {
		t.Errorf("GuildRank = %d, want %d", dto.GuildRank, 1)
	}
	if dto.GuildRankName != "Advisor" {
		t.Errorf("GuildRankName = %q, want %q", dto.GuildRankName, "Advisor")
	}
}

func TestDbPlayerFromDomain_noGuild(t *testing.T) {
	charID := uint32(999)
	player := &domain.Player{
		Nickname: "SoloChar",
		CharID:   &charID,
		Level:    1,
		Server:   5,
		MyGuild:  nil,
	}

	dto := dbPlayerFromDomain(player)

	if dto.GuildID != 0 {
		t.Errorf("GuildID = %d, want 0", dto.GuildID)
	}
	if dto.GuildName != "" {
		t.Errorf("GuildName = %q, want empty", dto.GuildName)
	}
	if dto.GuildRank != 0 {
		t.Errorf("GuildRank = %d, want 0", dto.GuildRank)
	}
	if dto.GuildRankName != "" {
		t.Errorf("GuildRankName = %q, want empty", dto.GuildRankName)
	}
}

func TestDbPlayerFromDomain_deletedChar(t *testing.T) {
	player := &domain.Player{
		Nickname: "DeletedChar",
		CharID:   nil,
		Deleted:  true,
	}

	dto := dbPlayerFromDomain(player)

	if dto.CharID.Valid {
		t.Errorf("expected CharID.Valid = false, got true")
	}
	if !dto.Deleted {
		t.Error("expected Deleted = true")
	}
}
