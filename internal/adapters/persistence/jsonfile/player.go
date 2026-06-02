package jsonfile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"pork-history-updater/internal/application"
	"pork-history-updater/internal/domain"
)

// Compile-time check.
var _ application.PlayerRepository = (*playerRepository)(nil)

// playerRepository stores player data as JSON files on disk.
// Each player has a single state file (player_<nick>_<server>.json)
// and an optional history file (player_history_<nick>_<server>.jsonl).
type playerRepository struct {
	baseDir string
	mutex   sync.Map
}

// NewPlayerRepository creates a file-based repository rooted at baseDir.
func NewPlayerRepository(baseDir string) *playerRepository {
	return &playerRepository{baseDir: baseDir}
}

// ============================================================================
// DTOs
// ============================================================================

// filePlayer is the on-disk JSON representation of a player state.
type filePlayer struct {
	Nickname         string  `json:"nickname"`
	CharID           *uint32 `json:"char_id,omitempty"`
	FirstName        string  `json:"first_name"`
	LastName         string  `json:"last_name"`
	GuildRank        int     `json:"guild_rank"`
	GuildRankName    string  `json:"guild_rank_name"`
	Level            int     `json:"level"`
	Faction          string  `json:"faction"`
	Profession       string  `json:"profession"`
	ProfessionTitle  string  `json:"profession_title"`
	Gender           string  `json:"gender"`
	Breed            string  `json:"breed"`
	DefenderRank     int     `json:"defender_rank"`
	DefenderRankName string  `json:"defender_rank_name"`
	GuildID          int     `json:"guild_id"`
	GuildName        string  `json:"guild_name"`
	Server           int     `json:"server"`
	LastChecked      int64   `json:"last_checked"`
	LastChanged      int64   `json:"last_changed"`
	Deleted          bool    `json:"deleted"`
}

// filePlayerFromDomain converts a domain.Player into its JSON file representation.
func filePlayerFromDomain(p *domain.Player) filePlayer {
	return filePlayer{
		Nickname:         p.Nickname,
		CharID:           p.CharID,
		FirstName:        p.FirstName,
		LastName:         p.LastName,
		Level:            p.Level,
		Faction:          p.Faction,
		Profession:       p.Profession,
		ProfessionTitle:  p.ProfessionTitle,
		Gender:           p.Gender,
		Breed:            p.Breed,
		DefenderRank:     p.DefenderRank,
		DefenderRankName: p.DefenderRankName,
		Server:           p.Server,
		LastChecked:      p.LastChecked.Unix(),
		LastChanged:      p.LastChanged.Unix(),
		Deleted:          p.Deleted,
		GuildID:          0,
		GuildName:        "",
		GuildRank:        0,
		GuildRankName:    "",
	}
}

// ToPlayer converts a filePlayer back into a domain.Player.
func (fp filePlayer) ToPlayer() domain.Player {
	p := domain.Player{
		Nickname:         fp.Nickname,
		CharID:           fp.CharID,
		FirstName:        fp.FirstName,
		LastName:         fp.LastName,
		Level:            fp.Level,
		Faction:          fp.Faction,
		Profession:       fp.Profession,
		ProfessionTitle:  fp.ProfessionTitle,
		Gender:           fp.Gender,
		Breed:            fp.Breed,
		DefenderRank:     fp.DefenderRank,
		DefenderRankName: fp.DefenderRankName,
		Server:           fp.Server,
		LastChecked:      time.Unix(fp.LastChecked, 0),
		LastChanged:      time.Unix(fp.LastChanged, 0),
		Deleted:          fp.Deleted,
	}
	if fp.GuildID != 0 {
		p.MyGuild = &domain.GuildMembership{
			ID:       fp.GuildID,
			Name:     fp.GuildName,
			Rank:     fp.GuildRank,
			RankName: fp.GuildRankName,
		}
	}
	return p
}

// ============================================================================
// Helpers
// ============================================================================

// playerPath returns the filesystem path for a player's state file.
func (r *playerRepository) playerPath(nickname string, server int) string {
	name := safeName(nickname)
	return filepath.Join(r.baseDir, fmt.Sprintf("player_%s_%d.json", name, server))
}

