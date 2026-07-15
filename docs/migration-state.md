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

### `/logout`

- PHP: `c.api.user->logout`
- Go: `internal/handler.UserHandler.Logout`
- Service: `internal/service/user.LogoutService`
- Repository: `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；token 为空、非法或 session 不存在也返回成功。旧 PHP 可能在无/非法 token 响应里写入动态 `data.xxx_api_auth`，Go 不生成该字段。
- DB: 执行 `DELETE FROM sessions WHERE sid=? AND type=0`；非法 sid 不访问 DB。
- 兼容规则：成功壳为 `retcode=0`、`errmsg=已退出`，空 `data` 省略。
- 测试：聚焦 `go test ./internal/service/user ./internal/server` 通过；PHP-Go 对比 GET `/logout` 无 token 和 POST `/logout` 非法 `x-cookie-auth` 分支忽略动态 `xxx_api_auth` 后一致；有效 token 删除分支由 service fake 覆盖，未直接删除共享测试 token。

### `/sms`、`/sms/index`、`/email`、`/email/index`

- PHP: `c.api.sms->index`、`c.api.email->index`
- Go: `internal/handler.EmptyHTML`
- Auth: 公共默认入口，不要求登录。
- 兼容规则：四个路径均返回 HTTP 200、`Content-Type: text/html`、空 body；`/sms/sendv`、`/sms/sendu`、`/email/send` 涉及验证码、平台发送、频控和风控，仍未接管。
- 测试：聚焦 `go test ./internal/server` 通过；PHP-Go 对比四个路径的 status/content-type/body 一致。

### `/aiundress/index`

- PHP: `c.api.aiundress->index`
- Go: `internal/handler.EmptyHTML`
- Auth: 按本地旧 PHP 运行时行为，`/aiundress/index` 不返回 JSON 业务数据。
- 兼容规则：返回 HTTP 200、`Content-Type: text/html`、空 body；上传、生成、第三方查询等 AI action 仍未迁移。
- 测试：聚焦 `go test ./internal/server` 通过；PHP-Go 对比 `/aiundress/index` 的 status/content-type/body 一致。

### `/captcha/req`

- PHP: `c.api.captcha->req`
- Go: `internal/handler.CaptchaHandler.Req`
- Service: `internal/service/captcha.Service`
- 稳定字段：`retcode=0`、`errmsg=""`、`data.smscaptcha=1`、`data.picurl` 前缀 `/captcha/picx?`
- 动态字段：`data.picurl` query secret
- 测试：`make ci` 通过；PHP-Go 对比通过，按动态 secret 规则只比较 shape。

### `/captcha/pic`、`/captcha/picx`

- PHP: `c.api.captcha->pic/picx`
- Go: `internal/handler.CaptchaHandler.Pic/PicX`
- Service: `internal/service/captcha.Service.PNG`
- Auth: 公共接口，不要求登录或验证码；旧 `c.api.__init__` 会跳过 `/captcha` 签名校验，仍可能创建游客并写动态 `xxx_api_auth`，Go 不生成该动态字段。
- 兼容规则：两个接口都读取完整 raw query string 作为 secret，不读取 `secret=` 命名参数；Go 复刻 PHP `encrypt/decrypt($value, "28ea4")` 的 base64 + hex secret 形态；无效或过期 secret 返回 HTTP 404、`retcode=-4`、`errmsg=验证码无效`、空 `data` 对象；有效 secret 返回 `Content-Type: image/png`、100x34 PNG。PNG 内容不做字节级一致，因为旧 PHP 使用随机颜色、干扰线和字体。
- 测试：聚焦 `go test ./internal/service/captcha ./internal/server` 通过；PHP-Go 对比 `/captcha/pic`、`/captcha/pic?bad`、`/captcha/picx`、`/captcha/picx?bad` 忽略动态 `xxx_api_auth` 后完全一致；固定 PHP secret `1234.2000000000` 和 Go `/captcha/req` 生成 secret 均可输出 100x34 PNG。

### `/test`

- PHP: `c.api.test->test`
- Go: `internal/handler.TestHandler.Test`
- Service: `internal/service/captcha.TestImageService`
- 兼容规则：旧 PHP 使用随机两个中文字符和 GD 输出 100x34 PNG，不返回 JSON，也不持久化验证码；Go 保持 `HTTP 200`、`Content-Type: image/png`、100x34 动态 PNG 输出。当前实现不硬编码 PHP 绝对字体路径，避免 Docker/K8s 环境缺资源。
- 动态字段：图片内容每次生成不同，只按 status、content-type、PNG magic 和 IHDR 尺寸做对比。
- 测试：聚焦 `go test ./internal/service/captcha ./internal/server` 通过；PHP-Go 形态对比通过。

### `/attach`、`/attach/index`、`/attach/upavatar`

- PHP: `c.api.attach->index/upavatar`
- Go: `internal/handler.AttachHandler`
- Service: `internal/service/attach.Service`
- Repository: `internal/repository/user.Repository`
- Auth: `upavatar` 需要登录，兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。旧 PHP 可能在未登录响应 `data.xxx_api_auth` 写入动态游客 token，Go 不生成该动态字段。
- DB: `upavatar` 成功分支执行 `UPDATE users SET avatar=? WHERE uid=?`，只接受纯数字系统头像 id。
- 兼容规则：`/attach` 和 `/attach/index` 是 PHP 空方法，返回 HTTP 200、`Content-Type: text/html`、空 body；`upavatar` 的非法 `avatarid` 返回 PHP 默认错误壳，`retcode=-1` 且不带 `data`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/attach ./internal/server` 通过；PHP-Go 对比 `/attach`、`/attach/index` 空响应一致；POST `/attach/upavatar avatarid=1` 未登录分支除旧 PHP 动态 `xxx_api_auth` 外语义一致；带测试 token 的 POST `/attach/upavatar avatarid=abc` 错误壳一致；成功更新分支由 service fake 覆盖，未直接改动本地 PHP/Go 共享测试用户头像。

### `/:size/:uri`

- PHP: `c.api.pic->index`
- Go: `internal/handler.PicHandler.Index`
- Service: `internal/service/pic.Service`
- Config: `UPLOAD_PATH`，默认 `/Users/canavs/xjProj/XJBackend/api/res`，对应旧 PHP `conf/upload.php` 的 `upload_path`。
- Auth: 公共图片资源入口，不要求登录；该路由在 Go 中按旧 PHP size 白名单显式注册 `C1..C9/T1..T9/R1..R9/M/N`，避免通配误吞其他业务路径。
- 兼容规则：保留 `getAbsPath` 风格的路径规整和 `jpg/jpeg/gif/png` 扩展校验；`N` 输出原图；`C*` 中心裁剪到固定尺寸；`T*` 宽度缩略且不放大小图；`R*` 白底等比缩放；无效 size、非法扩展或文件不存在返回 HTTP 404、`Content-Type: text/html`、空 body。`M` 当前按原图输出，未复刻旧 PHP 的 `data/waterlogo.png` 水印叠加，后续如果业务依赖水印字节效果需单独补齐。
- 测试：聚焦 `go test ./internal/service/pic ./internal/server` 通过；PHP-Go 对比 `/C1/missing.png`、`/N/missing.jpg`、`/C1/not-image.txt` 的 404 status/content-type/body 一致；图片生成分支用临时 PNG 覆盖原图、裁剪和缩略尺寸。

### `/shortcutstats/add`、`/adstats/add`、`/playstats/add`

- PHP: `c.api.shortcutstats->add`、`c.api.adstats->add`、`c.api.playstats->add`
- Go: `internal/handler.StatsHandler`
- Service: `internal/service/stats.Service`
- Repository: `internal/repository/stats.Repository` + `internal/repository/user.Repository`
- Auth/Guest: `shortcutstats/add` 不要求登录；`adstats/add`、`playstats/add` 需要用户 uid 或游客 sid。Go 局部复刻旧 `c.api.__init__` 无 token 游客副作用：按 `md5(sprintf("%x", crc32(IP)))` 生成 sid，并 `INSERT IGNORE user_guests(sid,goldcoin,timestamp)`；旧 PHP 成功响应可能带动态 `data.xxx_api_auth`，Go 不生成该字段。
- DB: `shortcutstats/add` 读写 `shortcut_created` 和 `shortcut_stats`；`adstats/add` 读写 `ad_stats`，同 `sid/title/url` 已存在时更新 `click/install`；`playstats/add` 读写 `play_stats`，已存在时仅当新 `played` 更大才更新。
- 兼容规则：成功返回 `retcode=0`、`errmsg=""`，空 `data` 省略；`adstats/add` 缺少 `title/url` 返回 `缺少参数`，`click<=0` 或 `pos<=0` 返回 `无效参数`；`install>=1` 时强制 `install=1`、`click=1`；`playstats/add` 的 `vid<=0` 或 `duration<=0` 返回 `无效参数`。
- 测试：聚焦 `go test ./internal/service/stats ./internal/server` 通过；PHP-Go 对比 `/shortcutstats/add`、无 token POST `/adstats/add title=a&url=b&pos=1&click=1` 成功分支忽略动态 `xxx_api_auth` 后一致；带测试 token 的 `/adstats/add` 缺 title 和 `/playstats/add` 无效参数错误壳一致。

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

### `/v2/vod/show/:vodid`

- PHP: `c.apiv2.vod->show`
- Go: `internal/handler.VODHandler.Show`
- Service: `internal/service/vod.ListingService.Show`
- Repository: `internal/repository/vod.ListingRepository`
- Auth: 公共详情接口，不要求登录；旧 PHP 可能在错误响应 `data.xxx_api_auth` 写入动态游客 token，Go 不生成该动态字段。
- 兼容规则：复用 `/vod/show/:vodid` 详情迁移，返回 `data.vodrow/parentrows/similarrows/likerows`；错误分支 `retcode=-1`、`errmsg=记录不存在或已删除`。
- 测试：聚焦 `go test ./internal/server ./internal/service/vod` 通过；PHP-Go 对比 `/v2/vod/show/0`、`/v2/vod/show/100` 错误分支忽略动态 `xxx_api_auth` 后一致；`/v2/vod/show/1` 成功详情 status/retcode/资源 URL 和核心字段形态对齐，随机相似/喜欢列表不做逐条完全相等。

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

### `/vod/up/:vodid`、`/vod/down/:vodid`、`/v2/vod/up/:vodid`、`/v2/vod/down/:vodid`

