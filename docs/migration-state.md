# 迁移状态压缩记录

用于每完成一个接口后压缩上下文。

## 当前约定

- Go 分层：`server -> handler -> service -> repository/client`。
- JSON 兼容：默认使用 `internal/legacyjson`。
- 旧 PHP 对比 host：`http://127.0.0.1:18765`。
- 客户端访问旧服务时用 `127.0.0.1`，不要用监听地址 `0.0.0.0`。
- 对比时忽略旧中间件动态字段：`data.xxx_api_auth`。

## 已完成接口

### `/sysavatar`

- PHP: `c.api.user->sysavatar`
- Go: `internal/handler.UserHandler.SysAvatar`
- Service: `internal/service/user.SysAvatarService`
- 测试：`make ci` 通过；PHP-Go 对比通过。

### `/captcha/req`

- PHP: `c.api.captcha->req`
- Go: `internal/handler.CaptchaHandler.Req`
- Service: `internal/service/captcha.Service`
- 稳定字段：`retcode=0`、`errmsg=""`、`data.smscaptcha=1`、`data.picurl` 前缀 `/captcha/picx?`
- 动态字段：`data.picurl` query secret
- 测试：`make ci` 通过；PHP-Go 对比通过，按动态 secret 规则只比较 shape。

### `/iploc/:ip`

- PHP: `c.api.index->iploc`
- Go: `internal/handler.IPLocHandler.Find`
- Service: `internal/service/iploc.Service`
- Config: `IPDB_PATH`，默认 `/Users/canavs/xjProj/XJBackend/api/data/ipipfree.ipdb`
- 稳定样例：`8.8.8.8 -> "GOOGLE.COM GOOGLE.COM"`，`127.0.0.1 -> "本机地址 本机地址"`
- 测试：`make ci` 通过；PHP-Go 对比通过。

### `/game/platforms`

- PHP: `c.api.game.index->index`
- Go: `internal/handler.GameHandler.Platforms`
- Service: `internal/service/game.PlatformService`
- Repository: `internal/repository/game.PlatformRepository`
- Config: `MYSQL_DSN`，默认连接本地 Docker MySQL `xj_comp`
- 兼容规则：保留 `game_platform` 原始字段和值类型；删除 `json` 字段；按 ``order`` DESC，limit 100。
- 测试：`make ci` 通过；PHP-Go 对比通过。

### `/game/categories`

- PHP: `c.api.game.index->categories`
- Go: `internal/handler.GameHandler.Categories`
- Service: `internal/service/game.CategoryService`
- Repository: `internal/repository/game.CategoryRepository`
- Config: `GAME_RESOURCE_BASE_URL`，默认 `https://image.xjdev.one`
- 兼容规则：保留 `game_category` 原始字段和值类型；`icon/image` 非空时拼接资源 URL；支持 `parent_id` query。
- 测试：`make ci` 通过；PHP-Go 对比 `/game/categories` 和 `/game/categories?parent_id=1` 均通过。

### `/v2/amazing/categories`

- PHP: `c.apiv2.amazing->categories`
- Go: `internal/handler.AmazingHandler.Categories`
- Service: `internal/service/amazing.CategoryService`
- Repository: `internal/repository/amazing.CategoryRepository`
- 兼容规则：查询 `amazing_category` 固定列 `id,parent_id,title,description,status,order`；支持 `parent_id` query；返回 `data.rows`。
- 测试：`make ci` 通过；PHP-Go 对比 `/v2/amazing/categories` 和 `/v2/amazing/categories?parent_id=1` 均通过。

### `/v2/so/list`

- PHP: `c.apiv2.so->index`
- Go: `internal/handler.SOHandler.List`
- Service: `internal/service/so.ConfigService`
- Repository: `internal/repository/so.ConfigRepository`
- DB: `server_so_config`，查询条件为 `arm=? AND version > ? AND channel=?`，不增加旧 PHP 没有的排序语义。
- 兼容规则：`version` 按整数解析；`arm/channel` 做轻量字符串清理；无记录、空字符串或非法 JSON 时 `data.data=null`；有记录时对 `value` 执行 JSON 解码后返回。
- 测试：`make ci` 通过；PHP-Go 对比 `/v2/so/list`、`?channel=xj&arm=v8a&version=510`、`?channel=Ali&arm=v8a&version=429`、`?channel=xj&arm=v8a&version=511` 均通过。

### `/v2/vod/{listing,recommend,hot,latest}`

