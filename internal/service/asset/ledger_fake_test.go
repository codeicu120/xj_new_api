package asset

import (
	"context"
	"errors"
	"testing"
	"time"

	domainasset "xj_comp/internal/domain/asset"
)

type fakeTx struct {
	active bool
}

func (tx fakeTx) InAssetTransaction() bool {
	return tx.active
}

type fakeCoinLedger struct {
	balance int64
	nextLog int64
}

func (l *fakeCoinLedger) ChangeCoin(_ context.Context, tx Transaction, req domainasset.Change) (domainasset.ChangeResult, error) {
	if err := RequireTransaction(tx); err != nil {
		return domainasset.ChangeResult{}, err
	}
	before := l.balance
	after := before
	switch req.Direction {
	case domainasset.DirectionCredit:
		after += req.Amount
	case domainasset.DirectionDebit:
		if before < req.Amount {
			return domainasset.ChangeResult{}, ErrInsufficientFunds
		}
		after -= req.Amount
	default:
		return domainasset.ChangeResult{}, errors.New("unsupported direction")
	}
	l.balance = after
	l.nextLog++
	return domainasset.ChangeResult{
		Kind:           domainasset.AssetKindCoin,
		Owner:          req.Owner,
		IdempotencyKey: req.IdempotencyKey,
		Applied:        true,
		LogID:          l.nextLog,
		BalanceBefore:  before,
		BalanceAfter:   after,
	}, nil
}

type fakeIdempotencyStore struct {
	records map[string]domainasset.IdempotencyRecord
	now     func() time.Time
}

func newFakeIdempotencyStore() *fakeIdempotencyStore {
	return &fakeIdempotencyStore{
		records: map[string]domainasset.IdempotencyRecord{},
		now:     time.Now,
	}
}

func (s *fakeIdempotencyStore) Reserve(_ context.Context, tx Transaction, record domainasset.IdempotencyRecord) (domainasset.IdempotencyRecord, error) {
	if err := RequireTransaction(tx); err != nil {
		return domainasset.IdempotencyRecord{}, err
	}
	existing, ok := s.records[record.Key]
	if ok {
		if existing.RequestHash != record.RequestHash {
			return existing, ErrDuplicateRequest
		}
		return existing, nil
	}
	record.Status = domainasset.IdempotencyReserved
	record.CreatedAt = s.now()
	record.UpdatedAt = record.CreatedAt
	s.records[record.Key] = record
	return record, nil
}

func (s *fakeIdempotencyStore) Complete(_ context.Context, tx Transaction, key string, result domainasset.ChangeResult) error {
	if err := RequireTransaction(tx); err != nil {
		return err
	}
	record := s.records[key]
	record.Status = domainasset.IdempotencyApplied
	record.Result = &result
	record.UpdatedAt = s.now()
	s.records[key] = record
	return nil
}

func (s *fakeIdempotencyStore) Fail(_ context.Context, tx Transaction, key string, message string) error {
	if err := RequireTransaction(tx); err != nil {
		return err
	}
	record := s.records[key]
	record.Status = domainasset.IdempotencyFailed
	record.ErrorMessage = message
	record.UpdatedAt = s.now()
	s.records[key] = record
	return nil
}

func (s *fakeIdempotencyStore) Lookup(_ context.Context, tx Transaction, key string) (domainasset.IdempotencyRecord, bool, error) {
	if err := RequireTransaction(tx); err != nil {
		return domainasset.IdempotencyRecord{}, false, err
	}
	record, ok := s.records[key]
	return record, ok, nil
}

func TestFakeCoinLedgerRequiresTransaction(t *testing.T) {
	ledger := &fakeCoinLedger{balance: 100}
	req := domainasset.Change{
		Owner:          domainasset.Owner{UID: 42},
		Kind:           domainasset.AssetKindCoin,
		Direction:      domainasset.DirectionDebit,
		Amount:         10,
		IdempotencyKey: "vod:reqplay:42:99",
	}

	if _, err := ledger.ChangeCoin(context.Background(), nil, req); !errors.Is(err, ErrMissingTransaction) {
		t.Fatalf("ChangeCoin without tx error = %v, want %v", err, ErrMissingTransaction)
	}
	if _, err := ledger.ChangeCoin(context.Background(), fakeTx{}, req); !errors.Is(err, ErrMissingTransaction) {
		t.Fatalf("ChangeCoin inactive tx error = %v, want %v", err, ErrMissingTransaction)
	}
}

func TestFakeCoinLedgerDebitWritesBalanceAndLogSnapshot(t *testing.T) {
	ledger := &fakeCoinLedger{balance: 100}
	req := domainasset.Change{
		Owner:          domainasset.Owner{UID: 42},
		Kind:           domainasset.AssetKindCoin,
		Direction:      domainasset.DirectionDebit,
		Amount:         30,
		Business:       "vod",
		ReferenceType:  "vodid",
		ReferenceID:    "99",
		IdempotencyKey: "vod:reqplay:42:99",
		LogType:        101,
	}

	got, err := ledger.ChangeCoin(context.Background(), fakeTx{active: true}, req)
	if err != nil {
		t.Fatalf("ChangeCoin error = %v", err)
	}
	if got.BalanceBefore != 100 || got.BalanceAfter != 70 || got.LogID == 0 || !got.Applied {
		t.Fatalf("ChangeCoin result = %+v, want applied log with 100 -> 70", got)
	}
}

func TestFakeIdempotencyStoreDuplicateRequest(t *testing.T) {
	ctx := context.Background()
	tx := fakeTx{active: true}
	store := newFakeIdempotencyStore()

	first := domainasset.IdempotencyRecord{Key: "asset:v1:order-1", RequestHash: "hash-a"}
	if _, err := store.Reserve(ctx, tx, first); err != nil {
		t.Fatalf("Reserve first error = %v", err)
	}
	if _, err := store.Reserve(ctx, tx, first); err != nil {
		t.Fatalf("Reserve same request error = %v", err)
	}

	second := domainasset.IdempotencyRecord{Key: "asset:v1:order-1", RequestHash: "hash-b"}
	if _, err := store.Reserve(ctx, tx, second); !errors.Is(err, ErrDuplicateRequest) {
		t.Fatalf("Reserve different request error = %v, want %v", err, ErrDuplicateRequest)
	}
}

func TestFakeIdempotencyStoreCompleteStoresResult(t *testing.T) {
	ctx := context.Background()
	tx := fakeTx{active: true}
	store := newFakeIdempotencyStore()
	key := "asset:v1:order-2"
	result := domainasset.ChangeResult{
		Kind:          domainasset.AssetKindCoin,
		Owner:         domainasset.Owner{UID: 42},
		Applied:       true,
		LogID:         7,
		BalanceBefore: 100,
		BalanceAfter:  120,
	}

	if _, err := store.Reserve(ctx, tx, domainasset.IdempotencyRecord{Key: key, RequestHash: "hash"}); err != nil {
		t.Fatalf("Reserve error = %v", err)
	}
	if err := store.Complete(ctx, tx, key, result); err != nil {
		t.Fatalf("Complete error = %v", err)
	}
	got, ok, err := store.Lookup(ctx, tx, key)
	if err != nil {
		t.Fatalf("Lookup error = %v", err)
	}
	if !ok || got.Status != domainasset.IdempotencyApplied || got.Result == nil || got.Result.LogID != 7 {
		t.Fatalf("Lookup = %+v, ok=%v; want applied result", got, ok)
	}
}
