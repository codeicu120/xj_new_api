# Asset Ledger Carver Draft

This is a narrow contract draft for future PHP success-path migrations that
touch coins, beans, or account balances. It does not connect any current
`ucp`, `vod`, `minivod`, `onego`, `invite`, `game`, or `respond` handler.

## Scope

- Define `CoinLedger`, `BeanLedger`, `AccountLedger`, and `IdempotencyStore`.
- Keep all real balance mutations behind transaction-aware interfaces.
- Provide fake tests that document transaction and idempotency expectations.
- Do not add route wiring, concrete DB repositories, or provider callbacks.

## Required Real Implementation Rules

1. The caller must open a database transaction before invoking a ledger.
2. The same transaction must reserve idempotency, update the balance, append the
   asset log, and mark idempotency as applied.
3. Balance and log writes must not be split across repositories without sharing
   the same transaction handle.
4. External platform calls must not be performed while asset rows are locked.
5. Lock order must stay stable: idempotency row, account balance, bean balance,
   coin balance, then append-only log rows.

## Non-goals

- No real MySQL repository.
- No automatic migration of existing success charge paths.
- No route or handler changes.
- No production config, key, or provider integration.