- PHP: `c.apiv2.vod->listing`
- Go: `internal/handler.VODHandler.Listing`
- Service: `internal/service/vod.ListingService`
- Repository: `internal/repository/vod.ListingRepository`
- 路由：注册 `/v2/vod/listing`、`/v2/vod/recommend`、`/v2/vod/hot`、`/v2/vod/latest` 及各自 `-:params` 形式；移除会与具体路由冲突的 `/v2/vod/*path` catch-all，占位保留 `/v2/vod/show/:vodid` 和 `up/down/reqplay/reqdown/buy`。
- Params：按 PHP 模板 `$cateid:0-$areaid:0-$yearid:0-$definition:0-$duration:0-$freetype:0-$mosaic:0-$langvoice:0-$orderby:0-$page:1` 解析；path page 为 `0` 时使用 query `page`。
- DB：读取 `vods`、`vod_categories`、`vod_areas`、`vod_years`、`vod_servers`、`vod_tags`；`pagesize=16`。
- 排序：`listing` 默认 `utimestamp DESC`，`orderby=1/2/3` 分别为 `upnum/playcount_total/scorenum DESC`；`hot=playcount_week DESC`；`latest=vodid DESC`；`recommend` 使用 `showtype=0` 随机列表并保持 `pageinfo.total=0` 语义。
- 兼容规则：输出 `now/action/sample_params/params/vodrows/pageinfo/orders/categories/areas/years/definitions/durations/freetypes/mosaics/langvoices`；`vodrows` 对齐 PHP `procRow2` 的核心字段、资源 URL、标签、播放/下载占位 URL、时长、时间和价格类型。
- 测试：`make ci` 通过；PHP-Go 对比 `/v2/vod/listing`、`/v2/vod/listing-0-0-0-0-0-0-0-0-0-2`、`/v2/vod/hot`、`/v2/vod/latest` 的 status/retcode/action/params/pageinfo/top5 vodid/首行关键字段均一致；`/v2/vod/recommend` 因旧 PHP 随机，只断言结构、数量和 `pageinfo.total=0`。

### `/vod/{listing,recommend,hot,latest}`

- PHP: `c.api.vod->listing`
- Go: `internal/handler.VODHandler.Listing`
- Service: `internal/service/vod.ListingService`
- Repository: `internal/repository/vod.ListingRepository`
- 路由：注册 `/vod/listing`、`/vod/recommend`、`/vod/hot`、`/vod/latest` 及各自 `-:params` 形式；与 `/v2/vod/*` 复用同一套 VOD 列表 service/repository。
- 兼容规则：`vodAction` 同时识别 `/vod/` 和 `/v2/vod/` 前缀；`pageinfo.pages` 按 PHP `kernel/lib/Page.php::pageSelector()` 算法生成；`scorenum` 保留 PHP `float(3,1)` 字符串表现，如 `9.0`、`10.0`。
- 测试：聚焦 `go test ./internal/service/vod ./internal/server` 通过；PHP-Go 对比 `/vod/listing`、`/vod/listing-0-0-0-0-0-0-0-0-0-2`、`/vod/hot`、`/vod/latest` 忽略 `data.xxx_api_auth` 和 `data.now` 后完全一致；`/vod/recommend` 因旧 PHP `newrc`/随机逻辑，只断言 status、retcode、action、`vodrows` 数量和 `pageinfo.total`。

### `/vod/show/:vodid`

- PHP: `c.api.vod->show`
- Go: `internal/handler.VODHandler.Show`
- Service: `internal/service/vod.ListingService.Show`
- Repository: `internal/repository/vod.ListingRepository`
- DB: 读取 `vods`、`vod_categories`、`vod_areas`、`vod_years`、`vod_servers`、`vod_tags`、`vod_tagmaps`；附件 `attachs` 在旧 PHP 中读取但未写入 JSON，Go 不额外暴露。
- 兼容规则：返回 `data.vodrow`、`data.categories`、`data.similarrows`、`data.likerows`；主视频行复用 `procRow2` 兼容字段；父级分类按 PHP `Category::getP()` 返回；相似视频按 tag + 90 天窗口随机，数量不足时随机补足；猜你喜欢按同分类随机 5 条。
- 测试：聚焦 `go test ./internal/service/vod ./internal/server` 通过；PHP-Go 对比 `/vod/show/1`、`/vod/show/10` 的 status、retcode、主视频关键字段、父级分类、相似/喜欢数量通过；`/vod/show/100` 错误分支完全一致。随机列表不做逐条完全相等。

