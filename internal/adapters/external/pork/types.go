package pork

type CharacterInfo struct {
	Name          string  `json:"NAME"`
	FirstName     string  `json:"FIRSTNAME"`
	LastName      string  `json:"LASTNAME"`
	Sex           string  `json:"SEX"`
	Breed         string  `json:"BREED"`
	Profession    string  `json:"PROF"`
	ProfName      string  `json:"PROFNAME"`
	Level         int     `json:"LEVELX"`
	AlienLevel    int     `json:"ALIENLEVEL"`
	Side          string  `json:"SIDE"`
	PvPRating     int     `json:"PVPRATING"`
	PvPTitle      *string `json:"PVPTITLE"` // null allowed
	RankName      string  `json:"RANK_name"`
	HeadID        int     `json:"HEADID"`
	CharInstance  int64   `json:"CHAR_INSTANCE"`
	CharDimension int     `json:"CHAR_DIMENSION"`
}

type OrgInfo struct {
	Name         string `json:"NAME"`
	Rank         int    `json:"RANK"`
	RankTitle    string `json:"RANK_TITLE"`
	OrgInstance  int    `json:"ORG_INSTANCE"`
	OrgDimension int    `json:"ORG_DIMENSION"`
}

type PlayerData struct {
	Character  CharacterInfo
	Org        *OrgInfo // optional
	LastUpdate int64
}
