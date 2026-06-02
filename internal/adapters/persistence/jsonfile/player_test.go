package jsonfile

import (
	"context"
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

//go:embed testdata/*
var testdataFS embed.FS

func copyTestdata(t *testing.T, dst string) {
	t.Helper()
	entries, err := testdataFS.ReadDir("testdata")
	if err != nil {
		t.Fatalf("reading embedded testdata: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := testdataFS.ReadFile("testdata/" + entry.Name())
		if err != nil {
			t.Fatalf("reading embedded %s: %v", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(dst, entry.Name()), data, 0o644); err != nil {
			t.Fatalf("writing %s to temp dir: %v", entry.Name(), err)
		}
	}
}

func TestPlayerRepository_roundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	copyTestdata(t, tmpDir)
	repo := NewPlayerRepository(tmpDir)

	// Pigtail already exists in testdata at level 199 with 2 history entries.
	loaded, err := repo.GetByName("Pigtail", 5)
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if loaded.Level != 199 {
		t.Errorf("initial Level = %d, want %d", loaded.Level, 199)
	}
	if loaded.MyGuild == nil || loaded.MyGuild.Name != "Troet" {
		t.Errorf("initial MyGuild.Name = %q, want %q", loaded.MyGuild.Name, "Troet")
	}

	// Update to level 200.
	player := *loaded
	player.Level = 200

	if err := repo.Update(context.Background(), &player); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Read back.
	loaded, err = repo.GetByName("Pigtail", 5)
	if err != nil {
		t.Fatalf("GetByName after update failed: %v", err)
	}
	if loaded.Level != 200 {
		t.Errorf("Level after update = %d, want %d", loaded.Level, 200)
	}

	// Append another history event.
	if err := repo.InsertHistoryEvent(context.Background(), &player); err != nil {
		t.Fatalf("InsertHistoryEvent failed: %v", err)
	}

	historyFile := filepath.Join(tmpDir, "player_history_Pigtail_5.jsonl")
	data, err := os.ReadFile(historyFile)
	if err != nil {
		t.Fatalf("reading history file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 history lines (2 from testdata + 1 new), got %d", len(lines))
	}
	if !strings.Contains(lines[2], `"level":200`) {
		t.Errorf("last history line missing updated level: %s", lines[2])
	}
}

func TestPlayerRepository_roundTripForGuildless(t *testing.T) {
	tmpDir := t.TempDir()
	copyTestdata(t, tmpDir)
	repo := NewPlayerRepository(tmpDir)

	// Pigtail already exists in testdata at level 199 with 2 history entries.
	loaded, err := repo.GetByName("Alice", 5)
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if loaded.Level != 10 {
		t.Errorf("initial Level = %d, want %d", loaded.Level, 199)
	}
	if loaded.Deleted != false {
		t.Errorf("initial deletes status = %v, want %v", loaded.Deleted, false)
	}
	if loaded.MyGuild != nil {
		t.Errorf("initial MyGuild.Name = %q, want <none>", loaded.MyGuild.Name)
	}

	// Update to level 200.
	player := *loaded
	player.Deleted = true

	if err := repo.Update(context.Background(), &player); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Read back.
	loaded, err = repo.GetByName("Alice", 5)
	if err != nil {
		t.Fatalf("GetByName after update failed: %v", err)
	}
	if loaded.Deleted != true {
		t.Errorf("updated Deleted status = %v, want %v", loaded.Deleted, true)
	}
	if loaded.MyGuild != nil {
		t.Errorf("updated MyGuild.Name = %q, want <none>", loaded.MyGuild.Name)
	}

	// Append another history event.
	if err := repo.InsertHistoryEvent(context.Background(), &player); err != nil {
		t.Fatalf("InsertHistoryEvent failed: %v", err)
	}

	historyFile := filepath.Join(tmpDir, "player_history_Alice_5.jsonl")
	data, err := os.ReadFile(historyFile)
	if err != nil {
		t.Fatalf("reading history file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 history line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], `"deleted":true`) {
		t.Errorf("last history line missing updated deleted: %s", lines[0])
	}
}

func TestPlayerRepository_StreamPlayers(t *testing.T) {
	timeout := time.Second * 5
	tmpDir := t.TempDir()
	copyTestdata(t, tmpDir)
	repo := NewPlayerRepository(tmpDir)

	ctx, cancelFunc := context.WithTimeout(t.Context(), timeout)
	defer cancelFunc()
	errorCh := make(chan error, 1)
	ch, err := repo.StreamPlayers(ctx)
	if err != nil {
		t.Fatalf("Unable to stream players: %v", err)
	}
	mainTimeout := time.NewTimer(timeout + time.Second*1)
	defer mainTimeout.Stop()

	var count int
loop:
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				break loop
			}
			count++
		case err := <-errorCh:
			if err != nil {
				t.Fatalf("StreamPlayer errors: %v", err)
			}
		case <-mainTimeout.C:
			t.Fatalf("StreamPlayers does not check for timeout")
		}
	}
	if count != 3 {
		t.Errorf("expected 3 players streamed, got %d", count)
	}
	select {
	case err := <-errorCh:
		if err != nil {
			t.Fatalf("StreamPlayers error: %v", err)
		}
	default:
	}
}

// TestEmbeddedFS ensures the embedded filesystem is readable.
func TestEmbeddedFS(t *testing.T) {
	entries, err := testdataFS.ReadDir("testdata")
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	if len(names) == 0 {
		t.Fatal("embedded testdata directory is empty")
	}
}

// Ensure embed.FS satisfies fs.FS (compile-time check).
var _ fs.FS = testdataFS