### `/vod/preView/:vodid/index.m3u8`

- PHP: `c.api.vod->preView`
- Go: `internal/handler.VODHandler.Preview`
- Service: `internal/service/vod.ListingService.Preview`
- DB/外部依赖：读取 `vods` 和 `vod_servers(type=play)`；m3u8 HTTP 拉取通过 `M3U8Fetcher` 接口注入，单元测试使用 fake，不连接真实 CDN。
- 兼容规则：返回 `Content-Type: vnd.apple.mpegurl`；不存在或已删除视频返回空 `200`；存在且有 `play_url` 但拉取/解析不到有效片段时返回 `#EXT-X-ENDLIST`；可解析时按旧 PHP 逻辑取子 m3u8 并截取 0-300 秒，KEY/TS 相对 URL 按域名根路径重写。
- 测试：聚焦 `go test ./internal/service/vod ./internal/server` 通过；PHP-Go 对比 `/vod/preView/100/index.m3u8` 空响应、`/vod/preView/1/index.m3u8` 和 `/vod/preView/10/index.m3u8` 的 `#EXT-X-ENDLIST` 响应完全一致。

### `/sendfile/play/:file`、`/sendfile/down/:file`

- PHP: `c.api.sendfile->play/down`
- Go: `internal/handler.SendfileHandler`
- Service: `internal/service/sendfile.Service`
- Repository: 复用 `internal/repository/user.Repository` 鉴权和 `internal/repository/vod.ListingRepository.VODByID`。
- 兼容规则：旧 PHP `play` 只检查登录、`vodid` 和视频存在性，成功分支没有 JSON 输出；Go 保持成功空 `200 text/html`。旧 PHP `down` 是空方法；Go 保持空 `200 text/html`。带点文件名如 `/sendfile/play/test.m3u8` 在旧路由中不匹配，Go 返回 404 保持 status 对齐。
- 测试：聚焦 `go test ./internal/service/sendfile ./internal/server` 通过；PHP-Go 对比 `/sendfile/play/test` 未登录、`/sendfile/play/test?vodid=100` 已登录不存在、`/sendfile/play/test?vodid=1` 已登录成功空响应、`/sendfile/down/test` 空响应、`/sendfile/play/test.m3u8` 404 通过。

### `/comment/listing-:params`

- PHP: `c.api.comment->listing`
- Go: `internal/handler.CommentHandler.Listing`
- Service: `internal/service/comment.Service`
- Repository: `internal/repository/comment.Repository`
- Params: `$vodid:0-$orderby:0-$page:1`；path page 为 `0` 时使用 query `page`。
- DB: 读取 `vods`、`vod_comments`、`users`、`user_groups`；只取 `vod.showtype <= 1` 且评论 `showtype=0/rootid=0`。
- 兼容规则：返回 `data.rows` 和 `data.pageinfo`；评论行对齐 `id/rootid/parentid/lft/rgt/depth/vodid/uid/sid/username/nickname/gender/gicon/isvip/content/upnum/downnum/avatar_url/addtime/__closenum__/subrows`；`orderby=1` 使用 `a.upnum DESC`，默认 `a.addtime DESC`。
- 测试：聚焦 `go test ./internal/service/comment ./internal/server` 通过；PHP-Go 对比 `/comment/listing-1-0-1`、`/comment/listing-61494-0-1`、`/comment/listing-61494-1-1`、`/comment/listing-999999-0-1` 忽略动态 `data.xxx_api_auth` 后完全一致。

### `/v2/amazing/{listing,recommend,hot,latest}`

- PHP: `c.apiv2.amazing->listing`
- Go: `internal/handler.AmazingHandler.Listing`
- Service: `internal/service/amazing.ListingService`
- Repository: `internal/repository/amazing.SoftwareRepository`
- 路由：注册 `/v2/amazing/listing`、`/v2/amazing/recommend`、`/v2/amazing/hot`、`/v2/amazing/latest` 及各自 `-:params` 形式。
- Params：按 PHP 模板 `$category_id:0-$orderby:0-$page:1` 解析；path page 为 `0` 时使用 query `page`。
- DB：读取 `amazing`；`pagesize=20`。
- 排序：`listing/recommend/latest` 使用 `id DESC`；`recommend` 额外 `is_recommend=1`；`hot` 使用 `dl_count DESC`；所有 action 都限定 `status=1`。
- 兼容规则：输出 `now/rows/pageinfo`；`icon/image` 非空时按资源 URL 拼接；保留 `amazing` 原始字段和值类型。
- 测试：`make ci` 通过；PHP-Go 对比 `/v2/amazing/listing`、`/v2/amazing/listing-3-0-1`、`/v2/amazing/recommend`、`/v2/amazing/hot`、`/v2/amazing/latest` 忽略动态 `xxx_api_auth` 和 `now` 后完全一致。

