# 接口重构总览

更新时间：2026-07-14

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
| `/sysavatar` | ANY | `UserHandler.SysAvatar` |
| `/captcha/req` | ANY | `CaptchaHandler.Req` |
| `/iploc/:ip` | ANY | `IPLocHandler.Find` |
| `/game/platforms` | ANY | `GameHandler.Platforms` |
| `/game/categories` | ANY | `GameHandler.Categories` |
| `/game/games` | ANY | `GameHandler.Games` |
| `/game/broadcasts` | ANY | `GameHandler.Broadcasts` |
| `/game/wali/gameList` | ANY | `GameHandler.WaliGames` |
| `/getLikeRows` | ANY | `VODHandler.LikeRows` |
| `/ucp/index` | ANY | `UCPHandler.Index` |
| `/ucp/feedback` | GET | `UCPHandler.FeedbackListing` |
| `/ucp/feedback/index` | GET | `UCPHandler.FeedbackIndex` |
| `/ucp/feedback/listing` | GET | `UCPHandler.FeedbackNewListing` |
| `/ucp/feedback/detail` | GET | `UCPHandler.FeedbackDetail` |
| `/ucp/msg`、`/ucp/msg/index` | GET | `UCPHandler.MsgListing` |
| `/ucp/myaff` | ANY | `UCPHandler.MyAff` |
| `/ucp/rolltitle` | ANY | `UCPHandler.RollTitle` |
| `/ucp/affcenter` | ANY | `UCPHandler.AffCenter` |
| `/ucp/payment`、`/ucp/payment/index`、`/ucp/payment/listing` | ANY | `UCPHandler.PaymentListing` |
| `/ucp/payment/safepaylog` | ANY | `UCPHandler.SafePayLog` |
| `/ucp/account`、`/ucp/account/index` | ANY | `UCPHandler.AccountIndex` |
| `/ucp/account/balancelog` | ANY | `UCPHandler.BalanceLog` |
| `/ucp/coinlog`、`/ucp/coinlog/index` | ANY | `UCPHandler.CoinLogIndex` |
| `/ucp/coinlog/bonuslog` | ANY | `UCPHandler.CoinLogBonusLog` |
| `/ucp/coinlog/invitelog` | ANY | `UCPHandler.CoinLogInviteLog` |
| `/vod/show/:vodid` | ANY | `VODHandler.Show` |
| `/vod/preView/:vodid/index.m3u8` | ANY | `VODHandler.Preview` |
| `/sendfile/play/:file`、`/sendfile/down/:file` | ANY | `SendfileHandler.Play/Down` |
| `/comment/listing-:params` | ANY | `CommentHandler.Listing` |
| `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last`、`/onego/hash`、`/onego/lucky`、`/onego/marquee` | ANY | `OneGoHandler` |
| `/vod/listing`、`/vod/recommend`、`/vod/hot`、`/vod/latest` | ANY | `VODHandler.Listing` |
| `/vod/listing-:params`、`/vod/recommend-:params`、`/vod/hot-:params`、`/vod/latest-:params` | ANY | `VODHandler.Listing` |
| `/v2/amazing/categories` | ANY | `AmazingHandler.Categories` |
| `/v2/amazing/listing`、`/v2/amazing/recommend`、`/v2/amazing/hot`、`/v2/amazing/latest` | ANY | `AmazingHandler.Listing` |
| `/v2/amazing/listing-:params`、`/v2/amazing/recommend-:params`、`/v2/amazing/hot-:params`、`/v2/amazing/latest-:params` | ANY | `AmazingHandler.Listing` |
| `/v2/so/list` | ANY | `SOHandler.List` |
| `/v2/vod/listing`、`/v2/vod/recommend`、`/v2/vod/hot`、`/v2/vod/latest` | ANY | `VODHandler.Listing` |
| `/v2/vod/listing-:params`、`/v2/vod/recommend-:params`、`/v2/vod/hot-:params`、`/v2/vod/latest-:params` | ANY | `VODHandler.Listing` |

