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
