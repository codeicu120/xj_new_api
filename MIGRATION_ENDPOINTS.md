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
| `/sms`、`/sms/index`、`/email`、`/email/index` | ANY | `handler.EmptyHTML` |
| `/sms/sendv`、`/sms/sendu`、`/email/send` | ANY | `VerificationHandler` |
| `/captcha/req` | ANY | `CaptchaHandler.Req` |
| `/captcha/pic`、`/captcha/picx` | ANY | `CaptchaHandler.Pic/PicX` |
| `/test` | ANY | `TestHandler.Test` |
| `/iploc/:ip` | ANY | `IPLocHandler.Find` |
| `/game/platforms` | ANY | `GameHandler.Platforms` |
| `/game/categories` | ANY | `GameHandler.Categories` |
| `/game/games` | ANY | `GameHandler.Games` |
| `/game/broadcasts` | ANY | `GameHandler.Broadcasts` |
| `/game/wali/gameList` | ANY | `GameHandler.WaliGames` |
| `/game/wali/test` | ANY | `GameHandler.WaliTest` |
| `/game/wali/balance` | ANY | `GameHandler.WaliBalance` |
| `/game/lottery/gameList` | ANY | `GameHandler.LotteryGames` |
| `/hgame/index` | ANY | `HGameHandler.Index` |
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
| `/payment/unpaid`、`/payment/success`、`/payment/failed` | ANY | `PaymentHandler.Unpaid/Success/Failed` |
| `/bought/listing`、`/bought/delete` | ANY | `BoughtHandler.Listing/Delete` |
| `/playlog`、`/playlog/index`、`/downlog`、`/downlog/index` | ANY | `handler.EmptyHTML` |
| `/playlog/listing`、`/playlog/remove`、`/downlog/listing`、`/downlog/remove` | ANY | `HistoryHandler` |
| `/miniplaylog/listing`、`/miniplaylog/remove` | ANY | `HistoryHandler` |
| `/favorite`、`/favorite/index`、`/minifavorite`、`/minifavorite/index` | ANY | `handler.EmptyHTML` |
| `/favorite/listing`、`/favorite/remove`、`/minifavorite/listing`、`/minifavorite/remove` | ANY | `FavoriteHandler` |
| `/minivod/listing`、`/minivod/recommend`、`/minivod/hot`、`/minivod/latest`、`/minivod/topzan`、`/minivod/topcomment`、`/minivod/topplay`、`/minivod/topcoin`、`/minivod/topnew`、`/minivod/topday`、`/minivod/topweek`、`/minivod/topmonth` | ANY | `MiniVODHandler.Listing` |
| `/minivod/*-:params`（上述 action） | ANY | `MiniVODHandler.Listing` |
| `/minivod/show/:vodid` | ANY | `MiniVODHandler.Show` |
| `/my/:authorid`、`/my/:authorid/:action` | ANY | `MiniVODHandler.Author` |
| `/community/list`、`/community/recommend`、`/community/hot`、`/community/latest`、`/community/favorite` | ANY | `CommunityHandler.Listing` |
| `/community/*-:params`（上述 action） | ANY | `CommunityHandler.Listing` |
| `/community/clisting`、`/community/clisting-:params` | ANY | `CommunityHandler.CommentListing` |
| `/explore/index` | ANY | `ExploreHandler.Index` |
| `/explore/notification`、`/explore/notification/index` | ANY | `ExploreHandler.EmptyOK` |
| `/explore/notification/clean` | ANY | `ExploreHandler.CleanNotification` |
| `/explore/signtask`、`/explore/signtask/index` | ANY | `ExploreHandler.EmptyOK` |
| `/explore/vodtask`、`/explore/vodtask/index` | ANY | `ExploreHandler.EmptyOK` |
| `/aiundress`、`/aiundress/listing` | ANY | `AIUndressHandler.Listing` |
| `/aiundress/index` | ANY | `handler.EmptyHTML` |
| `/getCertUuid` | ANY | `IndexHandler.GetCertUUID` |
| `/getGlobalData` | ANY | `IndexHandler.GetGlobalData` |
| `/init` | ANY | `IndexHandler.Init` |
| `/ucp/index` | ANY | `UCPHandler.Index` |
| `/ucp/user`、`/ucp/user/index` | ANY | `UCPHandler.UserIndex` |
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
| `/ucp/task/sharepic` | ANY | `UCPHandler.TaskSharePic` |
| `/ucp/task/qrlink` | ANY | `UCPHandler.TaskQRLink` |
| `/ucp/taskbox/index`、`/ucp/taskbox/taskboxlog`、`/ucp/taskbox/qrlink` | ANY | `UCPHandler.TaskboxIndex/TaskboxLog/TaskboxQRLink` |
| `/ucp/affcenter` | ANY | `UCPHandler.AffCenter` |
| `/ucp/payment`、`/ucp/payment/index`、`/ucp/payment/listing` | ANY | `UCPHandler.PaymentListing` |
| `/ucp/payment/safepaylog` | ANY | `UCPHandler.SafePayLog` |
| `/ucp/account`、`/ucp/account/index` | ANY | `UCPHandler.AccountIndex` |
| `/ucp/account/balancelog` | ANY | `UCPHandler.BalanceLog` |
| `/ucp/withdraw`、`/ucp/withdraw/index` | ANY | `UCPHandler.WithdrawIndex` |
| `/ucp/coinlog`、`/ucp/coinlog/index` | ANY | `UCPHandler.CoinLogIndex` |
| `/ucp/coinlog/bonuslog` | ANY | `UCPHandler.CoinLogBonusLog` |
| `/ucp/coinlog/invitelog` | ANY | `UCPHandler.CoinLogInviteLog` |
| `/ucp/vippkg`、`/ucp/vippkg/index` | ANY | `UCPHandler.VIPPkgIndex` |
| `/ucp/coinpkg`、`/ucp/coinpkg/index` | ANY | `UCPHandler.CoinPkgIndex` |
| `/ucp/beanpkg`、`/ucp/beanpkg/index` | ANY | `UCPHandler.BeanPkgIndex` |
| `/ucp/vodorder/myorders`、`/ucp/vodorder/mysupports`、`/ucp/vodorder/historyorders` | ANY | `UCPHandler.VODOrderMyOrders/MySupports/HistoryOrders` |
| `/vod/show/:vodid` | ANY | `VODHandler.Show` |
| `/vod/up/:vodid`、`/vod/down/:vodid` | ANY | `VODHandler.Up/Down` |
| `/vod/breaking` | ANY | `VODHandler.Breaking` |
| `/vod/preView/:vodid/index.m3u8` | ANY | `VODHandler.Preview` |
| `/sendfile/play/:file`、`/sendfile/down/:file` | ANY | `SendfileHandler.Play/Down` |
| `/comment`、`/comment/index` | ANY | `handler.EmptyHTML` |
| `/comment/listing-:params` | ANY | `CommentHandler.Listing` |
| `/comment/up`、`/comment/down` | ANY | `CommentHandler.Up/Down` |
| `/special/index` | ANY | `SpecialHandler.Index` |
| `/special/listing`、`/special/listing-:params` | ANY | `SpecialHandler.Listing` |
| `/special/detail/:spid`、`/special/detail/:spid-:params` | ANY | `SpecialHandler.Detail` |
| `/special/up/:spid`、`/special/down/:spid` | ANY | `SpecialHandler.Up/Down` |
| `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last`、`/onego/hash`、`/onego/history`、`/onego/lucky`、`/onego/bet_ranks`、`/onego/marquee` | ANY | `OneGoHandler` |
| `/vod/listing`、`/vod/recommend`、`/vod/hot`、`/vod/latest` | ANY | `VODHandler.Listing` |
| `/vod/listing-:params`、`/vod/recommend-:params`、`/vod/hot-:params`、`/vod/latest-:params` | ANY | `VODHandler.Listing` |
| `/v2/amazing/categories` | ANY | `AmazingHandler.Categories` |
| `/v2/amazing/listing`、`/v2/amazing/recommend`、`/v2/amazing/hot`、`/v2/amazing/latest` | ANY | `AmazingHandler.Listing` |
| `/v2/amazing/listing-:params`、`/v2/amazing/recommend-:params`、`/v2/amazing/hot-:params`、`/v2/amazing/latest-:params` | ANY | `AmazingHandler.Listing` |
| `/v2/so/list` | ANY | `SOHandler.List` |
| `/v2/vod/listing`、`/v2/vod/recommend`、`/v2/vod/hot`、`/v2/vod/latest` | ANY | `VODHandler.Listing` |
| `/v2/vod/listing-:params`、`/v2/vod/recommend-:params`、`/v2/vod/hot-:params`、`/v2/vod/latest-:params` | ANY | `VODHandler.Listing` |
| `/v2/vod/show/:vodid` | ANY | `VODHandler.Show` |
| `/v2/vod/up/:vodid`、`/v2/vod/down/:vodid` | ANY | `VODHandler.Up/Down` |

