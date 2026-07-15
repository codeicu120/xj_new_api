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
| `/logout` | `c.api.user->logout` | 本轮完成 | 退出登录；无 token/非法 token 分支对比通过，成功删除 session 由 service fake 覆盖。 |
| `/sms`、`/sms/index`、`/email`、`/email/index` | `c.api.sms/email->index` | 本轮完成 | 短信/邮件默认空入口，返回 `200 text/html` 空 body。 |
| `/sms/sendv`、`/sms/sendu`、`/email/send` | `c.api.sms/email->send*` | 本轮完成 | 验证码发送入口；手机号/邮箱/未登录错误分支 live 对比通过，成功发送由 sender/captcha/limiter fake 覆盖，默认不直连真实平台。 |
| `/captcha/req` | `c.api.captcha->req` | 本轮完成 | `picurl` secret 动态生成，对比稳定结构、前缀和 `smscaptcha`。 |
| `/captcha/pic`、`/captcha/picx` | `c.api.captcha->pic/picx` | 本轮完成 | 图片验证码输出；无效 secret 404 JSON 对比一致，有效 PHP/Go secret 均返回 `image/png`、100x34 PNG。 |
| `/attach`、`/attach/index`、`/attach/upavatar` | `c.api.attach->index/upavatar` | 本轮完成 | 附件空入口和系统头像更新；空响应、未登录、登录非法参数对比通过，成功更新由 service fake 覆盖。 |
| `/:size/:uri` | `c.api.pic->index` | 本轮完成 | 图片裁剪/缩略/原图输出入口；不存在和非法文件 404 对比通过，图片生成由 service 测试覆盖。 |
| `/test` | `c.api.test->test` | 本轮完成 | 动态 PNG 二进制输出，按 HTTP status、`image/png` 和 100x34 PNG 形态对比。 |
| `/iploc/:ip` | `c.api.index->iploc` | 本轮完成 | 忽略旧中间件动态字段后，稳定 IP 样例一致。 |
| `/shortcutstats/add`、`/adstats/add`、`/playstats/add` | `c.api.shortcutstats/adstats/playstats->add` | 本轮完成 | 统计写入接口；复刻无 token 游客 sid 创建，成功和参数错误分支对比通过。 |
| `/open`、`/open/index`、`/open/reqauth` | `c.api.open->index/reqauth` | 本轮完成 | 开放平台授权接口；`reqauth` 游客成功路径的 `authrow/openid/sign/time` 与旧 PHP 一致，旧 PHP 动态 `xxx_api_auth` 不回传。 |
| `/activity`、`/activity/index`、`/activity/details` | `c.api.activity->index/details` | 本轮完成 | 活动只读接口；当前无活动和无效 `aid` 错误分支对比通过，成功分支按源码读取活动与奖项。 |
| `/activity/luckyprizes` | `c.api.activity->luckyprizes` | 本轮完成 | 静态充值抽奖奖项列表；忽略旧 PHP 动态 `xxx_api_auth` 后一致。 |
| `/activity/newyear2020`、`/activity/luckydraw` | `c.api.activity->newyear2020/luckydraw` | 本轮完成 | 过期抽奖活动入口；当前日期下旧 PHP 直接返回活动结束错误，忽略动态 `xxx_api_auth` 后一致。 |
| `/activity/luckydrawhistory` | `c.api.activity->luckydrawhistory` | 本轮完成 | 登录只读充值抽奖历史；未登录和登录空历史分支对比通过。 |
| `/activity/ranking`、`/activity/receive` | `c.api.activity->ranking/receive` | 本轮完成 | 登录活动排名/领奖结果预览；未登录和无效活动分支对比通过，成功分支由 service fake 覆盖。 |
| `/activity/recommends` | `c.api.activity->recommends` | 本轮完成 | 登录邀请记录只读接口；未登录和无效活动分支对比通过，成功分支由 service fake 覆盖。 |
| `/invite/info` | `c.api.invite->info` | 本轮完成 | 登录只读当前绑定的邀请码；未登录和登录真实 token 分支对比通过。 |
| `/invite/bind` | `c.api.invite->bind` | 本轮完成 | 接管未登录、已绑定、缺邀请码、无效邀请码和无法绑定自己前置失败分支；绑定关系、VIP/金币奖励和事务写入成功分支未接管。 |
| `/payment/index`、`/payment/query` | `c.api.payment->index/query` | 本轮完成 | 只读订单状态查询；校验订单归属，未授权返回 `无权限`；裸 `/payment` 旧 PHP 为 404，不接管。 |
| `/payment/payways` | `c.api.payment->payways` | 本轮完成 | 只读订单支付方式列表；校验订单存在、未支付和归属，支付通道通过接口隔离，不伪造生产配置。 |
| `/payment/chpayway` | `c.api.payment->chpayway` | 本轮完成 | 修改未支付订单支付方式；保留本人校验、支付方式白名单校验和条件更新，避免已支付订单被修改。 |
| `/payment/unpaid` | `c.api.payment->unpaid` | 本轮完成 | 当前 PHP 运行代码固定返回 `data.total_count=0`；未执行的未支付订单查询分支暂不接管。 |
| `/payment/success`、`/payment/failed` | `c.api.payment->success/failed` | 本轮完成 | 固定支付状态 JSON 文案；不包含第三方支付回调验签或入账逻辑。 |
| `/payment/wappay1`、`/payment/wappay2`、`/payment/pay7submit`、`/payment/pay11` | `c.api.payment->wappay1/wappay2/pay7submit/pay11` | 本轮完成 | 支付返回页/只读 HTML 分支；`wappay2?payid=` 只读 `payhtml`，`pay7submit` 生成自动 POST 表单，`pay11` 支持二维码页。 |
| `/payment/pay7`、`/payment/pay8`、`/payment/pay9`、`/payment/pay10`、`/payment/pay10a`、`/payment/pay10b`、`/payment/pay12`、`/payment/gpay1`、`/payment/gpay2`、`/payment/newpay*` 页面 action | `c.api.payment->$action` | 本轮完成 | PHP public action 仅返回支付成功 HTML，Go 已批量接管；对应 `_action` 下单/写入分支未伪造。 |
| `/payment/shangfu`、`/payment/wappay3`、`/payment/wappay4`、`/payment/wappay4a`、`/payment/wappay5`、`/payment/hawpay`、`/payment/easypay`、`/payment/pay6` | `c.api.payment->$action` | 本轮完成 | 固定 public 成功回调 JSON 文案，返回 `retcode=0 errmsg=支付成功回调`；不包含网关请求或入账。 |
| `/payment/reqpay`、`/payment/pay12req` | `c.api.payment->reqpay/pay12req` | 本轮完成 | 接管缺失/已支付/过期/非本人和 known payway 支付方式不允许前置失败分支；`pay12req` 错误分支返回 payerror HTML，成功请求网关暂未接管。 |
| `/respond/*` 常见支付 provider 失败分支 | `c.respond.*` | 本轮完成 | 空请求/解析失败分支返回旧 provider `echoErr()` 文本；成功验签、锁单入账和 `payment->doAction()` 未接管。 |
| `/respond/chan1` | `c.respond.chan1` | 本轮完成 | 仅接管 `mobi|secret` token 校验失败分支，返回 `retcode=1 errmsg=校验失败`；成功短信/通知处理未接管。 |
| `/register`、`/login`、`/forgot`、`/delete`、`/changePhone`、`/v2/register`、`/v2/login`、`/v2/forgot` | `c.api.user/user2`、`c.apiv2.user` | 本轮完成 | 接管安全前置失败/只读推进分支：未同意协议、注册关闭、IP 频控、未登录、手机号/邮箱/用户名格式和查重、密码登录关闭、游客账号无需注销、空账号、无效 step、v2 账号不存在、v2 空密码、forgot/changePhone step1 等；成功注册/登录/改密/注销/换绑仍未接管。 |
| `/bought/delete` | `c.api.bought->delete` | 本轮完成 | 登录删除已购影片记录；未登录和登录空 `vodids` 分支对比通过。 |
| `/explore/notification`、`/explore/notification/index` | `c.api.explore.notification->index` | 本轮完成 | 旧 PHP 空 OK 入口；Go 不回传动态游客 token。 |
| `/explore/notification/clean` | `c.api.explore.notification->clean` | 本轮完成 | 发现页红点清理，`tabkey` 空/不存在错误分支对比通过，`all` 和指定 tab 更新由 fake 覆盖。 |
| `/explore/signtask`、`/explore/signtask/index`、`/explore/signtask/sign` | `c.api.explore.signtask->index/sign` | 本轮完成 | 空入口按旧 PHP 返回 OK；`sign` 已接管登录用户/游客签到事务、金币或 VIP 奖励、签到日志和连续天数更新。 |
| `/explore/vodtask`、`/explore/vodtask/index` | `c.api.explore.vodtask->index` | 本轮完成 | 旧 PHP 空 OK 入口。 |
| `/explore/vodtask/show/:vid` | `c.api.explore.vodtask->show` | 本轮完成 | 激励视频展示和当日领取日志创建/复用；错误分支 live 对比通过。 |
| `/explore/vodtask/reqcoin` | `c.api.explore.vodtask->reqcoin` | 本轮完成 | 激励视频金币领取；登录用户写金币日志，游客更新游客金币，重复领取和越权错误对齐。 |
| `/explore/index` | `c.api.explore.index->index` | 本轮完成 | 发现页入口，读取可见 tab，按用户/游客权限计算未来 7 天签到奖励和当前签到状态，对比通过。 |
| `/game/platforms` | `c.api.game.index->index` | 本轮完成 | 读 `game_platform`，保留 PHP 字段和值类型并剔除 `json`。 |
| `/game/categories` | `c.api.game.index->categories` | 本轮完成 | 读 `game_category`，保留 PHP 字段和值类型并拼接资源 URL。 |
| `/v2/amazing/categories` | `c.apiv2.amazing->categories` | 本轮完成 | 读 `amazing_category` 固定列，支持 `parent_id`。 |
| `/v2/so/list` | `c.apiv2.so->index` | 本轮完成 | 读 `server_so_config.value`，按 PHP `json_decode` 行为返回 `data.data`。 |
| `/v2/vod/listing`、`/v2/vod/recommend`、`/v2/vod/hot`、`/v2/vod/latest` | `c.apiv2.vod->listing` | 本轮完成 | 动态 action 路由组，支持 `-params`，迁移列表筛选、排序、分页和核心 `vodrows` 字段。 |
| `/v2/vod/show/:vodid` | `c.apiv2.vod->show` | 本轮完成 | v2 视频详情；复用详情 service，错误分支和真实样例对比通过。 |
| `/v2/amazing/listing`、`/v2/amazing/recommend`、`/v2/amazing/hot`、`/v2/amazing/latest` | `c.apiv2.amazing->listing` | 本轮完成 | 动态 action 路由组，支持 `-params`，迁移精彩推荐列表筛选、排序和分页。 |
| `/vod/listing`、`/vod/recommend`、`/vod/hot`、`/vod/latest` | `c.api.vod->listing` | 本轮完成 | 非 v2 动态 action 路由组，支持 `-params`；复用 VOD 列表服务并对齐 PHP 分页 selector。 |
| `/vod/show/:vodid` | `c.api.vod->show` | 本轮完成 | 视频详情只读接口，迁移主视频、父级分类、相似视频和猜你喜欢；随机列表按 shape 对比。 |
| `/vod/reqplay/:vodid`、`/vod/reqdown/:vodid` | `c.api.vod->reqplay/reqdown` | 本轮完成 | 长视频播放/下载地址请求的可控路径；记录/购买/权限/地址错误、免费/限免、已观看和权限额度内提供地址已接管，扣金币、日志和奖励分支后续事务化迁移。 |
| `/vod/preView/:vodid/index.m3u8` | `c.api.vod->preView` | 本轮完成 | m3u8 试看输出；HTTP 拉取通过 fetcher 注入，测试用 fixture，不依赖真实 CDN。 |
| `/sendfile/play/:file`、`/sendfile/down/:file` | `c.api.sendfile->play/down` | 本轮完成 | 兼容旧 PHP 空壳行为：play 只做登录和 vodid 存在性检查，成功空 200；down 空 200。 |
| `/comment/listing-:params` | `c.api.comment->listing` | 本轮完成 | 评论列表公共只读接口，支持评论树、排序、分页和用户头像/VIP 标识。 |
| `/comment/post` | `c.api.comment->post` | 本轮完成 | 登录评论发布；保留权限、长度、字符、回复、重复校验和评论树写入，金币奖励和回复通知保留后续接入点。 |
| `/playlog`、`/playlog/index`、`/downlog`、`/downlog/index` | `c.api.playlog/downlog->index` | 本轮完成 | 旧 PHP 空方法，返回 `200 text/html` 空 body。 |
| `/playlog/listing` | `c.api.playlog->listing` | 本轮完成 | 播放记录只读列表；不强制登录，游客按 sid 查询，支持 timeline/page 和 PHP 相对时间格式，游客 timeline 2/3 保留旧 PHP 边界反序行为。 |
| `/downlog/listing` | `c.api.downlog->listing` | 本轮完成 | 下载记录只读列表；不强制登录，游客按 sid 查询，支持 timeline/page 和 PHP 相对时间格式。 |
| `/playlog/remove`、`/downlog/remove` | `c.api.playlog/downlog->remove` | 本轮完成 | 播放/下载记录软删除；未登录按游客 sid，空 `vodids` 返回 `已删除0项`。 |
| `/comment`、`/comment/index` | `c.api.comment->index` | 本轮完成 | 旧 PHP 空方法，返回 `200 text/html` 空 body。 |
| `/favorite`、`/favorite/index`、`/minifavorite`、`/minifavorite/index` | `c.api.favorite/minifavorite->index` | 本轮完成 | 旧 PHP 空方法，返回 `200 text/html` 空 body。 |
| `/v2/minifavorite`、`/v2/minifavorite/index` | `c.apiv2.minifavorite->index` | 本轮完成 | 旧 PHP 空方法，返回 `200 text/html` 空 body。 |
| `/minivod/listing`、`/minivod/recommend`、`/minivod/hot`、`/minivod/latest` 及 `-params` | `c.api.minivod->listing` | 本轮完成 | 小视频公共列表，支持筛选、排序、分页、随机推荐和 latest 用户包装 rows。 |
| `/minivod/topzan`、`/minivod/topcomment`、`/minivod/topplay`、`/minivod/topcoin`、`/minivod/topnew`、`/minivod/topday`、`/minivod/topweek`、`/minivod/topmonth` 及 `-params` | `c.api.minivod->listing` | 本轮完成 | 小视频排行榜列表，setting 序列化 ID 排行和日/周/月榜对比通过。 |
| `/minivod/show/:vodid` | `c.api.minivod->show` | 本轮完成 | 小视频详情只读接口；返回详情、作者、分类层级、相关视频和猜你喜欢，错误分支 live 对比通过。 |
| `/minivod/up/:vodid`、`/minivod/down/:vodid` | `c.api.minivod->up/down` | 本轮完成 | 小视频赞踩；登录用户写 `vod_updowns`，游客用进程内 limiter，无效视频分支 live 对比通过。 |
| `/minivod/reqplay/:vodid`、`/minivod/reqdown/:vodid` | `c.api.minivod->reqplay/reqdown` | 本轮完成 | 小视频播放/下载地址请求的可控路径；记录/权限/地址错误、免费/限免、已观看和权限额度内提供地址已接管，扣金币和任务奖励分支后续事务化迁移。 |
| `/minivod/reqlist` | `c.api.minivod->reqlist` | 本轮完成 | 小视频请求列表的现有 viewlog 读取路径；返回 `rows[].vodrow/user`，拉取推荐、标记已浏览和广告插入副作用后续迁移。 |
| `/minivod/reqlong/:vodid` | `c.api.minivod->getLong2Mini` | 本轮完成 | 长视频转小视频播放地址；成功直接返回 `text/html` URL，错误分支 live 对比通过。 |
| `/minivod/parselong/:vodid/index.m3u8` | `c.api.minivod->parseM3u8` | 本轮完成 | 接管媒体 CDN 请求前的记录不存在和播放地址不存在错误分支；m3u8 拉取与裁剪成功分支未接管。 |
| `/miniplaylog/listing`、`/miniplaylog/remove` | `c.api.minivod->history/historyDelete` | 本轮完成 | 小视频播放记录列表和删除；列表按小视频分表读取，删除空参数 live 对比通过。 |
| `/my/:authorid`、`/my/:authorid/index`、`/my/:authorid/listing` | `c.api.my->index/listing` | 本轮完成 | 作者主页小视频列表；`/my/1` 关键字段 live 对比通过，用户不存在分支一致。 |
| `/community/list`、`/community/recommend`、`/community/hot`、`/community/latest`、`/community/favorite` 及 `-params` | `c.api.topic->list` | 本轮完成 | 社区主题列表；`favorite` 未登录分支、列表主结构和媒体字段 live 对比通过。 |
| `/community/show` | `c.api.topic->show` | 本轮完成 | 社区详情；返回主题、媒体和评论树，保留旧 PHP `visit_count+1` 副作用。 |
| `/community/clisting`、`/community/clisting-:params` | `c.api.topic->clisting` | 本轮完成 | 社区评论列表；`tid` 不存在分支和空评论分页 live 对比通过。 |
| `/community/categories` | `c.api.topic->categories` | 本轮完成 | 公共只读分类；支持 `parent_id` 过滤，只返回 `status=1` 分类。 |
| `/community/slides` | `c.api.topic->slides` | 本轮完成 | 公共只读轮播；读取 `global_adgroup_ad19` 并映射 article/link/game。 |
| `/community/search` | `c.api.topic->search` | 本轮完成 | 公共搜索；空关键词返回 `请输入关键词`，非空返回 `rows/hotwords/pageinfo`。 |
| `/ucp/withdraw/rule` | `c.api.ucp.withdraw->rule` | 本轮完成 | 公共只读提现规则；读取 `withdraw.rule` 的 HTML 内容。 |
| `/community/attention` | `c.api.topic->attention` | 本轮完成 | 登录收藏/取消收藏帖子，支持 `tids` 批量取消；未登录分支 live 对比通过，成功写入由 fake 覆盖。 |
| `/community/up`、`/community/up_comment` | `c.api.topic->up/up_comment` | 本轮完成 | 登录点赞/取消点赞帖子或评论；未登录分支 live 对比通过，成功写入由 fake 覆盖。 |
| `/community/comment` | `c.api.topic->comment` | 本轮完成 | 登录社区评论发布；未登录分支 live 对比通过，成功写入由 fake 覆盖。 |
| `/community/post` | `c.api.topic->post` | 本轮完成 | 登录发布社区主题；未登录分支 live 对比通过，无文件成功分支由 fake 覆盖，图片保存后续接管。 |
| `/game/games` | `c.api.game.index->games` | 本轮完成 | 读 `game`，支持 `platform_id/category_id`，保留 PHP 字段和值类型并拼接资源 URL。 |
| `/game/broadcasts` | `c.api.game.index->broadcasts` | 本轮完成 | 读 `game_broadcast`，按 PHP 替换 `{user}` 和 `{amount}` 占位符。 |
| `/getLikeRows` | `c.api.index->getLikeRows` | 本轮完成 | 复用 VOD 行处理，按旧 PHP 固定返回 6 条随机猜你喜欢。 |
| `/getCover` | `c.api.index->getCover` | 本轮完成 | 封面加密代理；缓存/外部服务/AES 成功分支由 fake 覆盖，非法 pic 返回旧错误壳且避免外部服务阻塞。 |
| `/getCertUuid` | `c.api.index->getCertUuid` | 本轮完成 | 证书 UUID 外部查询代理；本地错误分支对比一致，成功分支通过 fake client 覆盖。 |
| `/getGlobalData` | `c.api.index->getGlobalData` | 本轮完成 | 全局配置聚合；核心 key shape、版本覆盖、热门标签/分类和开关字段 live 对比通过，忽略旧 PHP 动态游客 token。 |
| `/init` | `c.api.index->init` | 本轮完成 | 客户端初始化聚合；游客/登录用户核心字段、顶层 appver、globalData appver 覆盖、通知和站点配置 live 对比通过。 |
| `/`、`/index` | `c.api.index->index` | 本轮完成 | 首页聚合；广告 slide、推荐、最新、猜你喜欢、分类分组、标签视频和热片核心 key/count live 对比通过。 |
| `/game/wali/gameList` | `c.api.game.wali->games` | 本轮完成 | 瓦力平台游戏列表，普通分类只读对齐；`category_id=5` 游客返回旧 PHP 未登录错误。 |
| `/game/wali/test` | `c.api.game.wali->ping` | 本轮完成 | 瓦力平台 ping；读取 `game_platform.json` 后 AES-ECB 加密、MD5 签名并外呼，live 对比一致。 |
| `/game/wali/balance` | `c.api.game.wali->getBalance` | 本轮完成 | 需要登录但无本地写入；复用瓦力 AES/签名外呼，返回外部平台余额。 |
| `/game/wali/topup`、`/game/wali/withdraw`、`/game/wali/enter` | `c.api.game.wali->topup/withdraw/enterGame` | 本轮完成 | 接管未登录、上分低于 `gamecoinlimit`、上分余额不足、下分金额不正确分支；金币事务、外部平台请求和进入游戏成功分支未接管。 |
| `/game/lottery/gameList` | `c.api.game.lottery->gameList` | 本轮完成 | 彩票普通分类只读列表；`category_id=5` 游客返回旧 PHP 未登录错误，登录常玩列表后续单独接管。 |
| `/game/lottery/topup`、`/game/lottery/withdraw`、`/game/lottery/enter`、`/game/lottery/balance` | `c.api.game.lottery->$action` | 本轮完成 | 接管未登录、上分低于 `gamecoinlimit`、上分余额不足、下分金额不正确分支；彩票平台资产、余额和进入游戏成功分支未接管。 |
| `/hgame/index` | `c.api.hgame->index` | 本轮完成 | HGame 公共只读列表，返回 `data.data.list/slide`，`/hgame` 本身保持旧 PHP 404 未接管。 |
| `/ucp/rolltitle` | `c.api.ucp.index->rolltitle` | 本轮完成 | 个人中心滚动消息公共只读接口，读 `roll_titles` 中 `status=1` 的最近 10 条。 |
| `/ucp/task/sharepic` | `c.api.ucp.task->sharepic` | 本轮完成 | 公共随机推广海报，只读 `poster.status=1`，随机行按 shape 对比。 |
| `/ucp/taskbox/index` | `c.api.ucp.taskbox->index` | 本轮完成 | 公共只读任务宝箱状态和最近开启记录；`/ucp/taskbox` 无稳定响应未接管，领奖 action 未接管。 |
| `/ucp/taskbox/share` | `c.api.ucp.taskbox->share` | 本轮完成 | 公共只读任务宝箱分享文案；随机游客邀请码/登录邀请码和每日推广 URL 形态对齐。 |
| `/ucp/taskbox/taskboxlog` | `c.api.ucp.taskbox->taskboxlog` | 本轮完成 | 登录只读本人任务宝箱日志；未登录错误、登录测试 token 分页和首行内容 live 对比通过。 |
| `/ucp/task/invite` | `c.api.ucp.task->invite` | 本轮完成 | 未登录返回旧错误；登录后 PHP 方法体为空，Go 返回 200 空 body。 |
| `/ucp/task/sign`、`/ucp/task/share`、`/ucp/task/qrcode`、`/ucp/task/qrcodeSave`、`/ucp/task/invitecodeInput`、`/ucp/task/adviewClick`、`/ucp/taskbox/taskboxopen`、`/ucp/taskbox/qrcode`、`/ucp/upgrade`、`/ucp/withdraw/create`、`/ucp/vippkg/placeorder`、`/ucp/vippkg/coinorder`、`/ucp/coinpkg/placeorder`、`/ucp/beanpkg/placeorder`、`/ucp/beanpkg/coinorder` | `c.api.ucp.*` | 本轮完成 | 接管高风险写入接口未登录及部分只读失败分支：`sign` 已签到/游客缺失、邀请码错误、二维码今日已保存、广告今日已送、宝箱不存在/金币为 0；`upgrade` 补齐已是尊贵会员、无效时长、终身 VIP 暂停和金币不足；`withdraw/create` 补齐金额、最小提现、限制提现、邀请人数、收款账号、日次数、渠道金额范围、游戏余额不足和普通提现兑换前置失败；套餐下单/金币兑换补齐套餐不存在/停用、余额不足和 VIP `rmbprice=3800` 失败分支；登录成功奖励、资产、支付下单、提现事务、二维码/图片生成仍未接管。 |
| `/ucp/coinlog/exchange` | `c.api.ucp.coinlog->exchange` | 本轮完成 | 接管兑换关闭、未登录、缺兑换类型、缺兑换数量、超过 100 万、金币换人民币最小金币和计算为 0 前置失败分支；金币兑换成功分支未接管。 |
| `/ucp/vodorder/create`、`/ucp/vodorder/support` | `c.api.ucp.vodorder->create/support` | 本轮完成 | 接管未登录、求片缺番号/名称、求片金币低于 100、金币不足、助力记录缺失、助力时间窗口、助力金币低于 1 和金币不足前置失败分支；求片扣费、助力写入和通知未接管。 |
| `/ucp/msg/show` | `c.api.ucp.msg->show` | 本轮完成 | 登录消息详情；返回会话、对方用户、消息列表并标记已读，错误壳和成功样例对比通过。 |
| `/ucp/msg/setread`、`/ucp/msg/cleanread`、`/ucp/msg/delete` | `c.api.ucp.msg->setread/cleanread/delete` | 本轮完成 | 登录消息状态写入；未登录和空数组成功分支对比通过，旧 PHP 动态 `xxx_api_auth` 不回传。 |
| `/ucp/user/checkemail`、`/ucp/user/sendemail`、`/ucp/user/verifyemail`、`/ucp/user/bindmobi` | `c.api.ucp.user->$action` | 本轮完成 | 仅接管未登录、邮箱格式错误、邮箱验证码缺失/失效和手机验证码错误分支；邮件发送、邮箱/手机绑定成功分支未接管。 |
| `/ucp/user/profile`、`/ucp/user/passwd` | `c.api.ucp.user->profile/passwd` | 本轮完成 | 接管未登录分支；`passwd` 额外接管密码长度和确认密码不一致前置失败分支；资料更新和密码校验写入未接管。 |
| `/onego` | `c.api.onego->rules`（旧路由默认行为） | 本轮完成 | 裸一元购入口按旧服务返回规则/未开放错误壳，忽略旧中间件动态 `data.xxx_api_auth`。 |
| `/onego/index` | `c.api.onego->index` | 本轮完成 | 旧 PHP 空方法，返回 `text/html` 空 body。 |
| `/onego/rules`、`/onego/rooms`、`/onego/current`、`/onego/last` | `c.api.onego->rules/rooms/current/last` | 本轮完成 | 一元购公共只读接口，读 `one_go`、`one_go_rooms`、`one_go_records`；本地错误分支和房间列表业务数据对齐，忽略旧中间件动态 `data.xxx_api_auth`。 |
| `/onego/hash` | `c.api.onego->hash` | 本轮完成 | 一元购公共哈希计算接口，复刻 PHP `hash('sha256')` 和末尾数字期号截取规则。 |
| `/onego/history` | `c.api.onego->history` | 本轮完成 | 登录只读本人投注历史；未登录和测试 token 空历史 live 对比通过。 |
| `/onego/bet` | `c.api.onego->bet` | 本轮完成 | 接管未登录、押注数量为 0、无效场次、无效期号、未开始、已结束、未知用户和余额不足前置失败分支；金币扣减、号码生成和订单写入成功分支未接管。 |
| `/onego/lucky` | `c.api.onego->lucky` | 本轮完成 | 一元购幸运榜公共只读接口，按获奖金币总数排序并附带各房间获奖次数；保留 PHP 未分页排行 SQL。 |
| `/onego/bet_ranks` | `c.api.onego->bet_ranks` | 本轮完成 | 押注排行只读接口；无效场次和无效期号 live 对比通过，本地无订单样本，成功分支由 service fake 覆盖。 |
| `/onego/marquee` | `c.api.onego->marquee` | 本轮完成 | 一元购跑马灯公共只读接口，读取最近已开奖期前 10 条记录并按规则模板生成中奖消息。 |
| `/special/index` | `c.api.special->index` | 本轮完成 | 旧 PHP 空方法，返回 `text/html` 空 body。 |
| `/special/listing`、`/special/detail/:spid` | `c.api.special->listing/detail` | 本轮完成 | 专题公共接口，复用 VOD 行处理；listing 含分页、前 4 个视频和第一页 actorrows，detail 含全量视频排序和浏览数写入副作用。 |
| `/special/up/:spid`、`/special/down/:spid` | `c.api.special->up/down` | 本轮完成 | 专题赞踩；不存在分支 live 对比通过，成功/重复投票写入分支由 service fake 覆盖。 |
| `/vod/breaking` | `c.api.vod->breaking` | 本轮完成 | 每日爆料公共只读接口，读取当天 `cateid=99` 视频并返回 `vodid/title`；本地无记录错误分支 live 对比通过。 |
| `/vod/errorreport`、`/v2/vod/errorreport` | `c.api.vod->errorreport`、`c.apiv2.vod->errorreport` | 本轮完成 | 视频报错反馈；校验必填参数、视频存在和重复提交后写 `vod_errors`，不涉及金币、支付或播放权限。 |
| `/search` | `c.api.search->index` | 本轮完成 | 普通视频搜索；空关键词返回热词/热片/猜你喜欢，带关键词返回分页视频列表并更新 `vod_schlogs`。 |
| `/minisearch` | `c.api.miniSearch->index` | 本轮完成 | 小视频搜索；空关键词返回小视频热词/热片/猜你喜欢，带关键词返回 `rows[].vodrow` 并更新 `minivod_schlogs`。 |
| `/art`、`/art/index` | `c.api.art->index` | 本轮完成 | 旧 PHP 空方法，返回 `text/html` 空 body。 |
| `/art/announce` | `c.api.art->announce` | 本轮完成 | 系统公告列表，读 `art_categories.uuid=announce` 和公开 `arts`；分页 URL 保留旧 PHP `/art/?page=[?]` 行为。 |
| `/art/show` | `c.api.art->show` | 本轮完成 | 公告/文章详情，读 `arts` 和 `arts_content`；成功、缺少 `artid`、不存在记录分支均对比一致。 |
| `/aiundress`、`/aiundress/listing` | `c.api.aiundress->listing` | 本轮完成 | 登录只读 AI 任务历史；未登录错误、`module/page` 分页、字段集合和 test 环境 R2 资源域名 live 对比通过。 |
| `/aiundress/index` | `c.api.aiundress->index` | 本轮完成 | 按本地旧 PHP 运行时行为返回 `200 text/html` 空 body；AI 上传/生成/查询 action 未接管。 |
| `/aiundress/moduleList`、`/aiundress/resourceTypeList`、`/aiundress/resourceList` | `c.api.aiundress->moduleList/resourceTypeList/resourceList` | 本轮完成 | 只读第三方资源查询；`channel_key` 不硬编码，需通过 `AIUNDRESS_THIRD_KEY` 注入，缺配置按旧 PHP 外部请求失败返回 `retcode=-1 errmsg=请求失败`。 |
| `/aiundress/upload`、`/aiundress/undress`、`/aiundress/delete` | `c.api.aiundress->upload/undress/delete` | 本轮完成 | 接管未登录失败分支，`delete` 额外接管记录不存在空 OK 分支；图片上传、AI 生成、删除外部资源和金豆扣减未接管。 |
| `/starLive/index` | `c.api.starlive->index` | 本轮完成 | 直播初始化；支持登录用户或游客 sid，读取 `starlive_info`，兼容 PHP AES-128-CBC/base64 的 `encryptUid` 和 `md5(appId_userId_secKey)` token。 |
| `/starLive/queryCoinBalance` | `c.api.starlive->queryCoinBalance` | 本轮完成 | 直播余额查询；返回 raw JSON `{code,data}`，游客长 memberId 余额为 0，用户余额按 `users_quota.goldcoin*10`。 |
| `/starLive/gameBet`、`/starLive/gameWin`、`/starLive/translate`、`/starLive/tryAgain` | `c.api.starlive->$action` | 本轮完成 | 接管 raw JSON 安全失败分支：游客长 `memberId`、未知用户和 `tryAgain` 未知业务类型；资产变更、下注结算、翻译扣款和外部回调未接管。 |
| `/minivod/reqcoin` | `c.api.minivod->reqcoin` | 本轮完成 | 小视频播放任务金币领取；登录用户写金币日志，游客更新游客金币，保留旧 PHP 未校验 log 归属行为。 |
| `/minivod/throwcoin/:vodid` | `c.api.minivod->throwcoin` | 本轮完成 | 接管未登录、视频不存在、作者不存在、GET 初始化 `mincoin/maxcoin/goldcoin`、POST 非正数和范围校验分支；金币投币事务和作者收益未接管。 |