- PHP: `c.api.vod->up/down`、`c.apiv2.vod->up/down`
- Go: `internal/handler.VODHandler.Up/Down`
- Service: `internal/service/vod.ListingService`
- Repository: `internal/repository/vod.ListingRepository`
- Auth: 旧 PHP 支持游客赞踩；Go 登录用户通过 `x-cookie-auth`/`xxx_api_auth` 查询 session，游客用稳定进程内 limiter key。
- DB: 成功时登录用户写 `vod_updowns`，并更新 `vods.upnum/downnum` 后按 `vod_updowns` 重新统计；游客分支只更新 `vods` 计数并使用 limiter 防重复。
- 兼容规则：视频不存在或 `showtype>0` 返回 `记录不存在或已被删除`；赞/踩/取消赞/取消踩返回 PHP 同款 `errmsg`。
- 测试：`go test ./internal/service/vod ./internal/server` 通过；PHP-Go live 对比 `/vod/up/0`、`/v2/vod/down/0` 无效视频分支一致；状态切换成功/重复分支由 service fake 覆盖。

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

### `/comment/up`、`/comment/down`

- PHP: `c.api.comment->up/down`
- Go: `internal/handler.CommentHandler.Up/Down`
- Service: `internal/service/comment.Service`
- Repository: `internal/repository/comment.Repository`
- Auth: 旧 PHP 允许游客赞踩，中间件通常会创建游客 sid；Go 无 token 时使用稳定游客 actor 仅做重复限制 key，不写游客表。
- DB: 读取 `vod_comments`，成功时更新 `upnum=upnum+1` 或 `downnum=downnum+1`。
- 兼容规则：评论不存在返回 `记录不存在或已被删除`；同一 actor 对同一评论重复赞/踩返回 `您已经赞/踩过了`；成功返回 `errmsg=已赞/已踩`。
- 测试：`go test ./internal/service/comment ./internal/server` 通过；PHP-Go live 对比 `/comment/up?id=0` 和登录态 `/comment/down?id=0` 错误分支一致；成功和重复分支由 service fake 覆盖。

### `/playlog`、`/playlog/index`、`/downlog`、`/downlog/index`

- PHP: `c.api.playlog->index`、`c.api.downlog->index`
- Go: `internal/handler.EmptyHTML`
- 兼容规则：旧 PHP 空方法，返回 HTTP 200、`Content-Type: text/html`、空 body。
- 测试：聚焦 `go test ./internal/server` 通过；PHP-Go 对比四个路径的 status/content-type/body 一致。

### `/playlog/listing`、`/playlog/remove`、`/downlog/listing`、`/downlog/remove`

- PHP: `c.api.playlog->listing`、`c.api.downlog->listing`
- Go: `internal/handler.HistoryHandler`
- Service: `internal/service/history.Service`
- Repository: `internal/repository/history.Repository`
- Auth: 不强制登录；有有效登录 token 时按 `uid` 查询用户记录，无用户时按清洗后的 sid 查询游客记录。旧 PHP 可由中间件创建游客 sid，Go 不生成动态 `xxx_api_auth`。
- Params: `page`、`timeline`；`pagesize=20`；分页 URL 分别为 `/playlog/listing?timeline=N&page=[?]`、`/downlog/listing?timeline=N&page=[?]`。
- DB: 播放读取 `vod_playlogs` 或 `vod_guest_playlogs`；下载读取 `vod_downlogs` 或 `vod_guest_downlogs`；再按 `vodid` 合并 `vods(showtype=0)`，缺失视频标题的日志按 PHP 行为过滤。
- 兼容规则：复用 VOD `ProcessRows` 对齐视频字段；追加 `logid` 和格式化后的 `playtime/downtime`；30 天内按 PHP `formatTime(..., '@d天前 @h小时前 @m分钟前 @s秒前')` 只显示最大非零单位，超过 30 天输出 `Y-m-d`。游客播放 timeline 2/3 保留旧 PHP `BETWEEN` 边界反序导致通常查不到的行为。
- 写入规则：`remove` 读取 `vodid` 或 `vodids`，`vodid>0` 时覆盖 `vodids`；登录用户按 `uid` 软删除，游客按 sid 软删除；播放记录用户态同时更新 `vod_playlogs` 和 `vod_playlogs_week`；返回 `errmsg=已删除N项`。
- 测试：聚焦 `go test ./internal/service/history ./internal/repository/history ./internal/server` 通过；repository 单测锁定 timeline 过滤，service 单测覆盖游客 sid、分页 URL、相对时间、VOD 行处理和 remove 计数。

### `/favorite`、`/favorite/index`、`/favorite/listing`、`/favorite/add`、`/favorite/remove`

- PHP: `c.api.favorite->index/listing/add/remove`
- Go: `internal/handler.FavoriteHandler`
- Service: `internal/service/favorite.Service`
- Repository: `internal/repository/favorite.Repository`
- Auth: `listing/add/remove` 要求登录；未登录返回 `retcode=-9999`、`errmsg=请登录后操作`。`index` 是旧 PHP 空方法。
- DB: `listing` 读取 `vod_favorites LEFT JOIN vods(showtype=0)`，`wd` 非空时按 `title LIKE` 搜索；`add` 校验 `vods.showtype=0` 后写入 `vod_favorites(uid,vodid,favtime)`；`remove` 删除 `vod_favorites WHERE uid=? AND vodid=?`。
- 兼容规则：`listing` 返回 `rows/pageinfo`，复用 VOD `ProcessRows`；分页 URL 为 `/favorite/listing?page=[?]` 或 `/favorite/listing?page=[?]&wd=$wd`；`add` 返回 `errmsg=已收藏`，重复收藏返回 `您已经收藏过了`；`remove` 返回 `errmsg=已删除N项`。
- 资产说明：旧 PHP `add` 可能按收藏任务写 `user_coinlogs` 奖励金币；Go 本轮只迁移收藏写入本身，金币奖励保留后续 rewarder/事务接入点，默认不改用户资产。
- 测试：聚焦 `go test ./internal/service/favorite ./internal/server ./internal/service/vod` 通过；service 覆盖未登录、关键词分页、行处理、add 不存在/重复/成功、删除计数；live 对比 add 未登录和无效 vodid 分支通过。

### `/minifavorite`、`/minifavorite/index`、`/minifavorite/listing`、`/minifavorite/add`、`/minifavorite/remove`

- PHP: `c.api.minifavorite->index/listing/add/remove`
- Go: `internal/handler.FavoriteHandler`
- Service: `internal/service/favorite.Service`
- Repository: `internal/repository/favorite.Repository`
- Auth: `listing/add/remove` 要求登录；`index` 是旧 PHP 空方法。
- DB: `listing` 读取 `minivod_favorites LEFT JOIN vods(showtype=1)`；`add` 校验 `vods.showtype=1` 后写入 `minivod_favorites(uid,vodid,favtime)`；`remove` 删除 `minivod_favorites WHERE uid=? AND vodid=?`。
- 兼容规则：`listing` 复用 VOD `ProcessMiniRows`，并按 PHP 补 `isfavorite=1`；分页 URL 为 `/minifavorite/listing?page=[?]`；`add` 返回 `errmsg=已收藏`，重复收藏返回 `您已经收藏过了`；`remove` 返回 `errmsg=已删除N项`。
- 资产说明：同 `/favorite/add`，旧 PHP 可能发放收藏奖励金币；Go 本轮默认不改用户资产。
- 测试：聚焦 `go test ./internal/service/favorite ./internal/server ./internal/service/vod` 通过；live 对比 add 未登录分支通过。

### `/minivod/{listing,recommend,hot,latest,topzan,topcomment,topplay,topcoin,topnew,topday,topweek,topmonth}`

- PHP: `c.api.minivod->listing`
- Go: `internal/handler.MiniVODHandler.Listing`
- Service: `internal/service/minivod.Service`
- Repository: `internal/repository/minivod.Repository`
- Auth: 公共列表接口，不要求登录。
- Params: `$cateid:0-$areaid:0-$yearid:0-$tagid:0-$definition:0-$duration:0-$freetype:0-$mosaic:0-$langvoice:0-$orderby:0-$page:1`；path page 为 `0` 时使用 query `page`。
- DB: 读取 `vods(showtype=1)`、`vod_categories`、`vod_areas`、`vod_years`、`vod_servers`、`vod_tags`；排行榜 `topzan/topcomment/topplay/topcoin` 读取 `settings` 中 PHP 序列化数组 `minivod.*_vodids` 并按 `FIELD(vodid, ...)` 排序；需要用户包装 rows 的 action 读取 `users`。
- 兼容规则：返回 `now/action/sample_params/params/rows/vodrows/pageinfo/orders/categories/areas/years/definitions/durations/freetypes/mosaics/langvoices`；`recommend` 保持 PHP 随机列表且 `pageinfo.total=0`；`latest/top*` 和 `tagid` 过滤时返回 `rows[].vodrow/user` 包装。
- 测试：聚焦 `go test ./internal/service/minivod ./internal/server ./...` 通过；PHP-Go 对比 `/minivod/listing`、`/minivod/latest`、`/minivod/topday`、`/minivod/recommend`、`/minivod/listing-...-2`，以及 `topzan/topcomment/topplay/topcoin` 的 total、数量和前 5 个 vodid 通过；`recommend` 随机内容只按 shape 对比。

### `/minivod/show/:vodid`

- PHP: `c.api.minivod->show`
- Go: `internal/handler.MiniVODHandler.Show`
- Service: `internal/service/minivod.Service.Show`
- Repository: `internal/repository/minivod.Repository`
- Auth: 公共详情接口，不要求登录。
- DB: 读取 `vods` 中 `showtype=1` 的小视频详情、`users` 作者、`vod_categories` 分类层级、`vod_tags/vod_tagmaps` 相关视频、随机猜你喜欢。
- 兼容规则：小视频不存在或 `showtype!=1` 返回 `记录不存在或已删除`；作者不存在返回 `作者不存在或已被删除`；成功返回 `vodrow/categories/similarrows/likerows/voduser`。PHP 详情未传 `minivod=true` 给 `procRow2`，Go 保持普通 `/vod` 播放/下载/预览 URL 前缀。
- 测试：`go test ./internal/service/minivod ./internal/repository/minivod ./internal/server` 通过；PHP-Go live 对比 `/minivod/show/0` 不存在分支和本地样本 `/minivod/show/56914`、`/minivod/show/76989` 作者缺失分支通过。本地导入库未找到作者存在的小视频样本，成功分支由 service 单测覆盖。

### `/minivod/up/:vodid`、`/minivod/down/:vodid`

