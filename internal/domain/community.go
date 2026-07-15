package domain

type CommunityListingData struct {
	Now          int64                    `json:"now"`
	Action       string                   `json:"action"`
	SampleParams string                   `json:"sample_params"`
	Params       map[string]string        `json:"params"`
	Rows         []map[string]interface{} `json:"rows"`
	PageInfo     map[string]interface{}   `json:"pageinfo"`
}

type CommunityCommentListingData struct {
	Now          int64                    `json:"now"`
	SampleParams string                   `json:"sample_params"`
	Params       map[string]interface{}   `json:"params"`
	Rows         []map[string]interface{} `json:"rows"`
	PageInfo     map[string]interface{}   `json:"pageinfo"`
}
