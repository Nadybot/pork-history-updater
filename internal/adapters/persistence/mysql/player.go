package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"pork-history-updater/internal/application"
	"pork-history-updater/internal/domain"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	ErrMySQLServerGoneAway   = 2006
	ErrMySQLLostConnection   = 2013
	ErrMySQLQueryInterrupted = 1317
	pageSize                 = 100
	numRetries               = 10
)

// Compile-time check.
var _ application.PlayerRepository = (*playerRepository)(nil)

type playerKey struct {
	Nickname string
	Server   int
}

// dbPlayer is a DTO with DB tags for sqlx.
type dbPlayer struct {
	Nickname         string        `db:"nickname"`
	CharID           sql.NullInt64 `db:"char_id"`
	FirstName        string        `db:"first_name"`
	LastName         string        `db:"last_name"`
	GuildRank        int           `db:"guild_rank"`
	GuildRankName    string        `db:"guild_rank_name"`
	Level            int           `db:"level"`
	Faction          string        `db:"faction"`
	Profession       string        `db:"profession"`
	ProfessionTitle  string        `db:"profession_title"`
	Gender           string        `db:"gender"`
	Breed            string        `db:"breed"`
	DefenderRank     int           `db:"defender_rank"`
	DefenderRankName string        `db:"defender_rank_name"`
	GuildID          int           `db:"guild_id"`
	GuildName        string        `db:"guild_name"`
	Server           int           `db:"server"`
	LastChecked      int           `db:"last_checked"`
	LastChanged      int           `db:"last_changed"`
	Deleted          bool          `db:"deleted"`
}

// ToDomain converts a database player DTO into a domain Player.
func (p dbPlayer) ToDomain() domain.Player {
	result := domain.Player{
		Nickname:         p.Nickname,
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
		LastChecked:      time.Unix(int64(p.LastChecked), 0),
		LastChanged:      time.Unix(int64(p.LastChanged), 0),
		Deleted:          p.Deleted,
	}
	if p.CharID.Valid {
		id := uint32(p.CharID.Int64)
		result.CharID = &id
	}
	if p.GuildID != 0 {
		result.MyGuild = &domain.GuildMembership{
			ID:       p.GuildID,
			Name:     p.GuildName,
			Rank:     p.GuildRank,
			RankName: p.GuildRankName,
		}
	}
	return result
}

// isConnectionError returns true if the error indicates a connection problem.
// This is used to determine whether we should retry the query.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if mysqlErr, ok := errors.AsType[*mysql.MySQLError](err); ok {
		switch mysqlErr.Number {
		case ErrMySQLServerGoneAway,
			ErrMySQLLostConnection,
			ErrMySQLQueryInterrupted: // Query interrupted
			return true
		}
	}
	return errors.Is(err, driver.ErrBadConn)
}

type playerRepository struct {
	db *sqlx.DB
}

// dbPlayerFromDomain converts a domain Player into its database DTO representation.
func dbPlayerFromDomain(p *domain.Player) dbPlayer {
	dto := dbPlayer{
		Nickname:         p.Nickname,
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
		LastChecked:      int(p.LastChecked.Unix()),
		LastChanged:      int(p.LastChanged.Unix()),
		Deleted:          p.Deleted,
	}
	if p.CharID != nil {
		dto.CharID = sql.NullInt64{Int64: int64(*p.CharID), Valid: true}
	} else {
		dto.CharID = sql.NullInt64{Valid: false}
	}
	if p.MyGuild != nil {
		dto.GuildRank = p.MyGuild.Rank
		dto.GuildRankName = p.MyGuild.RankName
		dto.GuildID = p.MyGuild.ID
		dto.GuildName = p.MyGuild.Name
	}
	return dto
}

// NewPlayerRepository creates a new MySQL-backed player repository.
func NewPlayerRepository(db *sqlx.DB) *playerRepository {
	return &playerRepository{db: db}
}