### 已注册占位

| 接口 | Method | Go handler |
| --- | --- | --- |
| `/v2/register` | ANY | `notImplemented("c.apiv2.user.register")` |
| `/v2/login` | ANY | `notImplemented("c.apiv2.user.login")` |
| `/v2/forgot` | ANY | `notImplemented("c.apiv2.user.forgot")` |
| `/v2/vod/show/:vodid` | ANY | `notImplemented("c.apiv2.vod.show")` |
| `/v2/vod/up/:vodid`、`/v2/vod/down/:vodid` | ANY | `notImplemented("c.apiv2.vod.up/down")` |
| `/v2/vod/reqplay/:vodid`、`/v2/vod/reqdown/:vodid` | ANY | `notImplemented("c.apiv2.vod.reqplay/reqdown")` |
| `/v2/vod/buy/:vodid` | ANY | `notImplemented("c.apiv2.vod.buy")` |

## 已重构接口

### 基础与公共接口

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/sysavatar` | `c.api.user->sysavatar` | `UserHandler.SysAvatar` | 已重构，对比通过 |
| `/captcha/req` | `c.api.captcha->req` | `CaptchaHandler.Req` | 已重构，动态 secret 按 shape 对比通过 |
| `/iploc/:ip` | `c.api.index->iploc` | `IPLocHandler.Find` | 已重构，对比通过 |
| `/getLikeRows` | `c.api.index->getLikeRows` | `VODHandler.LikeRows` | 已重构，对比通过 |

### 游戏公共接口

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/game/platforms` | `c.api.game.index->index` | `GameHandler.Platforms` | 已重构，对比通过 |
| `/game/categories` | `c.api.game.index->categories` | `GameHandler.Categories` | 已重构，对比通过 |
| `/game/games` | `c.api.game.index->games` | `GameHandler.Games` | 已重构，对比通过 |
| `/game/broadcasts` | `c.api.game.index->broadcasts` | `GameHandler.Broadcasts` | 已重构，随机广播按 shape 对比通过 |
| `/game/wali/gameList` | `c.api.game.wali->games` | `GameHandler.WaliGames` | 已重构，对比通过；`category_id=5` 游客未登录分支已对齐 |
| `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last` | `c.api.onego->rules/rooms/current/last` | `OneGoHandler` | 已重构，对比通过；一元购公共只读规则/房间/当前期数/上期记录，旧 PHP 动态 `xxx_api_auth` 忽略 |
| `/onego/hash` | `c.api.onego->hash` | `OneGoHandler.Hash` | 已重构；公共哈希计算接口，复刻 SHA256 后提取末尾数字期号规则 |
| `/onego/lucky` | `c.api.onego->lucky` | `OneGoHandler.Lucky` | 已重构，对比通过；一元购幸运榜公共只读，保留旧 PHP 排行 SQL 未分页行为 |
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
| `/vod/preView/:vodid/index.m3u8` | `c.api.vod->preView` | `VODHandler.Preview` | 已重构，m3u8 输出对比通过 |
| `/sendfile/play/:file` | `c.api.sendfile->play` | `SendfileHandler.Play` | 已重构，按旧 PHP 空壳行为对齐 |
| `/sendfile/down/:file` | `c.api.sendfile->down` | `SendfileHandler.Down` | 已重构，按旧 PHP 空响应对齐 |

