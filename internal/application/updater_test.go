package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"pork-history-updater/internal/domain"
)

// ============================================================================
// Mocks
// ============================================================================

type mockCharInfoFetcher struct {
	result *domain.Player
	err    error
}

func (m *mockCharInfoFetcher) FetchByNameAsPlayer(nickname string, server int) (*domain.Player, error) {
	return m.result, m.err
}

type mockPlayerRepo struct {
	insertErr         error
	updateErr         error
	checkErr          error
	insertCalled      bool
	updateCalled      bool
	markCheckedCalled bool
	insertedPlayer    *domain.Player
	updatedPlayer     *domain.Player
	checkedPlayer     *domain.Player
}

func (m *mockPlayerRepo) GetByID(id int, server int) (*domain.Player, error) {
	return nil, nil
}

func (m *mockPlayerRepo) GetByName(nickname string, server int) (*domain.Player, error) {
	return nil, nil
}

func (m *mockPlayerRepo) StreamPlayers(ctx context.Context) (<-chan PlayerResult, error) {
	return nil, nil
}

func (m *mockPlayerRepo) InsertHistoryEvent(ctx context.Context, player *domain.Player) error {
	m.insertCalled = true
	m.insertedPlayer = player
	return m.insertErr
}

func (m *mockPlayerRepo) Update(ctx context.Context, player *domain.Player) error {
	m.updateCalled = true
	m.updatedPlayer = player
	return m.updateErr
}

func (m *mockPlayerRepo) MarkChecked(ctx context.Context, player *domain.Player, checkedAt time.Time) error {
	m.markCheckedCalled = true
	m.checkedPlayer = player
	return m.checkErr
}

// ============================================================================
// processPlayer tests
// ============================================================================

func TestProcessPlayer_noChange(t *testing.T) {
	stored := domain.Player{Nickname: "Alice", Server: 5, Level: 10}
	mockFetch := &mockCharInfoFetcher{result: &domain.Player{Nickname: "Alice", Server: 5, Level: 10}}
	mockRepo := &mockPlayerRepo{}

	updater := NewUpdater(mockFetch, mockRepo, 1)
	result := updater.processPlayer(context.Background(), stored)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Updated {
		t.Error("expected Updated = false")
	}
	if mockRepo.insertCalled {
		t.Error("expected InsertHistoryEvent not to be called")
	}
	if mockRepo.updateCalled {
		t.Error("expected Update not to be called")
	}
	if !mockRepo.markCheckedCalled {
		t.Error("expected MarkChecked to be called")
	}
	if mockRepo.checkedPlayer == nil || mockRepo.checkedPlayer.Level != 10 {
		t.Errorf("expected checked player to have Level=10, got %+v", mockRepo.checkedPlayer)
	}
	if !mockRepo.checkedPlayer.LastChanged.Before(mockRepo.checkedPlayer.LastChecked) {
		t.Error("expected updated player to be changed before the last check")
	}
}

func TestProcessPlayer_changed(t *testing.T) {
	stored := domain.Player{Nickname: "Alice", Server: 5, Level: 10}
	remote := &domain.Player{Nickname: "Alice", Server: 5, Level: 99}
	mockFetch := &mockCharInfoFetcher{result: remote}
	mockRepo := &mockPlayerRepo{}

	updater := NewUpdater(mockFetch, mockRepo, 1)
	result := updater.processPlayer(context.Background(), stored)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.Updated {
		t.Error("expected Updated = true")
	}
	if !mockRepo.insertCalled {
		t.Error("expected InsertHistoryEvent to be called")
	}
	if mockRepo.insertedPlayer == nil || mockRepo.insertedPlayer.Level != 99 {
		t.Errorf("expected updated player to have Level=99, got %+v", mockRepo.updatedPlayer)
	}
	if !mockRepo.updateCalled {
		t.Error("expected Update to be called")
	}
	if mockRepo.updatedPlayer == nil || mockRepo.updatedPlayer.Level != 99 {
		t.Errorf("expected updated player to have Level=99, got %+v", mockRepo.updatedPlayer)
	}
	if !mockRepo.updatedPlayer.LastChanged.Before(mockRepo.updatedPlayer.LastChecked) {
		t.Error("expected updated player to be changed before the last check")
	}
	if mockRepo.markCheckedCalled {
		t.Error("expected MarkChecked not to be called")
	}
}

