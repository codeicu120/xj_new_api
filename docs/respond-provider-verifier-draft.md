# Respond Provider Verifier Draft

## pay7

- PHP source: `/Users/canavs/xjProj/XJBackend/api/src/c/respond/pay7.php` and `/Users/canavs/xjProj/XJBackend/api/src/m/payment/pay7.php`.
- Callback input: form/query fields accepted through `context->input()`.
- Signature: sort fields by key, skip `sign`, `sign_type`, empty values and arrays, join as `key=value` with `&`, append provider secret, then MD5 lowercase.
- Success status: `trade_status=TRADE_SUCCESS`.
- Extracted fields: `out_trade_no` as payment id, `trade_no` as third-party transaction id, `money` as RMB amount converted to cents.
- Go config gate: optional `RESPOND_PAY7_SECRET`; accounting remains disabled unless `RESPOND_ACCOUNTING_ENABLED=true`.
- Current behavior: verification can pass, but `/respond/pay7` still returns `echoErr` (`failed`) while accounting is disabled. It does not update `trade_payments` and does not credit accounts.
