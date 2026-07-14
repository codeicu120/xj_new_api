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
