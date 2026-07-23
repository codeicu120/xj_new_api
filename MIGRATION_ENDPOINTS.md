# 接口重构总览

更新时间：2026-07-15

旧 PHP 项目：`/Users/canavs/xjProj/XJBackend/api`

旧 PHP 本地对比地址：`http://127.0.0.1:18765`

Go 项目：`/Users/canavs/xjProj/xj_comp`

说明：

- “已重构”表示 Go/Gin 中已有真实 handler/service/repository 实现，并已做单元测试或 PHP-Go 对比。
- “占位”表示 Go 路由中注册了 `notImplemented`，但业务尚未迁移。
- “未重构”表示 Go 侧尚未实现对应旧 PHP 行为。
- 本文档已和 `internal/server/router.go` 对齐；新增/删除 Go 路由时必须同步更新“Go 代码实际注册路由”与下面各状态表。
- 公共接口和登录接口的详细压缩记录分别见 `docs/public-endpoints.md`、`docs/auth-required-endpoints.md`、`docs/migration-state.md`。

## Go 代码实际注册路由

来源：`internal/server/router.go`。

### 已接真实 handler

| 接口 | Method | Go handler |
| --- | --- | --- |
| `/healthz` | GET | `healthHandler` |
| `/readyz` | GET | `healthHandler` |
| `/`、`/index` | ANY | `IndexHandler.Index` |
| `/sysavatar` | ANY | `UserHandler.SysAvatar` |
| `/logout` | ANY | `UserHandler.Logout` |
| `/register`、`/login`、`/forgot`、`/delete`、`/changePhone` | ANY | `UserHandler` 账号失败分支 |
| `/sms`、`/sms/index`、`/email`、`/email/index` | ANY | `handler.EmptyHTML` |
| `/sms/sendv`、`/sms/sendu`、`/email/send` | ANY | `VerificationHandler` |
| `/captcha/req` | ANY | `CaptchaHandler.Req` |
| `/captcha/pic`、`/captcha/picx` | ANY | `CaptchaHandler.Pic/PicX` |
| `/captcha/verify` | ANY | `CaptchaHandler.Verify` |
| `/test` | ANY | `TestHandler.Test` |
| `/iploc/:ip` | ANY | `IPLocHandler.Find` |
| `/game/platforms` | ANY | `GameHandler.Platforms` |
| `/game/categories` | ANY | `GameHandler.Categories` |
| `/game/games` | ANY | `GameHandler.Games` |
| `/game/broadcasts` | ANY | `GameHandler.Broadcasts` |
| `/game/wali/gameList` | ANY | `GameHandler.WaliGames` |
| `/game/wali/test` | ANY | `GameHandler.WaliTest` |
| `/game/wali/balance` | ANY | `GameHandler.WaliBalance` |
| `/game/wali/topup`、`/game/wali/withdraw` | ANY | `GameHandler.TransferTopup/TransferWithdraw` |
| `/game/wali/enter` | ANY | `GameHandler.WaliEnter` |
| `/game/lottery/gameList` | ANY | `GameHandler.LotteryGames` |
| `/game/lottery/topup`、`/game/lottery/withdraw` | ANY | `GameHandler.TransferTopup/TransferWithdraw` |
| `/game/lottery/enter`、`/game/lottery/balance` | ANY | `GameHandler.LotteryEnter/LotteryBalance` |
| `/hgame/index` | ANY | `HGameHandler.Index` |
| `/starLive/index`、`/starLive/queryCoinBalance`、`/starLive/gameBet`、`/starLive/gameWin`、`/starLive/translate`、`/starLive/tryAgain` | ANY | `StarLiveHandler` |
| `/art`、`/art/index` | ANY | `ArtHandler.Index` |
| `/art/announce` | ANY | `ArtHandler.Announce` |
| `/art/show` | ANY | `ArtHandler.Show` |
| `/attach`、`/attach/index`、`/attach/upavatar` | ANY | `AttachHandler.Index/UpAvatar` |
| `/:size/:uri`（`C*`/`T*`/`R*`/`M`/`N`） | ANY | `PicHandler.Index` |
| `/getLikeRows` | ANY | `VODHandler.LikeRows` |
| `/getCover` | ANY | `IndexHandler.GetCover` |
| `/search` | ANY | `VODHandler.Search` |
| `/minisearch` | ANY | `VODHandler.MiniSearch` |
| `/shortcutstats/add`、`/adstats/add`、`/playstats/add` | ANY | `StatsHandler` |
| `/open`、`/open/index`、`/open/reqauth` | ANY | `OpenHandler.Index/ReqAuth` |
| `/activity`、`/activity/index`、`/activity/details` | ANY | `ActivityHandler.Index/Details` |
| `/activity/luckyprizes` | ANY | `ActivityHandler.LuckyPrizes` |
| `/activity/newyear2020`、`/activity/luckydraw` | ANY | `ActivityHandler.NewYear2020/LuckyDraw` |
| `/activity/luckydrawhistory` | ANY | `ActivityHandler.LuckyDrawHistory` |
| `/activity/ranking`、`/activity/receive` | ANY | `ActivityHandler.Ranking/Receive` |
| `/activity/recommends` | ANY | `ActivityHandler.Recommends` |
| `/invite/info` | ANY | `InviteHandler.Info` |
| `/invite/bind` | ANY | `InviteHandler.Bind` |
| `/payment/index`、`/payment/query` | ANY | `PaymentHandler.Query` |
| `/payment/payways` | ANY | `PaymentHandler.Payways` |
| `/payment/chpayway` | ANY | `PaymentHandler.ChPayway` |
| `/payment/unpaid`、`/payment/reqpay`、`/payment/pay12req`、`/payment/success`、`/payment/failed` | ANY | `PaymentHandler.Unpaid/ReqPay/Pay12Req/Success/Failed` |
| `/payment/wappay1`、`/payment/wappay2`、`/payment/pay7submit`、`/payment/pay11` | ANY | `PaymentHandler.WapPay1/WapPay2/Pay7Submit/Pay11` |
| `/payment/shangfu`、`/payment/wappay3`、`/payment/wappay4`、`/payment/wappay4a`、`/payment/wappay5`、`/payment/hawpay`、`/payment/easypay`、`/payment/pay6` | ANY | `PaymentHandler.Success` |
| `/payment/pay7`、`/payment/pay8`、`/payment/pay9`、`/payment/pay10`、`/payment/pay10a`、`/payment/pay10b`、`/payment/pay12` | ANY | `PaymentHandler.SuccessHTML` |
| `/payment/gpay1`、`/payment/gpay2`、`/payment/newpay*`（已注册页面 action） | ANY | `PaymentHandler.SuccessHTML` |
| `/respond/*`（已注册支付回调失败分支） | ANY | `RespondHandler.Failed/Chan1` |
| `/bought/listing`、`/bought/delete`、`/bought/buy` | ANY | `BoughtHandler.Listing/Delete/Buy` |
| `/playlog`、`/playlog/index`、`/downlog`、`/downlog/index` | ANY | `handler.EmptyHTML` |
| `/playlog/listing`、`/playlog/remove`、`/downlog/listing`、`/downlog/remove` | ANY | `HistoryHandler` |
| `/miniplaylog/listing`、`/miniplaylog/remove` | ANY | `HistoryHandler` |
| `/favorite`、`/favorite/index`、`/minifavorite`、`/minifavorite/index` | ANY | `handler.EmptyHTML` |
| `/favorite/listing`、`/favorite/remove`、`/minifavorite/listing`、`/minifavorite/remove` | ANY | `FavoriteHandler` |
| `/minivod/listing`、`/minivod/recommend`、`/minivod/hot`、`/minivod/latest`、`/minivod/topzan`、`/minivod/topcomment`、`/minivod/topplay`、`/minivod/topcoin`、`/minivod/topnew`、`/minivod/topday`、`/minivod/topweek`、`/minivod/topmonth` | ANY | `MiniVODHandler.Listing` |
| `/minivod/*-:params`（上述 action） | ANY | `MiniVODHandler.Listing` |
| `/minivod/show/:vodid` | ANY | `MiniVODHandler.Show` |
| `/minivod/throwcoin/:vodid` | ANY | `MiniVODHandler.ThrowCoin` |
| `/minivod/reqplay/:vodid`、`/minivod/reqdown/:vodid` | ANY | `MiniVODHandler.ReqPlay/ReqDown` |
| `/minivod/reqcoin` | ANY | `MiniVODHandler.ReqCoin` |
| `/minivod/reqlong/:vodid` | ANY | `MiniVODHandler.ReqLong` |
| `/minivod/parselong/:vodid/index.m3u8` | ANY | `MiniVODHandler.ParseLongM3U8` |
| `/my/:authorid`、`/my/:authorid/:action` | ANY | `MiniVODHandler.Author` |
| `/community/list`、`/community/recommend`、`/community/hot`、`/community/latest`、`/community/favorite` | ANY | `CommunityHandler.Listing` |
| `/community/*-:params`（上述 action） | ANY | `CommunityHandler.Listing` |
| `/community/show` | ANY | `CommunityHandler.Show` |
| `/community/clisting`、`/community/clisting-:params` | ANY | `CommunityHandler.CommentListing` |
| `/community/categories` | ANY | `CommunityHandler.Categories` |
| `/community/slides` | ANY | `CommunityHandler.Slides` |
| `/community/search` | ANY | `CommunityHandler.Search` |
| `/community/attention` | ANY | `CommunityHandler.Attention` |
| `/community/up`、`/community/up_comment` | ANY | `CommunityHandler.Up/UpComment` |
| `/community/comment` | ANY | `CommunityHandler.Comment` |
| `/community/post` | ANY | `CommunityHandler.Post` |
| `/explore/index` | ANY | `ExploreHandler.Index` |
| `/explore/notification`、`/explore/notification/index` | ANY | `ExploreHandler.EmptyOK` |
| `/explore/notification/clean` | ANY | `ExploreHandler.CleanNotification` |
| `/explore/signtask`、`/explore/signtask/index` | ANY | `ExploreHandler.EmptyOK` |
| `/explore/vodtask`、`/explore/vodtask/index` | ANY | `ExploreHandler.EmptyOK` |
| `/explore/vodtask/show/:vid` | ANY | `ExploreHandler.VodTaskShow` |
| `/explore/vodtask/reqcoin` | ANY | `ExploreHandler.VodTaskReqCoin` |
| `/aiundress`、`/aiundress/listing` | ANY | `AIUndressHandler.Listing` |
| `/aiundress/index` | ANY | `handler.EmptyHTML` |
| `/aiundress/upload`、`/aiundress/undress`、`/aiundress/delete` | ANY | `AIUndressHandler.Upload/Undress/Delete` |
| `/aiundress/moduleList`、`/aiundress/resourceTypeList`、`/aiundress/resourceList` | ANY | `AIUndressHandler.ModuleList/ResourceTypeList/ResourceList` |
| `/getCertUuid` | ANY | `IndexHandler.GetCertUUID` |
| `/getGlobalData` | ANY | `IndexHandler.GetGlobalData` |
| `/init` | ANY | `IndexHandler.Init` |
| `/ucp/index` | ANY | `UCPHandler.Index` |
| `/ucp/user`、`/ucp/user/index` | ANY | `UCPHandler.UserIndex` |
| `/ucp/user/profile`、`/ucp/user/passwd` | ANY | `UCPHandler.UserProfile/UserPasswd` |
| `/ucp/user/checkemail`、`/ucp/user/sendemail`、`/ucp/user/verifyemail`、`/ucp/user/bindmobi` | ANY | `UCPHandler.UserCheckEmail/SendEmail/VerifyEmail/BindMobi` |
| `/ucp/bankcard`、`/ucp/bankcard/index` | ANY | `UCPHandler.BankcardIndex` |
| `/ucp/bankcard/create`、`/ucp/bankcard/modify`、`/ucp/bankcard/delete` | ANY | `UCPHandler.BankcardCreate/Modify/Delete` |
| `/ucp/feedback` | ANY | `UCPHandler.FeedbackListing/FeedbackCreateLegacy` |
| `/ucp/feedback/index` | GET | `UCPHandler.FeedbackIndex` |
| `/ucp/feedback/listing` | GET | `UCPHandler.FeedbackNewListing` |
| `/ucp/feedback/detail` | GET | `UCPHandler.FeedbackDetail` |
| `/ucp/feedback/create` | ANY | `UCPHandler.FeedbackCreate` |
| `/ucp/msg`、`/ucp/msg/index` | GET | `UCPHandler.MsgListing` |
| `/ucp/msg/show` | ANY | `UCPHandler.MsgDetail` |
| `/ucp/msg/send` | ANY | `UCPHandler.MsgSend` |
| `/ucp/msg/setread`、`/ucp/msg/cleanread`、`/ucp/msg/delete` | ANY | `UCPHandler.MsgSetRead/CleanRead/Delete` |
| `/ucp/myaff` | ANY | `UCPHandler.MyAff` |
| `/ucp/rolltitle` | ANY | `UCPHandler.RollTitle` |
| `/ucp/task`、`/ucp/task/index` | ANY | `UCPHandler.TaskIndex` |
| `/ucp/task/sharepic` | ANY | `UCPHandler.TaskSharePic` |
| `/ucp/task/qrlink` | ANY | `UCPHandler.TaskQRLink` |
| `/ucp/task/invite` | ANY | `UCPHandler.TaskInvite` |
| `/ucp/task/sign`、`/ucp/task/invitecodeInput`、`/ucp/task/adviewClick` | ANY | `UCPHandler.TaskSign/TaskInviteCodeInput/TaskAdviewClick` |
| `/ucp/task/share`、`/ucp/task/qrcode` | ANY | `UCPHandler.TaskShare/TaskQRCode` |
| `/ucp/task/qrcodeSave` | ANY | `UCPHandler.TaskQRCodeSave` |
| `/ucp/taskbox/index`、`/ucp/taskbox/taskboxlog`、`/ucp/taskbox/share`、`/ucp/taskbox/qrlink` | ANY | `UCPHandler.TaskboxIndex/TaskboxLog/TaskboxShare/TaskboxQRLink` |
| `/ucp/taskbox/taskboxopen` | ANY | `UCPHandler.TaskboxOpen` |
| `/ucp/taskbox/qrcode` | ANY | `UCPHandler.TaskboxQRCode` |
| `/ucp/affcenter` | ANY | `UCPHandler.AffCenter` |
| `/ucp/upgrade` | ANY | `UCPHandler.Upgrade` |
| `/ucp/payment`、`/ucp/payment/index`、`/ucp/payment/listing` | ANY | `UCPHandler.PaymentListing` |
| `/ucp/payment/safepaylog` | ANY | `UCPHandler.SafePayLog` |
| `/ucp/account`、`/ucp/account/index` | ANY | `UCPHandler.AccountIndex` |
| `/ucp/account/balancelog` | ANY | `UCPHandler.BalanceLog` |
| `/ucp/withdraw`、`/ucp/withdraw/index` | ANY | `UCPHandler.WithdrawIndex` |
| `/ucp/withdraw/listing` | ANY | `UCPHandler.WithdrawListing` |
| `/ucp/withdraw/rule` | ANY | `UCPHandler.WithdrawRule` |
| `/ucp/withdraw/create` | ANY | `UCPHandler.WithdrawCreate` |
| `/ucp/coinlog`、`/ucp/coinlog/index` | ANY | `UCPHandler.CoinLogIndex` |
| `/ucp/coinlog/bonuslog` | ANY | `UCPHandler.CoinLogBonusLog` |
| `/ucp/coinlog/invitelog` | ANY | `UCPHandler.CoinLogInviteLog` |
| `/ucp/coinlog/exchange` | ANY | `UCPHandler.CoinLogExchange` |
| `/ucp/vippkg`、`/ucp/vippkg/index` | ANY | `UCPHandler.VIPPkgIndex` |
| `/ucp/vippkg/placeorder`、`/ucp/vippkg/coinorder` | ANY | `UCPHandler.VIPPkgPlaceOrder/VIPPkgCoinOrder` |
| `/ucp/coinpkg`、`/ucp/coinpkg/index` | ANY | `UCPHandler.CoinPkgIndex` |
| `/ucp/coinpkg/placeorder` | ANY | `UCPHandler.CoinPkgPlaceOrder` |
| `/ucp/beanpkg`、`/ucp/beanpkg/index` | ANY | `UCPHandler.BeanPkgIndex` |
| `/ucp/beanpkg/placeorder`、`/ucp/beanpkg/coinorder` | ANY | `UCPHandler.BeanPkgPlaceOrder/BeanPkgCoinOrder` |
| `/ucp/vodorder`、`/ucp/vodorder/index` | ANY | `UCPHandler.VODOrderIndex` |
| `/ucp/vodorder/myorders`、`/ucp/vodorder/mysupports`、`/ucp/vodorder/historyorders` | ANY | `UCPHandler.VODOrderMyOrders/MySupports/HistoryOrders` |
| `/ucp/vodorder/create`、`/ucp/vodorder/support` | ANY | `UCPHandler.VODOrderCreate/VODOrderSupport` |
| `/vod/show/:vodid` | ANY | `VODHandler.Show` |
| `/vod/up/:vodid`、`/vod/down/:vodid` | ANY | `VODHandler.Up/Down` |
| `/vod/reqplay/:vodid`、`/vod/reqdown/:vodid` | ANY | `VODHandler.ReqPlay/ReqDown` |
| `/vod/buy/:vodid` | ANY | `BoughtHandler.Buy` |
| `/vod/breaking` | ANY | `VODHandler.Breaking` |
| `/vod/errorreport`、`/v2/vod/errorreport` | ANY | `VODHandler.ErrorReport` |
| `/vod/preView/:vodid/index.m3u8` | ANY | `VODHandler.Preview` |
| `/sendfile/play/:file`、`/sendfile/down/:file` | ANY | `SendfileHandler.Play/Down` |
| `/comment`、`/comment/index` | ANY | `handler.EmptyHTML` |
| `/comment/listing-:params` | ANY | `CommentHandler.Listing` |
| `/comment/post` | ANY | `CommentHandler.Post` |
| `/comment/up`、`/comment/down` | ANY | `CommentHandler.Up/Down` |
| `/special/index` | ANY | `SpecialHandler.Index` |
| `/special/listing`、`/special/listing-:params` | ANY | `SpecialHandler.Listing` |
| `/special/detail/:spid`、`/special/detail/:spid-:params` | ANY | `SpecialHandler.Detail` |
| `/special/up/:spid`、`/special/down/:spid` | ANY | `SpecialHandler.Up/Down` |
| `/onego` | ANY | `OneGoHandler.Rules` |
| `/onego/index` | ANY | `handler.EmptyHTML` |
| `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last`、`/onego/hash`、`/onego/history`、`/onego/bet`、`/onego/lucky`、`/onego/bet_ranks`、`/onego/marquee` | ANY | `OneGoHandler` |
| `/vod/listing`、`/vod/recommend`、`/vod/hot`、`/vod/latest` | ANY | `VODHandler.Listing` |
| `/vod/listing-:params`、`/vod/recommend-:params`、`/vod/hot-:params`、`/vod/latest-:params` | ANY | `VODHandler.Listing` |
| `/v2/amazing/categories` | ANY | `AmazingHandler.Categories` |
| `/v2/amazing/listing`、`/v2/amazing/recommend`、`/v2/amazing/hot`、`/v2/amazing/latest` | ANY | `AmazingHandler.Listing` |
| `/v2/amazing/listing-:params`、`/v2/amazing/recommend-:params`、`/v2/amazing/hot-:params`、`/v2/amazing/latest-:params` | ANY | `AmazingHandler.Listing` |
| `/v2/captcha/req` | ANY | `CaptchaHandler.ReqV2` |
| `/v2/captcha/pic`、`/v2/captcha/picx` | ANY | `CaptchaHandler.Pic/PicX` |
| `/v2/captcha/verify` | ANY | `CaptchaHandler.Verify` |
| `/v2/captcha/test` | ANY | `TestHandler.Test` |
| `/v2/so/list` | ANY | `SOHandler.List` |
| `/v2/register`、`/v2/login`、`/v2/forgot` | ANY | `UserHandler` 账号 v2 失败分支 |
| `/v2/vod/listing`、`/v2/vod/recommend`、`/v2/vod/hot`、`/v2/vod/latest` | ANY | `VODHandler.Listing` |
| `/v2/vod/listing-:params`、`/v2/vod/recommend-:params`、`/v2/vod/hot-:params`、`/v2/vod/latest-:params` | ANY | `VODHandler.Listing` |
| `/v2/vod/show/:vodid` | ANY | `VODHandler.Show` |
| `/v2/vod/up/:vodid`、`/v2/vod/down/:vodid` | ANY | `VODHandler.Up/Down` |
| `/v2/vod/reqplay/:vodid`、`/v2/vod/reqdown/:vodid` | ANY | `VODHandler.ReqPlay/ReqDown` |
| `/v2/vod/buy/:vodid` | ANY | `BoughtHandler.Buy` |
| `/v2/minifavorite`、`/v2/minifavorite/index` | ANY | `handler.EmptyHTML` |
| `/v2/minifavorite/listing` | ANY | `FavoriteHandler.MiniV2Listing` |
| `/v2/minifavorite/add`、`/v2/minifavorite/remove` | ANY | `FavoriteHandler.MiniAdd/MiniRemove` |

