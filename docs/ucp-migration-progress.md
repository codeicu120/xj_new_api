# UCP 个人中心迁移进度

## PHP 核对范围

- PHP 路由范围：`/Users/canavs/xjProj/XJBackend/api/api.php` 中 `/ucp/*` 动态路由。
- PHP controller 范围：`/Users/canavs/xjProj/XJBackend/api/src/c/api/ucp/*.php`。
- 核对口径：只统计可通过路由访问的业务 action；`vodorder::error()` 是 public helper，不计入业务接口。

## 总体进度

| 项目 | 数量 |
| --- | ---: |
| PHP UCP 业务 action 总数 | 70 |
| 已完整迁移 | 70 |
| 未完全迁移 | 0 |
| 完全未注册到 Go | 0 |

## 未完全迁移 action

暂无。按当前 PHP `/ucp/*` 可达业务 action 核对，Go 侧均已注册并迁移主要旧行为。

## 风险归类

- 账号资料类：邮件发送使用 SMTP 外部依赖，已通过 `EmailSender` 隔离并补 fake 测试。
- 资产交易类：会员升级、提现冻结、金币/余额兑换、VIP/金币/金豆套餐下单或兑换已迁移，后续灰度需重点观察余额、冻结和订单状态。
- 求片类：求片金币扣减、助力写入、冻结金币统计已迁移，事务边界集中在 repository。

## 最新补充迁移

- `/ucp/user/profile`：已接管成功写入分支，更新当前用户 `gender/nickname`，返回 `资料设置成功`。
- `/ucp/user/sendemail`：已接管 SMTP 发送成功分支，成功后写入 `bindemail.$email.Ymd` 与 `email.$email.$code` 两条 keylimit。
- `/ucp/vippkg/placeorder`、`/ucp/coinpkg/placeorder`、`/ucp/beanpkg/placeorder`：已接管支付下单成功分支，写入 `trade_payments`，保留 `paytype/payway/paycode/pid/params` 兼容形态，随机支付通道会先解析为真实 paycode。
- `/ucp/vippkg/placeorder`：已补未支付订单冷却、新用户注册天数限制、新用户观影数限制、当日会员订单数上限、随机支付无可用通道分支。
- `/ucp/beanpkg/placeorder`：已补未支付订单冷却、当日金豆订单数上限、随机支付无可用通道分支。
- `/ucp/vodorder/create`、`/ucp/vodorder/support`：已接管求片/助力成功分支，事务扣金币、写金币流水并插入或更新 `user_vod_order/user_vod_support`。
- `/ucp/withdraw/create`：已接管提现申请成功分支，普通提现支持金币转余额后冻结，游戏提现冻结游戏余额，写 `user_withdraws/user_frozenlogs/user_balancelogs` 并通过可替换 notifier 发送 Telegram 通知。
- `/ucp/taskbox/taskboxopen`：已接管领奖成功分支，事务写入 `promotion_taskboxlogs`、`users_quota.goldcoin` 和 `user_coinlogs(cointype=19)`，成功返回 `宝箱成功开启` 与 `data.taskdone`。
- `/ucp/task/sign`、`/ucp/task/qrcodeSave`、`/ucp/task/invitecodeInput`、`/ucp/task/adviewClick`：已接管奖励成功分支；登录用户事务写 `users_quota/user_coinlogs` 并按 PHP 返回 `data.taskdone`，游客签到更新 `user_guests.goldcoin/signtime`。
- `/ucp/task/share`、`/ucp/task/qrcode`：已接管分享奖励、分享文案、二维码 keylimit 写入和 PNG 生成成功分支；`/ucp/task/*` 可达业务 action 均已覆盖。
- `/ucp/upgrade`：已接管金币升级尊贵会员成功分支，事务扣金币、写金币流水并更新 `users.sysgid/sysgid_exptime`。
- `/ucp/user/passwd`：已接管密码更新和重新登录成功分支，事务更新 `users.password/salt`、重建登录 session，并返回 `data.user/xxx_api_auth`。
- `/ucp/user/verifyemail`、`/ucp/user/bindmobi`：已接管邮箱验证绑定和手机验证绑定成功分支；邮箱绑定事务更新 `users.email` 并删除验证码 key，手机绑定事务释放旧手机号持有人并更新当前用户 `users.mobi`。
- `/ucp/vippkg/coinorder`：已接管金币购买 VIP 成功分支；成功事务扣金币、写 `user_coinlogs(cointype=103)` 并按当前 VIP 到期时间续期或从当前时间起算。
- `/ucp/beanpkg/coinorder`：已接管金币兑换金豆成功分支；成功事务扣金币、写 `user_coinlogs(cointype=103)`，同时增加 `users_goldbean.gold_bean` 并写 `user_beanlogs(bean_type=20)`。
- `/ucp/coinlog/exchange`：已接管金币和余额互换成功分支；金币转余额写 `user_coinlogs(cointype=104)` 与 `user_balancelogs(paytype=9)`，余额转金币写 `user_balancelogs(paytype=10)` 与 `user_coinlogs(cointype=8)`。

## 执行策略

1. 优先继续迁移低风险前置失败分支。
2. 写金币、金豆、VIP、提现冻结、支付订单、keylimit、Redis/session 或外部发送的成功路径必须先补齐事务、幂等和回滚策略；当前 UCP 可达 action 已按该策略补齐。
3. 每迁移一个 action 或明确一个阻断分支后，同步更新 `MIGRATION_ENDPOINTS.md`、本文件和 `docs/migration-state.md`。
4. 高风险写入成功路径必须启用 reviewer，并补充灰度/回滚说明。
