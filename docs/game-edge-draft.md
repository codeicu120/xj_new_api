# Game edge draft

## Scope

- PHP sources checked:
  - `/Users/canavs/xjProj/XJBackend/api/src/c/api/game/wali.php`
  - `/Users/canavs/xjProj/XJBackend/api/src/c/api/game/lottery.php`
- Go endpoints checked:
  - `/game/wali/withdraw`
  - `/game/lottery/topup`
  - `/game/lottery/withdraw`
  - `/game/lottery/balance`

## PHP behavior notes

- `wali->withdraw` order: require login, parse `amount` with two decimals, reject empty or non-positive amount with `金额输入不正确`, then call `transferV3`, save `game_order`, and on successful platform status add coins through `coinLog`.
- `lottery->withdraw` order: require login, convert `amount` to cents, reject empty or non-positive cents with `金额输入不正确`, then call `/api/player/transferOut`, save `game_order`, and on successful platform status add coins through `coinLog`.
- `lottery->balance` order: require login, call `/api/player/balance`, fail with `查询余额失败` if no result, then return nested `data` containing `status`, `balance`, `transferable`, and `currency`. `totalMoney` and `freeMoney` are cents formatted as two-decimal strings.

## Go draft status

- Confirmed safe precheck branches remain implemented for `wali` and `lottery` topup/withdraw: unauthenticated user, topup minimum amount, topup insufficient quota, and withdraw invalid amount.
- Added/confirmed lottery balance read-only branch through an interface client and fake tests. It does not write coins, orders, or local game assets, and does not call a real external platform in tests.
- `/game/lottery/balance` is wired to `GameHandler.LotteryBalance` instead of the generic high-risk placeholder.

## Remaining blockers

- `wali->withdraw` success path is blocked because it calls external `transferV3`, writes `game_order`, and may increase user coins.
- `lottery->topup` and `lottery->withdraw` success paths are blocked because they perform external transfer calls, write orders, and mutate coins with compensation logic.
- Those paths need transaction, idempotency/order, fake repository/client contracts, grey release, and rollback design before migration.
