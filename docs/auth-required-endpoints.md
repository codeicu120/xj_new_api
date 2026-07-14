# 登录接口迁移清单

旧 PHP host: `http://127.0.0.1:18765`

测试 token:

```text
xxx_api_auth=3235306637393062613731656332623964333835356634323464623232353965
```

说明：

- “登录接口”指旧 PHP 中会读取 `context->get('user')` 且要求 `uid > 0` 的接口。
- 当前阶段优先迁移不需要验证码、不涉及支付、不涉及用户资产写入、不调用外部平台的接口。
- Go 侧兼容旧 PHP 的 `x-cookie-auth` header 和 `xxx_api_auth` cookie；token 是 32 字节 sid 的 hex 编码。

## 已迁移

| 接口 | PHP handler | Go 状态 | 对比说明 |
| --- | --- | --- | --- |
| `/ucp/myaff` | `c.api.ucp.index->myaff` | 本轮完成 | 使用 `x-cookie-auth` 登录 token，读取推荐用户列表，用户行和分页结构与 PHP 一致。 |

## 暂缓

| 接口 | 原因 |
| --- | --- |
| `/ucp/index`、`/ucp/affcenter` | 依赖权限、观看/下载计数、金币、金豆、签到、用户组，适合作为下一批中等复杂度接口。 |
| `/ucp/feedback` POST、`/ucp/task/*` | 涉及写库、奖励或状态变更，需要单独测试和回滚策略。 |
| `/ucp/vippkg/*`、`/ucp/coinpkg/*`、`/ucp/beanpkg/*`、`/ucp/payment/*`、`/payment/*` | 会员、金币、金豆、支付相关，涉及资产和交易。 |
| `/game/wali/topup`、`/game/wali/withdraw`、`/game/wali/balance`、`/game/wali/enter`、`/game/lottery/*` | 游戏资产、余额或外部平台调用。 |