## 优先候选

| 优先级 | 接口 | PHP handler | 风险 |
| --- | --- | --- | --- |
| 1 | `/register`、`/login`、`/forgot` 成功路径 | `c.api.user` | 高；安全前置失败分支已迁，成功路径仍涉及账号、验证码、session 写入和风控。 |

## 暂缓

| 接口 | 原因 |
| --- | --- |
| `/register`、`/login`、`/forgot` 成功路径 | 失败和部分 step1 只读分支已迁；成功路径涉及账号、短信/邮箱验证码、风控、session 和写库。 |
| `/payment/*` 剩余下单/跳转 action、`/respond/*` 成功分支 | 支付页面、只读分支、unknown payway 选择支付方式返回和回调失败分支已迁；`/respond/chan1` 已接管 token 校验、用户不存在、已送过会员、套餐不存在/停用等失败分支；剩余涉及支付平台请求、下单状态写入、回调验签、锁单入账、短信通知或 `payment->doAction()`，需要独立 reviewer/灰度/回滚策略。 |
| `/ucp/*` 高风险写入接口的登录成功路径 | 未登录和部分参数前置失败分支已迁；剩余涉及任务奖励、VIP/金币/金豆资产、提现、支付订单、求片扣费、用户资料写入、密码校验、二维码生成或外部通知。 |
| `/game/*` 上下分、进入游戏和彩票余额成功路径 | 未登录分支已迁；剩余涉及金币资产、订单、外部平台和失败补偿。 |
| `/invite/bind` 成功路径 | 前置失败分支已迁；剩余涉及绑定推荐关系、VIP 赠送、金币奖励和事务回滚。 |
| `/onego/bet` 成功路径 | 前置失败分支已迁；剩余涉及投注金币扣减、号码生成、订单写入和事务回滚。 |
| `/sms/sendv`、`/sms/sendu`、`/email/send` | 验证码、短信/邮件平台、频控和风控。 |
| `/game/wali/topup`、`/game/wali/withdraw`、`/game/wali/enter`、`/game/lottery/topup`、`/game/lottery/withdraw`、`/game/lottery/enter`、`/game/lottery/balance` | 金额参数和 topup 余额不足前置失败已迁；剩余游戏资产、外部平台调用需要登录、事务、灰度和回滚策略。 |
| `/starLive/gameBet`、`/starLive/gameWin`、`/starLive/translate`、`/starLive/tryAgain` 成功路径 | 部分 raw JSON 安全失败分支已迁；剩余直播平台下注、结算、翻译扣款或外部回调涉及资产写入和平台幂等。 |
| `/minivod/throwcoin` 成功路径、`/minivod/parselong` 成功路径，以及 `/minivod/reqlist` 的拉取/标记/广告副作用、`/minivod/reqplay/reqdown` 的扣费奖励分支 | 小视频列表、排行榜、详情、播放记录、作者页、赞踩、请求列表读取路径、播放/下载可控路径、任务金币领取、投币前置/GET 分支、长视频地址转换和 parselong 前置错误已完成；剩余多涉及金币事务、奖励、媒体解析或推荐副作用。 |
| `/vod/reqplay/reqdown`、`/v2/vod/reqplay/reqdown` 的扣费日志奖励分支 | 长视频详情、赞踩、购买、播放/下载可控路径已完成；剩余涉及播放/下载扣费、日志写入和奖励。 |
| `/aiundress/upload`、`/aiundress/undress`、`/aiundress/delete` 成功路径及其他剩余 action | `/aiundress/listing`、只读资源查询、未登录失败和 delete 空记录 OK 分支已完成；剩余涉及图片上传、第三方 AI 生成、Redis 并发锁、删除外部资源和金豆扣减。 |