### `/game/games`

- PHP: `c.api.game.index->games`
- Go: `internal/handler.GameHandler.Games`
- Service: `internal/service/game.ListingService`
- Repository: `internal/repository/game.GameRepository`
- DB: 读取 `game`；`status=1`；支持 `platform_id`、`category_id` query；按 ``order`` DESC，limit 100。
- 兼容规则：保留 `game` 原始字段和值类型；`icon/image/cover` 非空时使用 `RESOURCE_BASE_URL` 拼接，默认动态域名与 PHP `runtime::resUrl` 对齐。
- 测试：`make ci` 通过；PHP-Go 对比 `/game/games`、`/game/games?platform_id=1`、`/game/games?platform_id=1&category_id=2` 忽略动态 `xxx_api_auth` 后完全一致。

### `/game/broadcasts`

- PHP: `c.api.game.index->broadcasts`
- Go: `internal/handler.GameHandler.Broadcasts`
- Service: `internal/service/game.BroadcastService`
- Repository: `internal/repository/game.BroadcastRepository`
- DB: 读取 `game_broadcast`；`status=1`；按 `id DESC`，limit 100。
- 兼容规则：返回 `data.data` 字符串数组；按旧 PHP 规则替换 `{user}` 为手机号样式脱敏串，替换 `{amount}` 为 `min_amount/max_amount` 区间金额。
- 测试：`make ci` 通过；PHP-Go 对比 `/game/broadcasts` 的 status/retcode/errmsg、条数、字符串类型和占位符替换完成度通过；具体文案不做完全相等，因为 PHP 每次请求会随机生成用户和金额。

### `/getLikeRows`

- PHP: `c.api.index->getLikeRows`
- Go: `internal/handler.VODHandler.LikeRows`
- Service: `internal/service/vod.ListingService.LikeRows`
- Repository: `internal/repository/vod.ListingRepository`
- DB: 读取 `vod_categories`、`vod_areas`、`vod_years`、`vod_servers`、`vod_tags`、`vods`；`vods` 使用 `showtype=0 ORDER BY RAND() LIMIT 6`。
- 兼容规则：返回 `data.likerows`；虽然 PHP 读取了 `pagesize` query，但实际调用 `randRows_slave(6)`，Go 保持固定 6 条；视频行复用 `/v2/vod/*` 的 `procRow2` 兼容字段。
- 测试：聚焦 `go test ./internal/service/vod ./internal/server` 通过；PHP-Go 对比 `/getLikeRows`、`/getLikeRows?pagesize=100` 的 status/retcode/errmsg、条数和核心字段 shape 通过；具体视频不做完全相等，因为旧 PHP 每次随机。

### `/game/wali/gameList`