- PHP: `c.api.minivod->up/down`
- Go: `internal/handler.MiniVODHandler.Up/Down`
- Service: `internal/service/minivod.Service`
- Repository: `internal/repository/minivod.Repository`
- Auth: 旧 PHP 支持游客赞踩但依赖 `user_guests/keylimit`；Go 登录用户通过 `x-cookie-auth`/`xxx_api_auth` 查询 session，游客分支先用进程内 limiter 保持重复限制。
- DB: 登录用户写 `vod_updowns` 并更新 `vods.upnum/downnum` 后按 `vod_updowns` 重新统计；游客分支只更新 `vods` 计数并用 limiter 防重复。
- 兼容规则：视频不存在或 `showtype!=1` 返回 `记录不存在或已被删除`；赞/踩/取消赞/取消踩返回 PHP 同款 `errmsg`。
- 测试：`go test ./internal/service/minivod ./internal/repository/minivod ./internal/server` 通过；PHP-Go live 对比 `/minivod/up/0`、`/minivod/down/0` 无效视频分支一致；状态切换成功/重复分支由 service fake 覆盖。

### `/minivod/reqlong/:vodid`

- PHP: `c.api.minivod->getLong2Mini`
- Go: `internal/handler.MiniVODHandler.ReqLong`
- Service: `internal/service/minivod.Service`
- Repository: `internal/repository/minivod.Repository`
- DB: 读取 `vods` 和播放服务器 `vod_servers`。
- 兼容规则：只接受普通长视频 `showtype=0`；不存在返回 JSON `retcode=1`、播放地址为空返回 `retcode=2`；成功直接返回 `text/html` URL。保留旧 PHP 的播放地址清洗、腾讯云/华为云 CDN 签名和相对地址播放服务器 host 补全。
- 测试：`go test ./internal/service/minivod ./internal/server` 通过；PHP-Go live 对比 `/minivod/reqlong/0` 错误分支一致，`/minivod/reqlong/1` 本地样本返回同一播放 URL。

### `/miniplaylog/listing`、`/miniplaylog/remove`

- PHP: `c.api.minivod->history/historyDelete`
- Go: `internal/handler.HistoryHandler.MiniPlayListing/MiniPlayRemove`
- Service: `internal/service/history.Service`
- Repository: `internal/repository/history.Repository`
- Auth: 不强制登录；登录用户按 `uid`，游客按 `sid`。
- DB: 登录用户读取 `minivod_viewlogs_%02d`，游客读取 `minivod_guestviewlogs_<sid首字符>`，再关联 `vods(showtype=1)`；删除按旧 PHP 模型语义从分表物理删除 `logid`。
- 兼容规则：列表 `pagesize=20`，timeline 1/2/3 分别对应最近一周、近一个月、一个月前到三个月内；返回 `rows/pageinfo`，`playtime` 保持 PHP 相对时间/日期格式；列表行使用小视频 URL 前缀。
- 测试：`go test ./internal/repository/history ./internal/service/history ./internal/server` 通过；`POST /miniplaylog/remove` 空参数 PHP-Go live 对比均为 `已删除0项`；本地旧 PHP `/miniplaylog/listing` 在 5 秒内无响应，新 Go smoke 返回空列表和正确分页 URL。

### `/my/:authorid`、`/my/:authorid/index`、`/my/:authorid/listing`

- PHP: `c.api.my->index/listing`
- Go: `internal/handler.MiniVODHandler.Author`
- Service: `internal/service/minivod.Service.AuthorListing`
- Repository: `internal/repository/minivod.Repository`
- Auth: 公共作者主页接口，不要求登录。
- DB: 读取 `users` 作者信息和 `vods(authorid=?, showtype=1)` 小视频列表。
- 兼容规则：用户不存在返回 `用户不存在或已被删除`；成功返回 `now/userrow/vodrows/pageinfo/orders`，`userrow.uniqkey` 使用 base36 大写，头像 URL 使用 `/C1/avatar/` 规则，列表行使用小视频 URL 前缀。
- 测试：`go test ./internal/service/minivod ./internal/server ./...` 通过；PHP-Go live 对比 `/my/1?page=1` 的 `uid/uniqkey/avatar_url/pageinfo/orders/vodrows count` 通过，`/my/999999999` 错误分支通过。

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

### `/getCertUuid`

- PHP: `c.api.index->getCertUuid`
- Go: `internal/handler.IndexHandler.GetCertUUID`
- Service: `internal/service/index.CertService`
- Repository/Client: `internal/repository/index.SettingsRepository` 读取 `settings.uuid='setting'`，`internal/service/index.CertHTTPClient` 封装外部 HTTP。
- 兼容规则：读取 PHP 序列化设置中的 `getCertUrl`，为空时使用默认 `https://api.apkcdn.cc/api/get_cert`；将 query `uuid` 透传给外部接口；外部响应 JSON `code==0` 时返回 `data.data`，否则返回 `retcode=-1`、`errmsg=记录不存在或已被删除`、空 `data` 对象。旧 PHP 中间件动态 `data.xxx_api_auth` 不生成。
- 测试：聚焦 `go test ./internal/service/index ./internal/server` 通过；PHP-Go 对比 `/getCertUuid`、`/getCertUuid?uuid=test` 忽略动态 `xxx_api_auth` 后完全一致；成功分支通过 fake client 单测覆盖，不依赖真实外部证书服务。

### `/search`

- PHP: `c.api.search->index`
- Go: `internal/handler.VODHandler.Search`
- Service: `internal/service/vod.ListingService.Search`
- Repository: `internal/repository/vod.ListingRepository`
- DB: 空关键词读取 `maintain_calldata(search.hotwords/search.hotvods)`、`vod_schlogs` 最高搜索记录和 `vods`；带关键词读取/写入 `vod_schlogs`，并从 `vods` 搜索 `title/tags/actor_tags/vodkey`。
- 兼容规则：`wd` 为空返回 `data.hotwords/hotrows/you_may_likes`；`wd` 非空返回 `data.vodrows/pageinfo`，`pagesize=16`；`free=1` 过滤免费影片；保留 PHP 的搜索日志 `REPLACE/UPDATE` 写入副作用；搜索页热片只用 `tags` 映射标签，关键词列表使用 `tags + actor_tags`；`vip_price` 使用 PHP 默认 100% 折扣，不使用 VOD 列表配置折扣。
- 测试：聚焦 `go test ./internal/service/vod ./internal/server` 通过；PHP-Go 对比 `/search`、`/search?wd=标题&page=1`、`/search?wd=标题&free=1&page=1` 忽略动态 `xxx_api_auth` 后递归一致。

### `/game/wali/gameList`

- PHP: `c.api.game.wali->games`
- Go: `internal/handler.GameHandler.WaliGames`
- Service: `internal/service/game.ListingService`
- Repository: `internal/repository/game.GameRepository`
- DB: 读取 `game`；固定 `platform_id=1`；普通分类支持 `category_id`；按 ``order`` DESC，limit 100。
- 兼容规则：普通分类返回 `data.data` 游戏数组并拼接资源 URL；`category_id=5` 是常玩游戏，需要登录历史，当前无鉴权上下文时按旧 PHP 游客行为返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- 测试：聚焦 `go test ./internal/server ./internal/service/game` 通过；PHP-Go 对比 `/game/wali/gameList`、`/game/wali/gameList?category_id=2`、`/game/wali/gameList?category_id=5` 忽略动态 `xxx_api_auth` 后完全一致。

### `/game/lottery/gameList`

- PHP: `c.api.game.lottery->gameList`
- Go: `internal/handler.GameHandler.LotteryGames`
- Service: `internal/service/game.ListingService`
- Repository: `internal/repository/game.GameRepository`
- DB: 读取 `game`；固定 `platform_id=0`；普通分类支持 `category_id`；按 ``order`` DESC，limit 100。
- 兼容规则：普通分类返回 `data.data` 游戏数组并拼接资源 URL；`category_id=5` 是常玩游戏，需要登录历史，当前无鉴权上下文时按旧 PHP 游客行为返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- 测试：聚焦 `go test ./internal/server` 通过；普通分类空库响应和 `category_id=5` 游客错误分支由路由测试覆盖。旧 PHP 本地请求在 `Lottery::__construct/initApi` 阶段依赖外部平台配置，`curl --max-time 5` 无响应，未完成 live 对比。

### `/game/wali/test`

- PHP: `c.api.game.wali->ping`
- Go: `internal/handler.GameHandler.WaliTest`
- Service: `internal/service/game.WaliService`
- Repository/Client: `internal/repository/game.PlatformRepository.PlatformByID` 读取 `game_platform.id=1` 的 `config_json`；`internal/service/game.WaliHTTPClient` 封装外部 HTTP。
- 兼容规则：读取 `json.url/account/aesKey/signKey/agentId`，使用 PHP 兼容 AES-128-ECB + PKCS#7 加密 `text=helloThere`，再按 `md5(p + unixTime + signKey)` 签名；外部返回 `code=0` 时把 `data.code/msg` 补入 `data.data`，失败时返回 `retcode=-1`、`errmsg=测试失败`。
- 测试：聚焦 `go test ./internal/service/game ./internal/server` 通过；单测断言 Go AES 密文与 PHP `openssl_encrypt(..., aes-128-ecb)` 一致；PHP-Go live 对比 `/game/wali/test` 忽略动态 `xxx_api_auth` 后完全一致。

### `/game/wali/balance`

- PHP: `c.api.game.wali->getBalance`
- Go: `internal/handler.GameHandler.WaliBalance`
- Service: `internal/service/game.WaliService.Balance`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`、空 `data` 对象。
- 兼容规则：登录后用当前用户 `uid` 调用瓦力 `getBalance`；复用 `game_platform.id=1` 的 AES-128-ECB 加密与 `md5(p + unixTime + signKey)` 签名；成功只返回 PHP 保留的 `status/balance/transferable` 三个字段，不暴露外部响应的 `code/msg`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/game ./internal/server` 通过；PHP-Go 对比未登录 `/game/wali/balance` 和登录 header token 分支均完全一致。

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

### `/ucp/withdraw`、`/ucp/withdraw/index`

- PHP: `c.api.ucp.withdraw->index`
- Go: `internal/handler.UCPHandler.WithdrawIndex`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository` + `internal/repository/index.SettingsRepository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `users_account`、`users_quota`、`user_bankcards`、`settings(setting/game.setting)`。
- 兼容规则：返回 `account/cardrows/goldcoin/exrate/topupmin/coin2rmb/max2rmb/game_withdrawmin/game_withdrawrate/alipay_withdraw_min/alipay_withdraw_max/bankcard_withdraw_min/bankcard_withdraw_max`；`topupmin/coin2rmb/max2rmb` 为元字符串，提现上下限保持旧配置原始整数值；`create` 提现写入仍未接管。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go live 对比 `/ucp/withdraw`、`/ucp/withdraw/index` 未登录分支和测试 token 登录成功分支字段值一致，忽略旧 PHP 动态游客 token。

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
- 兼容规则：GET 返回 `data.rows/pageinfo`；行字段对齐 PHP `misc.feedback->procRow2`，其中本旧入口未传 `payrow`，所以 `itemname=null`、`paidtime=""`；`ctimestamp/replytime` 使用 `Y-m-d H:i`。POST 已迁移到 `UCPHandler.FeedbackCreateLegacy`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 GET `/ucp/feedback?page=1`、`page=0`、cookie token 和未登录分支忽略动态 `xxx_api_auth` 后完全一致。