### 评论、收藏、播放/下载记录

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/comment/listing-:params` | `c.api.comment->listing` | `CommentHandler.Listing` | 已重构，对比通过 |

### 需要登录但不需要验证码

| 接口 | PHP handler | Go 入口 | 状态 |
| --- | --- | --- | --- |
| `/ucp/myaff` | `c.api.ucp.index->myaff` | `UCPHandler.MyAff` | 已重构，对比通过；支持 `x-cookie-auth` header 和 `xxx_api_auth` cookie |
| `/ucp/index` | `c.api.ucp.index->index` | `UCPHandler.Index` | 已重构；登录/游客只读个人中心聚合，旧 PHP 本地对比超时，已按源码契约和 Go 输出验证 |
| `/ucp/affcenter` | `c.api.ucp.index->affcenter` | `UCPHandler.AffCenter` | 已重构，对比通过；登录只读推广中心，合并用户组权限后计算播放/下载剩余额度 |
| `GET /ucp/feedback` | `c.api.ucp.index->feedback` | `UCPHandler.FeedbackListing` | 已重构，对比通过；登录只读历史反馈列表，POST 写入未接管 |
| `GET /ucp/feedback/index` | `c.api.ucp.feedback->index` | `UCPHandler.FeedbackIndex` | 已重构，对比通过；新版反馈初始化页，最近 30 天支付记录，POST 未接管 |
| `GET /ucp/feedback/listing` | `c.api.ucp.feedback->listing` | `UCPHandler.FeedbackNewListing` | 已重构，对比通过；新版反馈列表，支持 `type=0/1/2` 过滤，POST 未接管 |
| `GET /ucp/feedback/detail` | `c.api.ucp.feedback->detail` | `UCPHandler.FeedbackDetail` | 已重构，对比通过；新版反馈详情，校验反馈归属，附件图片和关联支付只读，POST 未接管 |
| `GET /ucp/msg`、`GET /ucp/msg/index` | `c.api.ucp.msg->index` | `UCPHandler.MsgListing` | 已重构，对比通过；登录只读消息会话列表，写状态 action 未接管 |
| `/ucp/payment`、`/ucp/payment/index` | `c.api.ucp.payment->index/listing` | `UCPHandler.PaymentListing` | 已重构，对比通过；兼容旧动态 action 默认入口 |
| `/ucp/payment/listing` | `c.api.ucp.payment->listing` | `UCPHandler.PaymentListing` | 已重构，对比通过；登录只读支付记录，支持 GET/POST page |
| `/ucp/payment/safepaylog` | `c.api.ucp.payment->safepaylog` | `UCPHandler.SafePayLog` | 已重构，对比通过；最近 7 天 safepay 记录 |
| `/ucp/account`、`/ucp/account/index` | `c.api.ucp.account->index` | `UCPHandler.AccountIndex` | 已重构，对比通过；登录只读资产主页 |
| `/ucp/account/balancelog` | `c.api.ucp.account->balancelog` | `UCPHandler.BalanceLog` | 已重构，对比通过；登录只读余额日志分页 |
| `/ucp/coinlog`、`/ucp/coinlog/index` | `c.api.ucp.coinlog->index` | `UCPHandler.CoinLogIndex` | 已重构，对比通过；登录只读金币日志首页，最近 10 条 |
| `/ucp/coinlog/bonuslog` | `c.api.ucp.coinlog->bonuslog` | `UCPHandler.CoinLogBonusLog` | 已重构，对比通过；登录只读收益金币日志分页和累计统计 |
| `/ucp/coinlog/invitelog` | `c.api.ucp.coinlog->invitelog` | `UCPHandler.CoinLogInviteLog` | 已重构，对比通过；登录只读邀请金币日志分页 |

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
| `/v2/vod/show/:vodid` | `c.apiv2.vod->show` |
| `/v2/vod/up/:vodid` | `c.apiv2.vod->up` |
| `/v2/vod/down/:vodid` | `c.apiv2.vod->down` |
| `/v2/vod/reqplay/:vodid` | `c.apiv2.vod->reqplay` |
| `/v2/vod/reqdown/:vodid` | `c.apiv2.vod->reqdown` |
| `/v2/vod/buy/:vodid` | `c.apiv2.vod->buy` |

## 未重构接口

### 首页、全局配置和工具类

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/init` | `c.api.index->init` | 未重构；依赖登录/游客初始化、系统设置、版本、广告、全局数据、通知等 |
| `/`、`/index` | `c.api.index->index` | 未重构；首页聚合，多表、多广告配置 |
| `/getGlobalData` | `c.api.index->getGlobalData` | 未重构；依赖系统设置、维护配置、广告、弹窗、版本和外链 |
| `/getCertUuid` | `c.api.index->getCertUuid` | 未重构；外部 HTTP 调用 |
| `/getCover` | `c.api.index->getCover` | 未重构；Redis、外部封面服务、AES 加密 |
| `/test` | `c.api.test->test` | 未重构；测试入口 |
| `/sms/:action?` | `c.api.sms->$action` | 未重构；短信相关 |
| `/email/:action?` | `c.api.email->$action` | 未重构；邮件相关 |