- PHP: `c.api.game.wali->games`
- Go: `internal/handler.GameHandler.WaliGames`
- Service: `internal/service/game.ListingService`
- Repository: `internal/repository/game.GameRepository`
- DB: 读取 `game`；固定 `platform_id=1`；普通分类支持 `category_id`；按 ``order`` DESC，limit 100。
- 兼容规则：普通分类返回 `data.data` 游戏数组并拼接资源 URL；`category_id=5` 是常玩游戏，需要登录历史，当前无鉴权上下文时按旧 PHP 游客行为返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- 测试：聚焦 `go test ./internal/server ./internal/service/game` 通过；PHP-Go 对比 `/game/wali/gameList`、`/game/wali/gameList?category_id=2`、`/game/wali/gameList?category_id=5` 忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/myaff`

- PHP: `c.api.ucp.index->myaff`
- Go: `internal/handler.UCPHandler.MyAff`
- Service: `internal/service/ucp.Service.MyAff`
- Repository: `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；hex token 解码为 32 字节 sid 后查 `sessions`，校验 `type=0` 和 `md5(users.password + "_" + users.salt)` token。
- DB: 读取 `sessions`、`users`、`user_groups`；推荐列表查询 `users a LEFT JOIN user_groups b/c WHERE a.recommend_uid=? ORDER BY a.uid DESC`；`pagesize=20`。
- 兼容规则：未登录返回 `retcode=-9999`、`errmsg=请登录后操作`、空 `data` 对象；登录后返回 `data.rows` 和 PHP Page `pageinfo`；用户行对齐 `procRow2` 的字段、类型、`gicon`、`isvip`、`avatar_url`、`duetime/dueday`、base36 `uniqkey`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 sid `250f790ba71ec2b9d3855f424db2259e`、`uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/myaff`、`/ucp/myaff?page=1`、`/ucp/myaff?page=2` 忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/rolltitle`

- PHP: `c.api.ucp.index->rolltitle`
- Go: `internal/handler.UCPHandler.RollTitle`
- Service: `internal/service/ucp.Service.RollTitle`
- Repository: `internal/repository/ucp.Repository`
- DB: 读取 `roll_titles`；查询条件 `status=1`；按 `id DESC`，limit 10。
- 兼容规则：公共只读，不要求登录；返回 `data.messages`，保留数据库原始字段和值类型；旧 PHP 中间件动态 `data.xxx_api_auth` 在 Go 中不返回。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/rolltitle` 忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/payment`、`/ucp/payment/index`、`/ucp/payment/listing`

- PHP: `c.api.ucp.payment->index/listing`
- Go: `internal/handler.UCPHandler.PaymentListing`
- Service: `internal/service/ucp.Service.PaymentListing`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- DB: 读取 `trade_payments`；查询条件 `uid=当前用户 uid`；`listing` 按 `createtime DESC`，`pagesize=20`。
- 兼容规则：`/ucp/payment` 和 `/ucp/payment/index` 走 `listing`；支持 GET query 和 POST form 的 `page`；金额按分转元字符串两位小数；`createtime/paidtime` 为 `Y-m-d H:i`；未知 `payway/paycode/paytype` 映射返回 `null`；分页 `plist/pages` 对齐 PHP `kernel/lib/Page.php`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/payment/listing?page=1`、`page=2`、`page=0`、`/ucp/payment`、`/ucp/payment/index` 和 POST `/ucp/payment/listing` 忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/payment/safepaylog`

- PHP: `c.api.ucp.payment->safepaylog`
- Go: `internal/handler.UCPHandler.SafePayLog`
- Service: `internal/service/ucp.Service.SafePayLog`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- DB: 读取 `trade_payments`；查询条件 `uid=当前用户 uid AND createtime > now-7天 AND payway='safepay'`；按旧 model 默认 `payid DESC`；limit 10。
- 兼容规则：返回 `data.payrows`，无 `pageinfo`；支付行字段复用 `payment.procRow2` 映射。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/payment/safepaylog` 和未登录分支忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/account`、`/ucp/account/index`

- PHP: `c.api.ucp.account->index`
- Go: `internal/handler.UCPHandler.AccountIndex`
- Service: `internal/service/ucp.Service.AccountIndex`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；普通请求按 PHP 行为 cookie 优先，OPTIONS 或无 cookie 时读 header；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- DB: 读取 `users_account`、`users_quota`、`user_balancelogs` 和 `settings.uuid='setting'` 中的 `exrate`；余额日志按 `trxtime DESC LIMIT 10`。
- 兼容规则：`account.balance/frozen/deposit/available_balance` 为两位小数字符串；`game_balance/game_frozen/game_available_balance` 为 int；`goldcoin/exrate` 为 int；`coin2rmb/max2rmb` 为两位小数字符串；`logrows.trxin/trxout/balance` 为两位小数字符串；`logrows.trxtime` 30 天内为相对时间，超过 30 天为 `Y-m-d`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/account`、`/ucp/account/index`、POST `/ucp/account/index`、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/account/balancelog`

- PHP: `c.api.ucp.account->balancelog`
- Go: `internal/handler.UCPHandler.BalanceLog`
- Service: `internal/service/ucp.Service.BalanceLog`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- DB: 读取 `user_balancelogs`；查询条件 `uid=当前用户 uid`；按 `trxtime DESC`；`pagesize=20`；`pageinfo` URL 为 `/ucp/account/balancelog?page=[?]`。
- 兼容规则：支持 GET query 和 POST form 的 `page`；`page=0` 归一到 1，过大页归一到最后页；`logrows.trxin/trxout/balance` 为两位小数字符串；`paytype` 按 PHP `procLogRow` 映射，未命中为 `--`；`trxtime` 30 天内为相对时间，超过 30 天为 `Y-m-d`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/account/balancelog?page=1`、`page=0`、`page=999999`、POST `/ucp/account/balancelog`、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/coinlog`、`/ucp/coinlog/index`

- PHP: `c.api.ucp.coinlog->index`
- Go: `internal/handler.UCPHandler.CoinLogIndex`
- Service: `internal/service/ucp.Service.CoinLogIndex`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- DB: 读取 `users_account`、`users_quota`、`settings.uuid='setting'` 中的 `exrate`，以及 `user_coinlogs LEFT JOIN users ON users.uid=user_coinlogs.invited_uid`；金币日志按 `logid DESC LIMIT 10`。
- 兼容规则：`/ucp/coinlog` 默认走 `index`；返回 `data.account/goldcoin/exrate/logrows`；`account` 金额字段复用 PHP `account.procRow`；`goldcoin/exrate` 为 int；`logrows.cointype` 按 PHP 映射，未命中为 `--`；`logrows.addtime` 30 天内使用 PHP 相对时间模板并保留天/小时/分钟后的尾随空格，超过 30 天为 `Y-m-d`；`mobi` 按 PHP `maskPhone` 遮罩。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/coinlog`、`/ucp/coinlog/index`、POST `/ucp/coinlog/index`、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/coinlog/invitelog`

- PHP: `c.api.ucp.coinlog->invitelog`
- Go: `internal/handler.UCPHandler.CoinLogInviteLog`
- Service: `internal/service/ucp.Service.CoinLogInviteLog`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- DB: 读取 `user_coinlogs LEFT JOIN users ON users.uid=user_coinlogs.invited_uid`；查询条件 `uid=当前用户 uid AND cointype IN (201,32,11)`；按 `addtime DESC`；`pagesize=20`；`pageinfo` URL 为 `/ucp/coinlog/invitelog?page=[?]`。
- 兼容规则：支持 GET query 和 POST form 的 `page`；`page=0` 归一到 1；`logrows` 复用 PHP `coinlog.procLogRow`，其中 `201` 映射为 `邀请好友赠送vip天数`，`mobi` 按 PHP `maskPhone` 遮罩。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/coinlog/invitelog?page=1`、`page=0`、POST `/ucp/coinlog/invitelog`、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/coinlog/bonuslog`

- PHP: `c.api.ucp.coinlog->bonuslog`
- Go: `internal/handler.UCPHandler.CoinLogBonusLog`
- Service: `internal/service/ucp.Service.CoinLogBonusLog`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- DB: 读取 `user_coinlogs LEFT JOIN users ON users.uid=user_coinlogs.invited_uid`；列表查询条件 `uid=当前用户 uid AND cointype IN (0,1,9,2,3,4,5,6,7,10,11,12,13,14,15,16,17,18,19,32,22)`；按 `addtime DESC`；`pagesize=20`；`pageinfo` URL 为 `/ucp/coinlog/bonuslog?page=[?]`。
- 兼容规则：返回 `data.logrows/addinfo/pageinfo`；`logrows` 复用 PHP `coinlog.procLogRow`；`addinfo.inviteTotal` 统计 `cointype=11` 的 `COUNT(DISTINCT invited_uid)`，`activeTotal` 统计 `cointype=15`，`bonusTotal` 只统计 `(0,1,9,2,3,4,5,6,7,10,11,12,13,14,15,16,17,18,19)`，保持 PHP 不含 `22/32` 的兼容行为；分页 `plist` 已按 `kernel/lib/Page.php` 的 `len0=5/len1=4` 算法对齐。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/coinlog/bonuslog?page=1`、`page=0`、POST `/ucp/coinlog/bonuslog`、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/affcenter`

- PHP: `c.api.ucp.index->affcenter`
- Go: `internal/handler.UCPHandler.AffCenter`
- Service: `internal/service/ucp.Service.AffCenter`
- Repository: `internal/repository/ucp.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- DB: 读取 `users_quota.goldcoin`、`users_goldbean.gold_bean`、`user_groups`、`vod_playlogs_week` 当日播放次数、`vod_downlogs` 当日下载次数。
- 兼容规则：登录用户通过 `initGids + initPerm` 合并 `user_groups.perms`，再按 `max.vod.play.daynum` 和 `max.vod.down.daynum` 计算 `uinfo.play_daily_remainders/down_daily_remainders`；`sysgid` 优先于 `gid`；`curr_group/next_group` 按 `minup ASC` 查找；`next_upgrade_need` 最小为 0；`data.user` 复用 PHP `user.procRow2` 字段。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/affcenter` GET header、POST header、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致。

### `/ucp/index`

- PHP: `c.api.ucp.index->index`
- Go: `internal/handler.UCPHandler.Index`
- Service: `internal/service/ucp.Service.Index`
- Repository: `internal/repository/ucp.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；旧 PHP 仅显式注册 `/ucp/index`，Go 不把裸 `/ucp` 等价到该接口。
- DB: 登录态读取 `users_quota`、`users_goldbean`、`user_groups`、`vod_playlogs_week`、`vod_downlogs`、`minivod_viewlogs_{uid%100}`、`user_coinlogs`；游客态读取 `user_guests`、`vod_guest_playlogs`、`vod_guest_downlogs`、`minivod_guestviewlogs_{sid首字符}`。
- 兼容规则：登录态返回 `data.user/uinfo/signed/groups`；游客态返回 `data.user/uinfo/signed` 且 `groups` 省略；游客缺 `user_guests` 时返回 `retcode=-1`、`errmsg=请登录后操作，客户端游客请先携带信息` 且省略 `data`；游客 `uinfo.curr_group/next_group` 为 `null`；登录态 `curr_group/next_group` 包含 `gid/gname/gicon/minup`；登录态 `mobi/email` 以 `~` 开头时清空；用户组能力列表只保留 `gicon` 非空组。
- 本地兼容：本地导入库缺部分 minivod 分表时，计数按 0 处理，避免只读接口因缺表失败；生产完整分表会按 PHP 分表规则查询。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`；游客对比使用本地 `user_guests.sid=0003b1c936338fd7871e3926db105db5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；旧 PHP `/ucp/index` 在本地 GET/POST/header/cookie/guest 场景均超时，未能完成 PHP-Go 响应体对比；新 Go 对应场景均返回 200，字段键和关键空值按源码契约验证。