### `POST /ucp/feedback`、`/ucp/feedback/create`

- PHP: `c.api.ucp.index->feedback`、`c.api.ucp.feedback->create`
- Go: `internal/handler.UCPHandler.FeedbackCreate`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository`
- Auth: 必须登录；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 新版创建校验最近一日反馈数量，读取 `trade_payments` 校验订单归属，并写入 `feedbacks`。
- 兼容规则：内容为空或超过 250 字符返回 `内容最多250个字符`；`cid=5` 必须提供本人订单；每日反馈超过限制返回 `当日反馈内容过多`；成功返回 `errmsg=信息已反馈`。
- 风险说明：本轮不落盘上传图片、不写 `attachs.ownerkey`、不发送 Telegram/外部告警；这些外部链路保留后续补齐。
- 测试：`go test ./internal/service/ucp ./internal/repository/ucp ./internal/server` 通过；PHP-Go live 对比两个创建入口未登录分支一致；成功、订单归属和内容校验由 service fake 覆盖。

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

### `/ucp/msg/show`

- PHP: `c.api.ucp.msg->show`
- Go: `internal/handler.UCPHandler.MsgDetail`
- Service: `internal/service/ucp.Service.MsgDetail`
- Repository: `internal/repository/ucp.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。旧 PHP 可能在未登录响应 `data.xxx_api_auth` 写入动态游客 token，Go 不生成该动态字段。
- DB: 读取 `msgc` 单条会话，要求 `uid=当前用户 uid AND cid=?`；读取对方 `users`；读取 `msg_maps LEFT JOIN msgs LEFT JOIN users`，按 `a.sendtime ASC`，`pagesize=100`；成功后执行 PHP `setRead` 副作用：当前会话 `newmsg=0`、对方会话 `risread=1`、当前用户 `users.newmsg -= 原会话 newmsg`。
- 兼容规则：会话不存在返回 `retcode=-1`、`errmsg=您的会话不存在` 且不带 `data`；`cuser` 不存在时返回空数组 `[]`；消息行复用 `procRow` 链接归一化和 `__url__=/ucp/msg/show?cid=<cid>`。
- 测试 token: `3235306637393062613731656332623964333835356634323464623232353965`，对应 `uid=5`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比未登录、`cid=0` 会话不存在、`cid=9776805` 成功分支通过；成功分支会按旧 PHP 标记已读，因此 `crow.newmsg` 可能随请求顺序变为 `0`。

### `/ucp/msg/send`

- PHP: `c.api.ucp.msg->send`
- Go: `internal/handler.UCPHandler.MsgSend`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository`
- Auth: 必须登录；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 会话内回复写入 `msgs`，为双方写入 `msg_maps`，更新或创建双方 `msgc`，并递增接收方 `users.newmsg`。
- 兼容规则：空内容返回 `请填写信息内容`；超过 2000 字返回 `信息内容不能超过2000字`；回复会话不存在返回 `您回复的会话不存在`；接收方不存在返回 `接收方不存在`；成功返回 `发送成功`。
- 风险说明：PHP 源码用户名群发分支存在变量遮蔽 bug，`$user['uid'] != $user['uid']` 永远为 false；Go 保持该分支不可用并返回 `请选择一个用户`，不误开群发。
- 测试：`go test ./internal/service/ucp ./internal/repository/ucp ./internal/server` 通过；PHP-Go live 对比 `/ucp/msg/send` 未登录分支一致；会话回复成功和群发 bug 分支由 service fake 覆盖。

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

### `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last`、`/onego/hash`、`/onego/lucky`、`/onego/marquee`

- PHP: `c.api.onego->rules/rooms/current/last/hash/lucky/marquee`
- Go: `internal/handler.OneGoHandler`
- Service: `internal/service/onego.Service`
- Repository: `internal/repository/onego.Repository`
- Auth: 公共接口，不要求登录；旧 `c.api.__init__` 可能在 `data.xxx_api_auth` 写入动态游客 token，Go 本轮不生成该动态字段。
- DB: `/onego/rules` 读取 `one_go LIMIT 1`；`/onego/rooms` 读取 `one_go_rooms ORDER BY id ASC LIMIT 10`；`/onego/current` 读取当前房间未开奖记录；`/onego/last` 读取最近已开奖 period 或指定 room 的已开奖记录；`/onego/hash` 不访问 DB；`/onego/lucky` 读取 `one_go_records` 的 `SUM(awards), winner` 排行和每个 winner 的 `COUNT(*), room_id` 获奖次数；`/onego/marquee` 先查最近已开奖记录，再查规则和最近 period 的 `one_go_records ORDER BY id DESC LIMIT 10`，按 `room_id` 查房间名。
- 兼容规则：规则表无数据时返回 PHP 默认错误壳 `retcode=-1`、`errmsg=系统尚未开放该活动`；`last/marquee` 无已开奖记录返回 `暂无数据`；记录行按 PHP `onego.record->procRow` 将核心数字字段转 int，并按 winner 查 `users` 或 `bot_users`；`hash` 对 `plaintext` trim 后计算 SHA256，提取 hash 中末尾 6 位数字，首位为 0 时继续向前取，空参数返回 `请传入参数`；`lucky` 保留旧 PHP `getRankWinCoins` 虽接收 page/pagesize 但 SQL 不分页的行为，并对 `total_awards/winner/wins/room_id` 转 int；`marquee` 过滤 `awards=0`，房间缺失跳过，按 `one_go.marquee` 替换 `{user}/{room}/{period}/{awards}/{win_rate}`；支持 GET/POST，匹配旧 `Route::any('/onego/?(:action)?')` 的 method 范围。
- 测试：聚焦 `go test ./internal/service/onego ./internal/server` 通过；PHP-Go 对比 `/onego/rules` 本地空表错误一致；`/onego/rooms` GET/POST 房间列表业务数据一致；`/onego/current?roomid=1` 和 `/onego/last` 本地错误分支一致；`/onego/hash?plaintext=abc` 和缺少 plaintext 错误分支一致；`/onego/lucky` GET/POST 业务数据一致；`/onego/marquee` GET/POST 本地无最新期错误一致；均忽略旧 PHP 动态 `data.xxx_api_auth`。

### `/special/index`、`/special/listing`、`/special/listing-:params`、`/special/detail/:spid`、`/special/detail/:spid-:params`

- PHP: `c.api.special->index/listing/detail`
- Go: `internal/handler.SpecialHandler`
- Service: `internal/service/vod.SpecialService`
- Repository: `internal/repository/vod.ListingRepository`
- Auth: 公共接口，不要求登录；旧 PHP 会从响应 data 中删除动态 `xxx_api_auth`，Go 不生成该字段。
- DB: `index` 不访问 DB；`listing` 读取 `vod_specials`，条件 `showtype=0 AND itemcount>=4`，可按 `sptype` 过滤，`pagesize=16`；每个专题取前 4 个 `vodids` 再查 `vods`；第一页额外读取 `sptype=1` 的 `actorrows` 最多 100 条；`detail` 读取单个 `vod_specials` 和全量 `vodids` 对应公开 `vods`。
- 兼容规则：`index` 返回 HTTP 200、`Content-Type: text/html`、空 body；`listing` 参数样例 `$sptype:0-$orderby:0-$page:1`，path page 为 0 时才读 query/form page；`orderby=1/2/3` 分别对应 addtime 降/升、randnum 升，`orderby=3` 保留 PHP 更新 randnum 的写入副作用；`detail` 参数样例 `$orderby:0`，默认最终按专题 `vodids` 原始顺序返回，`1/2/3` 分别按播放量降/升、vodid 升；`detail` 成功保留浏览数写入副作用；专题行和嵌套 VOD 行对齐 PHP 空值、类型、时长、`vip_price` 默认 100%。
- 测试：聚焦 `go test ./internal/service/vod ./internal/server` 通过；PHP-Go 对比 `/special/index` 空响应一致；递归对比 `/special/listing`、`/special/listing-0-0-0?page=1`、`/special/detail/0`、`/special/detail/0-1`、真实专题 `/special/detail/{spid}` 及 `-1/-2/-3` 均一致。

### `/special/up/:spid`、`/special/down/:spid`

- PHP: `c.api.special->up/down`
- Go: `internal/handler.SpecialHandler.Up/Down`
- Service: `internal/service/vod.SpecialService.Vote`
- Repository: `internal/repository/vod.ListingRepository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；旧 PHP 对游客不立即拒绝，但要求游客 sid 已存在于 `user_guests`，否则返回 `retcode=-9999`、`errmsg=请登录后操作，客户端游客请先携带信息`。Go 保持该分支；旧 PHP 可能在错误响应 `data.xxx_api_auth` 写入动态游客 token，Go 不生成该动态字段。
- DB: 读取 `vod_specials`，要求 `showtype=0`；按 `special.updown.{spid}.{uid|sid}` 做 `md5` 后读取/写入 `keylimits`；成功时更新 `vod_specials.upnum/downnum` 并按 PHP `reCount` 公式重算 `scorenum`。
- 兼容规则：`up` 和 `down` 都成功返回 `retcode=0`、`errmsg=已赞`，不带 `data`；重复投票返回 `retcode=-1`、`errmsg=您已经赞/踩过了`；记录不存在返回 PHP 默认错误壳。
- 测试：聚焦 `go test ./internal/service/vod ./internal/server` 通过；PHP-Go 对比 `/special/up/0`、`/special/down/0`、`/special/up/99999999` 不存在分支忽略动态 `xxx_api_auth` 后一致；成功和重复投票为写库分支，由 service fake 覆盖，未直接修改共享本地专题计数。

### `/art`、`/art/index`、`/art/announce`、`/art/show`