### 非 v2 视频接口

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/vod/up/:vodid`、`/vod/down/:vodid` | `c.api.vod->up/down` | 未重构；用户行为/写入 |
| `/vod/reqplay/:vodid`、`/vod/reqdown/:vodid` | `c.api.vod->reqplay/reqdown` | 未重构；播放/下载权限、日志、可能签名 |
| `/vod/buy/:vodid` | `c.api.vod->buy` | 未重构；购买/金币 |
| `/vod/:action?` | `c.api.vod->$action` | 未重构；动态 action |

### 评论、收藏、播放/下载记录

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/comment/:action?` | `c.api.comment->$action` | 未重构；评论发布涉及写库和权限 |
| `/favorite/:action?` | `c.api.favorite->$action` | 未重构；收藏写入 |
| `/minifavorite/:action?` | `c.api.minifavorite->$action` | 未重构 |
| `/playlog/:action?` | `c.api.playlog->$action` | 未重构 |
| `/downlog/:action?` | `c.api.downlog->$action` | 未重构 |

### 搜索、专题、公告

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/search` | `c.api.search->index` | 未重构；普通视频搜索 |
| `/minisearch` | `c.api.miniSearch->index` | 未重构；小视频搜索 |
| `/special/index` | `c.api.special->index` | 未重构 |
| `/special/listing`、`/special/listing-:params` | `c.api.special->listing` | 未重构 |
| `/special/detail/:spid`、`/special/detail/:spid-:params` | `c.api.special->detail` | 未重构 |
| `/special/up/:spid`、`/special/down/:spid` | `c.api.special->up/down` | 未重构；用户行为 |
| `/art/:action?` | `c.api.art->$action` | 未重构；公告/文章 |

### 小视频、作者页

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/minivod/listing`、`/minivod/recommend`、`/minivod/hot`、`/minivod/latest` | `c.api.minivod->listing` | 未重构 |
| `/minivod/topzan`、`/minivod/topcomment`、`/minivod/topplay`、`/minivod/topcoin`、`/minivod/topnew`、`/minivod/topday`、`/minivod/topweek`、`/minivod/topmonth` | `c.api.minivod->listing` | 未重构 |
| `/minivod/*-:params` | `c.api.minivod->listing` | 未重构 |
| `/minivod/reqlist` | `c.api.minivod->reqlist` | 未重构 |
| `/minivod/reqcoin` | `c.api.minivod->reqcoin` | 未重构 |
| `/minivod/show/:vodid`、`/minivod/up/:vodid`、`/minivod/down/:vodid` | `c.api.minivod->$action` | 未重构 |
| `/minivod/reqplay/:vodid`、`/minivod/reqdown/:vodid` | `c.api.minivod->$action` | 未重构；播放/下载权限 |
| `/minivod/throwcoin/:vodid` | `c.api.minivod->throwcoin` | 未重构；金币打赏 |
| `/miniplaylog/listing` | `c.api.minivod->history` | 未重构 |
| `/miniplaylog/remove` | `c.api.minivod->historyDelete` | 未重构；写入 |
| `/minivod/reqlong/:vodid` | `c.api.minivod->getLong2Mini` | 未重构 |
| `/minivod/parselong/:vodid/index.m3u8` | `c.api.minivod->parseM3u8` | 未重构；媒体解析 |
| `/my/:authorid/:action?` | `c.api.my->$action` | 未重构；作者页/小视频 |

