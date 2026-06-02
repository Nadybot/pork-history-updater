package pork

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new HTTP client for the Anarchy-Online people API.
func NewClient() *client {
	return &client{
		httpClient: &http.Client{},
		baseURL:    "https://people.anarchy-online.com",
	}
}

// FetchCharInfo fetches character and guild data for the given name and dimension.
// Returns nil when the character is not found.
func (p *client) FetchCharInfo(name string, dimension int) (*PlayerData, error) {
	url := fmt.Sprintf("%s/character/bio/d/%d/name/%s/bio.xml?data_type=json", p.baseURL, dimension, name)
	data, err := p.fetchURL(url)
	if err != nil {
		return nil, err
	}
	return parsePlayerData(data)
}

// fetchURL performs a GET request and returns the response body.
func (p *client) fetchURL(url string) ([]byte, error) {
	resp, err := p.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// parsePlayerData parses the Anarchy-Online people API JSON response.
// The response is a JSON array: [character, org, last_update].
// An empty array indicates the character was not found.
func parsePlayerData(data []byte) (*PlayerData, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if len(raw) < 2 {
		return nil, nil
	}

	result := &PlayerData{}

	if err := json.Unmarshal(raw[0], &result.Character); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(raw[1], &result.Org); err != nil {
		// Org could be missing
		result.Org = nil
	}

	if len(raw) > 2 {
		var lastUpdate string
		json.Unmarshal(raw[2], &lastUpdate)
		t, err := time.Parse("2006/01/02 15:04:05", lastUpdate)
		if err == nil {
			result.LastUpdate = t.Unix()
		}
	}

	return result, nil
}
