package domain

type Guild struct {
	GuildID     int
	GuildName   string
	Faction     string
	Server      int
	Deleted     bool
	LastChecked int
	LastChanged int
}
