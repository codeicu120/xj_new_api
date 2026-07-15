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

type SearchIndexData struct {
	HotWords    interface{}              `json:"hotwords"`
	HotRows     []map[string]interface{} `json:"hotrows"`
	YouMayLikes interface{}              `json:"you_may_likes"`
}

type SearchListData struct {
	VODRows  []map[string]interface{} `json:"vodrows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type MiniSearchListData struct {
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type MiniVODListingData struct {
	Now          int64                    `json:"now"`
	Action       string                   `json:"action"`
	SampleParams string                   `json:"sample_params"`
	Params       map[string]string        `json:"params"`
	Rows         []map[string]interface{} `json:"rows"`
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

type MiniVODShowData struct {
	VODRow      map[string]interface{}   `json:"vodrow"`
	Categories  []map[string]interface{} `json:"categories"`
	SimilarRows []map[string]interface{} `json:"similarrows"`
	LikeRows    []map[string]interface{} `json:"likerows"`
	VODUser     map[string]interface{}   `json:"voduser"`
}

type MiniAuthorListingData struct {
	Now      int64                    `json:"now"`
	UserRow  map[string]interface{}   `json:"userrow"`
	VODRows  []map[string]interface{} `json:"vodrows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
	Orders   []map[string]interface{} `json:"orders"`
}

type SpecialListingData struct {
	Rows         []map[string]interface{} `json:"rows"`
	PageInfo     map[string]interface{}   `json:"pageinfo"`
	SampleParams string                   `json:"sample_params"`
	Params       map[string]interface{}   `json:"params"`
	ActorRows    []map[string]interface{} `json:"actorrows"`
}

type SpecialDetailData struct {
	Row     map[string]interface{}   `json:"row"`
	VODRows []map[string]interface{} `json:"vodrows"`
}
