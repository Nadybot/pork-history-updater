package domain

import (
	"fmt"
	"time"
)

// GuildMembership represents a player's affiliation with a guild.
// It is a Value Object: it has no identity of its own and is immutable in usage.
type GuildMembership struct {
	ID       int    // Unique identifier of the guild.
	Name     string // Display name of the guild.
	Rank     int    // Numeric rank of the player within the guild.
	RankName string // Human-readable rank title (e.g. "Officer").
}

// String returns a human-readable representation of the membership.
// A nil receiver returns "<none>".
func (g *GuildMembership) String() string {
	if g == nil {
		return "<none>"
	}
	return fmt.Sprintf("{ID=%d, Name=%q, Rank=%d, RankName=%q}", g.ID, g.Name, g.Rank, g.RankName)
}

// ChangesTo returns the list of field-level differences between two memberships.
func (g *GuildMembership) ChangesTo(other *GuildMembership) []FieldChange {
	if other == nil {
		return nil
	}
	var changes []FieldChange
	if g.ID != other.ID {
		changes = append(changes, FieldChange{"ID", g.ID, other.ID})
	}
	if g.Name != other.Name {
		changes = append(changes, FieldChange{"Name", g.Name, other.Name})
	}
	if g.Rank != other.Rank {
		changes = append(changes, FieldChange{"Rank", g.Rank, other.Rank})
	}
	if g.RankName != other.RankName {
		changes = append(changes, FieldChange{"RankName", g.RankName, other.RankName})
	}
	return changes
}

// Player is the central domain entity representing a game character.
// It is identified by CharID and Server together.
type Player struct {
	Nickname         string           // In-game character name used for display. Character names are unique among a dimension/game server
	CharID           *uint32          // Unique character identifier; nil for deleted characters.
	FirstName        string           // RP first name.
	LastName         string           // RP last name.
	Level            int              // Character level.
	Faction          string           // Side the character belongs to (e.g. "Omni", "Clan", "Neutral").
	Profession       string           // Long profession name (e.g. "Martial Artist", "Enforcer").
	ProfessionTitle  string           // Profession title (e.g. "Don").
	Gender           string           // Character gender.
	Breed            string           // Character breed (e.g. "Solitus", "Opifex").
	DefenderRank     int              // Alien level / Defender rank.
	DefenderRankName string           // Human-readable defender rank title.
	MyGuild          *GuildMembership // Current guild affiliation; nil if not in a guild.
	Server           int              // Game server / dimension number.
	LastChecked      time.Time        // Timestamp of the most recent update check.
	LastChanged      time.Time        // Timestamp of the last actual change detected.
	Deleted          bool             // True if the character was removed from the remote source.
}

// FieldChange describes a single difference between two values.
type FieldChange struct {
	Name string // Field or property name that differs.
	Old  any    // Previous value.
	New  any    // Current value.
}

// ChangesTo returns a slice of differences between this player and another.
// LastChecked and LastChanged are intentionally omitted from the comparison.
func (p *Player) ChangesTo(other *Player) []FieldChange {
	if other == nil {
		return nil
	}
	var changes []FieldChange
	if p.Nickname != other.Nickname {
		changes = append(changes, FieldChange{"Nickname", p.Nickname, other.Nickname})
	}
	charIDsEqual := (p.CharID == nil && other.CharID == nil) ||
		(p.CharID != nil && other.CharID != nil && *p.CharID == *other.CharID)
	if !charIDsEqual {
		changes = append(changes, FieldChange{"CharID", p.CharID, other.CharID})
	}
	if p.FirstName != other.FirstName {
		changes = append(changes, FieldChange{"FirstName", p.FirstName, other.FirstName})
	}
	if p.LastName != other.LastName {
		changes = append(changes, FieldChange{"LastName", p.LastName, other.LastName})
	}
	if !p.MyGuild.Equals(other.MyGuild) {
		if p.MyGuild != nil && other.MyGuild != nil {
			for _, gc := range p.MyGuild.ChangesTo(other.MyGuild) {
				changes = append(changes, FieldChange{"MyGuild." + gc.Name, gc.Old, gc.New})
			}
		} else {
			changes = append(changes, FieldChange{"MyGuild", p.MyGuild, other.MyGuild})
		}
	}
	if p.Level != other.Level {
		changes = append(changes, FieldChange{"Level", p.Level, other.Level})
	}
	if p.Faction != other.Faction {
		changes = append(changes, FieldChange{"Faction", p.Faction, other.Faction})
	}
	if p.Profession != other.Profession {
		changes = append(changes, FieldChange{"Profession", p.Profession, other.Profession})
	}
	if p.ProfessionTitle != other.ProfessionTitle {
		changes = append(changes, FieldChange{"ProfessionTitle", p.ProfessionTitle, other.ProfessionTitle})
	}
	if p.Gender != other.Gender {
		changes = append(changes, FieldChange{"Gender", p.Gender, other.Gender})
	}
	if p.Breed != other.Breed {
		changes = append(changes, FieldChange{"Breed", p.Breed, other.Breed})
	}
	if p.DefenderRank != other.DefenderRank {
		changes = append(changes, FieldChange{"DefenderRank", p.DefenderRank, other.DefenderRank})
	}
	if p.DefenderRankName != other.DefenderRankName {
		changes = append(changes, FieldChange{"DefenderRankName", p.DefenderRankName, other.DefenderRankName})
	}
	if p.Server != other.Server {
		changes = append(changes, FieldChange{"Server", p.Server, other.Server})
	}
	if p.Deleted != other.Deleted {
		changes = append(changes, FieldChange{"Deleted", p.Deleted, other.Deleted})
	}
	return changes
}

// Equals reports whether this player is equal to another.
// LastChecked and LastChanged are intentionally ignored.
func (this *Player) Equals(that *Player) bool {
	if this == nil || that == nil {
		return this == that
	}
	charIDEqual := (this.CharID == nil && that.CharID == nil) ||
		(this.CharID != nil && that.CharID != nil && *this.CharID == *that.CharID)
	return this.MyGuild.Equals(that.MyGuild) &&
		this.Nickname == that.Nickname &&
		charIDEqual &&
		this.FirstName == that.FirstName &&
		this.LastName == that.LastName &&
		this.Level == that.Level &&
		this.Faction == that.Faction &&
		this.Profession == that.Profession &&
		this.ProfessionTitle == that.ProfessionTitle &&
		this.Gender == that.Gender &&
		this.Breed == that.Breed &&
		this.DefenderRank == that.DefenderRank &&
		this.DefenderRankName == that.DefenderRankName &&
		this.Server == that.Server &&
		this.Deleted == that.Deleted
}

// Equals reports whether this membership is equal to another.
// Two nil pointers are considered equal.
func (g *GuildMembership) Equals(other *GuildMembership) bool {
	if g == nil || other == nil {
		return g == other
	}
	return *g == *other
}