### 已注册占位

| 接口 | Method | Go handler |
| --- | --- | --- |
| `/v2/register` | ANY | `notImplemented("c.apiv2.user.register")` |
| `/v2/login` | ANY | `notImplemented("c.apiv2.user.login")` |
| `/v2/forgot` | ANY | `notImplemented("c.apiv2.user.forgot")` |
| `/v2/vod/reqplay/:vodid`、`/v2/vod/reqdown/:vodid` | ANY | `notImplemented("c.apiv2.vod.reqplay/reqdown")` |
| `/v2/vod/buy/:vodid` | ANY | `notImplemented("c.apiv2.vod.buy")` |

## 已重构接口

### 基础与公共接口

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/sysavatar` | `c.api.user->sysavatar` | `UserHandler.SysAvatar` | 已重构，对比通过 |
| `/logout` | `c.api.user->logout` | `UserHandler.Logout` | 已重构，对比通过；删除 type=0 session，非法/无 token 仍返回已退出 |
| `/sms`、`/sms/index`、`/email`、`/email/index` | `c.api.sms/email->index` | `handler.EmptyHTML` | 已重构，对比通过；默认空入口返回 `200 text/html` 空 body |
| `/sms/sendv`、`/sms/sendu`、`/email/send` | `c.api.sms/email->send*` | `VerificationHandler` | 已重构；手机号/邮箱/未登录错误分支 live 对比通过，成功发送通过 sender/captcha/limiter fake 覆盖，默认不直连真实短信/邮件平台 |
| `/captcha/req` | `c.api.captcha->req` | `CaptchaHandler.Req` | 已重构，动态 secret 按 shape 对比通过 |
| `/captcha/pic`、`/captcha/picx` | `c.api.captcha->pic/picx` | `CaptchaHandler.Pic/PicX` | 已重构；无效 secret 404 JSON 对比通过，有效 PHP secret 和 Go req secret 均输出 100x34 PNG |
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
| `/payment/unpaid` | `c.api.payment->unpaid` | `PaymentHandler.Unpaid` | 已重构；旧 PHP 当前直接返回 `data.total_count=0`，后续未执行的 24 小时未支付查询分支不接管 |
| `/payment/success`、`/payment/failed` | `c.api.payment->success/failed` | `PaymentHandler.Success/Failed` | 已重构；固定支付状态 JSON 文案，不包含平台回调验签 |
| `/bought/listing` | `c.api.bought->listing` | `BoughtHandler.Listing` | 已重构，对比通过；登录只读已购影片列表，复用 VOD 行处理和 PHP 分页 |
| `/bought/delete` | `c.api.bought->delete` | `BoughtHandler.Delete` | 已重构，对比通过；登录删除已购影片记录，空 `vodids` 成功 |
| `/comment`、`/comment/index` | `c.api.comment->index` | `handler.EmptyHTML` | 已重构；旧 PHP 空方法，返回 `200 text/html` 空 body |
| `/explore/notification`、`/explore/notification/index` | `c.api.explore.notification->index` | `ExploreHandler.EmptyOK` | 已重构，对比通过；旧 PHP 空 OK，动态 `xxx_api_auth` 不回传 |
| `/explore/notification/:action?`（除 `/explore/notification`、`/explore/notification/index`、`/explore/notification/clean`） | `c.api.explore.notification->$action` | 不接管 | PHP `notification` 仅定义 `index/clean`，未发现其他稳定 action |
| `/explore/signtask`、`/explore/signtask/index` | `c.api.explore.signtask->index` | `ExploreHandler.EmptyOK` | 已重构，对比通过；旧 PHP 空 OK，签到写入 action 未接管 |
| `/explore/vodtask`、`/explore/vodtask/index` | `c.api.explore.vodtask->index` | `ExploreHandler.EmptyOK` | 已重构，对比通过；旧 PHP 空 OK，reqcoin 未接管 |
| `/explore/vodtask/show/:vid` | `c.api.explore.vodtask->show` | `ExploreHandler.VodTaskShow` | 已重构；激励视频展示并创建/复用当日领取日志，错误分支 live 对比通过，成功分支 fake 覆盖；`reqcoin` 资产发放仍未接管 |
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
| `/game/lottery/gameList` | `c.api.game.lottery->gameList` | `GameHandler.LotteryGames` | 已重构；彩票普通分类只读列表，`category_id=5` 游客未登录分支已对齐 |
| `/hgame/index` | `c.api.hgame->index` | `HGameHandler.Index` | 已重构，对比通过；HGame 公共只读列表，`/hgame` 保持旧 PHP 404 未接管 |
| `/hgame/:action`（除 `/hgame/index`） | `c.api.hgame->$action` | 不接管 | PHP `c.api.hgame` 仅定义 `index`，未发现其他稳定 action；不伪造业务响应 |
| `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last` | `c.api.onego->rules/rooms/current/last` | `OneGoHandler` | 已重构，对比通过；一元购公共只读规则/房间/当前期数/上期记录，旧 PHP 动态 `xxx_api_auth` 忽略 |
| `/onego/hash` | `c.api.onego->hash` | `OneGoHandler.Hash` | 已重构；公共哈希计算接口，复刻 SHA256 后提取末尾数字期号规则 |
| `/onego/history` | `c.api.onego->history` | `OneGoHandler.History` | 已重构，对比通过；登录只读本人投注历史，未登录 `retcode=-9999` |
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
| `/v2/vod/listing` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/v2/vod/listing-:params` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/v2/vod/recommend` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构，随机列表按 shape 对比 |
| `/v2/vod/recommend-:params` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构 |
| `/v2/vod/hot` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/v2/vod/hot-:params` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构 |
| `/v2/vod/latest` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构，对比通过 |
| `/v2/vod/latest-:params` | `c.apiv2.vod->listing` | `VODHandler.Listing` | 已重构 |
| `/v2/vod/show/:vodid` | `c.apiv2.vod->show` | `VODHandler.Show` | 已重构，对比通过；复用视频详情实现 |

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
| `/vod/preView/:vodid/index.m3u8` | `c.api.vod->preView` | `VODHandler.Preview` | 已重构，m3u8 输出对比通过 |
| `/sendfile/play/:file` | `c.api.sendfile->play` | `SendfileHandler.Play` | 已重构，按旧 PHP 空壳行为对齐 |
| `/sendfile/down/:file` | `c.api.sendfile->down` | `SendfileHandler.Down` | 已重构，按旧 PHP 空响应对齐 |

