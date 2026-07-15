package domain

type StarLiveInfo struct {
	AppID    string
	SecKey   string
	APIHost  string
	Env      interface{}
	Src      string
	LiveHost string
}

type StarLiveIndexData struct {
	Data map[string]interface{} `json:"data"`
}

type StarLiveBalanceResponse struct {
	Code int                    `json:"code"`
	Data map[string]interface{} `json:"data"`
}
