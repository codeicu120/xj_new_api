# User Forgot Step2 Edge Draft

## Scope

- PHP sources:
  - `/Users/canavs/xjProj/XJBackend/api/src/c/api/user.php::forgot`
  - `/Users/canavs/xjProj/XJBackend/api/src/c/apiv2/user.php::forgot`
- Go scope:
  - `/forgot`
  - `/v2/forgot`
- Migrated only read-only step2 preflight branches. Step3 password update, session mutation, verification deletion, and all writes remain outside this draft.

## PHP Contract Notes

- v1 accepts mobile reset only.
- v2 prefers the mobile branch when `mobi` is non-empty; email branch is used only when `mobi` is empty and `email` is non-empty.
- `mobiprefix` defaults to `+86`; normalized mobile key is `trim(mobiprefix + "." + mobi, "+")`, for example `86.13800138000`.
- For `step1`, `step2`, and `step3`, PHP reads the user first:
  - missing mobile user: `retcode=-1 errmsg=输入的手机号码不存在`
  - missing email user: `retcode=-1 errmsg=输入的邮箱不存在`
- For `step2`, PHP checks `keylimit->get(key, 600)` unless PHP base env is `test`:
  - mobile key: `sms.<normalizedMobi>.<smscode>`
  - email key: `email.<email>.<verifyCode>`
  - email `verifyCode` is `emailcode`, or `smscode` when `emailcode` is empty.
- Verification failure messages:
  - mobile: `手机验证码不正确`
  - email: `邮箱验证码不正确`
- Valid step2 advances with `retcode=0 errmsg=step2->step3`.

## Go Draft Status

- Added service-level step2 read-only lookup and keylimit verification through the existing fakeable `AuthEdgePolicyStore`.
- No password update, session write, keylimit delete, DB update, Redis write, SMS, or email integration was added.
- Main migration tables are intentionally left for the harness/mainline update.