### 评论、收藏、播放/下载记录

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/comment/listing-:params` | `c.api.comment->listing` | `CommentHandler.Listing` | 已重构，对比通过 |
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
| `/minivod/reqlong/:vodid` | `c.api.minivod->getLong2Mini` | `MiniVODHandler.ReqLong` | 已重构；普通长视频转小视频播放地址，支持 CDN 签名/播放服务器 host 补全；错误分支和本地样本成功 URL live 对比通过 |
| `/miniplaylog/listing` | `c.api.minivod->history` | `HistoryHandler.MiniPlayListing` | 已重构；不强制登录，登录/游客按小视频分表读取，mini 行处理和相对时间格式 |
| `/miniplaylog/remove` | `c.api.minivod->historyDelete` | `HistoryHandler.MiniPlayRemove` | 已重构；按 PHP 模型语义用输入 `vodid/vodids` 删除 `logid`，空参数 live 对比通过 |
| `/my/:authorid`、`/my/:authorid/index`、`/my/:authorid/listing` | `c.api.my->index/listing` | `MiniVODHandler.Author` | 已重构，对比通过；作者主页小视频列表，返回 `now/userrow/vodrows/pageinfo/orders` |

### 社区、HGame、AI

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/community/list`、`/community/recommend`、`/community/hot`、`/community/latest` | `c.api.topic->list` | `CommunityHandler.Listing` | 已重构，对比通过；主题列表、推荐/热门/最新、分类/type/分页和媒体字段 |
| `/community/favorite` | `c.api.topic->list` | `CommunityHandler.Listing` | 已重构，对比通过；未登录 `retcode=-9999`，登录按 `topic_favorites` 过滤 |
| `/community/*-:params`（上述 action） | `c.api.topic->list` | `CommunityHandler.Listing` | 已重构；参数模板 `$category_id-$type-$orderby-$page` |
| `/community/clisting`、`/community/clisting-:params` | `c.api.topic->clisting` | `CommunityHandler.CommentListing` | 已重构，对比通过；评论树列表，`tid` 不存在分支一致 |