### 用户账号

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/register` | `c.api.user->register` | 未重构；验证码、注册风控、写库 |
| `/login` | `c.api.user->login` | 未重构；密码/短信登录、session |
| `/logout` | `c.api.user->logout` | 未重构；session 删除 |
| `/forgot` | `c.api.user->forgot` | 未重构；验证码、密码重置 |
| `/delete` | `c.api.user2->delAccount` | 未重构；账号注销 |
| `/changePhone` | `c.api.user2->changePhone` | 未重构；手机换绑、验证码 |

### 支付和回调

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/payment/:action` | `c.api.payment->$action` | 未重构；支付、查询、下单、回调跳转 |
| `/respond/:action` | `c.respond.*` | 未重构；支付平台回调 |
| `/respond/shangfu`、`/respond/wappay1`、`/respond/wappay2`、`/respond/wappay3`、`/respond/wappay4`、`/respond/wappay5` | `c.respond.*` | 未重构 |
| `/respond/hawpay`、`/respond/easypay`、`/respond/chan1`、`/respond/pay6`、`/respond/pay7` | `c.respond.*` | 未重构 |

### 个人中心

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `POST /ucp/feedback` | `c.api.ucp.index->feedback` | 未重构；提交反馈写入 |
| `/ucp/upgrade` | `c.api.ucp.index->upgrade` | 未重构；会员升级/金币 |
| `/ucp/user/:action?` | `c.api.ucp.user->$action` | 未重构 |
| `/ucp/msg/:action?`（除 `GET /ucp/msg`、`GET /ucp/msg/index`） | `c.api.ucp.msg->$action` | 未重构；show/setread/cleanread/delete/send 涉及详情、写入或已读状态 |
| `/ucp/task/:action?` | `c.api.ucp.task->$action` | 未重构；任务奖励/签到等 |
| `/ucp/account/:action?`（除 `/ucp/account`、`/ucp/account/index`、`/ucp/account/balancelog`） | `c.api.ucp.account->$action` | 未重构；账户其他 action |
| `/ucp/bankcard/:action?` | `c.api.ucp.bankcard->$action` | 未重构 |
| `/ucp/withdraw/:action?` | `c.api.ucp.withdraw->$action` | 未重构；提现 |
| `/ucp/coinlog/:action?`（除 `/ucp/coinlog`、`/ucp/coinlog/index`、`/ucp/coinlog/bonuslog`、`/ucp/coinlog/invitelog`） | `c.api.ucp.coinlog->$action` | 未重构；`exchange` 为金币兑换写入高风险 |
| `/ucp/taskbox/:action?` | `c.api.ucp.taskbox->$action` | 未重构 |
| `/ucp/vippkg/:action?` | `c.api.ucp.vippkg->$action` | 未重构；会员套餐/订单 |
| `/ucp/coinpkg/:action?` | `c.api.ucp.coinpkg->$action` | 未重构；金币套餐 |
| `/ucp/beanpkg/:action?` | `c.api.ucp.beanpkg->$action` | 未重构；金豆套餐 |
| `/ucp/payment/:action?`（除 `/ucp/payment`、`/ucp/payment/index`、`/ucp/payment/listing`、`/ucp/payment/safepaylog`） | `c.api.ucp.payment->$action` | 未重构；支付其他 action |
| `/ucp/feedback/:action?`（除 `GET /ucp/feedback/index`、`GET /ucp/feedback/listing`、`GET /ucp/feedback/detail`） | `c.api.ucp.feedback->$action` | 未重构；新版反馈剩余 `create` 写入待迁 |
| `/ucp/vodorder/:action?` | `c.api.ucp.vodorder->$action` | 未重构；视频订单 |