- PHP: `c.api.art->index/announce/show`
- Go: `internal/handler.ArtHandler`
- Service: `internal/service/art.Service`
- Repository: `internal/repository/art.Repository`
- Auth: 公共接口，不要求登录；旧 `c.api.__init__` 可能在错误响应 `data.xxx_api_auth` 写入动态游客 token，Go 不生成该动态字段。
- DB: `index` 不访问 DB；`announce` 读取 `art_categories` 中 `uuid=announce` 的分类，再读取 `arts` 中 `cateid=公告分类 AND showtype=0` 的记录，按 `utimestamp DESC` 分页；`show` 读取 `arts LEFT JOIN arts_content`，并要求 `showtype=0`。
- 兼容规则：`index` 返回 HTTP 200、`Content-Type: text/html`、空 body；`announce` 的 `pagesize=20`，`page=0` 归一为 1，分页 URL 保留旧 PHP 未定义 `$action` 后生成的 `/art/?page=[?]`；公告行复刻 `art.procRow2` 的 `coverpic/addtime/catename/content` 表现，列表中 `content=null`；`show` 缺少 `artid`、不存在或已删除均返回 `retcode=-1`、`errmsg=记录不存在或已被删除`、空 `data` 对象。
- 测试：聚焦 `go test ./internal/service/art ./internal/server` 通过；PHP-Go 对比 `/art`、`/art/index`、`/art/announce`、`/art/announce?page=0`、`/art/show`、`/art/show?artid=2`、`/art/show?artid=999999` 忽略动态 `xxx_api_auth` 后完全一致。

### `/minisearch`

- PHP: `c.api.miniSearch->index`
- Go: `internal/handler.VODHandler.MiniSearch`
- Service: `internal/service/vod.ListingService.MiniSearch`
- Repository: `internal/repository/vod.ListingRepository`
- Auth: 公共接口，不要求登录或验证码；旧 `c.api.__init__` 仍可能创建游客并写动态 `xxx_api_auth`，Go 不生成该动态字段。
- DB: 空关键词读取 `maintain_calldata(search.minihotwords/search.minihotvods)`、`minivod_schlogs` 最高搜索记录和 `vods.showtype=1`；带关键词读取/写入 `minivod_schlogs`，并从 `vods` 搜索 `title/tags/actor_tags/vodkey`。
- 兼容规则：`wd` 为空返回 `data.hotwords/hotrows/you_may_likes`，其中 `hotrows` 是 `[{vodrow:{...}}]` 包装结构；`wd` 非空返回 `data.rows/pageinfo`，`pagesize=16`；小视频行复用 `vod.procRow2` 但启用 `minivod=true`，所以 `play_url/down_url/preview_url` 使用 `/minivod` 前缀；时长保留 PHP `formatTime(..., 2)` 表现，56 秒为 `"56"`，2 分 19 秒为 `"02:19"`；分页 URL 保留旧 PHP `/search?wd=...&page=[?]`；搜索日志写入 `minivod_schlogs`。
- 测试：聚焦 `go test ./internal/service/vod ./internal/server` 通过；PHP-Go 对比 GET `/minisearch`、`/minisearch?wd=`、`/minisearch?wd=测试&page=1`、`/minisearch?wd=测试&page=2` 和 POST `/minisearch` 表单 `wd=测试&page=1`，忽略动态 `xxx_api_auth` 后完全一致。

### `/open`、`/open/index`、`/open/reqauth`

- PHP: `c.api.open->index/reqauth`
- Go: `internal/handler.OpenHandler`
- Service: `internal/service/open.Service`
- Repository: `internal/repository/user.Repository` + `internal/repository/stats.Repository`
- Auth: 公共接口，不要求登录；有 `x-cookie-auth` 或 `xxx_api_auth` 时按登录用户授权，无 token 时按旧 `__init__` 规则用客户端 IP 生成游客 sid 并确保 `user_guests` 存在。
- 兼容规则：`/open` 和 `/open/index` 按源码空方法返回空 body；本地旧 PHP `/open` 实测为 500 空体。`/open/reqauth` 仅支持旧 PHP 内置 appid `4b4131e49`，非法 appid 返回 `retcode=-1`、`errmsg=请输入正确的appid`；成功时使用 `AES-128-CBC` + `md5(md5Key)[:16]` 加密 `openid`，按 PHP `ksort` 后拼接 `key=value` 生成 MD5 签名；游客 `authrow` 含 `deviceString/headUrl/gender/nickName`，登录用户含 `phoneNumber/headUrl/gender/nickName`。
- 测试：聚焦 `go test ./internal/service/open ./internal/server` 通过；PHP-Go 对比 `/open/reqauth?appid=bad` 和 `/open/reqauth?appid=4b4131e49`，成功路径 `authrow/openid/sign/time` 一致，旧 PHP 游客响应中动态 `data.xxx_api_auth` 忽略。

### `/ucp/msg/setread`、`/ucp/msg/cleanread`、`/ucp/msg/delete`

- PHP: `c.api.ucp.msg->setread/cleanread/delete`
- Go: `internal/handler.UCPHandler.MsgSetRead/MsgCleanRead/MsgDelete`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: `setread` 复用 `SetMsgRead`，逐个 `cid` 清 `msgc.newmsg`、设置对方 `risread=1` 并递减当前用户 `users.newmsg`；`cleanread` 执行 `UPDATE users SET newmsg=0`；`delete` 删除当前用户 `msgc/msg_maps`，递减 `msgs.refcount` 并清理 `refcount=0` 的消息。
- 兼容规则：`cids` 支持 `cids[]`、重复 `cids` 和逗号字符串；空数组、非法 cid 与旧 PHP 一样返回成功；成功响应为 `retcode=0`、`errmsg=操作成功`，不带 `data`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比三接口未登录分支和登录空数组成功分支一致，忽略旧 PHP 游客动态 `data.xxx_api_auth`。

### `/activity/luckyprizes`

- PHP: `c.api.activity->luckyprizes`
- Go: `internal/handler.ActivityHandler.LuckyPrizes`
- Service: `internal/service/activity.LuckyPrizes`
- Auth: 公共接口，不要求登录；旧 PHP 可能在 `data.xxx_api_auth` 写入动态游客 token，Go 不生成该字段。
- 兼容规则：返回固定 5 个充值抽奖奖项，字段为 `keyid/prizename`，外层仍是 PHP 兼容 JSON 壳，业务数据位于 `data.data`。
- 测试：聚焦 `go test ./internal/service/activity ./internal/server` 通过；PHP-Go 对比 `/activity/luckyprizes` 忽略动态 `xxx_api_auth` 后一致。

### `/activity`、`/activity/index`、`/activity/details`

- PHP: `c.api.activity->index/details`
- Go: `internal/handler.ActivityHandler.Index/Details`
- Service: `internal/service/activity.Service`
- Repository: `internal/repository/activity.Repository`
- Auth: 公共接口，不要求登录；旧 PHP 游客响应可能包含动态 `data.xxx_api_auth`，Go 不生成该字段。
- DB: `index` 读取 `activity WHERE reward_expire_time > now ORDER BY id DESC LIMIT 1`；`details` 读取 `activity` 单条记录和 `activity_prizes WHERE aid=? ORDER BY ranking ASC`。
- 兼容规则：`/activity` 等价 `/activity/index`；无进行中活动返回 `retcode=-9999`、`errmsg=当前没有进行中的活动`；无效 `aid` 返回 `retcode=-9999`、`errmsg=获取活动信息失败`；详情奖项复刻 PHP `activityprizes.procRow` 的 `ranking` 区间和 `prize_users` 字符串。
- 测试：聚焦 `go test ./internal/service/activity ./internal/server` 通过；PHP-Go 对比 `/activity`、`/activity/index`、`/activity/details?aid=0` 忽略动态 `xxx_api_auth` 后一致。

### `/activity/newyear2020`、`/activity/luckydraw`

- PHP: `c.api.activity->newyear2020/luckydraw`
- Go: `internal/handler.ActivityHandler.NewYear2020/LuckyDraw`
- Service: `internal/service/activity.Service`
- Auth: 公共接口；当前日期下旧 PHP 在登录检查和抽奖写库前先判断活动已过期，Go 复刻该稳定分支。
- 兼容规则：`newyear2020` 过期截止为 `2020-02-08 23:59:59`；`luckydraw` 过期截止为 `2023-02-28 23:59:59`；当前日期为 2026 年，两个接口均返回 `retcode=-1`、`errmsg=抽奖活动已结束，谢谢支持`，不接管历史活动内的抽奖写入分支。
- 测试：聚焦 `go test ./internal/service/activity ./internal/server` 通过；PHP-Go 对比 `/activity/newyear2020` 和 `/activity/luckydraw` 忽略旧 PHP 动态 `data.xxx_api_auth` 后一致。

### `/activity/luckydrawhistory`

- PHP: `c.api.activity->luckydrawhistory`
- Go: `internal/handler.ActivityHandler.LuckyDrawHistory`
- Service: `internal/service/activity.Service`
- Repository: `internal/repository/activity.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=请登录后操作`。
- DB: 读取 `activity_prizelogs WHERE uid=? ORDER BY createtime DESC LIMIT 20 OFFSET ?`；不返回分页信息，保持旧 PHP 只返回 `data.data`。
- 兼容规则：按旧 PHP 固定映射给 `keyid` 追加 `prizename`，未知 key 返回空字符串。
- 测试：聚焦 `go test ./internal/service/activity ./internal/server` 通过；PHP-Go 对比 `/activity/luckydrawhistory?page=1` 未登录和登录空历史分支一致，忽略旧 PHP 动态 `data.xxx_api_auth`。

### `/activity/ranking`、`/activity/receive`

- PHP: `c.api.activity->ranking/receive`
- Go: `internal/handler.ActivityHandler.Ranking/Receive`
- Service: `internal/service/activity.Service`
- Repository: `internal/repository/activity.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: `ranking` 读取活动、奖项和 `activity_records LEFT JOIN users`，按 `score DESC` 返回前 20 条；机器人用户 `uid<0` 时按 `bot_users.uid=abs(uid)` 替换用户名/头像。`receive` 读取当前用户在活动中的排名，按奖项区间计算 `prize_level/prize_money`，不更新领取状态。
- 兼容规则：无效活动返回 `retcode=-9999`、`errmsg=获取活动信息失败` 且不带 `data`；`receive` 超过领奖截止返回 `超过该活动领奖截止日期`；`ranking` 的奖项映射按 PHP `activityrecords.procRow2` 第一条匹配后 `break`，`receive` 保留 PHP 循环不 `break` 的行为。
- 测试：聚焦 `go test ./internal/service/activity ./internal/server` 通过；PHP-Go 对比 `/activity/ranking?aid=0`、`/activity/receive?aid=0` 的未登录和登录无效活动分支一致，成功分支由 service fake 覆盖。

### `/activity/recommends`

- PHP: `c.api.activity->recommends`
- Go: `internal/handler.ActivityHandler.Recommends`
- Service: `internal/service/activity.Service`
- Repository: `internal/repository/activity.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取活动的 `effect_time/expire_time` 后，统计并读取 `users` 中由当前用户推荐且注册时间落入活动窗口的用户；同时读取 `user_groups` 生成 `gicon`。
- 兼容规则：返回 `data.data` 用户行和 `data.total`；用户行复刻 PHP `user.procRow2` 的核心字段、`uniqkey` 大写 base36、`avatar_url`、`isvip/gicon`、金币/金豆整型。
- 测试：聚焦 `go test ./internal/service/activity ./internal/server` 通过；PHP-Go 对比 `/activity/recommends?aid=0&page=1` 的未登录和登录无效活动分支一致，成功分支由 service fake 覆盖。

