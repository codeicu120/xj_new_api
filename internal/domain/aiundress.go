package domain

type AIUndressListingData struct {
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type AIUndressExternalData struct {
	Data interface{} `json:"data"`
}

type AIUndressResourceListInput struct {
	Module   string
	TypeID   string
	PageSize int
	Current  string
}
