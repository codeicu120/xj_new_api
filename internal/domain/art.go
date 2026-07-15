package domain

type ArtListingData struct {
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type ArtShowData struct {
	Row map[string]interface{} `json:"row"`
}