func TestProcessPlayer_deleted(t *testing.T) {
	charID := uint32(42)
	stored := domain.Player{Nickname: "Alice", Server: 5, Level: 10, Deleted: false, CharID: &charID}
	mockFetch := &mockCharInfoFetcher{result: nil}
	mockRepo := &mockPlayerRepo{}

	updater := NewUpdater(mockFetch, mockRepo, 1)
	result := updater.processPlayer(context.Background(), stored)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.Updated {
		t.Error("expected Updated = true")
	}
	if !result.Deleted {
		t.Error("expected Deleted = true")
	}
	if !mockRepo.insertCalled {
		t.Error("expected InsertHistoryEvent to be called")
	}
	if mockRepo.insertedPlayer == nil || !mockRepo.insertedPlayer.Deleted {
		t.Error("expected inserted player to be marked as deleted")
	}
	if mockRepo.insertedPlayer != nil && mockRepo.insertedPlayer.CharID == nil {
		t.Error("expected deleted player to keep valid CharID")
	}
	if mockRepo.insertedPlayer != nil && *mockRepo.insertedPlayer.CharID != 42 {
		t.Errorf("expected CharID to remain 42, got %d", *mockRepo.insertedPlayer.CharID)
	}
	if !mockRepo.updateCalled {
		t.Error("expected Update to be called")
	}
	if mockRepo.updatedPlayer == nil || !mockRepo.updatedPlayer.Deleted {
		t.Error("expected updated player to be marked as deleted")
	}
	if mockRepo.updatedPlayer != nil && mockRepo.updatedPlayer.CharID == nil {
		t.Error("expected updated player to keep valid CharID")
	}
	if mockRepo.updatedPlayer != nil && *mockRepo.updatedPlayer.CharID != 42 {
		t.Errorf("expected CharID to remain 42, got %d", *mockRepo.updatedPlayer.CharID)
	}
	if !mockRepo.updatedPlayer.LastChanged.Before(mockRepo.updatedPlayer.LastChecked) {
		t.Error("expected updated player to be changed before the last check")
	}
	if mockRepo.markCheckedCalled {
		t.Error("expected MarkChecked not to be called")
	}
}

func TestProcessPlayer_undeleted(t *testing.T) {
	stored := domain.Player{Nickname: "Alice", Server: 5, Level: 10, Deleted: true}
	remoteCharID := uint32(99)
	remote := &domain.Player{Nickname: "Alice", Server: 5, Level: 1, CharID: &remoteCharID}
	mockFetch := &mockCharInfoFetcher{result: remote}
	mockRepo := &mockPlayerRepo{}

	updater := NewUpdater(mockFetch, mockRepo, 1)
	result := updater.processPlayer(context.Background(), stored)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.Updated {
		t.Error("expected Updated = true")
	}
	if result.Deleted {
		t.Error("expected Deleted = false")
	}
	if !mockRepo.insertCalled {
		t.Error("expected InsertHistoryEvent to be called")
	}
	if mockRepo.insertedPlayer == nil || mockRepo.insertedPlayer.Deleted {
		t.Error("expected inserted player to be marked as not deleted")
	}
	if mockRepo.insertedPlayer != nil && mockRepo.insertedPlayer.CharID == nil {
		t.Error("expected undeleted player to have valid CharID")
	}
	if mockRepo.insertedPlayer != nil && *mockRepo.insertedPlayer.CharID != 99 {
		t.Errorf("expected CharID to be 99, got %d", *mockRepo.insertedPlayer.CharID)
	}
	if !mockRepo.updateCalled {
		t.Error("expected Update to be called")
	}
	if mockRepo.updatedPlayer == nil || mockRepo.updatedPlayer.Deleted {
		t.Error("expected updated player to be marked as not deleted")
	}
	if mockRepo.updatedPlayer != nil && mockRepo.updatedPlayer.CharID == nil {
		t.Error("expected undeleted player to have valid CharID")
	}
	if mockRepo.updatedPlayer != nil && *mockRepo.updatedPlayer.CharID != 99 {
		t.Errorf("expected CharID to be 99, got %d", *mockRepo.updatedPlayer.CharID)
	}
	if !mockRepo.updatedPlayer.LastChanged.Before(mockRepo.updatedPlayer.LastChecked) {
		t.Errorf("expected updated player to be changed before the last check: %s !< %s", mockRepo.updatedPlayer.LastChanged, mockRepo.updatedPlayer.LastChecked)
	}
	if mockRepo.markCheckedCalled {
		t.Error("expected MarkChecked not to be called")
	}
}