### `GET /ucp/feedback`

- PHP: `c.api.ucp.index->feedback` 的 GET 分支
- Go: `internal/handler.UCPHandler.FeedbackListing`
- Service: `internal/service/ucp.Service.FeedbackListing`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `feedbacks`；查询条件 `uid=当前用户 uid`；按 `id DESC`；`pagesize=20`；`pageinfo` URL 为 `/ucp/feedback?page=[?]`。
- 兼容规则：只接管 GET，`POST /ucp/feedback` 本轮不注册新 handler；返回 `data.rows/pageinfo`；行字段对齐 PHP `misc.feedback->procRow2`，其中本旧入口未传 `payrow`，所以 `itemname=null`、`paidtime=""`；`ctimestamp/replytime` 使用 `Y-m-d H:i`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 GET `/ucp/feedback?page=1`、`page=0`、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致；Go `POST /ucp/feedback` 返回 404，未接管写入。

### `GET /ucp/msg`、`GET /ucp/msg/index`

- PHP: `c.api.ucp.msg->index`
- Go: `internal/handler.UCPHandler.MsgListing`
- Service: `internal/service/ucp.Service.MsgListing`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `msgc a LEFT JOIN msgs b ON b.msgid=a.last_msgid LEFT JOIN users c ON c.uid=a.ruid`；查询条件 `a.uid=当前用户 uid`；按 `a.last_sendtime DESC`；`pagesize=20`；`pageinfo` URL 为 `/ucp/msg?page=[?]`。
- 兼容规则：只接管 GET `/ucp/msg` 和 GET `/ucp/msg/index`；返回 `data.rows/pageinfo`；行字段保留 PHP SQL 输出字符串/NULL 表现并追加 `__url__=/ucp/msg/show?cid=<cid>`；`sendtime/last_sendtime` 不格式化；旧 PHP 在 `total==0` 时会 `cleanRead(uid)` 写 `users.newmsg=0`，Go 本轮为只读迁移不复刻该副作用。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 GET `/ucp/msg`、`/ucp/msg/index`、`/ucp/msg?page=0`、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致；Go `POST /ucp/msg` 返回 404，未接管旧 `Route::any` 可能触发的写状态行为。