### 搜索、专题、公告

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/special/index` | `c.api.special->index` | `SpecialHandler.Index` | 已重构，对比通过；旧 PHP 空方法，返回 text/html 空 body |
| `/special/listing`、`/special/listing-:params` | `c.api.special->listing` | `SpecialHandler.Listing` | 已重构，对比通过；专题列表，含前 4 个视频、分页和第一页 actorrows |
| `/special/detail/:spid`、`/special/detail/:spid-:params` | `c.api.special->detail` | `SpecialHandler.Detail` | 已重构，对比通过；专题详情，含全量视频排序和浏览数写入副作用 |
| `/special/up/:spid`、`/special/down/:spid` | `c.api.special->up/down` | `SpecialHandler.Up/Down` | 已重构；不存在分支对比通过，成功和重复投票分支由 service fake 覆盖 |
| `/search` | `c.api.search->index` | `VODHandler.Search` | 已重构，对比通过；空关键词搜索页和关键词结果列表，保留搜索统计写入 |
| `/minisearch` | `c.api.miniSearch->index` | `VODHandler.MiniSearch` | 已重构，对比通过；小视频搜索，保留 `showtype=1`、`/minivod` URL、`minivod_schlogs` 搜索日志写入和旧分页 URL |
| `/art`、`/art/index` | `c.api.art->index` | `ArtHandler.Index` | 已重构，对比通过；旧 PHP 空方法，返回 text/html 空 body |
| `/art/announce` | `c.api.art->announce` | `ArtHandler.Announce` | 已重构，对比通过；公告列表，保留旧 PHP 未定义 `$action` 导致的 `/art/?page=[?]` 分页 URL |
| `/art/show` | `c.api.art->show` | `ArtHandler.Show` | 已重构，对比通过；公告/文章详情，成功和不存在错误分支一致 |
| `/aiundress`、`/aiundress/listing` | `c.api.aiundress->listing` | `AIUndressHandler.Listing` | 已重构，对比通过；登录只读 AI 任务历史，支持 `module/page`，未登录 `retcode=-1` |
| `/aiundress/index` | `c.api.aiundress->index` | `handler.EmptyHTML` | 已重构，对比通过；按本地旧 PHP 运行时行为返回 `200 text/html` 空 body，AI 业务 action 未接管 |
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
| `/ucp/coinlog`、`/ucp/coinlog/index` | `c.api.ucp.coinlog->index` | `UCPHandler.CoinLogIndex` | 已重构，对比通过；登录只读金币日志首页，最近 10 条 |
| `/ucp/coinlog/bonuslog` | `c.api.ucp.coinlog->bonuslog` | `UCPHandler.CoinLogBonusLog` | 已重构，对比通过；登录只读收益金币日志分页和累计统计 |
| `/ucp/coinlog/invitelog` | `c.api.ucp.coinlog->invitelog` | `UCPHandler.CoinLogInviteLog` | 已重构，对比通过；登录只读邀请金币日志分页 |
| `/ucp/user`、`/ucp/user/index` | `c.api.ucp.user->index` | `UCPHandler.UserIndex` | 已重构，对比通过；登录只读当前用户资料，复用 PHP user row 字段 |
| `/ucp/bankcard`、`/ucp/bankcard/index` | `c.api.ucp.bankcard->index` | `UCPHandler.BankcardIndex` | 已重构，对比通过；登录只读提款地址和后台银行列表 |
| `/ucp/bankcard/create` | `c.api.ucp.bankcard->create` | `UCPHandler.BankcardCreate` | 已重构；登录新增提款地址，保留 PHP 的类型到支付宝/微信映射、最多 5 条判断和旧错误文案 |
| `/ucp/bankcard/modify` | `c.api.ucp.bankcard->modify` | `UCPHandler.BankcardModify` | 已重构；登录修改本人提款地址，缺失记录返回 `修改的记录不存在` |
| `/ucp/bankcard/delete` | `c.api.ucp.bankcard->delete` | `UCPHandler.BankcardDelete` | 已重构；登录删除本人提款地址，返回 `操作成功` |
| `/ucp/bankcard/:action?`（除已列 action） | `c.api.ucp.bankcard->$action` | 不接管 | PHP `ucp/bankcard.php` 仅定义 `index/create/modify/delete`，均已覆盖 |
| `/ucp/task/sharepic` | `c.api.ucp.task->sharepic` | `UCPHandler.TaskSharePic` | 已重构，对比通过；公共随机推广海报，只读无奖励写入 |
| `/ucp/task/qrlink` | `c.api.ucp.task->qrlink` | `UCPHandler.TaskQRLink` | 已重构，对比通过；登录只读推广二维码链接，读取推广 URL 和邀请码，不生成图片、不写 keylimit |
| `/ucp/taskbox/index` | `c.api.ucp.taskbox->index` | `UCPHandler.TaskboxIndex` | 已重构，对比通过；公共只读任务宝箱状态和最近开启记录，领奖 action 未接管 |
| `/ucp/taskbox/taskboxlog` | `c.api.ucp.taskbox->taskboxlog` | `UCPHandler.TaskboxLog` | 已重构，对比通过；登录只读本人任务宝箱日志，分页和日志行处理一致 |
| `/ucp/taskbox/qrlink` | `c.api.ucp.taskbox->qrlink` | `UCPHandler.TaskboxQRLink` | 已重构，对比通过；登录只读任务宝箱推广二维码链接，不生成图片、不发奖励 |
| `/ucp/vippkg`、`/ucp/vippkg/index` | `c.api.ucp.vippkg->index` | `UCPHandler.VIPPkgIndex` | 已重构，对比通过；登录只读 VIP 套餐列表和 safepayurl，支付通道通过接口隔离，默认不伪造旧 PHP 配置 |
| `/ucp/coinpkg`、`/ucp/coinpkg/index` | `c.api.ucp.coinpkg->index` | `UCPHandler.CoinPkgIndex` | 已重构，对比通过；登录只读金币套餐列表和 safepayurl，支付通道通过接口隔离 |
| `/ucp/beanpkg`、`/ucp/beanpkg/index` | `c.api.ucp.beanpkg->index` | `UCPHandler.BeanPkgIndex` | 已重构，对比通过；登录只读金豆套餐列表和 safepayurl，支付通道通过接口隔离 |
| `/ucp/vodorder/myorders` | `c.api.ucp.vodorder->myorders` | `UCPHandler.VODOrderMyOrders` | 已重构，对比通过；登录只读我的求片记录、累计消耗和当前冻结金币 |
| `/ucp/vodorder/mysupports` | `c.api.ucp.vodorder->mysupports` | `UCPHandler.VODOrderMySupports` | 已重构，对比通过；登录只读我的助力求片记录 |
| `/ucp/vodorder/historyorders` | `c.api.ucp.vodorder->historyorders` | `UCPHandler.VODOrderHistoryOrders` | 已重构，对比通过；登录只读成功的历史求片记录 |
| `/vod/breaking` | `c.api.vod->breaking` | `VODHandler.Breaking` | 已重构，对比通过；公共只读每日爆料，返回当天 cateid=99 的 vodid/title |

### 个人中心公开只读

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/ucp/rolltitle` | `c.api.ucp.index->rolltitle` | `UCPHandler.RollTitle` | 已重构，对比通过；公共只读，返回滚动消息 |