// GetByName loads a player from the database by nickname and server.
func (r *playerRepository) GetByName(nickname string, server int) (*domain.Player, error) {
	var p dbPlayer
	stmt, err := r.db.Preparex(`SELECT * FROM player WHERE nickname=? AND server=?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRowx(nickname, server)
	err = row.StructScan(&p)
	if err != nil {
		return nil, err
	}
	result := p.ToDomain()
	return &result, nil
}

// retryWrap executes the given function with exponential backoff on connection errors.
func retryWrap[T any](ctx context.Context, numRetries int, wrapped func(ctx context.Context) (T, error)) (T, error) {
	var zero T
	var lastErr error
	for attempts := range numRetries {
		result, err := wrapped(ctx)
		lastErr = err
		if err == nil {
			return result, nil
		}
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		if attempts >= numRetries-1 {
			lastErr = fmt.Errorf("max retries exceeded: %w", err)
			break
		}
		if isConnectionError(err) {
			backoff := time.Duration(attempts*attempts+1) * time.Second
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(backoff):
			}
			continue
		}
		return zero, err
	}
	return zero, lastErr
}

// fetchPage loads the next page of players ordered by (nickname, server).
func (r *playerRepository) fetchPage(ctx context.Context, lastKey playerKey) ([]dbPlayer, error) {
	return retryWrap(ctx, numRetries, func(ctx context.Context) ([]dbPlayer, error) {
		var players []dbPlayer
		err := r.db.SelectContext(ctx, &players, `
				SELECT * FROM player
				WHERE (nickname, server) > (?, ?)
				ORDER BY nickname, server
				LIMIT ?
			`, lastKey.Nickname, lastKey.Server, pageSize)
		return players, err
	})
}

// StreamPlayers streams all players from the database to a channel.
// It stops streaming when the context is canceled or an error occurs.
func (r *playerRepository) StreamPlayers(ctx context.Context) (<-chan application.PlayerResult, error) {
	const channelBuffer = 100
	playerCh := make(chan application.PlayerResult, channelBuffer)

	go func() {
		defer close(playerCh)
		lastKey := playerKey{}

		for {
			players, err := r.fetchPage(ctx, lastKey)
			if err != nil {
				select {
				case playerCh <- application.PlayerResult{Err: fmt.Errorf("querying players: %w", err)}:
				case <-ctx.Done():
				}
				return
			}

			if len(players) == 0 {
				return
			}

			for _, p := range players {
				select {
				case playerCh <- application.PlayerResult{Player: p.ToDomain()}:
				case <-ctx.Done():
					return
				}
			}

			last := players[len(players)-1]
			lastKey = playerKey{Nickname: last.Nickname, Server: last.Server}

			if len(players) < pageSize {
				return // last page
			}
		}
	}()

	return playerCh, nil
}

// InsertHistoryEvent writes a player snapshot to the history table.
func (r *playerRepository) InsertHistoryEvent(ctx context.Context, player *domain.Player) error {
	dto := dbPlayerFromDomain(player)
	query, args, err := sqlx.Named(`
		INSERT INTO player_history (
			nickname, char_id, first_name, last_name, guild_rank, guild_rank_name,
			level, faction, profession, profession_title, gender, breed, defender_rank,
			defender_rank_name, guild_id, guild_name, server, last_checked, last_changed, delete
		) VALUES (
			:nickname, :char_id, :first_name, :last_name, :guild_rank, :guild_rank_name,
			:level, :faction, :profession, :profession_title, :gender, :breed, :defender_rank,
			:defender_rank_name, :guild_id, :guild_name, :server, :last_checked, :last_changed, :deleted
		)`,
		dto,
	)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// Update persists the current player state to the database.
func (r *playerRepository) Update(ctx context.Context, player *domain.Player) error {
	dto := dbPlayerFromDomain(player)
	query, args, err := sqlx.Named(`
	    UPDATE player SET
	        char_id = :char_id,
	        first_name = :first_name,
	        last_name = :last_name,
	        guild_rank = :guild_rank,
	        guild_rank_name = :guild_rank_name,
	        level = :level,
	        faction = :faction,
	        profession = :profession,
	        profession_title = :profession_title,
	        gender = :gender,
	        breed = :breed,
	        defender_rank = :defender_rank,
	        defender_rank_name = :defender_rank_name,
	        guild_id = :guild_id,
	        guild_name = :guild_name,
	        last_checked = :last_checked,
	        last_changed = :last_changed,
	        deleted = :deleted
	    WHERE nickname = :nickname
		  AND server = :server
	`, dto)
	if err != nil {
		return err
	}

	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// MarkChecked marks the player as checked at the given time.
// This is useful to avoid checking the same player multiple times in a short period of time.
func (r *playerRepository) MarkChecked(ctx context.Context, player *domain.Player, checkedAt time.Time) error {
	dto := dbPlayerFromDomain(player)
	dto.LastChecked = int(checkedAt.Unix())
	query, args, err := sqlx.Named(`
	    UPDATE player SET
	        last_checked = :last_checked
	    WHERE nickname = :nickname AND server = :server
	`, dto)
	if err != nil {
		return err
	}

	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}
