package domain

type AIUndressListingData struct {
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}