### `GET /ucp/feedback/index`

- PHP: `c.api.ucp.feedback->index`
- Go: `internal/handler.UCPHandler.FeedbackIndex`
- Service: `internal/service/ucp.Service.FeedbackIndex`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `trade_payments`；查询条件 `uid=当前用户 uid AND createtime > now-30天`；按 `payid DESC`；最多 100 条。
- 兼容规则：只接管 GET `/ucp/feedback/index`；返回 `data.payrows`；支付行复用 PHP `payment.procRow2` 兼容字段；裸 `/ucp/feedback` 仍是旧版 legacy feedback 列表，不指向新版 index。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 GET `/ucp/feedback/index`、`/ucp/feedback/index?page=1`、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致；Go `POST /ucp/feedback/index` 返回 404，未接管旧 `Route::any`。

### `GET /ucp/feedback/listing`

- PHP: `c.api.ucp.feedback->listing`
- Go: `internal/handler.UCPHandler.FeedbackNewListing`
- Service: `internal/service/ucp.Service.FeedbackNewListing`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `feedbacks`；基础条件 `uid=当前用户 uid`；`type=1` 增加 `cid IN (0,1,2,3,4)`；`type=2` 增加 `cid IN (5,6,7)`；其他 type 归一为 0；按 `id DESC`；`pagesize=20`。
- 兼容规则：只接管 GET `/ucp/feedback/listing`；返回 `data.rows/pageinfo`；行字段复用 PHP `misc.feedback->procRow2`，本入口未传 `payrow`，所以 `itemname=null`、`paidtime=""`；`pageinfo` URL 为 `/ucp/feedback/listing?type={0|1|2}&page=[?]`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 GET `/ucp/feedback/listing`、`type=1&page=1`、`type=2&page=1`、非法 type + `page=0`、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致；Go `POST /ucp/feedback/listing` 返回 404，未接管旧 `Route::any`。