## 已重构接口

### 基础与公共接口

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/sysavatar` | `c.api.user->sysavatar` | `UserHandler.SysAvatar` | 已重构，对比通过 |
| `/logout` | `c.api.user->logout` | `UserHandler.Logout` | 已重构，对比通过；删除 type=0 session，非法/无 token 仍返回已退出 |
| `/register`、`/v2/register` | `c.api.user->register`、`c.apiv2.user->register` | `UserHandler.Register` | 部分已重构；安全前置失败分支，覆盖未同意协议、已登录、注册关闭、IP 频控、v1/v2 手机注册手机号格式和查重、v2 邮箱格式/查重、v2 用户名格式/查重、v2 账号注册密码长度，不执行验证码、注册写库、邀请奖励或 session |
| `/login`、`/v2/login` | `c.api.user->login`、`c.apiv2.user->login` | `UserHandler.Login/LoginV2` | 部分已重构；已登录、v1 密码登录关闭、v2 空账号、账号不存在、空密码、密码校验失败、锁定用户、v1/v2 密码登录成功和 v2 手机/邮箱验证码登录成功均已迁移，成功写 `sessions` 并返回 `data.user/xxx_api_auth`；v1 短信登录自动注册仍未接管 |
| `/forgot`、`/v2/forgot` | `c.api.user->forgot`、`c.apiv2.user->forgot` | `UserHandler.Forgot/ForgotV2` | 已重构；手机号/邮箱找回密码 step1/step2/step3 均已接管，step3 使用 PHP 兼容 40 位密码 hash 更新 `users.password/salt` 并返回 `密码已成功设置` |
| `/delete` | `c.api.user2->delAccount` | `UserHandler.Delete` | 已重构；未登录、重复注销、游客账号、验证码失败和成功注销申请均已迁移，成功写 Redis `delAccountList` 并删除当前 session |
| `/changePhone` | `c.api.user2->changePhone` | `UserHandler.ChangePhone` | 已重构；未登录、手机号格式、步骤错误、相同手机号、手机号存在、step1 推进、step2 验证码失败和事务换绑成功分支均已迁移 |
| `/sms`、`/sms/index`、`/email`、`/email/index` | `c.api.sms/email->index` | `handler.EmptyHTML` | 已重构，对比通过；默认空入口返回 `200 text/html` 空 body |
| `/sms/sendv`、`/sms/sendu`、`/email/send` | `c.api.sms/email->send*` | `VerificationHandler` | 已重构；手机号/邮箱/未登录错误分支 live 对比通过，成功发送通过 sender/captcha/limiter fake 覆盖，默认不直连真实短信/邮件平台 |
| `/captcha/req` | `c.api.captcha->req` | `CaptchaHandler.Req` | 已重构，动态 secret 按 shape 对比通过 |
| `/captcha/pic`、`/captcha/picx` | `c.api.captcha->pic/picx` | `CaptchaHandler.Pic/PicX` | 已重构；无效 secret 404 JSON 对比通过，有效 PHP secret 和 Go req secret 均输出 100x34 PNG |
| `/captcha/verify` | `c.api.captcha->verify` | `CaptchaHandler.Verify` | 已重构；v1 图形验证码校验路由已接管，错误分支返回 `验证失败` |
| `/test` | `c.api.test->test` | `TestHandler.Test` | 已重构，动态 PNG 按 status/content-type/PNG 尺寸对比通过 |
| `/attach`、`/attach/index`、`/attach/upavatar` | `c.api.attach->index/upavatar` | `AttachHandler.Index/UpAvatar` | 已重构；空响应、未登录和登录非法头像分支对比通过，成功更新分支由 service fake 覆盖 |
| `/:size/:uri`（`C*`/`T*`/`R*`/`M`/`N`） | `c.api.pic->index` | `PicHandler.Index` | 已重构；无效/不存在文件 404 分支对比通过，图片生成由 service 测试覆盖 |
| `/iploc/:ip` | `c.api.index->iploc` | `IPLocHandler.Find` | 已重构，对比通过 |
| `/`、`/index` | `c.api.index->index` | `IndexHandler.Index` | 已重构，对比通过；首页广告、推荐、最新、猜你喜欢和视频分组聚合，核心 key/count 与旧 PHP 一致 |
| `/getLikeRows` | `c.api.index->getLikeRows` | `VODHandler.LikeRows` | 已重构，对比通过 |
| `/getCover` | `c.api.index->getCover` | `IndexHandler.GetCover` | 已重构；缓存/外部服务/AES 成功分支由 fake 覆盖，非法 pic 错误壳对齐并避免外部服务阻塞 |
| `/getCertUuid` | `c.api.index->getCertUuid` | `IndexHandler.GetCertUUID` | 已重构，本地错误分支对比通过；成功分支用 fake client 覆盖 |
| `/shortcutstats/add` | `c.api.shortcutstats->add` | `StatsHandler.ShortcutAdd` | 已重构，对比通过；按 IP 去重写入快捷方式统计 |
| `/adstats/add` | `c.api.adstats->add` | `StatsHandler.AdAdd` | 已重构，对比通过；复刻无 token 游客 sid 创建和广告点击/安装统计 |
| `/playstats/add` | `c.api.playstats->add` | `StatsHandler.PlayAdd` | 已重构，对比通过；复刻无 token 游客 sid 创建和播放进度统计 |
| `/open`、`/open/index` | `c.api.open->index` | `OpenHandler.Index` | 已重构；按 PHP 源码空响应实现，本地旧 PHP `/open` 返回 500 空体 |
| `/open/reqauth` | `c.api.open->reqauth` | `OpenHandler.ReqAuth` | 已重构，对比通过；游客授权 `authrow/openid/sign/time` 与旧 PHP 一致，动态 `xxx_api_auth` 不回传 |
| `/activity`、`/activity/index` | `c.api.activity->index` | `ActivityHandler.Index` | 已重构，对比通过；当前无进行中活动错误分支一致，成功分支按源码读取活动表 |
| `/activity/details` | `c.api.activity->details` | `ActivityHandler.Details` | 已重构，对比通过；无效 `aid` 错误分支一致，成功分支读取活动和奖项 |
| `/activity/luckyprizes` | `c.api.activity->luckyprizes` | `ActivityHandler.LuckyPrizes` | 已重构，对比通过；静态充值抽奖奖项列表 |
| `/activity/newyear2020`、`/activity/luckydraw` | `c.api.activity->newyear2020/luckydraw` | `ActivityHandler.NewYear2020/LuckyDraw` | 已重构，对比通过；按当前日期复刻旧 PHP 过期活动错误 |
| `/activity/luckydrawhistory` | `c.api.activity->luckydrawhistory` | `ActivityHandler.LuckyDrawHistory` | 已重构，对比通过；登录只读充值抽奖历史并补 `prizename` |
| `/activity/ranking` | `c.api.activity->ranking` | `ActivityHandler.Ranking` | 已重构，对比通过；登录活动排名只读，支持无效活动错误和记录奖项计算 |
| `/activity/receive` | `c.api.activity->receive` | `ActivityHandler.Receive` | 已重构，对比通过；登录领奖结果预览只读，按源码未写入领取状态 |
| `/activity/recommends` | `c.api.activity->recommends` | `ActivityHandler.Recommends` | 已重构，对比通过；登录邀请记录只读，复刻用户行处理 |
| `/invite/info` | `c.api.invite->info` | `InviteHandler.Info` | 已重构，对比通过；登录只读当前绑定邀请码 |
| `/invite/bind` | `c.api.invite->bind` | `InviteHandler.Bind` | 已重构；未登录、已绑定、缺邀请码、无效邀请码、无法绑定自己和绑定成功分支均已迁移，成功事务写 `user_recommend/users/users_quota/user_coinlogs`，返回 `data.data=invitecode` |
| `/payment/index`、`/payment/query` | `c.api.payment->index/query` | `PaymentHandler.Query` | 已重构；只读订单状态查询，校验订单归属后返回 `payrow`；裸 `/payment` 旧 PHP 为 404，不接管 |
| `/payment/payways` | `c.api.payment->payways` | `PaymentHandler.Payways` | 已重构；只读订单支付方式列表，校验订单存在、未支付和归属后返回 `payrow/payments`；支付通道通过接口隔离，不伪造生产配置 |
| `/payment/chpayway` | `c.api.payment->chpayway` | `PaymentHandler.ChPayway` | 已重构；修改未支付订单支付方式，保留本人校验、支付通道校验和条件更新防已支付订单被修改 |
| `/payment/unpaid` | `c.api.payment->unpaid` | `PaymentHandler.Unpaid` | 已重构；旧 PHP 当前直接返回 `data.total_count=0`，后续未执行的 24 小时未支付查询分支不接管 |
| `/payment/reqpay` | `c.api.payment->reqpay` | `PaymentHandler.ReqPay` | 部分已重构；缺失/已支付/过期/非本人、`walletpay` 未登录、known payway 支付方式不允许、unknown payway `retcode=1` 并返回 `payrow/payments` 的前置分支已迁移，钱包支付扣费、第三方网关请求和下单成功分支暂未接管 |
| `/payment/pay12req` | `c.api.payment->pay12req` | `PaymentHandler.Pay12Req` | 部分已重构；缺失/已支付订单返回 payerror HTML，成功请求 pay12 网关分支暂未接管 |
| `/payment/success`、`/payment/failed` | `c.api.payment->success/failed` | `PaymentHandler.Success/Failed` | 已重构；固定支付状态 JSON 文案，不包含平台回调验签 |
| `/payment/shangfu`、`/payment/wappay3`、`/payment/wappay4`、`/payment/wappay4a`、`/payment/wappay5`、`/payment/hawpay`、`/payment/easypay`、`/payment/pay6` | `c.api.payment->$action` | `PaymentHandler.Success` | 已重构；PHP public action 固定返回 `retcode=0 errmsg=支付成功回调`，不包含第三方请求和入账 |
| `/payment/wappay1` | `c.api.payment->wappay1` | `PaymentHandler.WapPay1` | 已重构；固定 `retcode=0 errmsg=支付成功回调`，不涉及回调入账 |
| `/payment/wappay2` | `c.api.payment->wappay2` | `PaymentHandler.WapPay2` | 已重构；无 `payid` 返回固定成功回调文案，有 `payid` 时只读 `trade_payments.payhtml` 并返回 HTML |
| `/payment/pay7submit` | `c.api.payment->pay7submit` | `PaymentHandler.Pay7Submit` | 已重构；解码 `p` 后生成自动 POST 表单 HTML，不请求支付平台 |
| `/payment/pay11` | `c.api.payment->pay11` | `PaymentHandler.Pay11` | 已重构；无 `qrlink` 返回支付成功页，有 `qrlink` 返回二维码 HTML 页 |
| `/payment/pay7`、`/payment/pay8`、`/payment/pay9`、`/payment/pay10`、`/payment/pay10a`、`/payment/pay10b`、`/payment/pay12` | `c.api.payment->$action` | `PaymentHandler.SuccessHTML` | 已重构；PHP public action 均为读取 `templates/api/paysuccess.html` 并返回 HTML |
| `/payment/gpay1`、`/payment/gpay2`、`/payment/newpay`、`/payment/newpayff`、`/payment/newpayxxx`、`/payment/newpayqk`、`/payment/newpayxyf`、`/payment/newpaykf`、`/payment/newpaypi`、`/payment/newpaygs`、`/payment/newpaylep`、`/payment/newpayys`、`/payment/newpayyswx`、`/payment/newpayhw`、`/payment/newpayhs`、`/payment/newpaypx`、`/payment/newpaypxwx`、`/payment/newpay99`、`/payment/newpayxy`、`/payment/newpayjd`、`/payment/newpaycr`、`/payment/newpaylu`、`/payment/newpayluwx`、`/payment/newpaymyr`、`/payment/newpaymyrz`、`/payment/newpaylh`、`/payment/newpaylai`、`/payment/newpayxh`、`/payment/newpayya`、`/payment/newpayyh`、`/payment/newpayhf`、`/payment/newpaydd`、`/payment/newpaykk`、`/payment/newpayrq` | `c.api.payment->$action` | `PaymentHandler.SuccessHTML` | 已重构；PHP public action 均为支付成功 HTML 页面；其中 `_newpayhw` 下单分支写 `out_trxid` 未接管，只接管 public 返回页 |
| `/respond/shangfu`、`/respond/wappay1`、`/respond/wappay2`、`/respond/wappay3`、`/respond/wappay4`、`/respond/wappay4a`、`/respond/wappay5` | `c.respond.*` | `RespondHandler.Provider` | 部分已重构；provider registry/verifier 骨架已接管，空请求、缺配置、验签失败和验签成功但 accounting disabled 均返回 provider `echoErr()`；`shangfu` MD5 form fixture 已覆盖，锁单入账未接管 |
| `/respond/hawpay`、`/respond/easypay`、`/respond/gpay1`、`/respond/gpay2`、`/respond/pay6`、`/respond/pay7`、`/respond/pay8`、`/respond/pay9`、`/respond/pay10`、`/respond/pay10a`、`/respond/pay10b`、`/respond/pay11`、`/respond/pay12` | `c.respond.*` | `RespondHandler.Provider` | 部分已重构；缺配置/解析失败仍返回 `failed`/`Err`/`FAILED` 等旧失败文本，`pay7` MD5 query+secret 验签 fixture 已覆盖但 accounting disabled 仍返回失败文本，其他 provider 验签和所有入账事务未接管 |
| `/respond/newpay*`（除未注册 `chan1`） | `c.respond.*` | `RespondHandler.Provider` | 部分已重构；缺配置/解析失败返回旧 provider 失败文本，具体 provider 验签和入账事务未接管 |
| `/respond/chan1` | `c.respond.chan1` | `RespondHandler.Chan1` | 部分已重构；token 校验失败、合法 token 后用户不存在、已送过会员、套餐不存在或停用只读失败分支已迁移，下单、充值入账、支付记录变更和赠送 VIP 成功分支暂未接管 |
| `/bought/listing` | `c.api.bought->listing` | `BoughtHandler.Listing` | 已重构，对比通过；登录只读已购影片列表，复用 VOD 行处理和 PHP 分页 |
| `/bought/delete` | `c.api.bought->delete` | `BoughtHandler.Delete` | 已重构，对比通过；登录删除已购影片记录，空 `vodids` 成功 |
| `/bought/buy` | `c.api.bought->buy` | `BoughtHandler.Buy` | 已重构；登录购买付费影片，复刻记录不存在、已购成功、VIP 折扣、金豆余额和金豆事务扣费写 `user_beanlogs/user_bought` |
| `/comment`、`/comment/index` | `c.api.comment->index` | `handler.EmptyHTML` | 已重构；旧 PHP 空方法，返回 `200 text/html` 空 body |
| `/explore/notification`、`/explore/notification/index` | `c.api.explore.notification->index` | `ExploreHandler.EmptyOK` | 已重构，对比通过；旧 PHP 空 OK，动态 `xxx_api_auth` 不回传 |
| `/explore/notification/:action?`（除 `/explore/notification`、`/explore/notification/index`、`/explore/notification/clean`） | `c.api.explore.notification->$action` | 不接管 | PHP `notification` 仅定义 `index/clean`，未发现其他稳定 action |
| `/explore/signtask`、`/explore/signtask/index` | `c.api.explore.signtask->index` | `ExploreHandler.EmptyOK` | 已重构，对比通过；旧 PHP 空 OK |
| `/explore/signtask/sign` | `c.api.explore.signtask->sign` | `ExploreHandler.SignTaskSign` | 已重构；登录用户/游客签到事务，锁定 `users/user_guests`，按连续天数发金币或 VIP、写 `explore_signlogs/explore_guestsignlogs` 并更新签到状态 |
| `/explore/vodtask`、`/explore/vodtask/index` | `c.api.explore.vodtask->index` | `ExploreHandler.EmptyOK` | 已重构，对比通过；旧 PHP 空 OK |
| `/explore/vodtask/show/:vid` | `c.api.explore.vodtask->show` | `ExploreHandler.VodTaskShow` | 已重构；激励视频展示并创建/复用当日领取日志，错误分支 live 对比通过，成功分支 fake 覆盖 |
| `/explore/vodtask/reqcoin` | `c.api.explore.vodtask->reqcoin` | `ExploreHandler.VodTaskReqCoin` | 已重构；领取激励视频金币，事务锁定日志，登录用户写 `users_quota/user_coinlogs`，游客更新 `user_guests.goldcoin` |
| `/explore/index` | `c.api.explore.index->index` | `ExploreHandler.Index` | 已重构，对比通过；发现页 tab、7 日签到奖励和签到状态只读聚合 |
| `/explore/notification/clean` | `c.api.explore.notification->clean` | `ExploreHandler.CleanNotification` | 已重构，对比通过；清理发现页红点，仅更新 `notification_all` |

### 游戏公共接口

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/game/platforms` | `c.api.game.index->index` | `GameHandler.Platforms` | 已重构，对比通过 |
| `/game/categories` | `c.api.game.index->categories` | `GameHandler.Categories` | 已重构，对比通过 |
| `/game/games` | `c.api.game.index->games` | `GameHandler.Games` | 已重构，对比通过 |
| `/game/broadcasts` | `c.api.game.index->broadcasts` | `GameHandler.Broadcasts` | 已重构，随机广播按 shape 对比通过 |
| `/game/wali/gameList` | `c.api.game.wali->games` | `GameHandler.WaliGames` | 已重构，对比通过；`category_id=5` 游客未登录分支已对齐 |
| `/game/wali/test` | `c.api.game.wali->ping` | `GameHandler.WaliTest` | 已重构，对比通过；读取平台配置后 AES-ECB 加密、签名并调用瓦力 ping |
| `/game/wali/balance` | `c.api.game.wali->getBalance` | `GameHandler.WaliBalance` | 已重构，对比通过；登录后外部只读余额查询 |
| `/game/wali/topup`、`/game/wali/withdraw`、`/game/wali/enter` | `c.api.game.wali->topup/withdraw/enterGame` | `GameHandler.TransferTopup/TransferWithdraw/WaliEnter` | 部分已重构；未登录、上分低于 `gamecoinlimit`、上分余额不足、下分金额不正确和进入游戏外部成功/失败分支已迁移，上下分金币事务和外部转账请求暂未接管 |
| `/game/lottery/gameList` | `c.api.game.lottery->gameList` | `GameHandler.LotteryGames` | 已重构；彩票普通分类只读列表，`category_id=5` 游客未登录分支已对齐 |
| `/game/lottery/topup`、`/game/lottery/withdraw`、`/game/lottery/enter`、`/game/lottery/balance` | `c.api.game.lottery->$action` | `GameHandler.TransferTopup/TransferWithdraw/LotteryEnter/LotteryBalance` | 部分已重构；未登录、上分低于 `gamecoinlimit`、上分余额不足、下分金额不正确、进入游戏外部成功/失败和余额只读查询分支已迁移，彩票平台上下分订单和金币写入暂未接管 |
| `/hgame/index` | `c.api.hgame->index` | `HGameHandler.Index` | 已重构，对比通过；HGame 公共只读列表，`/hgame` 保持旧 PHP 404 未接管 |
| `/hgame/:action`（除 `/hgame/index`） | `c.api.hgame->$action` | 不接管 | PHP `c.api.hgame` 仅定义 `index`，未发现其他稳定 action；不伪造业务响应 |
| `/onego` | `c.api.onego->rules`（旧路由默认行为） | `OneGoHandler.Rules` | 已重构，对比通过；裸路径与旧服务一致返回一元购规则/未开放错误壳 |
| `/onego/index` | `c.api.onego->index` | `handler.EmptyHTML` | 已重构，对比通过；旧 PHP 空方法，返回 `text/html` 空 body |
| `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last` | `c.api.onego->rules/rooms/current/last` | `OneGoHandler` | 已重构，对比通过；一元购公共只读规则/房间/当前期数/上期记录，旧 PHP 动态 `xxx_api_auth` 忽略 |
| `/onego/hash` | `c.api.onego->hash` | `OneGoHandler.Hash` | 已重构；公共哈希计算接口，复刻 SHA256 后提取末尾数字期号规则 |
| `/onego/history` | `c.api.onego->history` | `OneGoHandler.History` | 已重构，对比通过；登录只读本人投注历史，未登录 `retcode=-9999` |
| `/onego/bet` | `c.api.onego->bet` | `OneGoHandler.Bet` | 已重构；未登录、押注数量为 0、无效场次、无效期号、未开始、已结束、未知用户、余额不足和投注成功分支均已迁移，成功事务扣金币、生成号码、累加期数并写订单 |
| `/onego/lucky` | `c.api.onego->lucky` | `OneGoHandler.Lucky` | 已重构，对比通过；一元购幸运榜公共只读，保留旧 PHP 排行 SQL 未分页行为 |
| `/onego/bet_ranks` | `c.api.onego->bet_ranks` | `OneGoHandler.BetRanks` | 已重构；押注排行只读，错误分支 live 对比通过，本地无订单样本成功分支由 fake 覆盖 |
| `/onego/marquee` | `c.api.onego->marquee` | `OneGoHandler.Marquee` | 已重构，对比通过；一元购跑马灯公共只读，按最近已开奖期生成中奖消息 |