### `/invite/info`

- PHP: `c.api.invite->info`
- Go: `internal/handler.InviteHandler.Info`
- Service: `internal/service/invite.Service`
- Repository: `internal/repository/invite.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `user_recommend AS r LEFT JOIN users AS u ON r.recommend_uid=u.uid WHERE r.uid=?`。
- 兼容规则：未绑定返回 `data.data=null`；已绑定返回推荐人 `uniqkey` 的 base36 小写字符串。
- 测试：聚焦 `go test ./internal/service/invite ./internal/server` 通过；PHP-Go 对比 `/invite/info` 未登录分支和登录真实 token 分支一致，忽略旧 PHP 动态 `data.xxx_api_auth`。

### `/bought/listing`

- PHP: `c.api.bought->listing`
- Go: `internal/handler.BoughtHandler.Listing`
- Service: `internal/service/bought.Service`
- Repository: `internal/repository/bought.Repository` + `internal/repository/user.Repository`，并复用 `internal/service/vod.ListingService.ProcessRows`。
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=请登录后操作`。
- DB: 先统计 `user_bought WHERE uid=?`，再按旧 PHP 查询 `user_bought AS a LEFT JOIN vods AS b ON b.vodid=a.vodid WHERE a.uid=? AND b.showtype=0 ORDER BY a.buytime DESC`，`pagesize=20`。
- 兼容规则：返回 `data.rows` 和 `data.pageinfo`；视频行复用普通 VOD `procRow2` 兼容字段，分页 URL 为 `/bought/listing?page=[?]`。
- 测试：聚焦 `go test ./internal/service/bought ./internal/service/vod ./internal/server` 通过；PHP-Go 对比 `/bought/listing?page=1` 未登录分支忽略旧 PHP 动态 `data.xxx_api_auth` 后一致，登录测试 token 返回空 `rows` 和分页结构一致。

### `/bought/delete`

- PHP: `c.api.bought->delete`
- Go: `internal/handler.BoughtHandler.Delete`
- Service: `internal/service/bought.Service`
- Repository: `internal/repository/bought.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=请登录后操作`。
- DB: 对 `vodids` 逗号列表逐个执行 `DELETE FROM user_bought WHERE uid=? AND vodid=?`。
- 兼容规则：`vodids` 为空时与旧 PHP 一样直接成功；成功响应 `retcode=0`、`errmsg=""`，不带 `data`。
- 测试：聚焦 `go test ./internal/service/bought ./internal/server` 通过；PHP-Go 对比 `/bought/delete` 未登录和登录空 `vodids` 分支一致，忽略旧 PHP 动态 `data.xxx_api_auth`。

### `/explore/notification`、`/explore/signtask`、`/explore/vodtask` 空入口

- PHP: `c.api.explore.notification->index`、`c.api.explore.signtask->index`、`c.api.explore.vodtask->index`
- Go: `internal/handler.ExploreHandler.EmptyOK`
- Service: `internal/service/explore.Service`
- Auth: 公共接口，不要求登录；旧 PHP 会带动态游客 token，Go 不生成该字段。
- 兼容规则：仅接管默认 `index` 空入口，包括 `/explore/notification`、`/explore/notification/index`、`/explore/signtask`、`/explore/signtask/index`、`/explore/vodtask`、`/explore/vodtask/index`；响应为 `retcode=0`、`errmsg=""`，不带业务 `data`。`signtask/sign`、`vodtask/reqcoin` 仍未接管。
- 测试：聚焦 `go test ./internal/server` 通过；PHP-Go 对比 `/explore/notification`、`/explore/signtask`、`/explore/vodtask` 忽略动态 `xxx_api_auth` 后一致。

### `/explore/vodtask/show/:vid`

- PHP: `c.api.explore.vodtask->show`
- Go: `internal/handler.ExploreHandler.VodTaskShow`
- Service: `internal/service/explore.Service`
- Repository: `internal/repository/explore.Repository`
- DB: 读取 `explore_vods`；按登录用户或游客读取/创建 `explore_vodlogs`、`explore_guestvodlogs` 当日日志。
- 兼容规则：视频不存在或 `showtype!=0` 返回 `记录不存在或已被删除`；无当日日志时按 `mincoin/maxcoin` 随机 `reqcoin` 并创建日志；已有日志复用 `logid/reqcoin/reqtime`；返回 `vodrow` 字段对齐 `vodmgr->procRow2`。
- 风险说明：本接口只创建领取日志，不发放金币；剩余 `reqcoin` 涉及金币/游客金币更新和事务锁，仍留在未重构高风险清单。
- 测试：`go test ./internal/service/explore ./internal/repository/explore ./internal/server` 通过；PHP-Go live 对比 `/explore/vodtask/show/0` 错误分支一致；成功和日志复用分支由 service fake 覆盖。

### `/explore/index`

- PHP: `c.api.explore.index->index`
- Go: `internal/handler.ExploreHandler.Index`
- Service: `internal/service/explore.Service`
- Repository: `internal/repository/explore.Repository` + `internal/repository/user.Repository`
- Auth: 公共接口；无 token 按游客组权限计算，登录时兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie 并按用户组权限计算。
- DB: 读取 `explore_tabs WHERE showtype=0 ORDER BY sortnum ASC LIMIT 100` 和 `user_groups`；登录分支读取 `sessions/users`。
- 兼容规则：返回 `data.tabrows/dayrows/signdata`；`tabrows` 复刻 `tabmgr.procRow2` 字段，`extjson` 空值为 `null`；`dayrows` 从中国时区今日开始连续 7 天，金币数按 `max.signtask.coinnumN` 权限；`signdata` 返回今日是否已签到、历史最高连续、当前连续和本循环完成天数。
- 测试：聚焦 `go test ./internal/service/explore ./internal/server` 通过；PHP-Go 对比 `/explore/index` 游客分支忽略旧 PHP 动态 `data.xxx_api_auth` 后一致，登录测试 token 的 7 日奖励和签到状态一致。

### `/explore/notification/clean`

- PHP: `c.api.explore.notification->clean`
- Go: `internal/handler.ExploreHandler.CleanNotification`
- Service: `internal/service/explore.Service`
- Repository: `internal/repository/explore.Repository` + `internal/repository/user.Repository`
- Auth: 公共接口；无 token 按游客处理，登录时兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie。
- DB: `tabkey=all` 时把 `notification_all` 写为 JSON `null`；指定 tab 时要求该键已存在，再置为 `0`；登录用户更新 `users.notification_all`，游客更新 `user_guests.notification_all`。
- 兼容规则：`tabkey` 为空返回 `retcode=-1`、`errmsg=请提供频道键名`；指定不存在 tab 返回 `指定的频道键名不存在`；成功返回 `data.notification_all`。
- 测试：聚焦 `go test ./internal/service/explore ./internal/server` 通过；PHP-Go 对比空 `tabkey` 和不存在 tab 错误分支，忽略旧 PHP 动态 `data.xxx_api_auth` 后一致；成功写入分支由 service fake 覆盖，未直接改动共享数据。

### `/hgame/index`

- PHP: `c.api.hgame->index`
- Go: `internal/handler.HGameHandler.Index`
- Service: `internal/service/hgame.Service`
- Repository: `internal/repository/hgame.Repository`
- Auth: 公共接口，不要求登录；旧 PHP 会在响应中追加动态游客 `data.xxx_api_auth`，Go 不生成该字段。
- DB: 先统计 `hgame` 总数，若为 0 返回 `暂未开放`；列表读取 `status=0 AND show_type!=1 ORDER BY sort ASC`，幻灯片读取 `status=0 AND show_type!=0 ORDER BY sort ASC`，`pagesize=20`。
- 兼容规则：返回 `data.data.list` 和 `data.data.slide`；行字段复刻 `hgame.procRow2`，`remark` 能 JSON 解码时返回数组/对象，否则保留原字符串；资源字段复用 `RESOURCE_BASE_URL` 拼接。
- 测试：聚焦 `go test ./internal/service/hgame ./internal/server` 通过；PHP-Go 对比 `/hgame/index` 成功分支，忽略旧 PHP 动态 `data.xxx_api_auth` 后核心字段一致；旧 PHP `/hgame` 为 404，Go 未注册该路径。PHP `c.api.hgame` 仅定义 `index`，其他动态 action 不伪造业务响应。

### `/ucp/task/sharepic`

- PHP: `c.api.ucp.task->sharepic`
- Go: `internal/handler.UCPHandler.TaskSharePic`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository`
- Auth: 公共接口，不要求登录；旧 PHP 会在游客响应中追加动态 `data.xxx_api_auth`，Go 不生成该字段。
- DB: 读取 `poster WHERE status=1`；无记录返回 `data.data=[]`，有记录时随机返回一条 poster 原始行。
- 兼容规则：保持旧 PHP `data.data` 包装；随机行不做逐条完全一致，只要求字段 shape 和 `retcode/errmsg` 对齐。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/task/sharepic` 成功分支，忽略动态 `xxx_api_auth` 和随机行差异后一致。

### `/ucp/task/qrlink`

- PHP: `c.api.ucp.task->qrlink`
- Go: `internal/handler.UCPHandler.TaskQRLink`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/index.SettingsRepository` 通过 `ucpStore` 只读注入。
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `maintain_calldata.uuid=(pid.)global.qrcode.link` 和 `settings.uuid=baseset`；渠道配置缺失时回退默认 `global.qrcode.link`。
- 兼容规则：`pid` 只允许英文数字；`inviteUrls` 按旧 PHP 的 `-` 分组并按当月日期选择，替换 `{inviteUrl}` 和 `{inviteCode}`，邀请码为用户 `uniqkey` 的大写 base36。
- 测试：`go test ./internal/service/ucp ./internal/server` 通过；PHP-Go live 对比 `/ucp/task/qrlink` 未登录分支 retcode/errmsg 一致，旧 PHP 额外返回动态游客 token，新 Go 按既有策略返回空 `data`。

### `/ucp/taskbox/index`

- PHP: `c.api.ucp.taskbox->index`
- Go: `internal/handler.UCPHandler.TaskboxIndex`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository`
- Auth: 公共接口，不强制登录；登录用户会按 uid 查询本人任务开启记录，游客任务状态按未完成计算。
- DB: 读取 `promotion_taskboxs ORDER BY taskid ASC`、每个任务对应的 `promotion_taskboxlogs`、最近 30 条 `taskstatus=2` 的宝箱日志并关联用户。
- 兼容规则：返回 `data.taskrows` 和 `data.logrows`；任务状态复刻推广任务、每日 22:00 五分钟、每周六 22:00 五分钟规则；缺失用户的 `username/nickname` 保持 `null`。
- 测试：聚焦 `go test ./internal/service/ucp ./internal/server` 通过；PHP-Go 对比 `/ucp/taskbox/index`，忽略动态 `xxx_api_auth` 后字段和状态一致；旧 PHP `/ucp/taskbox` 为空响应，Go 未注册该路径。

