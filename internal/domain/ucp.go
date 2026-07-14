package domain

type UCPMyAffData struct {
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type UCPRollTitleData struct {
	Messages []map[string]interface{} `json:"messages"`
}

type UCPAffCenterData struct {
	User  map[string]interface{} `json:"user"`
	UInfo map[string]interface{} `json:"uinfo"`
}

type UCPIndexData struct {
	User   map[string]interface{}   `json:"user"`
	UInfo  map[string]interface{}   `json:"uinfo"`
	Signed int                      `json:"signed"`
	Groups []map[string]interface{} `json:"groups,omitempty"`
}

type UCPFeedbackData struct {
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type UCPFeedbackIndexData struct {
	PayRows []map[string]interface{} `json:"payrows"`
}

type UCPFeedbackDetailData struct {
	Row     map[string]interface{} `json:"row"`
	PicURLs interface{}            `json:"picurls"`
}

type UCPMsgListingData struct {
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type UCPPaymentListingData struct {
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type UCPSafePayLogData struct {
	PayRows []map[string]interface{} `json:"payrows"`
}

type UCPAccountIndexData struct {
	Account  map[string]interface{}   `json:"account"`
	GoldCoin int                      `json:"goldcoin"`
	ExRate   int                      `json:"exrate"`
	Coin2RMB string                   `json:"coin2rmb"`
	Max2RMB  string                   `json:"max2rmb"`
	LogRows  []map[string]interface{} `json:"logrows"`
}

type UCPBalanceLogData struct {
	LogRows  []map[string]interface{} `json:"logrows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type UCPCoinLogIndexData struct {
	Account  map[string]interface{}   `json:"account"`
	GoldCoin int                      `json:"goldcoin"`
	ExRate   int                      `json:"exrate"`
	LogRows  []map[string]interface{} `json:"logrows"`
}

type UCPCoinLogListingData struct {
	LogRows  []map[string]interface{} `json:"logrows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type UCPCoinLogBonusData struct {
	LogRows  []map[string]interface{} `json:"logrows"`
	AddInfo  map[string]interface{}   `json:"addinfo"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}
