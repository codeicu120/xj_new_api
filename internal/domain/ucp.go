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

type FeedbackCreateInput struct {
	UID        int
	CID        int
	Content    string
	PayID      int
	PayName    string
	PayAccount string
	AIDs       string
	CreatedAt  int64
	IP         string
	Device     string
	LongIDs    string
	ShortIDs   string
}

type PackagePaymentInput struct {
	PayID     int64
	PayType   int
	Payway    string
	Paycode   string
	NoCheck   int
	ItemName  string
	Amount    int
	UID       int
	PID       string
	CreatedAt int64
	Params    map[string]interface{}
}

type VODOrderCreateInput struct {
	UID       int
	Serial    string
	Name      string
	Coins     int
	Issue     int64
	Period    int
	Support   int
	CreatedAt int64
}

type VODOrderSupportInput struct {
	UID       int
	OrderID   int
	OrderUID  int
	Coins     int
	CreatedAt int64
}

type WithdrawCreateInput struct {
	UID            int
	Username       string
	WDType         int
	WithdrawAmount int
	CoinNum        int
	ConvertAmount  int
	CardName       string
	CardNum        string
	BankName       string
	CardType       int
	GameRate       float64
	CreatedAt      int64
}

type UCPMsgListingData struct {
	Rows     []map[string]interface{} `json:"rows"`
	PageInfo map[string]interface{}   `json:"pageinfo"`
}

type UCPMsgDetailData struct {
	Crow     map[string]interface{}   `json:"crow"`
	CUser    interface{}              `json:"cuser"`
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

type UCPWithdrawIndexData struct {
	Account             map[string]interface{}   `json:"account"`
	CardRows            []map[string]interface{} `json:"cardrows"`
	GoldCoin            int                      `json:"goldcoin"`
	ExRate              int                      `json:"exrate"`
	TopupMin            string                   `json:"topupmin"`
	Coin2RMB            string                   `json:"coin2rmb"`
	Max2RMB             string                   `json:"max2rmb"`
	GameWithdrawMin     int                      `json:"game_withdrawmin"`
	GameWithdrawRate    float64                  `json:"game_withdrawrate"`
	AlipayWithdrawMin   int                      `json:"alipay_withdraw_min"`
	AlipayWithdrawMax   int                      `json:"alipay_withdraw_max"`
	BankcardWithdrawMin int                      `json:"bankcard_withdraw_min"`
	BankcardWithdrawMax int                      `json:"bankcard_withdraw_max"`
}

type UCPWithdrawListingData struct {
	Rows          []map[string]interface{} `json:"rows"`
	WithdrawTotal string                   `json:"withdrawTotal"`
	PageInfo      map[string]interface{}   `json:"pageinfo"`
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