### `/ucp/taskbox/qrlink`

- PHP: `c.api.ucp.taskbox->qrlink`
- Go: `internal/handler.UCPHandler.TaskboxQRLink`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/index.SettingsRepository` 通过 `ucpStore` 只读注入。
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `maintain_calldata.uuid=(pid.)taskbox.qrcode.link` 和 `settings.uuid=baseset`；渠道配置缺失时回退默认 `taskbox.qrcode.link`。
- 兼容规则：与 `/ucp/task/qrlink` 相同的 pid 校验、每日 inviteUrls 分组和 `{inviteCode}` 替换；不生成 PNG，也不写 keylimit 或发奖励。
- 测试：`go test ./internal/service/ucp ./internal/server` 通过；PHP-Go live 对比 `/ucp/taskbox/qrlink` 未登录分支 retcode/errmsg 一致，旧 PHP 额外返回动态游客 token，新 Go 按既有策略返回空 `data`。

### `/community/{list,recommend,hot,latest,favorite}`、`/community/*-:params`

- PHP: `c.api.topic->list`
- Go: `internal/handler.CommunityHandler.Listing`
- Service: `internal/service/community.Service`
- Repository: `internal/repository/community.Repository`
- Auth: 公共列表接口不要求登录；`favorite` 未登录返回 `retcode=-9999`、`errmsg=请登录后操作`，登录后按 `topic_favorites` 过滤。
- Params: `$category_id:0-$type:0-$orderby:0-$page:1`；path page 为 `0` 时使用 query `page`。
- DB: 读取 `topics`、`topic_images`、`topic_videos`、`vod_servers`；登录时补充 `topic_favorites/topic_ups` 标志。
- 兼容规则：返回 `now/action/sample_params/params/rows/pageinfo`；支持 `recommend`、`hot`、`latest` 排序和 `category_id/type/orderby` 筛选；行内补 `images/content_images/videos/is_favorite/is_up`。
- 测试：`go test ./internal/service/community ./internal/server ./...` 通过；PHP-Go live 对比 `/community/list` 主结构、首行 key/媒体字段、`/community/favorite` 未登录分支通过。

### `/community/clisting`、`/community/clisting-:params`

- PHP: `c.api.topic->clisting`
- Go: `internal/handler.CommunityHandler.CommentListing`
- Service: `internal/service/community.Service`
- Repository: `internal/repository/community.Repository`
- Auth: 公共评论列表接口；登录时补充评论 `is_up`。
- Params: `$orderby:0-$page:1`；query `orderby/page` 会覆盖默认值。
- DB: 读取 `topics` 存在性、`topic_comments LEFT JOIN users` 根评论和子评论、`topic_comments_ups`。
- 兼容规则：`tid` 不存在返回 `记录不存在或已被删除`；成功返回 `now/sample_params/params/rows/pageinfo`，评论行包含 `subrows`、相对时间和头像 URL。
- 测试：`go test ./internal/service/community ./internal/server ./...` 通过；PHP-Go live 对比 `/community/clisting?tid=0` 错误分支和 `/community/clisting?tid=<列表首条 tid>` 空评论分页通过；修正默认 `params.orderby` 为数字 `0` 以匹配 PHP。

### `/getGlobalData`

- PHP: `c.api.index->getGlobalData`
- Go: `internal/handler.IndexHandler.GetGlobalData`
- Service: `internal/service/index.GlobalService`
- Repository: `internal/repository/index.SettingsRepository`
- Auth: 公共配置接口；旧 PHP 游客响应会追加动态 `xxx_api_auth`，Go 不生成该字段。
- DB: 读取 `maintain_calldata` 中 `global.appver/search.hotwords/global.hottags/global.hotcategories/global.ads/global.popup*/*adgroup*` 等配置，读取 `settings` 中 `setting/baseset/user.regopt`。
- 兼容规则：返回版本、热词、热门标签/分类、下载链接、弹窗、广告组、启动广告、推广文案、游戏/直播/AI/验证码等开关；`ver` query 会覆盖 `appver.AndroidVer/iOSVer`；资源路径用 `RESOURCE_BASE_URL` 拼接。
- 测试：`go test ./internal/service/index ./internal/server ./...` 通过；PHP-Go live 对比 `/getGlobalData` 的核心 key shape、`webreg/appver/hottags/hotcategories/appdownurl/gameDisabled/smscaptcha` 和 `/getGlobalData?ver=1.2.3` 版本覆盖通过，忽略旧 PHP 动态 `xxx_api_auth` 和随机广告/推广 URL 差异。

### `/aiundress`、`/aiundress/listing`

- PHP: `c.api.aiundress->listing`
- Go: `internal/handler.AIUndressHandler.Listing`
- Service: `internal/service/aiundress.Service`
- Repository: `internal/repository/aiundress.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录按旧 PHP 返回 `retcode=-1`、`errmsg=请先登录`。
- DB: 读取 `ai_undress WHERE uid=? AND status<=4 AND status>0`，`module` 非 0 时追加筛选，`ORDER BY id DESC`，`pagesize=10`；同时读取 `settings.uuid=setting` 中的 `resurl_h5_ai`。
- 兼容规则：返回 `data.rows/pageinfo`；`image/output` 空值保持空，已有 URL 原样返回，相对路径在 `APP_ENV=test` 下按旧 PHP 拼接 R2 域名，其他环境优先使用 `resurl_h5_ai` 并替换 `{rand}` 为中国时区 `yyyyMMddHH`。
- 测试：`go test ./internal/service/aiundress ./internal/server` 通过；PHP-Go live 对比 `/aiundress`、`/aiundress/listing` 未登录分支一致，登录测试 token 下 `/aiundress/listing?page=1&module=4` 的 `retcode/total/rows/page_url/首行字段/image` 一致。

### `/ucp/taskbox/taskboxlog`

- PHP: `c.api.ucp.taskbox->taskboxlog`
- Go: `internal/handler.UCPHandler.TaskboxLog`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 统计并读取 `promotion_taskboxlogs AS a LEFT JOIN users AS b ON b.uid=a.uid WHERE a.uid=? ORDER BY a.addtime DESC`，`pagesize=20`。
- 兼容规则：返回 `data.logrows/pageinfo`；日志行复用任务宝箱首页的 `procRow2` 兼容字段，分页 URL 为 `/ucp/taskbox/taskboxlog?page=[?]`。
- 测试：`go test ./internal/service/ucp ./internal/server` 通过；PHP-Go live 对比 `/ucp/taskbox/taskboxlog?page=1` 未登录分支一致，登录测试 token 下 `rows/total/page_url/首行内容` 一致。

### `/onego/history`

- PHP: `c.api.onego->history`
- Go: `internal/handler.OneGoHandler.History`
- Service: `internal/service/onego.Service`
- Repository: `internal/repository/onego.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取当前用户 `one_go_orders`，按 PHP 的 `GROUP BY period, room_id ORDER BY id DESC` 获取代表订单；再读取同一期同房间的用户投注号和 `one_go_records` 开奖信息。
- 兼容规则：返回 `data.data`；代表订单的 `id/uid/room_id/bet_coins` 转 int，补 `bet_no/win_no/win_coins`，不额外返回分页字段。
- 测试：`go test ./internal/service/onego ./internal/server` 通过；PHP-Go live 对比 `/onego/history?page=1` 未登录分支和登录测试 token 空历史一致。

### `/onego/bet_ranks`

- PHP: `c.api.onego->bet_ranks`
- Go: `internal/handler.OneGoHandler.BetRanks`
- Service: `internal/service/onego.Service`
- Repository: `internal/repository/onego.Repository`
- Auth: 公共只读排行接口，不要求登录。
- DB: 校验 `one_go_rooms.id` 和 `one_go_records(period, room_id)`，读取 `one_go_orders` 按 uid 汇总 `SUM(bet_coins)` 与 `COUNT(*)`。
- 兼容规则：无效房间返回 `无效场次`，无效期号返回 `无效的活动期号`；成功行转 int，补 `room_name/user`，`total_bets` 按房间金币单价折算。
- 测试：`go test ./internal/service/onego ./internal/server` 通过；PHP-Go live 对比无效房间、有效房间无效期号分支一致，当前本地库无投注订单样本，成功分支由 service fake 覆盖。

### `/ucp/user`、`/ucp/user/index`

- PHP: `c.api.ucp.user->index`
- Go: `internal/handler.UCPHandler.UserIndex`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 通过 session 获取 uid 后重新读取 `users` 当前行，并读取 `user_groups` 计算 `gicon/isvip`。
- 兼容规则：返回 `data.user`；复用 PHP `user.procRow2` 字段，包括 `uniqkey` 大写 base36、头像 URL、VIP 到期字段、金币/金豆缺失时为 0。
- 测试：`go test ./internal/service/ucp ./internal/server` 通过；PHP-Go live 对比 `/ucp/user` 未登录分支一致，登录测试 token 下 user 字段集合和核心值一致。

### `/ucp/bankcard`、`/ucp/bankcard/index`

- PHP: `c.api.ucp.bankcard->index`
- Go: `internal/handler.UCPHandler.BankcardIndex`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `user_bankcards WHERE uid=? ORDER BY isdef DESC` 和 `banks WHERE showtype=0 ORDER BY sortnum ASC, bankid ASC`。
- 兼容规则：返回 `data.cardrows/maxallow/allowtype/banknames/bankRows`；`maxallow=3`，`allowtype=7`，`bankRows` 只保留 `bankid/bankname/coverpic` 并按 PHP 转换 `bankid` 为 int。
- 测试：`go test ./internal/service/ucp ./internal/server` 通过；PHP-Go live 对比 `/ucp/bankcard` 未登录分支一致，登录测试 token 下 cardrows、bankRows、限制字段和首行内容一致。

### `/ucp/bankcard/create`、`/ucp/bankcard/modify`、`/ucp/bankcard/delete`

- PHP: `c.api.ucp.bankcard->create/modify/delete`
- Go: `internal/handler.UCPHandler.BankcardCreate/BankcardModify/BankcardDelete`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository` + `internal/repository/user.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: `create/modify/delete` 写入 `user_bankcards`；设置默认地址时先清空当前用户 `isdef`，再将目标 `cardid` 置为默认。
- 兼容规则：`type=0/1` 时强制 `bankname=支付宝`，`type=3` 时强制 `bankname=微信`；保留 PHP 的姓名、开户银行、账号长度校验文案；新增时沿用 PHP `count >= 5` 但错误文案为 `最多可以设置3个地址`。
- 测试：`go test ./internal/service/ucp ./internal/repository/ucp ./internal/server` 通过；覆盖未登录、创建校验、支付宝类型映射、默认地址、修改缺失记录和删除成功分支。写入成功 live 分支未直接打旧库，避免污染本地导入数据。

### `/ucp/vippkg`、`/ucp/vippkg/index`、`/ucp/coinpkg`、`/ucp/coinpkg/index`、`/ucp/beanpkg`、`/ucp/beanpkg/index`

- PHP: `c.api.ucp.vippkg->index`、`c.api.ucp.coinpkg->index`、`c.api.ucp.beanpkg->index`
- Go: `internal/handler.UCPHandler.VIPPkgIndex/CoinPkgIndex/BeanPkgIndex`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository` + `internal/repository/index.SettingsRepository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: 读取 `trade_vippkgs/trade_coinpkgs/trade_beanpkgs WHERE showtype=0 ORDER BY sortnum ASC, pkgid ASC`；读取 `settings.uuid=setting` 取得 `safepayurl`。
- 兼容规则：套餐字段按 PHP `procRow2` 输出，金额按分转元字符串两位小数；支付通道通过 `PaymentChannels` 接口隔离，当前 repository 默认返回空列表，不伪造旧 PHP `conf/payment.php` 和后台支付通道配置。
- 测试：`go test ./internal/service/ucp ./internal/repository/ucp ./internal/server` 通过；PHP-Go live 对比三类套餐未登录分支 retcode/errmsg 一致，旧 PHP 额外返回动态游客 token，新 Go 按既有策略返回空 `data`。

### `/ucp/vodorder/myorders`、`/ucp/vodorder/mysupports`、`/ucp/vodorder/historyorders`

- PHP: `c.api.ucp.vodorder->myorders/mysupports/historyorders`
- Go: `internal/handler.UCPHandler.VODOrderMyOrders/VODOrderMySupports/VODOrderHistoryOrders`
- Service: `internal/service/ucp.Service`
- Repository: `internal/repository/ucp.Repository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；未登录返回 `retcode=-9999`、`errmsg=您还没有登录`。
- DB: `myorders` 读取 `user_vod_order` 本人记录、成功消耗金币、冻结金币和本人助力金币汇总；`mysupports` 读取 `user_vod_support` 按 `void` 汇总后关联 `user_vod_order`；`historyorders` 读取 `user_vod_order.status=1`。
- 兼容规则：保留 PHP 的 `data`、`total_cost`、`current_frozen` 字段；分页 pageSize 为 20；只读 action 不扣金币、不创建求片、不助力。
- 测试：`go test ./internal/service/ucp ./internal/repository/ucp ./internal/server` 通过；PHP-Go live 对比三条接口未登录分支 retcode/errmsg 一致，旧 PHP 额外返回动态游客 token，新 Go 按既有策略返回空 `data`。

