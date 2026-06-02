package pork

import (
	"strings"
	"time"

	"pork-history-updater/internal/application"
	"pork-history-updater/internal/domain"
)

// Compile-time check that Adapter implements the port.
var _ application.CharInfoFetcher = (*adapter)(nil)

type charInfoProvider interface {
	FetchCharInfo(name string, dimension int) (*PlayerData, error)
}

type adapter struct {
	client charInfoProvider
}

// NewAdapter creates a new pork API adapter wrapping the given provider.
func NewAdapter(client charInfoProvider) *adapter {
	return &adapter{client: client}
}

// FetchByNameAsPlayer fetches player info by name and server
// and returns a domain.Player struct.
func (a *adapter) FetchByNameAsPlayer(nickname string, server int) (*domain.Player, error) {
	lastChecked := time.Now()
	charInfo, err := a.client.FetchCharInfo(nickname, server)
	if err != nil || charInfo == nil {
		return nil, err
	}

	id := uint32(charInfo.Character.CharInstance)
	result := &domain.Player{
		CharID:           &id,
		Nickname:         charInfo.Character.Name,
		Server:           server,
		FirstName:        strings.TrimSpace(charInfo.Character.FirstName),
		LastName:         strings.TrimSpace(charInfo.Character.LastName),
		Level:            charInfo.Character.Level,
		Faction:          strings.TrimSpace(charInfo.Character.Side),
		Profession:       strings.TrimSpace(charInfo.Character.Profession),
		ProfessionTitle:  strings.TrimSpace(charInfo.Character.ProfName),
		Gender:           strings.TrimSpace(charInfo.Character.Sex),
		Breed:            strings.TrimSpace(charInfo.Character.Breed),
		DefenderRank:     charInfo.Character.AlienLevel,
		DefenderRankName: strings.TrimSpace(charInfo.Character.RankName),
		LastChanged:      time.Unix(charInfo.LastUpdate, 0),
		LastChecked:      lastChecked,
	}
	if charInfo.Org != nil {
		result.MyGuild = &domain.GuildMembership{
			ID:       charInfo.Org.OrgInstance,
			Name:     charInfo.Org.Name,
			Rank:     charInfo.Org.Rank,
			RankName: strings.TrimSpace(charInfo.Org.RankTitle),
		}
	}

	return result, nil
}
