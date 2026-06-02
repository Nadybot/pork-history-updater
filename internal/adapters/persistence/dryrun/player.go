package dryrun

import (
	"context"
	"log/slog"
	"time"

	"pork-history-updater/internal/application"
	"pork-history-updater/internal/domain"
)

// Compile-time check.
var _ application.PlayerRepository = (*playerRepository)(nil)

type playerRepository struct {
	wrapped application.PlayerRepository
}

// NewPlayerRepository wraps another PlayerRepository and logs actions instead of persisting them.
func NewPlayerRepository(wrapped application.PlayerRepository) *playerRepository {
	return &playerRepository{
		wrapped: wrapped,
	}
}

// GetByName delegates to the wrapped repository.
func (r *playerRepository) GetByName(nickname string, server int) (*domain.Player, error) {
	return r.wrapped.GetByName(nickname, server)
}

// StreamPlayers delegates to the wrapped repository.
func (r *playerRepository) StreamPlayers(ctx context.Context) (<-chan application.PlayerResult, error) {
	return r.wrapped.StreamPlayers(ctx)
}

// InsertHistoryEvent logs the action and returns nil without persisting.
func (r *playerRepository) InsertHistoryEvent(ctx context.Context, player *domain.Player) error {
	slog.InfoContext(ctx, "[DRY RUN] Would insert history event", "player", player.Nickname, "server", player.Server)
	return nil
}

// Update logs the action and returns nil without persisting.
func (r *playerRepository) Update(ctx context.Context, player *domain.Player) error {
	slog.InfoContext(ctx, "[DRY RUN] Would update player", "player", player.Nickname, "server", player.Server)
	return nil
}

// MarkChecked logs the action and returns nil without persisting.
func (r *playerRepository) MarkChecked(ctx context.Context, player *domain.Player, checkedAt time.Time) error {
	slog.InfoContext(ctx, "[DRY RUN] Would mark player as checked", "player", player.Nickname, "server", player.Server)
	return nil
}
