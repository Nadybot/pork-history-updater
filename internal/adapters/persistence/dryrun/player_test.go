package dryrun

import (
	"context"
	"errors"
	"testing"
	"time"

	"pork-history-updater/internal/application"
	"pork-history-updater/internal/domain"
)

// mockRepo is a test double that tracks which methods were called.
// InsertHistoryEvent and Update return errors to prove they are never reached.
type mockRepo struct {
	getByIDCalled       bool
	getByNameCalled     bool
	streamPlayersCalled bool
	insertCalled        bool
	updateCalled        bool
	markCheckedCalled   bool

	getByIDResult    *domain.Player
	getByIDErr       error
	getByNameResult  *domain.Player
	getByNameErr     error
	streamPlayersErr error
}

func (m *mockRepo) GetByName(nickname string, server int) (*domain.Player, error) {
	m.getByNameCalled = true
	return m.getByNameResult, m.getByNameErr
}

func (m *mockRepo) StreamPlayers(ctx context.Context) (<-chan application.PlayerResult, error) {
	m.streamPlayersCalled = true
	return nil, m.streamPlayersErr
}

func (m *mockRepo) InsertHistoryEvent(ctx context.Context, player *domain.Player) error {
	m.insertCalled = true
	return errors.New("insert must not be called in dry run")
}

func (m *mockRepo) Update(ctx context.Context, player *domain.Player) error {
	m.updateCalled = true
	return errors.New("update must not be called in dry run")
}

func (m *mockRepo) MarkChecked(ctx context.Context, player *domain.Player, checkedAt time.Time) error {
	m.markCheckedCalled = true
	return errors.New("markChecked must not be called in dry run")
}

// ============================================================================
// Delegation tests
// ============================================================================

func TestPlayerRepository_GetByName_delegates(t *testing.T) {
	mock := &mockRepo{getByNameResult: &domain.Player{Nickname: "Bob"}}
	repo := NewPlayerRepository(mock)

	player, err := repo.GetByName("Bob", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.getByNameCalled {
		t.Error("expected GetByName to be delegated")
	}
	if player.Nickname != "Bob" {
		t.Errorf("Nickname = %q, want %q", player.Nickname, "Bob")
	}
}

func TestPlayerRepository_StreamPlayers_delegates(t *testing.T) {
	mock := &mockRepo{}
	repo := NewPlayerRepository(mock)

	ctx := context.Background()
	_, _ = repo.StreamPlayers(ctx)
	if !mock.streamPlayersCalled {
		t.Error("expected StreamPlayers to be delegated")
	}
}

// ============================================================================
// No-op tests
// ============================================================================

func TestPlayerRepository_InsertHistoryEvent_noOp(t *testing.T) {
	mock := &mockRepo{}
	repo := NewPlayerRepository(mock)

	player := &domain.Player{Nickname: "Charlie", Server: 5}
	err := repo.InsertHistoryEvent(context.Background(), player)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.insertCalled {
		t.Fatal("expected InsertHistoryEvent NOT to be delegated to wrapped repository")
	}
}

func TestPlayerRepository_Update_noOp(t *testing.T) {
	mock := &mockRepo{}
	repo := NewPlayerRepository(mock)

	player := &domain.Player{Nickname: "Dana", Server: 5}
	err := repo.Update(context.Background(), player)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.updateCalled {
		t.Fatal("expected Update NOT to be delegated to wrapped repository")
	}
}