### v2 公共接口

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/v2/so/list` | `c.apiv2.so->index` | `SOHandler.List` | 已重构，对比通过 |
| `/v2/amazing/categories` | `c.apiv2.amazing->categories` | `AmazingHandler.Categories` | 已重构，对比通过 |
| `/v2/amazing/listing` | `c.apiv2.amazing->listing` | `AmazingHandler.Listing` | 已重构，对比通过 |
| `/v2/amazing/listing-:params` | `c.apiv2.amazing->listing` | `AmazingHandler.Listing` | 已重构，对比通过 |
| `/v2/amazing/recommend` | `c.apiv2.amazing->listing` | `AmazingHandler.Listing` | 已重构，对比通过 |
| `/v2/amazing/recommend-:params` | `c.apiv2.amazing->listing` | `AmazingHandler.Listing` | 已重构 |
| `/v2/amazing/hot` | `c.apiv2.amazing->listing` | `AmazingHandler.Listing` | 已重构，对比通过 |
| `/v2/amazing/hot-:params` | `c.apiv2.amazing->listing` | `AmazingHandler.Listing` | 已重构 |
| `/v2/amazing/latest` | `c.apiv2.amazing->listing` | `AmazingHandler.Listing` | 已重构，对比通过 |
| `/v2/amazing/latest-:params` | `c.apiv2.amazing->listing` | `AmazingHandler.Listing` | 已重构 |
| `/v2/captcha/req` | `c.apiv2.captcha->req` | `CaptchaHandler.ReqV2` | 已重构；返回 URL encode 后的 base64 PNG、`smscaptcha` 和 `captcha_key` |
| `/v2/captcha/pic`、`/v2/captcha/picx` | `c.apiv2.captcha->pic/picx` | `CaptchaHandler.Pic/PicX` | 已重构；无效 secret 返回 HTTP 404 + `retcode=-4`，有效 secret 输出 100x34 PNG |
| `/v2/captcha/verify` | `c.apiv2.captcha->verify` | `CaptchaHandler.Verify` | 已重构；本地 `captcha_key/captcha_code` 分支对齐，Google/Tencent/自建验证码外部票据不伪造成功 |
| `/v2/captcha/test` | `c.apiv2.captcha->test` | `TestHandler.Test` | 已重构；输出 100x34 PNG |
| `/v2/vod/listing` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/v2/vod/listing-:params` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/v2/vod/recommend` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构，随机列表按 shape 对比 |
| `/v2/vod/recommend-:params` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构 |
| `/v2/vod/hot` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/v2/vod/hot-:params` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构 |
| `/v2/vod/latest` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/v2/vod/latest-:params` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构 |
| `/v2/vod/show/:vodid` | `c.apiv2.vod->show` | `VODHandler.Show` | 已重构，对比通过；复用视频详情实现 |
| `/v2/vod/reqplay/:vodid`、`/v2/vod/reqdown/:vodid` | `c.apiv2.vod->reqplay/reqdown` | `VODHandler.ReqPlay/ReqDown` | 已接管可控路径；复用普通视频播放/下载地址请求实现，记录/购买/权限/地址错误、免费/限免、已观看/下载和额度内分支可用，非扣费成功路径写播放/下载日志与 `vods` 计数；超限扣金币、扣费日志和奖励分支暂不写资产 |
| `/v2/vod/buy/:vodid` | `c.apiv2.vod->buy` | `BoughtHandler.Buy` | 已重构；复用购买付费影片事务，返回码按 v2 PHP：未登录 `-9999`、余额不足 `4`、不存在 `-1` |
| `/v2/minifavorite`、`/v2/minifavorite/index` | `c.apiv2.minifavorite->index` | `handler.EmptyHTML` | 已重构；旧 PHP 空方法，返回 `200 text/html` 空 body |
| `/v2/minifavorite/listing` | `c.apiv2.minifavorite->listing` | `FavoriteHandler.MiniV2Listing` | 已重构；登录只读小视频收藏，支持 `wd` 搜索，rows 按 v2 PHP 包装为 `{vodrow,user}` |
| `/v2/minifavorite/add` | `c.apiv2.minifavorite->add` | `FavoriteHandler.MiniAdd` | 已重构；复用小视频收藏新增逻辑，登录、视频不存在、重复收藏和成功写入分支迁移；金币奖励默认不改资产 |
| `/v2/minifavorite/remove` | `c.apiv2.minifavorite->remove` | `FavoriteHandler.MiniRemove` | 已重构；复用小视频取消收藏逻辑，空 `vodids` 返回 `已删除0项` |

### 非 v2 视频列表接口

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/vod/listing` | `c.api.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/vod/listing-:params` | `c.api.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/vod/recommend` | `c.api.vod->listing` | `VODHandler.Listing` | 已重构，推荐/随机列表按 shape 对比 |
| `/vod/recommend-:params` | `c.api.vod->listing` | `VODHandler.Listing` | 已重构 |
| `/vod/hot` | `c.api.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/vod/hot-:params` | `c.api.vod->listing` | `VODHandler.Listing` | 已重构 |
| `/vod/latest` | `c.api.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/vod/latest-:params` | `c.api.vod->listing` | `VODHandler.Listing` | 已重构 |
| `/vod/show/:vodid` | `c.api.vod->show` | `VODHandler.Show` | 已重构，详情主字段对比通过；相似/喜欢随机列表按 shape 对比 |
| `/vod/up/:vodid`、`/vod/down/:vodid`、`/v2/vod/up/:vodid`、`/v2/vod/down/:vodid` | `c.api.vod/apiv2.vod->up/down` | `VODHandler.Up/Down` | 已重构；普通视频赞踩状态切换，登录用户写 `vod_updowns`，游客用进程内 limiter；无效视频分支 live 对比通过 |
| `/vod/reqplay/:vodid`、`/vod/reqdown/:vodid` | `c.api.vod->reqplay/reqdown` | `VODHandler.ReqPlay/ReqDown` | 已接管可控路径；记录/购买/权限/地址错误、免费/限免、已观看/下载和权限额度内提供地址，非扣费成功路径写播放/下载日志与 `vods` 计数；超限扣金币、扣费日志和奖励分支暂不写资产 |
| `/vod/buy/:vodid` | `c.api.vod->buy` | `BoughtHandler.Buy` | 已重构；复用购买付费影片事务，金豆扣减和已购写入保持同一事务 |
| `/vod/preView/:vodid/index.m3u8` | `c.api.vod->preView` | `VODHandler.Preview` | 已重构，m3u8 输出对比通过 |
| `/sendfile/play/:file` | `c.api.sendfile->play` | `SendfileHandler.Play` | 已重构，按旧 PHP 空壳行为对齐 |
| `/sendfile/down/:file` | `c.api.sendfile->down` | `SendfileHandler.Down` | 已重构，按旧 PHP 空响应对齐 |