func TestProcessPlayer_alreadyDeleted(t *testing.T) {
	stored := domain.Player{Nickname: "Alice", Server: 5, Level: 10, Deleted: true}
	mockFetch := &mockCharInfoFetcher{result: nil}
	mockRepo := &mockPlayerRepo{}

	updater := NewUpdater(mockFetch, mockRepo, 1)
	result := updater.processPlayer(context.Background(), stored)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Updated {
		t.Error("expected Updated = false for already deleted player")
	}
	if mockRepo.insertCalled {
		t.Error("expected InsertHistoryEvent not to be called")
	}
	if mockRepo.updateCalled {
		t.Error("expected Update not to be called")
	}
	if !mockRepo.markCheckedCalled {
		t.Error("expected MarkChecked to be called")
	}
	if mockRepo.checkedPlayer == nil || mockRepo.checkedPlayer.Level != 10 {
		t.Errorf("expected checked player to have Level=10, got %+v", mockRepo.checkedPlayer)
	}
	if !mockRepo.checkedPlayer.LastChanged.Before(mockRepo.checkedPlayer.LastChecked) {
		t.Error("expected updated player to be changed before the last check")
	}
}

func TestProcessPlayer_fetchError(t *testing.T) {
	stored := domain.Player{Nickname: "Alice", Server: 5}
	mockFetch := &mockCharInfoFetcher{err: errors.New("network timeout")}
	mockRepo := &mockPlayerRepo{}

	updater := NewUpdater(mockFetch, mockRepo, 1)
	result := updater.processPlayer(context.Background(), stored)

	if result.Err == nil {
		t.Fatal("expected error, got nil")
	}
	if result.Updated {
		t.Error("expected Updated = false")
	}
	if mockRepo.insertCalled {
		t.Error("expected InsertHistoryEvent not to be called")
	}
}

func TestProcessPlayer_insertError(t *testing.T) {
	stored := domain.Player{Nickname: "Alice", Server: 5, Level: 10}
	remote := &domain.Player{Nickname: "Alice", Server: 5, Level: 99}
	mockFetch := &mockCharInfoFetcher{result: remote}
	mockRepo := &mockPlayerRepo{insertErr: errors.New("db locked")}

	updater := NewUpdater(mockFetch, mockRepo, 1)
	result := updater.processPlayer(context.Background(), stored)

	if result.Err == nil {
		t.Fatal("expected error, got nil")
	}
	if result.Updated {
		t.Error("expected Updated = false when insert fails")
	}
	if !mockRepo.insertCalled {
		t.Error("expected InsertHistoryEvent to be called")
	}
	if mockRepo.updateCalled {
		t.Error("expected Update not to be called when insert fails")
	}
}

func TestProcessPlayer_updateError(t *testing.T) {
	stored := domain.Player{Nickname: "Alice", Server: 5, Level: 10}
	remote := &domain.Player{Nickname: "Alice", Server: 5, Level: 99}
	mockFetch := &mockCharInfoFetcher{result: remote}
	mockRepo := &mockPlayerRepo{updateErr: errors.New("db timeout")}

	updater := NewUpdater(mockFetch, mockRepo, 1)
	result := updater.processPlayer(context.Background(), stored)

	if result.Err == nil {
		t.Fatal("expected error, got nil")
	}
	if result.Updated {
		t.Error("expected Updated = false when update fails")
	}
	if !mockRepo.insertCalled {
		t.Error("expected InsertHistoryEvent to be called")
	}
	if !mockRepo.updateCalled {
		t.Error("expected Update to be called")
	}
}
