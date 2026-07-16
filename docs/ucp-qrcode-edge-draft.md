# UCP qrcode edge migration draft

Scope: `/ucp/task/qrcode`, `/ucp/task/qrcodeSave`, and `/ucp/taskbox/qrcode`.

## PHP behavior

- `src/c/api/ucp/task.php::qrcode`
  - Order: init `maintain.calldata` and cache, read user and `pid`, sanitize non-alnum `pid`, require login, set `keylimit` key `task.qrcode.{uid}.{Ymd}`, build `global.qrcode.link`, replace `{inviteUrl}` and `{inviteCode}`, generate QR PNG, compose it on `data/qrbg.png`, overlay `data/logo.png`, return raw PNG.
  - Response: `Content-Type: image/png`, `X-Served-By: newbie`, raw image body on success; JSON `retcode=-9999 errmsg=您还没有登录` when not logged in.
  - Blocked boundary: success path writes `keylimit` before image generation. That key is consumed by sign reward logic, so Go should not take over without the keylimit write/rollback contract.

- `src/c/api/ucp/task.php::qrcodeSave`
  - Order: require login, count today `COINTYPE_SAVE_QRCODE`, return already-saved error if count is positive, read `max.goldcoin.saveqrcode.num`, begin transaction, write `coinlog`, commit on success and return `taskdone`, otherwise rollback, then OK message.
  - Response: legacy JSON. Already migrated safe branches: not logged in and already saved today.
  - Blocked boundary: success path writes coinlog and user coins. No reward or asset mutation was migrated in this pass.

- `src/c/api/ucp/taskbox.php::qrcode`
  - Order: init `maintain.calldata` and cache, read user and `pid`, sanitize non-alnum `pid`, require login, build `taskbox.qrcode.link`, replace `{inviteUrl}` and `{inviteCode}`, generate QR PNG, compose it on `data/qrbg.png`, return raw PNG.
  - Response: `Content-Type: image/png`, `X-Served-By: newbie`, raw image body on success; JSON `retcode=-9999 errmsg=您还没有登录` when not logged in.
  - Safe branch: PHP keylimit writes are commented out and no reward, coinlog, or persisted image file is written. Go migrates this branch using a fakeable renderer and in-memory PNG response.

## Go status

- Implemented `/ucp/taskbox/qrcode` via `UCPHandler.TaskboxQRCode` and `Service.TaskboxQRCode`.
- The renderer is injectable through `WithQRCodeRenderer`; service tests use a fake renderer, so no production assets or files are written during tests.
- Compatibility note: Go returns an in-memory 400x400 PNG with 5px padding and does not read PHP `data/qrbg.png`; this avoids asset-file dependency but may differ from a customized legacy background image.
- `/ucp/task/qrcode` remains blocked by keylimit write semantics.
- `/ucp/task/qrcodeSave` remains blocked at the reward/coinlog transaction success path.

## Validation

- `go test ./internal/service/ucp` passed during implementation.
- `go test ./internal/service/ucp ./internal/handler ./internal/server` passed after gofmt.
