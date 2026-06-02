package application

import (
	"context"
	"log/slog"
	"os"
	"pork-history-updater/internal/domain"
	"sync"
	"time"
)

type WorkerResult struct {
	Deleted bool
	Updated bool
	Err     error
	Player  *domain.Player
}

type Updater struct {
	CharInfo   CharInfoFetcher
	PlayerRepo PlayerRepository
	MaxWorkers int
}

// NewUpdater creates a new Updater with the given dependencies.
func NewUpdater(charInfo CharInfoFetcher, playerRepo PlayerRepository, maxWorkers int) *Updater {
	return &Updater{
		CharInfo:   charInfo,
		PlayerRepo: playerRepo,
		MaxWorkers: maxWorkers,
	}
}

// Run starts the update process, streaming all players and processing them concurrently.
func (app *Updater) Run() {
	slog.Info("Starting pork history updater...")

	// Create a new context with a timeout or cancellation mechanism.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChannel := make(chan WorkerResult)
	playerChannel, err := app.PlayerRepo.StreamPlayers(ctx)
	if err != nil {
		slog.Error("Error streaming players", "error", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	for i := 0; i < app.MaxWorkers; i++ {
		wg.Go(func() {
			app.processPlayers(ctx, playerChannel, resultChannel)
		})
	}

	// Close the channel if all results are in
	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	// Process results
	for result := range resultChannel {
		if result.Err != nil {
			slog.Warn("Player error", "player", result.Player.Nickname, "server", result.Player.Server, "error", result.Err)
			continue
		}
		if result.Updated {
			if result.Deleted {
				slog.Info("Player deleted", "player", result.Player.Nickname, "server", result.Player.Server)
			} else {
				slog.Info("Player updated", "player", result.Player.Nickname, "server", result.Player.Server)
			}
		} else {
			slog.Debug("Player skipped", "player", result.Player.Nickname, "server", result.Player.Server)
		}
	}
}

// processPlayers processes players from the player channel and sends results to the result channel.
// It uses a context to handle cancellation and timeouts.
// It returns when the player channel is closed.
func (app *Updater) processPlayers(ctx context.Context, playerChannel <-chan PlayerResult, resultChannel chan<- WorkerResult) {
	var result WorkerResult
	for storedPlayer := range playerChannel {
		if storedPlayer.Err != nil {
			result = WorkerResult{Err: storedPlayer.Err}
		} else {
			result = app.processPlayer(ctx, storedPlayer.Player)
		}
		resultChannel <- result
	}
}

// processPlayer updates a single player in the database.
// It fetches the player's data from the remote source and compares it with the stored data.
// If the data is different, it updates the stored data.
// If the player is deleted, it marks the player as deleted in the database.
// It returns a WorkerResult with the player and the error (if any).
func (app *Updater) processPlayer(ctx context.Context, storedPlayer domain.Player) WorkerResult {
	var remotePlayer *domain.Player
	result := WorkerResult{Player: &storedPlayer}
	storedPlayer.LastChecked = time.Now()
	remotePlayer, result.Err = app.CharInfo.FetchByNameAsPlayer(storedPlayer.Nickname, storedPlayer.Server)
	if remotePlayer != nil {
		remotePlayer.LastChecked = storedPlayer.LastChecked
	}
	if result.Err != nil {
		return result
	}
	if storedPlayer.Equals(remotePlayer) {
		app.PlayerRepo.MarkChecked(ctx, &storedPlayer, storedPlayer.LastChecked)
		return result
	}
	if remotePlayer == nil && storedPlayer.Deleted {
		app.PlayerRepo.MarkChecked(ctx, &storedPlayer, storedPlayer.LastChecked)
		return result
	}
	if remotePlayer == nil {
		remotePlayer = &domain.Player{}
		*remotePlayer = storedPlayer
		remotePlayer.Deleted = true
		result.Deleted = true
	}
	result.Err = app.PlayerRepo.InsertHistoryEvent(ctx, remotePlayer)
	if result.Err != nil {
		return result
	}
	result.Err = app.PlayerRepo.Update(ctx, remotePlayer)
	if result.Err != nil {
		return result
	}
	result.Updated = true
	changes := storedPlayer.ChangesTo(remotePlayer)
	slog.Debug("Calculated changes", "player", storedPlayer.Nickname, "changes", changes)
	return result
}
