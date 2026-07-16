# UCP 个人中心迁移进度

## PHP 核对范围

- PHP 路由范围：`/Users/canavs/xjProj/XJBackend/api/api.php` 中 `/ucp/*` 动态路由。
- PHP controller 范围：`/Users/canavs/xjProj/XJBackend/api/src/c/api/ucp/*.php`。
- 核对口径：只统计可通过路由访问的业务 action；`vodorder::error()` 是 public helper，不计入业务接口。

## 总体进度

| 项目 | 数量 |
| --- | ---: |
| PHP UCP 业务 action 总数 | 70 |
| 已完整迁移 | 49 |
| 未完全迁移 | 21 |
| 完全未注册到 Go | 0 |

## 未完全迁移 action

| 模块 | 数量 | 接口 |
| --- | ---: | --- |
| `/ucp/upgrade` | 1 | `/ucp/upgrade` |
| `/ucp/user` | 4 | `/ucp/user/passwd`、`/ucp/user/sendemail`、`/ucp/user/verifyemail`、`/ucp/user/bindmobi` |
| `/ucp/task` | 6 | `/ucp/task/sign`、`/ucp/task/share`、`/ucp/task/qrcode`、`/ucp/task/qrcodeSave`、`/ucp/task/invitecodeInput`、`/ucp/task/adviewClick` |
| `/ucp/withdraw` | 1 | `/ucp/withdraw/create` |
| `/ucp/coinlog` | 1 | `/ucp/coinlog/exchange` |
| `/ucp/taskbox` | 1 | `/ucp/taskbox/taskboxopen` |
| `/ucp/vippkg` | 2 | `/ucp/vippkg/placeorder`、`/ucp/vippkg/coinorder` |
| `/ucp/coinpkg` | 1 | `/ucp/coinpkg/placeorder` |
| `/ucp/beanpkg` | 2 | `/ucp/beanpkg/placeorder`、`/ucp/beanpkg/coinorder` |
| `/ucp/vodorder` | 2 | `/ucp/vodorder/create`、`/ucp/vodorder/support` |

## 风险归类

- 账号资料类：密码更新、重新登录、邮件发送、邮箱/手机绑定成功写入。
- 任务奖励类：签到、分享、二维码保存、邀请码输入、广告点击、任务宝箱开启成功奖励。
- 资产交易类：会员升级、提现冻结、金币/余额兑换、VIP/金币/金豆套餐下单或兑换。
- 求片类：求片金币扣减、助力写入、冻结金币统计更新。

## 最新补充迁移

- `/ucp/user/profile`：已接管成功写入分支，更新当前用户 `gender/nickname`，返回 `资料设置成功`。
- `/ucp/vippkg/placeorder`、`/ucp/coinpkg/placeorder`、`/ucp/beanpkg/placeorder`：已接管支付方式错误或不被允许分支，只读校验支付通道、`paycode`、paytype 和金额范围；订单创建仍暂缓。
- `/ucp/vippkg/placeorder`：已补未支付订单冷却、新用户注册天数限制、新用户观影数限制、当日会员订单数上限、随机支付无可用通道分支。
- `/ucp/beanpkg/placeorder`：已补未支付订单冷却、当日金豆订单数上限、随机支付无可用通道分支。

## 执行策略

1. 优先继续迁移低风险前置失败分支。
2. 不接管会写金币、金豆、VIP、提现冻结、支付订单、keylimit、Redis/session 或外部发送的成功路径，除非先补齐事务、幂等和回滚策略。
3. 每迁移一个 action 或明确一个阻断分支后，同步更新 `MIGRATION_ENDPOINTS.md`、本文件和 `docs/migration-state.md`。
4. 高风险写入成功路径必须启用 reviewer，并补充灰度/回滚说明。