### Go 服务自有接口

| 接口 | 说明 |
| --- | --- |
| `/healthz` | Go 服务健康检查，不是 PHP 旧接口 |
| `/readyz` | Go 服务就绪检查，不是 PHP 旧接口 |

## Go 已注册但仍是占位

这些路由在 Go 中已经注册，但当前返回 `not_implemented`，业务还没有迁移。

| 接口 | 旧 PHP handler |
| --- | --- |
| `/v2/register` | `c.apiv2.user->register` |
| `/v2/login` | `c.apiv2.user->login` |
| `/v2/forgot` | `c.apiv2.user->forgot` |
| `/v2/vod/reqplay/:vodid` | `c.apiv2.vod->reqplay` |
| `/v2/vod/reqdown/:vodid` | `c.apiv2.vod->reqdown` |
| `/v2/vod/buy/:vodid` | `c.apiv2.vod->buy` |

## 未重构接口

### 首页、全局配置和工具类

| 接口 | PHP handler | 备注 |
| --- | --- | --- |

### 非 v2 视频接口

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/vod/reqplay/:vodid`、`/vod/reqdown/:vodid` | `c.api.vod->reqplay/reqdown` | 未重构；播放/下载权限、日志、可能签名 |
| `/vod/buy/:vodid` | `c.api.vod->buy` | 未重构；购买/金币 |
| `/vod/:action?`（除已列 action） | `c.api.vod->$action` | 未重构；剩余 `reqplay/reqdown/buy/errorreport` 等涉及播放权限、购买、反馈写入或媒体处理 |

### 评论、收藏、播放/下载记录

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/comment/post` | `c.api.comment->post` | 未重构；评论发布涉及敏感词、树结构、金币奖励和通知 |