### 评论、收藏、播放/下载记录

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/comment/listing-:params` | `c.api.comment->listing` | `CommentHandler.Listing` | 已重构，对比通过 |
| `/comment/post` | `c.api.comment->post` | `CommentHandler.Post` | 已重构；登录评论发布，保留权限、长度、字符、回复、重复校验和评论树写入；金币奖励和回复通知保留后续接口接入点 |
| `/comment/up`、`/comment/down` | `c.api.comment->up/down` | `CommentHandler.Up/Down` | 已重构；游客/登录赞踩、重复限制和计数自增；无效评论分支 live 对比通过 |
| `/playlog`、`/playlog/index` | `c.api.playlog->index` | `handler.EmptyHTML` | 已重构，对比通过；旧 PHP 空方法，返回 `200 text/html` 空 body |
| `/playlog/listing` | `c.api.playlog->listing` | `HistoryHandler.PlayListing` | 已重构；播放记录只读列表，支持登录/游客、timeline、分页和 PHP 相对时间格式；游客 timeline 2/3 保留旧 PHP 边界反序行为 |
| `/playlog/remove` | `c.api.playlog->remove` | `HistoryHandler.PlayRemove` | 已重构；按登录 uid 或游客 sid 软删除播放记录，空 `vodids` 返回 `已删除0项` |
| `/downlog`、`/downlog/index` | `c.api.downlog->index` | `handler.EmptyHTML` | 已重构，对比通过；旧 PHP 空方法，返回 `200 text/html` 空 body |
| `/downlog/listing` | `c.api.downlog->listing` | `HistoryHandler.DownListing` | 已重构；下载记录只读列表，支持登录/游客、timeline、分页和 PHP 相对时间格式 |
| `/downlog/remove` | `c.api.downlog->remove` | `HistoryHandler.DownRemove` | 已重构；按登录 uid 或游客 sid 软删除下载记录，空 `vodids` 返回 `已删除0项` |
| `/favorite`、`/favorite/index` | `c.api.favorite->index` | `handler.EmptyHTML` | 已重构，对比通过；旧 PHP 空方法，返回 `200 text/html` 空 body |
| `/favorite/listing` | `c.api.favorite->listing` | `FavoriteHandler.Listing` | 已重构；登录只读普通视频收藏，支持 `wd` 搜索和分页，复用 VOD 行处理 |
| `/favorite/add` | `c.api.favorite->add` | `FavoriteHandler.Add` | 已重构；登录、视频不存在、重复收藏和成功写入分支迁移；金币奖励保留 rewarder 扩展点，默认不改资产 |
| `/favorite/remove` | `c.api.favorite->remove` | `FavoriteHandler.Remove` | 已重构；登录删除普通视频收藏，空 `vodids` 返回 `已删除0项` |
| `/minifavorite`、`/minifavorite/index` | `c.api.minifavorite->index` | `handler.EmptyHTML` | 已重构，对比通过；旧 PHP 空方法，返回 `200 text/html` 空 body |
| `/minifavorite/listing` | `c.api.minifavorite->listing` | `FavoriteHandler.MiniListing` | 已重构；登录只读小视频收藏，复用 mini VOD 行处理并补 `isfavorite=1` |
| `/minifavorite/add` | `c.api.minifavorite->add` | `FavoriteHandler.MiniAdd` | 已重构；登录、视频不存在、重复收藏和成功写入分支迁移；金币奖励保留 rewarder 扩展点，默认不改资产 |
| `/minifavorite/remove` | `c.api.minifavorite->remove` | `FavoriteHandler.MiniRemove` | 已重构；登录删除小视频收藏，空 `vodids` 返回 `已删除0项` |

### 小视频、作者页

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/minivod/listing`、`/minivod/recommend`、`/minivod/hot`、`/minivod/latest` | `c.api.minivod->listing` | `MiniVODHandler.Listing` | 已重构，对比通过；支持筛选、排序、分页、随机推荐和 latest 用户包装 rows |
| `/minivod/topzan`、`/minivod/topcomment`、`/minivod/topplay`、`/minivod/topcoin`、`/minivod/topnew`、`/minivod/topday`、`/minivod/topweek`、`/minivod/topmonth` | `c.api.minivod->listing` | `MiniVODHandler.Listing` | 已重构，对比通过；setting 序列化排行榜、日/周/月榜和 rows 用户包装 |
| `/minivod/*-:params`（上述 action） | `c.api.minivod->listing` | `MiniVODHandler.Listing` | 已重构；参数模板 `$cateid-$areaid-$yearid-$tagid-$definition-$duration-$freetype-$mosaic-$langvoice-$orderby-$page` |
| `/minivod/show/:vodid` | `c.api.minivod->show` | `MiniVODHandler.Show` | 已重构；读取 `showtype=1` 小视频详情、作者、分类层级、相关视频和猜你喜欢；本地旧库缺作者样本的错误分支对比通过，成功分支单测覆盖 |
| `/minivod/up/:vodid`、`/minivod/down/:vodid` | `c.api.minivod->up/down` | `MiniVODHandler.Up/Down` | 已重构；小视频赞踩状态切换，登录用户写 `vod_updowns`，游客用进程内 limiter；无效视频分支 live 对比通过 |
| `/minivod/reqplay/:vodid`、`/minivod/reqdown/:vodid` | `c.api.minivod->reqplay/reqdown` | `MiniVODHandler.ReqPlay/ReqDown` | 已接管可控路径；记录/权限/地址错误、免费/限免、已观看/下载和权限额度内提供地址，非扣费成功路径写 `minivod_viewlogs/minivod_guestviewlogs` 与 `vods` 计数；超限扣金币、扣费标记和任务奖励分支暂不写资产 |
| `/minivod/reqcoin` | `c.api.minivod->reqcoin` | `MiniVODHandler.ReqCoin` | 已重构；领取小视频播放任务金币，事务锁定任务日志，登录用户写 `users_quota/user_coinlogs(cointype=25)`，游客更新 `user_guests.goldcoin`；保留旧 PHP 未校验 log 归属行为 |
| `/minivod/throwcoin/:vodid` | `c.api.minivod->throwcoin` | `MiniVODHandler.ThrowCoin` | 已重构；未登录、视频不存在、作者不存在、GET 初始化、POST 参数校验和投币成功事务均已迁移，成功扣用户金币、加作者金币并写 `user_coinlogs/minivod_coinlogs` |
| `/minivod/reqlist` | `c.api.minivod->reqlist` | `MiniVODHandler.ReqList` | 已重构；待展示 viewlog 不足 100 时拉取推荐池，返回随机 10 条并包装 `vodrow/user/isfavorite`，返回前按 `debug` 标记 `reqtime/showtype=1`，并按 PHP 规则随机插入 `minivod.ads` 广告行；按 PHP 三库语义，主库为 `MYSQL_DSN`、`*_slave` 查询通过 `MYSQL_READ_DSN`、推荐日志分表通过 `MYSQL_LOG_DSN`，生产必须注入三套连接 |
| `/minivod/reqlong/:vodid` | `c.api.minivod->getLong2Mini` | `MiniVODHandler.ReqLong` | 已重构；普通长视频转小视频播放地址，支持 CDN 签名/播放服务器 host 补全；错误分支和本地样本成功 URL live 对比通过 |
| `/minivod/parselong/:vodid/index.m3u8` | `c.api.minivod->parseM3u8` | `MiniVODHandler.ParseLongM3U8` | 已重构；复用长转短前置校验，成功时返回 `vnd.apple.mpegurl`，按 `vod_map_ls.start/end` 裁剪子 m3u8 并重写 KEY/TS 绝对 URL；媒体源拉取失败或无子 m3u8 时按 PHP 返回空 m3u8 body |
| `/miniplaylog/listing` | `c.api.minivod->history` | `HistoryHandler.MiniPlayListing` | 已重构；不强制登录，登录/游客按小视频分表读取，mini 行处理和相对时间格式 |
| `/miniplaylog/remove` | `c.api.minivod->historyDelete` | `HistoryHandler.MiniPlayRemove` | 已重构；按 PHP 模型语义用输入 `vodid/vodids` 删除 `logid`，空参数 live 对比通过 |
| `/my/:authorid`、`/my/:authorid/index`、`/my/:authorid/listing` | `c.api.my->index/listing` | `MiniVODHandler.Author` | 已重构，对比通过；作者主页小视频列表，返回 `now/userrow/vodrows/pageinfo/orders` |

