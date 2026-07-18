package domain

type InviteBindInput struct {
	UID        int
	InviterUID int
	InviteCode string
	Now        int64
	NoReward   bool
	Bonus      map[string]int
	Groups     []map[string]interface{}
}

type MiniVODThrowCoinInput struct {
	UID       int
	AuthorUID int
	VODID     int
	CoinNum   int
	Now       int64
}
