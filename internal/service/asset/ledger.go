package asset

import (
	"context"
	"errors"

	domainasset "xj_comp/internal/domain/asset"
)

var (
	ErrMissingTransaction = errors.New("asset ledger requires transaction")
	ErrDuplicateRequest   = errors.New("asset idempotency key already applied to a different request")
	ErrInsufficientFunds  = errors.New("asset balance is insufficient")
)

// Transaction is the minimum handle accepted by ledger implementations.
// Real adapters should wrap their concrete *sql.Tx or unit-of-work object and
// return true only while all balance and log writes share that transaction.
type Transaction interface {
	InAssetTransaction() bool
}

// Lock order for real implementations:
//  1. Reserve or load the idempotency row by key using the same transaction.
//  2. Lock owner balance rows in stable order: account, bean, coin.
//  3. Update the balance row and append the matching asset log row together.
//  4. Mark the idempotency row applied with the resulting log/balance snapshot.
//
// Do not call external providers while holding these locks. Platform calls
// should happen before reservation or after commit via an outbox/compensation
// flow owned by the caller.

// CoinLedger changes user or guest coin balances. Implementations must write
// the coin balance and coin log atomically in the supplied transaction.
type CoinLedger interface {
	ChangeCoin(ctx context.Context, tx Transaction, req domainasset.Change) (domainasset.ChangeResult, error)
}

// BeanLedger changes bean balances. Implementations must write the bean balance
// and bean log atomically in the supplied transaction.
type BeanLedger interface {
	ChangeBean(ctx context.Context, tx Transaction, req domainasset.Change) (domainasset.ChangeResult, error)
}

// AccountLedger changes account/RMB-style balances. Implementations must write
// the account balance and account log atomically in the supplied transaction.
type AccountLedger interface {
	ChangeAccount(ctx context.Context, tx Transaction, req domainasset.Change) (domainasset.ChangeResult, error)
}

// IdempotencyStore owns dedupe records for asset mutations. Reserve and
// Complete must be committed or rolled back with the matching ledger change.
type IdempotencyStore interface {
	Reserve(ctx context.Context, tx Transaction, record domainasset.IdempotencyRecord) (domainasset.IdempotencyRecord, error)
	Complete(ctx context.Context, tx Transaction, key string, result domainasset.ChangeResult) error
	Fail(ctx context.Context, tx Transaction, key string, message string) error
	Lookup(ctx context.Context, tx Transaction, key string) (domainasset.IdempotencyRecord, bool, error)
}

func RequireTransaction(tx Transaction) error {
	if tx == nil || !tx.InAssetTransaction() {
		return ErrMissingTransaction
	}
	return nil
}