### `GET /ucp/feedback/detail`

- PHP: `c.api.ucp.feedback->detail`
- Go: `internal/handler.UCPHandler.FeedbackDetail`
- Service: `internal/service/ucp.Service.FeedbackDetail`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `feedbacks` 单条记录并校验 `row.uid=当前用户 uid`；`row.aids` 非空时按原顺序读取 `attachs`，SQL 对齐 PHP `ORDER BY FIELD(aid, ...)`；`row.payid>0` 时读取 `trade_payments` 单条记录。
- 兼容规则：只接管 GET `/ucp/feedback/detail`；返回 `data.row/picurls`；`row` 字段复用 PHP `misc.feedback->procRow2`，关联支付只用于 `itemname/paidtime`，不额外校验 payment uid；空附件返回 `picurls=null`；记录不存在或越权返回 PHP 默认错误壳，不带 `data`；Go `POST /ucp/feedback/detail` 返回 404，未接管旧 `Route::any`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 GET `/ucp/feedback/detail?id=1917132` 成功响应一致，`picurls=null`、无 payment 时 `itemname=null/paidtime=""`；GET `id=0` 的不存在错误壳一致；未登录分支除旧 PHP 动态 `xxx_api_auth` 外语义一致。

### `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last`、`/onego/hash`

- PHP: `c.api.onego->rules/rooms/current/last/hash`
- Go: `internal/handler.OneGoHandler`
- Service: `internal/service/onego.Service`
- Repository: `internal/repository/onego.Repository`
- Auth: 公共接口，不要求登录；旧 `c.api.__init__` 可能在 `data.xxx_api_auth` 写入动态游客 token，Go 本轮不生成该动态字段。
- DB: `/onego/rules` 读取 `one_go LIMIT 1`；`/onego/rooms` 读取 `one_go_rooms ORDER BY id ASC LIMIT 10`；`/onego/current` 读取当前房间未开奖记录；`/onego/last` 读取最近已开奖 period 或指定 room 的已开奖记录；`/onego/hash` 不访问 DB。
- 兼容规则：规则表无数据时返回 PHP 默认错误壳 `retcode=-1`、`errmsg=系统尚未开放该活动`；`last` 无已开奖记录返回 `暂无数据`；记录行按 PHP `onego.record->procRow` 将核心数字字段转 int，并按 winner 查 `users` 或 `bot_users`；`hash` 对 `plaintext` trim 后计算 SHA256，提取 hash 中末尾 6 位数字，首位为 0 时继续向前取，空参数返回 `请传入参数`；支持 GET/POST，匹配旧 `Route::any('/onego/?(:action)?')` 的 method 范围。
- 测试：聚焦 `go test ./internal/service/onego ./internal/server` 通过；PHP-Go 对比 `/onego/rules` 本地空表错误一致；`/onego/rooms` GET/POST 房间列表业务数据一致；`/onego/current?roomid=1` 和 `/onego/last` 本地错误分支一致；`/onego/hash?plaintext=abc` 和缺少 plaintext 错误分支一致；均忽略旧 PHP 动态 `data.xxx_api_auth`。
