package domain

type CommentListingData struct {
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type CommentCreateInput struct {
	RootID   int
	ParentID int
	Left     int
	Right    int
	Depth    int
	VODID    int
	UID      int
	SID      string
	Content  string
	AddTime  int64
	IP       string
	ShowType int
}
