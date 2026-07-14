package domain

type VODListingData struct {
	Now          int64                    `json:"now"`
	Action       string                   `json:"action"`
	SampleParams string                   `json:"sample_params"`
	Params       map[string]string        `json:"params"`
	VODRows      []map[string]interface{} `json:"vodrows"`
	PageInfo     map[string]interface{}   `json:"pageinfo"`
	Orders       []map[string]interface{} `json:"orders"`
	Categories   []map[string]interface{} `json:"categories"`
	Areas        []map[string]interface{} `json:"areas"`
	Years        []map[string]interface{} `json:"years"`
	Definitions  []map[string]interface{} `json:"definitions"`
	Durations    []map[string]interface{} `json:"durations"`
	FreeTypes    []map[string]interface{} `json:"freetypes"`
	Mosaics      []map[string]interface{} `json:"mosaics"`
	LangVoices   []map[string]interface{} `json:"langvoices"`
}

type VODLikeRowsData struct {
	LikeRows []map[string]interface{} `json:"likerows"`
}

type VODShowData struct {
	VODRow      map[string]interface{}   `json:"vodrow"`
	Categories  []map[string]interface{} `json:"categories"`
	SimilarRows []map[string]interface{} `json:"similarrows"`
	LikeRows    []map[string]interface{} `json:"likerows"`
}
