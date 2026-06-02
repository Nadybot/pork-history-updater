package pork

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:embed testdata/valid_character_in_org.json
var validCharacterInOrg []byte

//go:embed testdata/valid_character_without_org.json
var validCharacterWithoutOrg []byte

//go:embed testdata/character_invalid.json
var invalidCharacter []byte

//go:embed testdata/character_empty.json
var emptyCharacter []byte

//go:embed testdata/character_too_short.json
var tooShortCharacter []byte

// TestParsePlayerData_full verifies parsing of a complete PORK response
// containing character data, org data and a last-update timestamp.
func TestParsePlayerData_full(t *testing.T) {
	pd, err := parsePlayerData(validCharacterInOrg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pd == nil {
		t.Fatal("expected PlayerData, got nil")
	}

	// Character assertions
	if pd.Character.Name != "Pigtail" {
		t.Errorf("Character.Name = %q, want %q", pd.Character.Name, "TestChar")
	}
	if pd.Character.Level != 200 {
		t.Errorf("Character.Level = %d, want %d", pd.Character.Level, 200)
	}
	if pd.Character.CharInstance != 1178213 {
		t.Errorf("Character.CharInstance = %d, want %d", pd.Character.CharInstance, 12345)
	}

	// Org assertions
	if pd.Org == nil {
		t.Fatal("expected Org, got nil")
	}
	if pd.Org.Name != "Troet" {
		t.Errorf("Org.Name = %q, want %q", pd.Org.Name, "TestGuild")
	}
	if pd.Org.Rank != 1 {
		t.Errorf("Org.Rank = %d, want %d", pd.Org.Rank, 5)
	}

	// Timestamp assertion
	if pd.LastUpdate == 0 {
		t.Error("expected LastUpdate to be parsed, got 0")
	}
}

// TestParsePlayerData_noOrg verifies that a missing org block (null as second
// element) is handled gracefully (Org == nil).
func TestParsePlayerData_noOrg(t *testing.T) {
	pd, err := parsePlayerData(validCharacterWithoutOrg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pd == nil {
		t.Fatal("expected PlayerData, got nil")
	}
	if pd.Org != nil {
		t.Errorf("expected Org=nil, got %+v", pd.Org)
	}
}

// TestParsePlayerData_tooShort verifies that responses with fewer than 2
// elements return nil without error (character not found or empty).
func TestParsePlayerData_tooShort(t *testing.T) {
	for _, json := range [][]byte{emptyCharacter, tooShortCharacter} {
		pd, err := parsePlayerData(json)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", json, err)
		}
		if pd != nil {
			t.Errorf("expected nil for %q, got %+v", json, pd)
		}
	}
}

func TestParsePlayerData_invalidCharacter(t *testing.T) {
	pd, err := parsePlayerData(invalidCharacter)
	if err != nil {
		t.Fatalf("expected no error for invalid characters, got %v", err)
	}
	if pd != nil {
		t.Fatalf("expected nil for invalid characters, got %v", pd)
	}
}

// TestParsePlayerData_invalidJSON verifies that malformed JSON returns an error.
func TestParsePlayerData_invalidJSON(t *testing.T) {
	_, err := parsePlayerData([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestClientFetchCharInfo_success demonstrates how httptest.Server replaces
// the real PORK endpoint. No external HTTP calls are made.
func TestClientFetchCharInfo_success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validCharacterInOrg))
	}))
	defer ts.Close()

	c := &client{
		httpClient: http.DefaultClient,
		baseURL:    ts.URL,
	}

	pd, err := c.FetchCharInfo("ServerChar", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pd == nil {
		t.Fatal("expected PlayerData, got nil")
	}
	if pd.Character.Name != "Pigtail" {
		t.Errorf("Character.Name = %q, want %q", pd.Character.Name, "ServerChar")
	}
	if pd.Org == nil || pd.Org.Name != "Troet" {
		t.Errorf("Org.Name = %q, want %q", pd.Org.Name, "ServerGuild")
	}
}

// TestClientFetchCharInfo_notFound verifies that non-200 status codes are
// surfaced as errors.
func TestClientFetchCharInfo_notFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := &client{
		httpClient: http.DefaultClient,
		baseURL:    ts.URL,
	}

	_, err := c.FetchCharInfo("Unknown", 5)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

// TestClientFetchCharInfo_invalidJSON verifies that invalid JSON from the
// server is propagated as an error.
func TestClientFetchCharInfo_invalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid`))
	}))
	defer ts.Close()

	c := &client{
		httpClient: http.DefaultClient,
		baseURL:    ts.URL,
	}

	_, err := c.FetchCharInfo("BadJson", 5)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