// historyPath returns the filesystem path for a player's history file.
func (r *playerRepository) historyPath(nickname string, server int) string {
	name := safeName(nickname)
	return filepath.Join(r.baseDir, fmt.Sprintf("player_history_%s_%d.jsonl", name, server))
}

// safeName replaces filesystem-unfriendly characters in a nickname.
func safeName(nickname string) string {
	return strings.ReplaceAll(nickname, "/", "_")
}

// ============================================================================
// Repository implementation
// ============================================================================

// GetByName loads a player from disk by nickname and server.
func (r *playerRepository) GetByName(nickname string, server int) (*domain.Player, error) {
	return r.loadFile(r.playerPath(nickname, server))
}

// loadFile reads and unmarshals a single player state file.
func (r *playerRepository) loadFile(name string) (*domain.Player, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var fp filePlayer
	if err := json.Unmarshal(data, &fp); err != nil {
		return nil, err
	}
	result := fp.ToPlayer()
	return &result, nil
}

// StreamPlayers streams all player files from disk.
// Files that cannot be parsed are silently skipped.
func (r *playerRepository) StreamPlayers(ctx context.Context) (<-chan application.PlayerResult, error) {
	entries, err := os.ReadDir(r.baseDir)
	if err != nil {
		return nil, err
	}
	playerChannel := make(chan application.PlayerResult, 1)
	go func() {
		defer close(playerChannel)
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasPrefix(entry.Name(), "player_") || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			p, err := r.loadFile(filepath.Join(r.baseDir, entry.Name()))
			if err != nil {
				continue // skip unreadable files
			}
			select {
			case playerChannel <- application.PlayerResult{Player: *p}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return playerChannel, nil
}

// getMutex returns a mutex for the given file path, creating one if necessary.
func (r *playerRepository) getMutex(path string) *sync.Mutex {
	v, _ := r.mutex.LoadOrStore(path, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// InsertHistoryEvent appends a JSON line to the player's history file.
func (r *playerRepository) InsertHistoryEvent(ctx context.Context, player *domain.Player) error {
	fp := filePlayerFromDomain(player)
	if player.MyGuild != nil {
		fp.GuildID = player.MyGuild.ID
		fp.GuildName = player.MyGuild.Name
		fp.GuildRank = player.MyGuild.Rank
		fp.GuildRankName = player.MyGuild.RankName
	}
	data, err := json.Marshal(fp)
	if err != nil {
		return err
	}
	fileName := r.historyPath(player.Nickname, player.Server)
	mu := r.getMutex(fileName)
	mu.Lock()
	defer mu.Unlock()
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	_, err = f.WriteString("\n")
	return err
}

// Update overwrites the player's state file.
func (r *playerRepository) Update(ctx context.Context, player *domain.Player) error {
	fp := filePlayerFromDomain(player)
	if player.MyGuild != nil {
		fp.GuildID = player.MyGuild.ID
		fp.GuildName = player.MyGuild.Name
		fp.GuildRank = player.MyGuild.Rank
		fp.GuildRankName = player.MyGuild.RankName
	}
	data, err := json.MarshalIndent(fp, "", "  ")
	if err != nil {
		return err
	}
	fileName := r.playerPath(player.Nickname, player.Server)
	mu := r.getMutex(fileName)
	mu.Lock()
	defer mu.Unlock()
	return os.WriteFile(fileName, data, 0644)
}

// MarkChecked marks the player as checked at the given time.
func (r *playerRepository) MarkChecked(ctx context.Context, player *domain.Player, checkedAt time.Time) error {
	fp := filePlayerFromDomain(player)
	if player.MyGuild != nil {
		fp.GuildID = player.MyGuild.ID
		fp.GuildName = player.MyGuild.Name
		fp.GuildRank = player.MyGuild.Rank
		fp.GuildRankName = player.MyGuild.RankName
	}
	fp.LastChecked = checkedAt.Unix()
	data, err := json.MarshalIndent(fp, "", "  ")
	if err != nil {
		return err
	}
	fileName := r.playerPath(player.Nickname, player.Server)
	mu := r.getMutex(fileName)
	mu.Lock()
	defer mu.Unlock()
	return os.WriteFile(fileName, data, 0644)
}
