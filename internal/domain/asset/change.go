package asset

import "time"

// AssetKind identifies the balance table/log family that a ledger
// implementation owns. It is intentionally small until real repositories are
// wired in from each migrated success path.
type AssetKind string

const (
	AssetKindCoin    AssetKind = "coin"
	AssetKindBean    AssetKind = "bean"
	AssetKindAccount AssetKind = "account"
)

type Direction string

const (
	DirectionCredit Direction = "credit"
	DirectionDebit  Direction = "debit"
)

// Owner identifies the asset holder. UID is the primary PHP user identifier.
// GuestID is reserved for legacy guest coin flows that cannot be represented by
// UID alone.
type Owner struct {
	UID     int64
	GuestID string
}

// Change is the common request shape for coin, bean and account ledgers.
// Amount is always positive; Direction carries add/subtract semantics.
type Change struct {
	Owner          Owner
	Kind           AssetKind
	Direction      Direction
	Amount         int64
	Business       string
	ReferenceType  string
	ReferenceID    string
	IdempotencyKey string
	LogType        int
	Remark         string
	Metadata       map[string]string
	OccurredAt     time.Time
}

type ChangeResult struct {
	Kind           AssetKind
	Owner          Owner
	IdempotencyKey string
	Applied        bool
	Duplicate      bool
	LogID          int64
	BalanceBefore  int64
	BalanceAfter   int64
}

type IdempotencyStatus string

const (
	IdempotencyReserved IdempotencyStatus = "reserved"
	IdempotencyApplied  IdempotencyStatus = "applied"
	IdempotencyFailed   IdempotencyStatus = "failed"
)

type IdempotencyRecord struct {
	Key          string
	RequestHash  string
	Status       IdempotencyStatus
	Result       *ChangeResult
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
