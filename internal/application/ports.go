package application

import (
	"context"
	"time"

	"pork-history-updater/internal/domain"
)

type PlayerResult struct {
	Player domain.Player
	Err    error
}

type PlayerRepository interface {
	GetByName(nickname string, server int) (*domain.Player, error)
	Update(ctx context.Context, player *domain.Player) error
	MarkChecked(ctx context.Context, player *domain.Player, checkedAt time.Time) error
	StreamPlayers(ctx context.Context) (<-chan PlayerResult, error)
	InsertHistoryEvent(ctx context.Context, player *domain.Player) error
}

type CharInfoFetcher interface {
	FetchByNameAsPlayer(nickname string, server int) (*domain.Player, error)
}
