package domain

type ExploreIndexData struct {
	TabRows  []map[string]interface{} `json:"tabrows"`
	DayRows  []map[string]interface{} `json:"dayrows"`
	SignData map[string]interface{}   `json:"signdata"`
}