### 小视频、作者页

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/minivod/reqlist` | `c.api.minivod->reqlist` | 未重构 |
| `/minivod/reqcoin` | `c.api.minivod->reqcoin` | 未重构 |
| `/minivod/reqplay/:vodid`、`/minivod/reqdown/:vodid` | `c.api.minivod->$action` | 未重构；播放/下载权限 |
| `/minivod/throwcoin/:vodid` | `c.api.minivod->throwcoin` | 未重构；金币打赏 |
| `/minivod/parselong/:vodid/index.m3u8` | `c.api.minivod->parseM3u8` | 未重构；媒体解析 |

### 用户账号

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/register` | `c.api.user->register` | 未重构；验证码、注册风控、写库 |
| `/login` | `c.api.user->login` | 未重构；密码/短信登录、session |
| `/forgot` | `c.api.user->forgot` | 未重构；验证码、密码重置 |
| `/delete` | `c.api.user2->delAccount` | 未重构；账号注销 |
| `/changePhone` | `c.api.user2->changePhone` | 未重构；手机换绑、验证码 |

### 支付和回调

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/payment/:action`（除 `/payment/unpaid`、`/payment/success`、`/payment/failed`） | `c.api.payment->$action` | 未重构；剩余 `index/query/payways/chpayway/reqpay` 等涉及支付查询、下单、支付方式切换或平台请求 |
| `/respond/:action` | `c.respond.*` | 未重构；支付平台回调 |
| `/respond/shangfu`、`/respond/wappay1`、`/respond/wappay2`、`/respond/wappay3`、`/respond/wappay4`、`/respond/wappay5` | `c.respond.*` | 未重构 |
| `/respond/hawpay`、`/respond/easypay`、`/respond/chan1`、`/respond/pay6`、`/respond/pay7` | `c.respond.*` | 未重构 |

### 个人中心

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/ucp/upgrade` | `c.api.ucp.index->upgrade` | 未重构；会员升级/金币 |
| `/ucp/user/:action?`（除 `/ucp/user`、`/ucp/user/index`） | `c.api.ucp.user->$action` | 未重构；资料修改、密码、邮箱/手机绑定等写入或验证码相关 |
| `/ucp/task/:action?`（除 `/ucp/task/sharepic`、`/ucp/task/qrlink`） | `c.api.ucp.task->$action` | 未重构；任务奖励/签到、二维码图片生成或 keylimit 写入等 |
| `/ucp/withdraw/:action?`（除 `/ucp/withdraw`、`/ucp/withdraw/index`） | `c.api.ucp.withdraw->$action` | 未重构；剩余 `create` 为提现写入，涉及账户余额、银行卡、风控和事务 |
| `/ucp/coinlog/:action?`（除 `/ucp/coinlog`、`/ucp/coinlog/index`、`/ucp/coinlog/bonuslog`、`/ucp/coinlog/invitelog`） | `c.api.ucp.coinlog->$action` | 未重构；`exchange` 为金币兑换写入高风险 |
| `/ucp/taskbox/:action?`（除 `/ucp/taskbox/index`、`/ucp/taskbox/taskboxlog`、`/ucp/taskbox/qrlink`） | `c.api.ucp.taskbox->$action` | 未重构；`/ucp/taskbox` 本身旧 PHP 无稳定响应未接管，`taskboxopen/share/qrcode` 涉及奖励写入或图片生成 |
| `/ucp/vippkg/:action?`（除 `/ucp/vippkg`、`/ucp/vippkg/index`） | `c.api.ucp.vippkg->$action` | 未重构；`placeorder/coinorder` 涉及支付下单、金币兑换和会员资产 |
| `/ucp/coinpkg/:action?`（除 `/ucp/coinpkg`、`/ucp/coinpkg/index`） | `c.api.ucp.coinpkg->$action` | 未重构；`placeorder` 涉及支付下单和金币资产 |
| `/ucp/beanpkg/:action?`（除 `/ucp/beanpkg`、`/ucp/beanpkg/index`） | `c.api.ucp.beanpkg->$action` | 未重构；`placeorder/coinorder` 涉及支付下单、金豆和金币兑换 |
| `/ucp/vodorder/:action?`（除 `/ucp/vodorder/myorders`、`/ucp/vodorder/mysupports`、`/ucp/vodorder/historyorders`） | `c.api.ucp.vodorder->$action` | 未重构；`index/create/support` 涉及榜单聚合、求片金币扣减或助力写入 |