### 活动、邀请、发现页

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/activity/:action?` | `c.api.activity->$action` | 未重构 |
| `/invite/:action?` | `c.api.invite->$action` | 未重构 |
| `/explore/index` | `c.api.explore.index->index` | 未重构 |
| `/explore/notification/:action?` | `c.api.explore.notification->$action` | 未重构 |
| `/explore/signtask/:action?` | `c.api.explore.signtask->$action` | 未重构；签到任务 |
| `/explore/vodtask/show/:vid` | `c.api.explore.vodtask->show` | 未重构 |
| `/explore/vodtask/:action?` | `c.api.explore.vodtask->$action` | 未重构 |

### 开放平台、统计

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/open/:action?` | `c.api.open->$action` | 未重构 |
| `/shortcutstats/add` | `c.api.shortcutstats->add` | 未重构；统计写入 |
| `/adstats/add` | `c.api.adstats->add` | 未重构；统计写入 |
| `/playstats/add` | `c.api.playstats->add` | 未重构；统计写入 |

### 游戏、直播、一元购

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/game/wali/topup` | `c.api.game.wali->topup` | 未重构；上分、金币扣减、外部平台 |
| `/game/wali/withdraw` | `c.api.game.wali->withdraw` | 未重构；下分、金币增加、外部平台 |
| `/game/wali/balance` | `c.api.game.wali->getBalance` | 未重构；外部平台余额 |
| `/game/wali/enter` | `c.api.game.wali->enterGame` | 未重构；外部平台进入游戏 |
| `/game/wali/test` | `c.api.game.wali->ping` | 未重构；外部平台测试 |
| `/game/lottery/gameList`、`/game/lottery/topup`、`/game/lottery/withdraw`、`/game/lottery/enter`、`/game/lottery/balance` | `c.api.game.lottery->$action` | 未重构；彩票游戏平台 |
| `/starLive/:action` | `c.api.starlive->$action` | 未重构；直播平台、部分回调/扣款 |
| `/onego/:action?`（除 `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last`、`/onego/hash`、`/onego/lucky`、`/onego/marquee`） | `c.api.onego->$action` | 未重构；一元购剩余 history/bet/bet_ranks 等登录/投注写入 |
| `/bought/:action?` | `c.api.bought->$action` | 未重构；付费影片 |

### 社区、HGame、AI

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/community/list`、`/community/recommend`、`/community/hot`、`/community/latest`、`/community/favorite` | `c.api.topic->list` | 未重构 |
| `/community/*-:params` | `c.api.topic->list` | 未重构 |
| `/community/clisting`、`/community/clisting-:params` | `c.api.topic->clisting` | 未重构 |
| `/community/:action?` | `c.api.topic->$action` | 未重构；发帖/评论等写入 |
| `/hgame/:action` | `c.api.hgame->$action` | 未重构 |
| `/aiundress/:action?` | `c.api.aiundress->$action` | 未重构；AI/外部服务/任务 |

### 图片、附件和通配资源

| 接口 | PHP handler | 备注 |
| --- | --- | --- |
| `/:size/:uri` | `c.api.pic->index` | 未重构；图片处理、缩略图 |
| `/captcha/pic`、`/captcha/picx` | `c.api.captcha->$action` | 未重构；图片验证码二进制输出 |
| `/attach/:action?` | `c.api.attach->$action` | 未重构；附件 |

## 建议后续顺序

1. 继续按风险优先级评估 `/ucp/msg/show`：旧 PHP 读取详情同时可能涉及已读状态，需要明确是否复刻写入副作用。
2. 继续公共只读接口：`/special/listing`、`/special/detail`；`/search` 会写搜索日志，放入普通写入批次评估。
3. 中风险接口：`/getGlobalData`、`/init`、`/vod/show/:vodid`。
4. 高风险接口最后迁移：支付、金币/金豆、购买、任务奖励、提现、游戏上分/下分、验证码注册/登录。

## 当前验证命令

```shell
go test ./...
go vet ./...
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/xj-comp-api ./cmd/api
make ci
```
