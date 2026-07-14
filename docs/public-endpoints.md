# 公共接口迁移清单

旧 PHP host: `http://127.0.0.1:18765`

说明：

- “公共”指不需要已登录用户。
- 旧 `c.api.__init__` 中间件仍可能创建游客、写 cookie/auth 字段、读取 Redis/缓存/系统设置。
- Go 迁移时按一个接口一个接口推进；每个接口完成后记录状态，作为后续上下文压缩点。
- 本清单已和 `internal/server/router.go` 中的公共真实 handler 路由对齐；Go 新增公共路由后同步更新此文件和根目录 `MIGRATION_ENDPOINTS.md`。

## 已迁移

| 接口 | PHP handler | Go 状态 | 对比说明 |
| --- | --- | --- | --- |
| `/sysavatar` | `c.api.user->sysavatar` | 已完成 | 忽略旧中间件动态字段 `data.xxx_api_auth` 后一致。 |
| `/captcha/req` | `c.api.captcha->req` | 本轮完成 | `picurl` secret 动态生成，对比稳定结构、前缀和 `smscaptcha`。 |
| `/iploc/:ip` | `c.api.index->iploc` | 本轮完成 | 忽略旧中间件动态字段后，稳定 IP 样例一致。 |
| `/game/platforms` | `c.api.game.index->index` | 本轮完成 | 读 `game_platform`，保留 PHP 字段和值类型并剔除 `json`。 |
| `/game/categories` | `c.api.game.index->categories` | 本轮完成 | 读 `game_category`，保留 PHP 字段和值类型并拼接资源 URL。 |
| `/v2/amazing/categories` | `c.apiv2.amazing->categories` | 本轮完成 | 读 `amazing_category` 固定列，支持 `parent_id`。 |
| `/v2/so/list` | `c.apiv2.so->index` | 本轮完成 | 读 `server_so_config.value`，按 PHP `json_decode` 行为返回 `data.data`。 |
| `/v2/vod/listing`、`/v2/vod/recommend`、`/v2/vod/hot`、`/v2/vod/latest` | `c.apiv2.vod->listing` | 本轮完成 | 动态 action 路由组，支持 `-params`，迁移列表筛选、排序、分页和核心 `vodrows` 字段。 |
| `/v2/amazing/listing`、`/v2/amazing/recommend`、`/v2/amazing/hot`、`/v2/amazing/latest` | `c.apiv2.amazing->listing` | 本轮完成 | 动态 action 路由组，支持 `-params`，迁移精彩推荐列表筛选、排序和分页。 |
| `/vod/listing`、`/vod/recommend`、`/vod/hot`、`/vod/latest` | `c.api.vod->listing` | 本轮完成 | 非 v2 动态 action 路由组，支持 `-params`；复用 VOD 列表服务并对齐 PHP 分页 selector。 |
| `/vod/show/:vodid` | `c.api.vod->show` | 本轮完成 | 视频详情只读接口，迁移主视频、父级分类、相似视频和猜你喜欢；随机列表按 shape 对比。 |
| `/vod/preView/:vodid/index.m3u8` | `c.api.vod->preView` | 本轮完成 | m3u8 试看输出；HTTP 拉取通过 fetcher 注入，测试用 fixture，不依赖真实 CDN。 |
| `/sendfile/play/:file`、`/sendfile/down/:file` | `c.api.sendfile->play/down` | 本轮完成 | 兼容旧 PHP 空壳行为：play 只做登录和 vodid 存在性检查，成功空 200；down 空 200。 |
| `/comment/listing-:params` | `c.api.comment->listing` | 本轮完成 | 评论列表公共只读接口，支持评论树、排序、分页和用户头像/VIP 标识。 |
| `/game/games` | `c.api.game.index->games` | 本轮完成 | 读 `game`，支持 `platform_id/category_id`，保留 PHP 字段和值类型并拼接资源 URL。 |
| `/game/broadcasts` | `c.api.game.index->broadcasts` | 本轮完成 | 读 `game_broadcast`，按 PHP 替换 `{user}` 和 `{amount}` 占位符。 |
| `/getLikeRows` | `c.api.index->getLikeRows` | 本轮完成 | 复用 VOD 行处理，按旧 PHP 固定返回 6 条随机猜你喜欢。 |
| `/game/wali/gameList` | `c.api.game.wali->games` | 本轮完成 | 瓦力平台游戏列表，普通分类只读对齐；`category_id=5` 游客返回旧 PHP 未登录错误。 |
| `/ucp/rolltitle` | `c.api.ucp.index->rolltitle` | 本轮完成 | 个人中心滚动消息公共只读接口，读 `roll_titles` 中 `status=1` 的最近 10 条。 |
| `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last` | `c.api.onego->rules/rooms/current/last` | 本轮完成 | 一元购公共只读接口，读 `one_go`、`one_go_rooms`、`one_go_records`；本地错误分支和房间列表业务数据对齐，忽略旧中间件动态 `data.xxx_api_auth`。 |
| `/onego/hash` | `c.api.onego->hash` | 本轮完成 | 一元购公共哈希计算接口，复刻 PHP `hash('sha256')` 和末尾数字期号截取规则。 |
| `/onego/lucky` | `c.api.onego->lucky` | 本轮完成 | 一元购幸运榜公共只读接口，按获奖金币总数排序并附带各房间获奖次数；保留 PHP 未分页排行 SQL。 |

## 优先候选

| 优先级 | 接口 | PHP handler | 风险 |
| --- | --- | --- | --- |
| 1 | `/special/listing`、`/special/detail/:spid` | `c.api.special->listing/detail` | 中；专题只读，但会聚合视频行和详情浏览数。 |
| 2 | `/onego/marquee` | `c.api.onego->marquee` | 低到中；公共只读，一元购跑马灯。 |
| 3 | `/init` | `c.api.index->init` | 中；依赖系统设置、游客初始化、较多字段。 |
| 4 | `/getGlobalData` | `c.api.index->getGlobalData` | 中；依赖系统设置/缓存。 |
| 5 | `/search` | `c.api.search->index` | 中；搜索结果会写入/更新搜索日志。 |

## 暂缓

| 接口 | 原因 |
| --- | --- |
| `/register`、`/login`、`/forgot` | 公共但涉及账号、短信、风控和写库。 |
| `/payment/*`、`/respond/*` | 支付相关，需要独立 reviewer/灰度/回滚策略。 |
| `/game/wali/topup`、`/game/wali/withdraw`、`/game/wali/balance`、`/game/wali/enter`、`/game/lottery/*` | 游戏资产、余额或外部平台调用，需要登录、事务、灰度和回滚策略。 |
| `/captcha/pic`、`/captcha/picx` | 图片二进制输出，需要字体和图片生成兼容。 |
| `/minivod/*` | 多数依赖游客/用户权限、播放记录、金币或数据库。 |
| `/vod/up`、`/vod/down`、`/vod/reqplay`、`/vod/reqdown`、`/vod/buy` 及对应 `/v2/vod/*` 高风险动作 | 涉及播放/下载请求、点赞踩、购买和用户/游客记录，需单独迁移。 |
