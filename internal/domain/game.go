package domain

type GamePlatformsData struct {
	Data []map[string]interface{} `json:"data"`
}

type GameCategoriesData struct {
	Data []map[string]interface{} `json:"data"`
}

type GamesData struct {
	Data []map[string]interface{} `json:"data"`
}

type GameBroadcastsData struct {
	Data []string `json:"data"`
}

type GameWaliData struct {
	Data map[string]interface{} `json:"data"`
}

type OneGoData struct {
	Data interface{} `json:"data"`
}

type OneGoBetInput struct {
	UID      int
	Period   string
	RoomID   int
	Quantity int
	BetCoins int
	Now      int64
}

type OneGoBetResult struct {
	BetNo      []int `json:"bet_no"`
	TotalBetNo []int `json:"total_bet_no"`
}