### 活动、邀请、发现页

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/invite/:action?`（除 `/invite/info`） | `c.api.invite->$action` | 未重构；剩余 `bind` 涉及绑定关系、VIP/金币奖励写入 |
| `/explore/signtask/:action?`（除 `/explore/signtask`、`/explore/signtask/index`） | `c.api.explore.signtask->$action` | 未重构；签到任务 |
| `/explore/vodtask/:action?`（除 `/explore/vodtask`、`/explore/vodtask/index`、`/explore/vodtask/show/:vid`） | `c.api.explore.vodtask->$action` | 未重构；剩余 `reqcoin` 涉及金币发放和日志加锁 |

### 游戏、直播、一元购

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/game/wali/topup` | `c.api.game.wali->topup` | 未重构；上分、金币扣减、外部平台 |
| `/game/wali/withdraw` | `c.api.game.wali->withdraw` | 未重构；下分、金币增加、外部平台 |
| `/game/wali/enter` | `c.api.game.wali->enterGame` | 未重构；外部平台进入游戏 |
| `/game/lottery/topup`、`/game/lottery/withdraw`、`/game/lottery/enter`、`/game/lottery/balance` | `c.api.game.lottery->$action` | 未重构；彩票游戏平台资产、余额或外部进入游戏 |
| `/starLive/:action` | `c.api.starlive->$action` | 未重构；直播平台、部分回调/扣款 |
| `/onego/:action?`（除 `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last`、`/onego/hash`、`/onego/history`、`/onego/lucky`、`/onego/bet_ranks`、`/onego/marquee`） | `c.api.onego->$action` | 未重构；一元购剩余 `bet` 投注写入涉及金币扣减 |
| `/bought/:action?`（除 `/bought/listing`、`/bought/delete`） | `c.api.bought->$action` | 未重构；剩余 `buy` 涉及金豆扣费 |

### 社区、HGame、AI

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/community/:action?` | `c.api.topic->$action` | 未重构；发帖/评论等写入 |
| `/aiundress/:action?`（除 `/aiundress`、`/aiundress/listing`、`/aiundress/index`） | `c.api.aiundress->$action` | 未重构；上传、生成、资源查询等依赖外部 AI 服务、Redis 锁和金豆扣减 |

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