### `/vod/breaking`

- PHP: `c.api.vod->breaking`
- Go: `internal/handler.VODHandler.Breaking`
- Service: `internal/service/vod.ListingService`
- Repository: `internal/repository/vod.ListingRepository`
- Auth: 公共只读接口，不要求登录；旧 PHP 游客响应会带动态 `xxx_api_auth`，Go 不生成该字段。
- DB: 读取 `vods WHERE cateid=99 AND utimestamp>=今日零点 LIMIT 1`。
- 兼容规则：成功返回 `data.vodid/title` 且 `errmsg=ok`；无记录或 `showtype>0` 返回 `记录不存在或已被删除`。
- 测试：`go test ./internal/service/vod ./internal/repository/vod ./internal/server` 通过；PHP-Go live 对比当前本地库无记录错误分支 retcode/errmsg 一致，忽略旧 PHP 动态游客 token。

### `/payment/unpaid`

- PHP: `c.api.payment->unpaid`
- Go: `internal/handler.PaymentHandler.Unpaid`
- Service: `internal/service/payment.Service`
- Auth: 当前 PHP 函数在登录检查前直接返回 `total_count=0`，因此 Go 也不要求登录。
- DB/Cache: 无；旧 PHP 后续的 24 小时未支付订单查询和 Redis 标记分支不可达，本次不接管。
- 兼容规则：返回 `retcode=0`、`errmsg=""`、`data.total_count=0`；旧 PHP 游客中间件可能追加动态 `xxx_api_auth`，Go 不生成。
- 测试：`go test ./internal/service/payment ./internal/server` 通过；路由测试覆盖 JSON 壳和 `total_count=0`。

### `/payment/success`、`/payment/failed`

- PHP: `c.api.payment->success/failed`
- Go: `internal/handler.PaymentHandler.Success/Failed`
- Service: `internal/service/payment.Service`
- Auth/DB/External: 无；两个接口只返回固定 JSON 文案，不读取订单、不验签、不调用支付平台。
- 兼容规则：`success` 返回 `retcode=0`、`errmsg=支付成功回调`、无 `data`；`failed` 返回 `retcode=-1`、`errmsg=支付失败回调`、无 `data`。
- 测试：`go test ./internal/service/payment ./internal/server` 通过；PHP-Go live 对比两个接口 retcode/errmsg 一致，忽略旧 PHP 游客中间件动态字段。

### `/comment`、`/comment/index`

- PHP: `c.api.comment->index`
- Go: `handler.EmptyHTML`
- Auth/DB/External: 无；旧 PHP 方法体为空。
- 兼容规则：返回 `200 text/html` 空 body；评论发布 `/comment/post` 仍未接管。
- 测试：`go test ./internal/server` 通过；路由测试覆盖空 body。

### `/init`

- PHP: `c.api.index->init`
- Go: `internal/handler.IndexHandler.Init`
- Service: `internal/service/index.InitService`
- Repository: `internal/server.indexStore` 组合 `user.Repository`、`ucp.Repository`、`index.SettingsRepository`
- Auth: 兼容 `x-cookie-auth` header 和 `xxx_api_auth` cookie；无 token 返回游客 user 壳。
- DB: 读取 session/user、user_groups、users_quota、users_goldbean、`settings(setting/baseset/promotion.bonus)` 和 `maintain_calldata(global.appver/playHeaders/全局配置)`。
- 兼容规则：返回 `globalData/invite_bonus/user/appver/notification_all/inviteCodeUrl/inviteCodeAppid/playHeaders/urlHosts/csurl/sitelogo/isclosed/closetips/externalUrlDating`；顶层 `appver` 不受 `ver` query 覆盖，`globalData.appver` 保持 `/getGlobalData` 的覆盖规则。
- 测试：`go test ./internal/service/index ./internal/server` 通过；PHP-Go live 对比 `/init?ver=1.2.3` 游客和测试 token 登录分支，核心 key、user、appver/globalData.appver、通知空值和站点配置一致，忽略旧 PHP 动态 `xxx_api_auth`。

### `/`、`/index`

- PHP: `c.api.index->index`
- Go: `internal/handler.IndexHandler.Index`
- Service: `internal/service/index.HomeService`
- Repository: `internal/repository/vod.ListingRepository` 复用 `maintain_calldata` 和 `vods` 查询。
- Auth: 公共首页接口；当前 Go 不写旧 PHP 的 `appversion` 访问记录副作用。
- DB: 读取 `maintain_calldata(index.slide/index.slide.v2/index.slide.pc/index.slide.mb/index.recommend.vods/index.tagvods)`，并读取 `vods` 生成 `dayrows/latestrows/likerows/a_vodrows/b_vodrows/c_vodrows/d_vodrows/tagvodrows/hotrows`。
- 兼容规则：返回 `sliderows/v2sliderows/pcsliderows/mbsliderows/dayrows/latestrows/likerows/a_vodrows/b_vodrows/c_vodrows/d_vodrows/tagvodrows/hotrows`；资源路径按 `RESOURCE_BASE_URL` 拼接，视频行复用 VOD `ProcessRows`。
- 测试：`go test ./internal/service/index ./internal/repository/vod ./internal/server` 通过；PHP-Go live 对比 `/index` 和 `/` 的 `retcode/errmsg/key 集合/主要 count` 一致，忽略旧 PHP 动态游客 token 和首页访问版本写入。

### `/getCover`

- PHP: `c.api.index->getCover`
- Go: `internal/handler.IndexHandler.GetCover`
- Service: `internal/service/index.CoverService`
- Cache/Client: 默认内存 TTL cache，外部封面服务通过 `CoverFetcher` 接口注入；后续接 Redis 时替换 `CoverCache`。
- DB: 读取 `settings.uuid=setting` 的 `getCoverUrl`，缺省为旧 PHP 的 `http://172.22.0.7:8026/coverpic`。
- 兼容规则：缓存命中直接返回 `data.data`；缓存未命中请求 `getCoverUrl?pic=...`，用 `substr(pic,10,32)` 等价 key 做 AES-256-CBC/PKCS7 加密，再 base64，返回 `Cache-Control: max-age=86400`；非法/缺失/外部空响应返回 `记录不存在或已被删除`。
- 测试：`go test ./internal/service/index ./internal/server` 通过；成功加密和缓存分支由 fake 覆盖；live 验证 `/getCover?pic=short` Go 返回旧错误壳且不阻塞，旧 PHP 在缺省内网封面服务不可达时 3 秒超时无响应。

### `/sms/sendv`、`/sms/sendu`、`/email/send`

- PHP: `c.api.sms->sendv/sendu`、`c.api.email->send`
- Go: `internal/handler.VerificationHandler`
- Service: `internal/service/verification.Service`
- Repository: `internal/server.indexStore` 读取 `settings.uuid=setting` 和登录用户 session。
- External: `CaptchaVerifier`、`SMSSender`、`MailSender`、`Limiter` 均为接口；默认 sender 不直连真实短信/邮件平台，生产接入时替换具体 client。
- 兼容规则：保留手机号/邮箱格式错误、sendu 未登录、缺图形验证码、验证码失败、频控、平台未配置等 legacy errmsg；成功响应使用 PHP `Json::ok($errmsg)` 形态，`retcode=0` 且消息在 `errmsg`。
- 测试：`go test ./internal/service/verification ./internal/server` 通过；PHP-Go live 对比 `/sms/sendv?mobi=bad`、`/email/send?email=bad`、`/sms/sendu` 未登录错误分支一致；成功发送分支由 fake sender/captcha/limiter 覆盖。
