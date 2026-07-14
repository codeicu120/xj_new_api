package domain

type AmazingCategoriesData struct {
	Rows []map[string]interface{} `json:"rows"`
}

type AmazingListingData struct {
	Now      int64                    `json:"now"`
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}
