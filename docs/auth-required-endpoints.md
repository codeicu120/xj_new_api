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
- 本清单已和 `internal/server/router.go` 中的登录态真实 handler 路由对齐；新增登录接口后同步更新此文件和根目录 `MIGRATION_ENDPOINTS.md`。

## 已迁移

| 接口 | PHP handler | Go 状态 | 对比说明 |
| --- | --- | --- | --- |
| `/ucp/myaff` | `c.api.ucp.index->myaff` | 本轮完成 | 使用 `x-cookie-auth` 登录 token，读取推荐用户列表，用户行和分页结构与 PHP 一致。 |
| `/ucp/index` | `c.api.ucp.index->index` | 本轮完成 | 登录/游客个人中心只读聚合；登录返回 `user/uinfo/signed/groups`，游客返回 `user/uinfo/signed`；旧 PHP 本地请求超时，已按源码契约和 Go 输出验证。 |
| `/ucp/affcenter` | `c.api.ucp.index->affcenter` | 本轮完成 | 登录只读推广中心，读取金币、金豆、播放/下载当日计数和用户组权限，用户与 `uinfo` 字段和 PHP 一致。 |
| `GET /ucp/feedback` | `c.api.ucp.index->feedback` | 本轮完成 | 登录只读历史反馈列表，读取 `feedbacks`，分页和 `procRow2` 字段映射与 PHP 一致；POST 写入未接管。 |
| `POST /ucp/feedback`、`/ucp/feedback/create` | `c.api.ucp.index->feedback`、`c.api.ucp.feedback->create` | 本轮完成 | 登录反馈创建；未登录分支 live 对比通过，内容、订单归属、每日次数和写入由 fake 覆盖；图片上传保存和告警通知暂未接管。 |
| `/ucp/msg/send` | `c.api.ucp.msg->send` | 本轮完成 | 登录站内信发送；会话内回复写库由 fake 覆盖，未登录分支 live 对比通过；用户名群发按 PHP 源码 bug 保持不可用。 |
| `GET /ucp/feedback/index` | `c.api.ucp.feedback->index` | 本轮完成 | 新版反馈初始化页，读取最近 30 天最多 100 条 `trade_payments`，支付行映射与 PHP 一致；POST 未接管。 |
| `GET /ucp/feedback/listing` | `c.api.ucp.feedback->listing` | 本轮完成 | 新版反馈列表，读取 `feedbacks`，支持 `type=0/1/2` 过滤，分页和字段映射与 PHP 一致；POST 未接管。 |
| `GET /ucp/feedback/detail` | `c.api.ucp.feedback->detail` | 本轮完成 | 新版反馈详情，读取单条 `feedbacks`、按 `aids` 顺序读取 `attachs`、按 `payid` 读取关联 `trade_payments`；POST 未接管。 |
| `GET /ucp/msg`、`GET /ucp/msg/index` | `c.api.ucp.msg->index` | 本轮完成 | 登录只读消息会话列表，读取 `msgc/msg/users`，分页和 `procRow` 字段映射与 PHP 一致；POST 和写状态 action 未接管。 |
| `/ucp/payment`、`/ucp/payment/index`、`/ucp/payment/listing` | `c.api.ucp.payment->index/listing` | 本轮完成 | 登录只读支付记录，读取 `trade_payments`，分页和 `procRow2` 字段映射与 PHP 一致。 |
| `/ucp/payment/safepaylog` | `c.api.ucp.payment->safepaylog` | 本轮完成 | 登录只读最近 7 天 safepay 记录，最多 10 条，字段映射与 PHP 一致。 |
| `/ucp/account`、`/ucp/account/index` | `c.api.ucp.account->index` | 本轮完成 | 登录只读资产主页，读取账户、金币、汇率和最近余额日志，金额与时间格式对齐 PHP。 |
| `/ucp/account/balancelog` | `c.api.ucp.account->balancelog` | 本轮完成 | 登录只读余额日志分页，读取 `user_balancelogs`，分页和日志字段映射与 PHP 一致。 |
| `/ucp/withdraw`、`/ucp/withdraw/index` | `c.api.ucp.withdraw->index` | 本轮完成 | 登录只读提现初始化页，返回账户、收款地址、金币折算和提现配置；提现 `create` 未接管。 |
| `/ucp/coinlog`、`/ucp/coinlog/index` | `c.api.ucp.coinlog->index` | 本轮完成 | 登录只读金币日志首页，读取账户、金币、汇率和最近 10 条 `user_coinlogs`，类型、时间、手机号遮罩与 PHP 一致。 |
| `/ucp/coinlog/bonuslog` | `c.api.ucp.coinlog->bonuslog` | 本轮完成 | 登录只读收益金币日志分页，列表包含 22/32，累计收益统计按 PHP 保持不含 22/32。 |
| `/ucp/coinlog/invitelog` | `c.api.ucp.coinlog->invitelog` | 本轮完成 | 登录只读邀请金币日志分页，过滤 `cointype IN (201,32,11)`，分页和日志字段映射与 PHP 一致。 |
| `/bought/listing` | `c.api.bought->listing` | 本轮完成 | 登录只读已购影片列表，读取 `user_bought LEFT JOIN vods`，复用 VOD `procRow2` 兼容字段和 `/bought/listing?page=[?]` 分页，对比通过。 |
| `/bought/delete` | `c.api.bought->delete` | 本轮完成 | 登录删除已购影片记录；空 `vodids` 与 PHP 一样返回成功，删除写入按 uid 限定。 |
| `/bought/buy`、`/vod/buy/:vodid`、`/v2/vod/buy/:vodid` | `c.api.bought->buy`、`c.api.vod->buy`、`c.apiv2.vod->buy` | 本轮完成 | 登录购买付费影片；校验已购、影片状态、VIP 折扣和金豆余额，并在事务内扣 `users_goldbean`、写 `user_beanlogs` 和 `user_bought`。 |
| `/playlog/listing`、`/downlog/listing` | `c.api.playlog/downlog->listing` | 本轮完成 | 登录用户读取用户播放/下载记录；未登录按游客 sid 返回空或游客记录，不强制登录。 |
| `/playlog/remove`、`/downlog/remove` | `c.api.playlog/downlog->remove` | 本轮完成 | 登录用户软删除自己的播放/下载记录；未登录按游客 sid 软删除，不强制登录。 |
| `/favorite/listing`、`/minifavorite/listing` | `c.api.favorite/minifavorite->listing` | 本轮完成 | 登录只读收藏列表；普通视频支持 `wd` 搜索，小视频补 `isfavorite=1`。 |
| `/favorite/add`、`/minifavorite/add` | `c.api.favorite/minifavorite->add` | 本轮完成 | 登录新增收藏；未登录、视频不存在、重复收藏分支 live 对比通过，成功写入由 fake 覆盖；金币奖励默认不改资产，保留后续 rewarder 接入点。 |
| `/favorite/remove`、`/minifavorite/remove` | `c.api.favorite/minifavorite->remove` | 本轮完成 | 登录删除收藏记录；空 `vodids` 与 PHP 一样返回 `已删除0项`。 |
| `/comment/post` | `c.api.comment->post` | 本轮完成 | 登录评论发布；未登录分支 live 对比通过，成功写入由 fake 覆盖；金币奖励和回复通知保留后续接入点。 |
| `/community/attention` | `c.api.topic->attention` | 本轮完成 | 登录收藏/取消收藏帖子，支持 `tids` 批量取消；未登录分支 live 对比通过，成功写入由 fake 覆盖。 |
| `/community/up`、`/community/up_comment` | `c.api.topic->up/up_comment` | 本轮完成 | 登录点赞/取消点赞帖子或评论；未登录分支 live 对比通过，成功写入由 fake 覆盖。 |
| `/community/comment` | `c.api.topic->comment` | 本轮完成 | 登录社区评论发布；未登录分支 live 对比通过，成功写入由 fake 覆盖。 |
| `/community/post` | `c.api.topic->post` | 本轮完成 | 登录发布社区主题；未登录分支 live 对比通过，无文件成功分支由 fake 覆盖，图片保存后续接管。 |
| `/ucp/task/sharepic` | `c.api.ucp.task->sharepic` | 本轮完成 | 此 action 在 UCP 下但不要求登录，只读随机推广海报；奖励/签到 task action 未接管。 |
| `/ucp/task`、`/ucp/task/index` | `c.api.ucp.task->index` | 本轮完成 | 登录只读任务中心聚合；统计分享、评论、收藏、观看、保存二维码、广告点击、小视频下载任务进度，不发奖励。 |
| `/ucp/task/qrlink` | `c.api.ucp.task->qrlink` | 本轮完成 | 登录只读推广二维码链接；保留 pid 校验、渠道配置回退、每日 inviteUrls 分组选择和 `{inviteCode}` 替换。 |
| `/ucp/taskbox/index` | `c.api.ucp.taskbox->index` | 本轮完成 | 此 action 在 UCP 下但不要求登录，只读任务宝箱状态；开启宝箱奖励写入未接管。 |
| `/ucp/taskbox/taskboxlog` | `c.api.ucp.taskbox->taskboxlog` | 本轮完成 | 登录只读本人任务宝箱日志，分页 URL 为 `/ucp/taskbox/taskboxlog?page=[?]`；开启宝箱奖励写入未接管。 |
| `/ucp/taskbox/qrlink` | `c.api.ucp.taskbox->qrlink` | 本轮完成 | 登录只读任务宝箱推广二维码链接；复用 qrlink 兼容逻辑但读取 `taskbox.qrcode.link`。 |
| `/onego/history` | `c.api.onego->history` | 本轮完成 | 登录只读本人一元购投注历史；投注写入 `/onego/bet` 未接管。 |
| `/ucp/user`、`/ucp/user/index` | `c.api.ucp.user->index` | 本轮完成 | 登录只读当前用户资料；资料修改、密码和邮箱/手机验证码 action 未接管。 |
| `/ucp/bankcard`、`/ucp/bankcard/index` | `c.api.ucp.bankcard->index` | 本轮完成 | 登录只读提款地址和后台银行列表。 |
| `/ucp/bankcard/create`、`/ucp/bankcard/modify`、`/ucp/bankcard/delete` | `c.api.ucp.bankcard->create/modify/delete` | 本轮完成 | 登录新增、修改、删除本人提款地址；保留 PHP 旧文案、类型映射和默认地址逻辑。 |
| `/ucp/vippkg`、`/ucp/vippkg/index` | `c.api.ucp.vippkg->index` | 本轮完成 | 登录只读 VIP 套餐；支付通道通过接口隔离，当前不伪造旧 PHP `conf/payment.php`。 |
| `/ucp/coinpkg`、`/ucp/coinpkg/index` | `c.api.ucp.coinpkg->index` | 本轮完成 | 登录只读金币套餐；保留套餐字段和 `safepayurl`。 |
| `/ucp/beanpkg`、`/ucp/beanpkg/index` | `c.api.ucp.beanpkg->index` | 本轮完成 | 登录只读金豆套餐；保留套餐字段和 `safepayurl`。 |
| `/ucp/vodorder`、`/ucp/vodorder/index` | `c.api.ucp.vodorder->index` | 本轮完成 | 登录只读求片榜单；返回当前期数、榜单、top 助力人和本人助力数，不执行金币写入。 |
| `/ucp/vodorder/myorders`、`/ucp/vodorder/mysupports`、`/ucp/vodorder/historyorders` | `c.api.ucp.vodorder->myorders/mysupports/historyorders` | 本轮完成 | 登录只读求片/助力/历史成功记录；`create/support` 未接管。 |

## 暂缓

| 接口 | 原因 |
| --- | --- |
| `/ucp/task/*`（除 `/ucp/task`、`/ucp/task/index`、`/ucp/task/sharepic`、`/ucp/task/qrlink`） | 涉及写库、奖励、图片生成或状态变更，需要单独测试和回滚策略。 |
| `/ucp/vippkg/*`、`/ucp/coinpkg/*`、`/ucp/beanpkg/*` 其他 action、`/ucp/payment/*` 其他 action、`/ucp/withdraw/create`、`/ucp/coinlog/exchange`、`/payment/reqpay` 和 `/respond/*` | 会员、金币、金豆、支付和提现写入相关，涉及资产和交易；套餐 `index`、`/ucp/payment/listing`、`/ucp/payment/safepaylog`、`/ucp/withdraw/index`、`/ucp/coinlog/index`、`/ucp/coinlog/bonuslog`、`/ucp/coinlog/invitelog`、`/payment/payways` 和 `/payment/chpayway` 已迁移。 |
| `/game/wali/topup`、`/game/wali/withdraw`、`/game/wali/enter`、`/game/lottery/topup`、`/game/lottery/withdraw`、`/game/lottery/enter`、`/game/lottery/balance` | 游戏资产、余额或外部平台调用。 |