### 社区、HGame、AI

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/community/list`、`/community/recommend`、`/community/hot`、`/community/latest` | `c.api.topic->list` | `CommunityHandler.Listing` | 已重构，对比通过；主题列表、推荐/热门/最新、分类/type/分页和媒体字段 |
| `/community/favorite` | `c.api.topic->list` | `CommunityHandler.Listing` | 已重构，对比通过；未登录 `retcode=-9999`，登录按 `topic_favorites` 过滤 |
| `/community/*-:params`（上述 action） | `c.api.topic->list` | `CommunityHandler.Listing` | 已重构；参数模板 `$category_id-$type-$orderby-$page` |
| `/community/show` | `c.api.topic->show` | `CommunityHandler.Show` | 已重构；社区详情，读取主题、媒体和评论树，并保留旧 PHP `visit_count+1` 副作用 |
| `/community/clisting`、`/community/clisting-:params` | `c.api.topic->clisting` | `CommunityHandler.CommentListing` | 已重构，对比通过；评论树列表，`tid` 不存在分支一致 |
| `/community/categories` | `c.api.topic->categories` | `CommunityHandler.Categories` | 已重构；公共只读分类，保留 `parent_id` 过滤、`status=1` 和 ``order`` DESC/id ASC 排序 |
| `/community/slides` | `c.api.topic->slides` | `CommunityHandler.Slides` | 已重构；读取 `global_adgroup_ad19` 并映射 article/link/game 为 post/ad/game |
| `/community/search` | `c.api.topic->search` | `CommunityHandler.Search` | 已重构；空关键词返回 `请输入关键词`，非空按 title/tags 搜索、返回 `rows/hotwords/pageinfo` |
| `/community/attention` | `c.api.topic->attention` | `CommunityHandler.Attention` | 已重构；登录收藏/取消收藏帖子，支持 `tids` 批量取消收藏并按实际删除更新 `fav_count`；未登录分支 live 对比通过 |
| `/community/up`、`/community/up_comment` | `c.api.topic->up/up_comment` | `CommunityHandler.Up/UpComment` | 已重构；登录点赞/取消点赞帖子或评论，更新去重表和 `upnum` 计数；未登录分支 live 对比通过 |
| `/community/comment` | `c.api.topic->comment` | `CommunityHandler.Comment` | 已重构；登录评论发布，保留权限、内容、回复、重复校验和评论树写入，成功评论进入待审核状态 |
| `/community/post` | `c.api.topic->post` | `CommunityHandler.Post` | 已重构；登录发布主题，无文件分支写入 `topics`；图片上传保存暂未接管，超过 3 张图片写入前拒绝以避免半成品主题 |

### 搜索、专题、公告

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/special/index` | `c.api.special->index` | `SpecialHandler.Index` | 已重构，对比通过；旧 PHP 空方法，返回 text/html 空 body |
| `/special/listing`、`/special/listing-:params` | `c.api.special->listing` | `SpecialHandler.Listing` | 已重构，对比通过；专题列表，含前 4 个视频、分页和第一页 actorrows；avatar/coverpic 已按数据库 setting 的 resurl/H5/地区/鉴权规则动态生成 |
| `/special/detail/:spid`、`/special/detail/:spid-:params` | `c.api.special->detail` | `SpecialHandler.Detail` | 已重构，对比通过；专题详情，含全量视频排序和浏览数写入副作用 |
| `/special/up/:spid`、`/special/down/:spid` | `c.api.special->up/down` | `SpecialHandler.Up/Down` | 已重构；不存在分支对比通过，成功和重复投票分支由 service fake 覆盖 |
| `/search` | `c.api.search->index` | `VODHandler.Search` | 已重构，对比通过；空关键词搜索页和关键词结果列表，保留搜索统计写入 |
| `/minisearch` | `c.api.miniSearch->index` | `VODHandler.MiniSearch` | 已重构，对比通过；小视频搜索，保留 `showtype=1`、`/minivod` URL、`minivod_schlogs` 搜索日志写入和旧分页 URL |
| `/art`、`/art/index` | `c.api.art->index` | `ArtHandler.Index` | 已重构，对比通过；旧 PHP 空方法，返回 text/html 空 body |
| `/art/announce` | `c.api.art->announce` | `ArtHandler.Announce` | 已重构，对比通过；公告列表，保留旧 PHP 未定义 `$action` 导致的 `/art/?page=[?]` 分页 URL |
| `/art/show` | `c.api.art->show` | `ArtHandler.Show` | 已重构，对比通过；公告/文章详情，成功和不存在错误分支一致 |
| `/aiundress`、`/aiundress/listing` | `c.api.aiundress->listing` | `AIUndressHandler.Listing` | 已重构，对比通过；登录只读 AI 任务历史，支持 `module/page`，未登录 `retcode=-1` |
| `/aiundress/index` | `c.api.aiundress->index` | `handler.EmptyHTML` | 已重构，对比通过；按本地旧 PHP 运行时行为返回 `200 text/html` 空 body，AI 业务 action 未接管 |
| `/aiundress/moduleList`、`/aiundress/resourceTypeList`、`/aiundress/resourceList` | `c.api.aiundress->moduleList/resourceTypeList/resourceList` | `AIUndressHandler.ModuleList/ResourceTypeList/ResourceList` | 已重构；只读外部资源查询，第三方 host/key 通过 `AIUNDRESS_THIRD_HOST`/`AIUNDRESS_THIRD_KEY` 注入，缺配置或外部请求失败返回 `retcode=-1 errmsg=请求失败` |
| `/aiundress/upload`、`/aiundress/undress`、`/aiundress/delete` | `c.api.aiundress->$action` | `AIUndressHandler.Upload/Undress/Delete` | 部分已重构；`upload` 已接管登录校验、本地保存、可配置 R2 上传、`ai_undress` 新增/刷新和 `data.file` 响应；`undress` 无效图片路径和 `delete` 空记录/成功软删除分支已迁移，`undress` Redis 锁/金豆扣减/第三方生成暂未接管 |
| `/starLive/index` | `c.api.starlive->index` | `StarLiveHandler.Index` | 已重构；直播初始化，兼容登录用户或游客 sid，读取 `starlive_info`，按 PHP AES-128-CBC/base64 生成 `encryptUid` 并返回嵌套 `data.data` |
| `/starLive/queryCoinBalance` | `c.api.starlive->queryCoinBalance` | `StarLiveHandler.QueryCoinBalance` | 已重构；直播平台余额查询 raw JSON 响应，游客长 memberId 返回 0，用户金币余额按 `goldcoin*10` 返回 |
| `/starLive/gameBet`、`/starLive/gameWin`、`/starLive/translate`、`/starLive/tryAgain` | `c.api.starlive->$action` | `StarLiveHandler.GameBet/GameWin/Translate/TryAgain` | 部分已重构；raw JSON 游客长 `memberId`、空/非法 memberId 和未知 `busiType` 安全失败分支已迁移，下注/中奖/钻石兑换资产事务和重复订单查询暂未接管 |
| `/getGlobalData` | `c.api.index->getGlobalData` | `IndexHandler.GetGlobalData` | 已重构；全局配置/版本/广告/弹窗/开关聚合，核心 key shape 和版本覆盖对比通过 |
| `/init` | `c.api.index->init` | `IndexHandler.Init` | 已重构；客户端初始化聚合，复用 globalData，登录/游客 user、appver、通知、邀请和站点配置 live 对比通过 |

### 需要登录但不需要验证码

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/ucp/myaff` | `c.api.ucp.index->myaff` | `UCPHandler.MyAff` | 已重构，对比通过；支持 `x-cookie-auth` header 和 `xxx_api_auth` cookie |
| `/ucp/index` | `c.api.ucp.index->index` | `UCPHandler.Index` | 已重构；登录/游客只读个人中心聚合，旧 PHP 本地对比超时，已按源码契约和 Go 输出验证 |
| `/ucp/affcenter` | `c.api.ucp.index->affcenter` | `UCPHandler.AffCenter` | 已重构，对比通过；登录只读推广中心，合并用户组权限后计算播放/下载剩余额度 |
| `GET /ucp/feedback` | `c.api.ucp.index->feedback` | `UCPHandler.FeedbackListing` | 已重构，对比通过；登录只读历史反馈列表 |
| `POST /ucp/feedback`、`/ucp/feedback/create` | `c.api.ucp.index->feedback`、`c.api.ucp.feedback->create` | `UCPHandler.FeedbackCreate` | 已重构；登录反馈创建、内容/订单/每日次数校验和写入 `feedbacks`；未登录分支 live 对比通过，图片上传保存和告警通知未接管 |
| `GET /ucp/feedback/index` | `c.api.ucp.feedback->index` | `UCPHandler.FeedbackIndex` | 已重构，对比通过；新版反馈初始化页，最近 30 天支付记录，POST 未接管 |
| `GET /ucp/feedback/listing` | `c.api.ucp.feedback->listing` | `UCPHandler.FeedbackNewListing` | 已重构，对比通过；新版反馈列表，支持 `type=0/1/2` 过滤，POST 未接管 |
| `GET /ucp/feedback/detail` | `c.api.ucp.feedback->detail` | `UCPHandler.FeedbackDetail` | 已重构，对比通过；新版反馈详情，校验反馈归属，附件图片和关联支付只读，POST 未接管 |
| `/ucp/feedback/:action?`（除已列 action） | `c.api.ucp.feedback->$action` | 不接管 | PHP `ucp/feedback.php` 仅定义 `index/create/listing/detail`，均已覆盖；图片上传保存和告警通知作为 `create` 外部链路后续增强 |
| `GET /ucp/msg`、`GET /ucp/msg/index` | `c.api.ucp.msg->index` | `UCPHandler.MsgListing` | 已重构，对比通过；登录只读消息会话列表，写状态 action 未接管 |
| `/ucp/msg/show` | `c.api.ucp.msg->show` | `UCPHandler.MsgDetail` | 已重构，对比通过；读取会话详情并复刻 setRead 已读副作用 |
| `/ucp/msg/send` | `c.api.ucp.msg->send` | `UCPHandler.MsgSend` | 已重构；会话内回复写入 `msgs/msg_maps/msgc/users.newmsg`，未登录分支 live 对比通过；按 PHP 源码变量遮蔽 bug 保持用户名群发不可用 |
| `/ucp/msg/setread` | `c.api.ucp.msg->setread` | `UCPHandler.MsgSetRead` | 已重构，对比通过；批量会话设为已读，空 `cids` 仍返回操作成功 |
| `/ucp/msg/cleanread` | `c.api.ucp.msg->cleanread` | `UCPHandler.MsgCleanRead` | 已重构，对比通过；清空当前用户 `newmsg` |
| `/ucp/msg/delete` | `c.api.ucp.msg->delete` | `UCPHandler.MsgDelete` | 已重构，对比通过；删除当前用户会话、消息映射并递减消息引用计数 |
| `/ucp/msg/:action?`（除已列 action） | `c.api.ucp.msg->$action` | 不接管 | PHP `ucp/msg.php` 仅定义 `index/show/send/setread/cleanread/delete`，均已覆盖 |
| `/ucp/payment`、`/ucp/payment/index` | `c.api.ucp.payment->index/listing` | `UCPHandler.PaymentListing` | 已重构，对比通过；兼容旧动态 action 默认入口 |
| `/ucp/payment/listing` | `c.api.ucp.payment->listing` | `UCPHandler.PaymentListing` | 已重构，对比通过；登录只读支付记录，支持 GET/POST page |
| `/ucp/payment/safepaylog` | `c.api.ucp.payment->safepaylog` | `UCPHandler.SafePayLog` | 已重构，对比通过；最近 7 天 safepay 记录 |
| `/ucp/payment/:action?`（除已列 action） | `c.api.ucp.payment->$action` | 不接管 | PHP `ucp/payment.php` 仅定义 `index/listing/safepaylog`，均已覆盖 |
| `/ucp/account`、`/ucp/account/index` | `c.api.ucp.account->index` | `UCPHandler.AccountIndex` | 已重构，对比通过；登录只读资产主页 |
| `/ucp/account/balancelog` | `c.api.ucp.account->balancelog` | `UCPHandler.BalanceLog` | 已重构，对比通过；登录只读余额日志分页 |
| `/ucp/account/:action?`（除已列 action） | `c.api.ucp.account->$action` | 不接管 | PHP `ucp/account.php` 仅定义 `index/balancelog`，均已覆盖 |
| `/ucp/withdraw`、`/ucp/withdraw/index` | `c.api.ucp.withdraw->index` | `UCPHandler.WithdrawIndex` | 已重构；登录只读提现初始化页，返回账户、收款地址、金币折算和提现配置 |
| `/ucp/withdraw/listing` | `c.api.ucp.withdraw->listing` | `UCPHandler.WithdrawListing` | 已重构；登录只读提现记录，返回 `rows/withdrawTotal/pageinfo`，金额和时间按旧 PHP `procRow` 格式化 |
| `/ucp/withdraw/rule` | `c.api.ucp.withdraw->rule` | `UCPHandler.WithdrawRule` | 已重构；公共只读提现规则，读取 `withdraw.rule` 的 html 内容 |
| `/ucp/coinlog`、`/ucp/coinlog/index` | `c.api.ucp.coinlog->index` | `UCPHandler.CoinLogIndex` | 已重构，对比通过；登录只读金币日志首页，最近 10 条 |
| `/ucp/coinlog/bonuslog` | `c.api.ucp.coinlog->bonuslog` | `UCPHandler.CoinLogBonusLog` | 已重构，对比通过；登录只读收益金币日志分页和累计统计 |
| `/ucp/coinlog/invitelog` | `c.api.ucp.coinlog->invitelog` | `UCPHandler.CoinLogInviteLog` | 已重构，对比通过；登录只读邀请金币日志分页 |
| `/ucp/user`、`/ucp/user/index` | `c.api.ucp.user->index` | `UCPHandler.UserIndex` | 已重构，对比通过；登录只读当前用户资料，复用 PHP user row 字段 |
| `/ucp/user/profile` | `c.api.ucp.user->profile` | `UCPHandler.UserProfile` | 已重构；未登录、昵称长度/字符集/白名单校验和资料写入成功分支已接管，写入 `users.gender/nickname` |
| `/ucp/user/passwd` | `c.api.ucp.user->passwd` | `UCPHandler.UserPasswd` | 已重构；未登录、原密码错误、密码长度、确认不一致、密码更新和重新登录成功分支已迁移，成功返回 `data.user/xxx_api_auth` |
| `/ucp/user/checkemail`、`/ucp/user/sendemail`、`/ucp/user/verifyemail`、`/ucp/user/bindmobi` | `c.api.ucp.user->$action` | `UCPHandler.UserCheckEmail/SendEmail/VerifyEmail/BindMobi` | 已重构；邮箱可用检查、SMTP 发送验证码、验证码 keylimit 写入、邮箱验证绑定和手机绑定成功分支均已接管 |
| `/ucp/bankcard`、`/ucp/bankcard/index` | `c.api.ucp.bankcard->index` | `UCPHandler.BankcardIndex` | 已重构，对比通过；登录只读提款地址和后台银行列表 |
| `/ucp/bankcard/create` | `c.api.ucp.bankcard->create` | `UCPHandler.BankcardCreate` | 已重构；登录新增提款地址，保留 PHP 的类型到支付宝/微信映射、最多 5 条判断和旧错误文案 |
| `/ucp/bankcard/modify` | `c.api.ucp.bankcard->modify` | `UCPHandler.BankcardModify` | 已重构；登录修改本人提款地址，缺失记录返回 `修改的记录不存在` |
| `/ucp/bankcard/delete` | `c.api.ucp.bankcard->delete` | `UCPHandler.BankcardDelete` | 已重构；登录删除本人提款地址，返回 `操作成功` |
| `/ucp/bankcard/:action?`（除已列 action） | `c.api.ucp.bankcard->$action` | 不接管 | PHP `ucp/bankcard.php` 仅定义 `index/create/modify/delete`，均已覆盖 |
| `/ucp/task`、`/ucp/task/index` | `c.api.ucp.task->index` | `UCPHandler.TaskIndex` | 已重构；登录只读任务中心聚合，统计分享、评论、收藏、观看、保存二维码、广告点击、小视频下载任务进度 |
| `/ucp/task/sharepic` | `c.api.ucp.task->sharepic` | `UCPHandler.TaskSharePic` | 已重构，对比通过；公共随机推广海报，只读无奖励写入 |
| `/ucp/task/qrlink` | `c.api.ucp.task->qrlink` | `UCPHandler.TaskQRLink` | 已重构，对比通过；登录只读推广二维码链接，读取推广 URL 和邀请码，不生成图片、不写 keylimit |
| `/ucp/task/invite` | `c.api.ucp.task->invite` | `UCPHandler.TaskInvite` | 已重构；未登录错误分支对齐，登录后按 PHP 空方法体返回 200 空 body |
| `/ucp/task/sign`、`/ucp/task/invitecodeInput`、`/ucp/task/adviewClick` | `c.api.ucp.task->$action` | `UCPHandler.TaskSign/TaskInviteCodeInput/TaskAdviewClick` | 已重构；`sign` 游客/登录成功路径已接管，登录成功事务写 `users_quota/user_coinlogs` 并返回 `data.taskdone`，游客成功更新 `user_guests.goldcoin/signtime`；`invitecodeInput/adviewClick` 已接管每日限制、参数失败和奖励成功分支 |
| `/ucp/task/share`、`/ucp/task/qrcode` | `c.api.ucp.task->$action` | `UCPHandler.TaskShare/TaskQRCode` | 已重构；`share` 支持游客随机邀请码、登录用户分享奖励和 `data.sharetext/taskdone`；`qrcode` 登录后写 `keylimits` 并返回 `image/png` 二维码 |
| `/ucp/task/qrcodeSave` | `c.api.ucp.task->qrcodeSave` | `UCPHandler.TaskQRCodeSave` | 已重构；未登录、今日已保存二维码和保存二维码奖励成功分支已迁移，成功事务写 `users_quota/user_coinlogs(cointype=6)` 并返回 `保存二维码已送金币: N` 与 `data.taskdone` |
| `/ucp/taskbox/index` | `c.api.ucp.taskbox->index` | `UCPHandler.TaskboxIndex` | 已重构，对比通过；公共只读任务宝箱状态和最近开启记录，领奖 action 已由 `/ucp/taskbox/taskboxopen` 接管 |
| `/ucp/taskbox/taskboxlog` | `c.api.ucp.taskbox->taskboxlog` | `UCPHandler.TaskboxLog` | 已重构，对比通过；登录只读本人任务宝箱日志，分页和日志行处理一致 |
| `/ucp/taskbox/share` | `c.api.ucp.taskbox->share` | `UCPHandler.TaskboxShare` | 已重构；公共只读任务宝箱分享文案，替换随机/登录邀请码和每日推广 URL，按 shape 对比通过 |
| `/ucp/taskbox/qrlink` | `c.api.ucp.taskbox->qrlink` | `UCPHandler.TaskboxQRLink` | 已重构，对比通过；登录只读任务宝箱推广二维码链接，不生成图片、不发奖励 |
| `/ucp/taskbox/taskboxopen` | `c.api.ucp.taskbox->taskboxopen` | `UCPHandler.TaskboxOpen` | 已重构；未登录、任务不存在/停用、宝箱赠送金币为 0、每日/每周宝箱时间窗、已领取、推广人数不足和领奖成功分支已迁移；成功分支事务写 `promotion_taskboxlogs/users_quota/user_coinlogs`，返回 `data.taskdone` |
| `/ucp/taskbox/qrcode` | `c.api.ucp.taskbox->qrcode` | `UCPHandler.TaskboxQRCode` | 已重构；登录后读取 `taskbox.qrcode.link` 和每日推广 URL，替换邀请码后生成 `image/png` 二维码，不写 keylimit/coinlog |
| `/ucp/upgrade` | `c.api.ucp.index->upgrade` | `UCPHandler.Upgrade` | 已重构；未登录、已经是尊贵会员、无效时长、终身 VIP 暂停升级、金币不足和升级成功分支已迁移，成功事务扣金币、写 `user_coinlogs(cointype=103)` 并更新 `users.sysgid/sysgid_exptime` |
| `/ucp/withdraw/create` | `c.api.ucp.withdraw->create` | `UCPHandler.WithdrawCreate` | 已重构；支持 JSON/form/query 输入，`withdraw_amount` 按 PHP `checkRMB` 以十进制定点从元转分；前置校验、普通提现金币转余额、提现单创建、免费提现额度扣减、普通/游戏余额冻结、冻结日志和 Telegram 通知均已接管 |
| `/ucp/coinlog/exchange` | `c.api.ucp.coinlog->exchange` | `UCPHandler.CoinLogExchange` | 已重构；兑换关闭、未登录、兑换类型、兑换数量、100 万上限、金币换人民币最小金币、计算为 0、金币转余额和余额转金币成功事务均已迁移 |
| `/ucp/vippkg`、`/ucp/vippkg/index` | `c.api.ucp.vippkg->index` | `UCPHandler.VIPPkgIndex` | 已重构，对比通过；登录只读 VIP 套餐列表和 safepayurl，支付通道通过接口隔离，默认不伪造旧 PHP 配置 |
| `/ucp/vippkg/coinorder` | `c.api.ucp.vippkg->coinorder` | `UCPHandler.VIPPkgCoinOrder` | 已重构；未登录、套餐不存在/停用、金币不足和金币购买 VIP 成功分支已迁移，成功事务扣金币、写 `user_coinlogs(cointype=103)` 并更新 `users.sysgid/sysgid_exptime` |
| `/ucp/vippkg/placeorder` | `c.api.ucp.vippkg->placeorder` | `UCPHandler.VIPPkgPlaceOrder` | 已重构；支付前置限制、随机通道解析、测试账号 `pay12.925` 特例和 `trade_payments(paytype=8)` 下单成功分支均已接管 |
| `/ucp/coinpkg`、`/ucp/coinpkg/index` | `c.api.ucp.coinpkg->index` | `UCPHandler.CoinPkgIndex` | 已重构，对比通过；登录只读金币套餐列表和 safepayurl，支付通道通过接口隔离 |
| `/ucp/coinpkg/placeorder` | `c.api.ucp.coinpkg->placeorder` | `UCPHandler.CoinPkgPlaceOrder` | 已重构；套餐校验、支付方式校验和 `trade_payments(paytype=7)` 金币套餐下单成功分支均已接管 |
| `/ucp/beanpkg`、`/ucp/beanpkg/index` | `c.api.ucp.beanpkg->index` | `UCPHandler.BeanPkgIndex` | 已重构，对比通过；登录只读金豆套餐列表和 safepayurl，支付通道通过接口隔离 |
| `/ucp/beanpkg/coinorder` | `c.api.ucp.beanpkg->coinorder` | `UCPHandler.BeanPkgCoinOrder` | 已重构；未登录、套餐不存在/停用、金币不足和金币兑换金豆成功分支已迁移，成功事务扣金币、增加金豆并写 `user_coinlogs/user_beanlogs` |
| `/ucp/beanpkg/placeorder` | `c.api.ucp.beanpkg->placeorder` | `UCPHandler.BeanPkgPlaceOrder` | 已重构；套餐限制、随机通道解析、支付方式校验和 `trade_payments(paytype=21)` 金豆套餐下单成功分支均已接管 |
| `/ucp/vodorder`、`/ucp/vodorder/index` | `c.api.ucp.vodorder->index` | `UCPHandler.VODOrderIndex` | 已重构；登录只读求片榜单，按当前期数返回榜单、top 助力人和本人助力数，不执行求片或助力写入 |
| `/ucp/vodorder/myorders` | `c.api.ucp.vodorder->myorders` | `UCPHandler.VODOrderMyOrders` | 已重构，对比通过；登录只读我的求片记录、累计消耗和当前冻结金币 |
| `/ucp/vodorder/mysupports` | `c.api.ucp.vodorder->mysupports` | `UCPHandler.VODOrderMySupports` | 已重构，对比通过；登录只读我的助力求片记录 |
| `/ucp/vodorder/historyorders` | `c.api.ucp.vodorder->historyorders` | `UCPHandler.VODOrderHistoryOrders` | 已重构，对比通过；登录只读成功的历史求片记录 |
| `/ucp/vodorder/create`、`/ucp/vodorder/support` | `c.api.ucp.vodorder->$action` | `UCPHandler.VODOrderCreate/VODOrderSupport` | 已重构；求片/助力前置校验、期数创建、金币扣减、金币流水、求片单创建、本人追加和他人助力写入成功分支均已接管 |
| `/vod/breaking` | `c.api.vod->breaking` | `VODHandler.Breaking` | 已重构，对比通过；公共只读每日爆料，返回当天 cateid=99 的 vodid/title |
| `/vod/errorreport`、`/v2/vod/errorreport` | `c.api.vod->errorreport`、`c.apiv2.vod->errorreport` | `VODHandler.ErrorReport` | 已重构；视频报错反馈写入 `vod_errors`，不涉及金币、支付或播放权限 |

### 个人中心公开只读

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/ucp/rolltitle` | `c.api.ucp.index->rolltitle` | `UCPHandler.RollTitle` | 已重构，对比通过；公共只读，返回滚动消息 |

### Go 服务自有接口

| 接口 | 说明 |
| --- | --- |
| `/healthz` | Go 服务健康检查，不是 PHP 旧接口 |
| `/readyz` | Go 服务就绪检查，不是 PHP 旧接口 |

## 未重构接口

### 非 v2 视频接口

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/vod/reqplay/:vodid`、`/vod/reqdown/:vodid` 的扣费/奖励分支 | `c.api.vod->reqplay/reqdown` | 部分未重构；免费/限免、已观看/下载和权限额度内非扣费成功路径的播放/下载日志与 `vods` 计数已迁移；剩余超限扣金币、扣费日志、播放/下载任务奖励、推荐奖励仍需事务化迁移 |

### 小视频、作者页

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/minivod/reqplay/:vodid`、`/minivod/reqdown/:vodid` 的扣费/任务奖励分支 | `c.api.minivod->$action` | 部分未重构；免费/限免、已观看/下载和权限额度内非扣费成功路径的 viewlog 与 `vods` 计数已迁移；剩余超限扣金币、扣费标记、播放任务、推荐奖励和喜好分析仍需事务化迁移 |

### 用户账号

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/register`、`/v2/register` | `c.api.user->register`、`c.apiv2.user->register` | 部分未重构；未同意协议、已登录、注册关闭、IP 频控、手机号/邮箱/用户名格式和查重等失败分支已迁移，验证码、成功注册、邀请奖励和写库仍未迁移 |
| `/login`、`/v2/login` | `c.api.user->login`、`c.apiv2.user->login` | 部分未重构；账号密码登录、v2 手机/邮箱验证码登录、session 写入、quota/goldbean 返回和 Redis 注销标记清理已迁移；v1 短信登录不存在用户时自动注册、注册联动邀请/奖励仍未迁移 |

### 支付和回调

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/payment/:action`（除已列 payment action） | `c.api.payment->$action` | 部分未重构；`reqpay/pay12req` 失败分支、known payway 支付方式不允许、unknown payway 选择支付方式返回、常见 public 成功回调文案和支付页面已迁移，钱包支付扣费、第三方网关请求、订单状态写入和成功跳转仍需 provider 配置与网关接口 |
| `/respond/:action` 成功验签/入账分支 | `c.respond.*` | 阻断未重构；provider registry/verifier 骨架已落，当前只允许缺配置、验签失败、验签成功但 accounting disabled 返回 provider `echoErr()`，绝不返回 OK/SUCCESS；后续仍需要 provider 密钥/RSA 配置补齐、`trade_payments SELECT ... FOR UPDATE` 锁单、幂等入账和 `payment->doAction()` 事务 |
| `/respond/chan1` | `c.respond.chan1` | 部分未重构；token 校验失败、用户不存在、已送过会员、套餐不存在或停用分支已迁移，合法 token 后下单、充值入账、支付记录变更和赠送 VIP 仍需事务化迁移 |

### 个人中心

按 `/Users/canavs/xjProj/XJBackend/api/src/c/api/ucp/*.php` 当前可达业务 action 核对，UCP 个人中心 70 个 action 已全部注册并迁移；动态 `:action?` 中 PHP 未定义的 action 不接管。

### 活动、邀请、发现页

按 PHP `c.api.invite` 当前可达 action 核对，`/invite/info` 和 `/invite/bind` 已全部迁移；动态 `:action?` 中 PHP 未定义的 action 不接管。

### 游戏、直播、一元购

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/game/wali/topup` | `c.api.game.wali->topup` | 部分未重构；未登录、低于最低转入金币和余额不足分支已迁移，上分金币扣减、外部平台请求和失败归还金币仍需事务化迁移 |
| `/game/wali/withdraw` | `c.api.game.wali->withdraw` | 部分未重构；未登录和金额输入不正确分支已迁移，下分外部平台请求、金币增加和订单写入仍需迁移 |
| `/game/lottery/topup`、`/game/lottery/withdraw` | `c.api.game.lottery->$action` | 部分未重构；topup/withdraw 的未登录、最低转入金币、余额不足和金额输入不正确分支已迁移，彩票平台上下分订单和金币写入仍需迁移 |
| `/starLive/:action`（除 `/starLive/index`、`/starLive/queryCoinBalance`、`/starLive/gameBet`、`/starLive/gameWin`、`/starLive/translate`、`/starLive/tryAgain`） | `c.api.starlive->$action` | 部分未重构；资产 action 的游客/未知用户/未知业务类型前置失败分支已迁移，重复订单查询、下注扣款、中奖加钱和钻石兑换事务仍需迁移 |

### 社区、HGame、AI

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/aiundress/undress` 的成功外部调用分支 | `c.api.aiundress->undress` | 部分未重构；未登录分支、`upload` 成功写入分支、`undress` 无效图片路径和 `delete` 空记录/成功软删除分支已迁移，Redis 锁、第三方 AI 和金豆扣减仍需迁移 |

### 图片、附件和通配资源

| 接口 | PHP handler | 备注 |
| --- | --- | --- |

## 建议后续顺序

1. 高风险写入接口：评论发布、收藏新增、点赞踩、播放/下载授权、购买、任务奖励。
2. 资产/外部平台接口：支付、金币/金豆、提现、游戏上分/下分、AI 上传/生成、直播平台。

## 当前验证命令

```shell
go test ./...
go vet ./...
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/xj-comp-api ./cmd/api
make ci
```
