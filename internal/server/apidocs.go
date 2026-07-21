package server

import (
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type routeDoc struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Handler     string `json:"handler"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

// registerAPIDocs exposes lightweight documentation endpoints without adding a
// Swagger dependency. The OpenAPI document is generated from Gin's route table,
// so newly registered routes appear in the docs automatically.
func registerAPIDocs(router *gin.Engine) {
	router.GET("/docs", apiDocsIndex)
	router.GET("/docs/routes.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "xj-comp-api",
			"routes":  collectRouteDocs(router),
		})
	})
	router.GET("/docs/openapi.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, buildOpenAPIDoc(router))
	})
}

// apiDocsIndex is intentionally static HTML. API clients should consume
// /docs/openapi.json, while humans can use this page as a small entry point.
func apiDocsIndex(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>xj-comp-api 接口文档</title>
  <style>
    body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; color: #202124; background: #f6f7f9; }
    main { max-width: 960px; margin: 0 auto; padding: 40px 20px; }
    h1 { margin: 0 0 12px; font-size: 28px; }
    p { line-height: 1.7; }
    a { color: #0b57d0; }
    code { background: #eef1f5; border-radius: 4px; padding: 2px 6px; }
    .panel { background: #fff; border: 1px solid #e4e7ec; border-radius: 8px; padding: 20px; margin-top: 20px; }
    li { margin: 8px 0; }
  </style>
</head>
<body>
  <main>
    <h1>xj-comp-api 接口文档</h1>
    <p>文档由当前 Gin 路由表自动生成，适合迁移期核对已注册接口。详细业务字段仍以 PHP 兼容响应和迁移记录为准。</p>
    <section class="panel">
      <ul>
        <li><a href="/docs/openapi.json">/docs/openapi.json</a>：OpenAPI 3.0 JSON，可导入 Apifox、Postman、Swagger Editor。</li>
        <li><a href="/docs/routes.json">/docs/routes.json</a>：Go 服务实际注册路由清单。</li>
        <li><a href="/healthz">/healthz</a>：健康检查。</li>
      </ul>
    </section>
    <section class="panel">
      <p>登录态接口兼容旧 PHP 请求头：<code>xxx_api_auth: &lt;token&gt;</code>。默认 JSON 响应壳兼容 <code>retcode</code>、<code>errmsg</code>、<code>data</code>。</p>
    </section>
  </main>
</body>
</html>`)
}

// collectRouteDocs returns the actual routes registered in the current Gin
// engine. HEAD and OPTIONS are omitted to keep the migration-facing docs focused
// on callable business endpoints.
func collectRouteDocs(router *gin.Engine) []routeDoc {
	routes := router.Routes()
	docs := make([]routeDoc, 0, len(routes))
	for _, route := range routes {
		if route.Method == http.MethodHead || route.Method == http.MethodOptions {
			continue
		}
		summary, description := routeSummaryAndDescription(route.Method, route.Path)
		docs = append(docs, routeDoc{
			Method:      route.Method,
			Path:        route.Path,
			Handler:     route.Handler,
			Summary:     summary,
			Description: description,
		})
	}
	sort.Slice(docs, func(i, j int) bool {
		if docs[i].Path == docs[j].Path {
			return docs[i].Method < docs[j].Method
		}
		return docs[i].Path < docs[j].Path
	})
	return docs
}

// buildOpenAPIDoc creates a generic OpenAPI contract from route metadata. It
// documents paths, methods, path parameters, tags, and the legacy response shell;
// field-level request/response schemas remain in migration records and tests.
func buildOpenAPIDoc(router *gin.Engine) map[string]any {
	paths := make(map[string]map[string]any)
	tags := make(map[string]map[string]string)
	for _, route := range collectRouteDocs(router) {
		path := openAPIPath(route.Path)
		if _, ok := paths[path]; !ok {
			paths[path] = make(map[string]any)
		}
		tag := routeTag(route.Path)
		tags[tag] = map[string]string{"name": tag}
		paths[path][strings.ToLower(route.Method)] = map[string]any{
			"tags":        []string{tag},
			"summary":     route.Summary,
			"description": route.Description,
			"operationId": operationID(route.Method, route.Path),
			"parameters":  pathParameters(path),
			"responses": map[string]any{
				"200": map[string]any{
					"description": "PHP-compatible response envelope, empty response, text response, or binary body depending on the legacy endpoint.",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/LegacyResponse",
							},
						},
					},
				},
				"404": map[string]any{
					"description": "Route not found.",
				},
			},
		}
	}

	tagList := make([]map[string]string, 0, len(tags))
	for _, tag := range tags {
		tagList = append(tagList, tag)
	}
	sort.Slice(tagList, func(i, j int) bool {
		return tagList[i]["name"] < tagList[j]["name"]
	})

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "xj-comp-api",
			"version":     "0.1.0",
			"description": "Go/Gin migration API documentation generated from registered routes.",
		},
		"servers": []map[string]string{
			{"url": "/"},
		},
		"tags":  tagList,
		"paths": paths,
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"LegacyAuthToken": map[string]any{
					"type": "apiKey",
					"in":   "header",
					"name": "xxx_api_auth",
				},
			},
			"schemas": map[string]any{
				"LegacyResponse": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"retcode": map[string]string{"type": "integer"},
						"errmsg":  map[string]string{"type": "string"},
						"data":    map[string]string{"type": "object"},
					},
					"additionalProperties": true,
				},
			},
		},
	}
}

// routeSummaryAndDescription gives every registered route a Chinese explanation.
// Exact entries are used for high-risk or special endpoints; the derived fallback
// keeps newly migrated routes documented as soon as they are registered.
func routeSummaryAndDescription(method string, path string) (string, string) {
	if copy, ok := exactRouteDescriptions[path]; ok {
		return copy.Summary, copy.Description
	}

	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "首页入口", "旧 PHP 首页入口兼容路由，返回初始化或默认页面相关数据。"
	}

	parts := strings.Split(trimmed, "/")
	version := ""
	if parts[0] == "v2" && len(parts) > 1 {
		version = "v2 "
		parts = parts[1:]
	}

	module := parts[0]
	if isPictureSizeModule(module) {
		return "图片资源输出", "按旧 PHP 图片资源路径输出指定尺寸或原图资源，路径参数 uri 表示资源相对地址。"
	}

	moduleName := labelFor(moduleLabels, module, module)
	actionName := "默认入口"
	if len(parts) > 1 {
		actionName = actionLabel(parts[len(parts)-1])
	}
	if len(parts) > 2 {
		actionName = sectionLabel(parts[1]) + actionName
	}

	summary := strings.TrimSpace(version + moduleName + actionName)
	description := summary + "接口，兼容旧 PHP 路由 " + path + "。"
	if strings.Contains(path, ":") || strings.Contains(path, "*") {
		description += "路径中包含动态参数，OpenAPI 会转换为 {param} 形式。"
	}
	if usesLegacyAnyMethod(method) {
		description += "旧接口使用 router.Any 兼容多种 HTTP method，业务语义以旧 PHP 行为为准。"
	}
	return summary, description
}

type routeDescription struct {
	Summary     string
	Description string
}

var exactRouteDescriptions = map[string]routeDescription{
	"/docs":               {"接口文档首页", "浏览器入口，提供 OpenAPI、路由清单和健康检查链接。"},
	"/docs/openapi.json":  {"OpenAPI 文档", "输出由当前 Gin 路由表生成的 OpenAPI 3.0 JSON，可导入 Apifox、Postman 或 Swagger Editor。"},
	"/docs/routes.json":   {"路由清单", "输出当前 Go 服务实际注册的路由、handler、summary 和 description，用于迁移核对。"},
	"/healthz":            {"健康检查", "返回服务健康状态、运行环境和当前时间，供容器探针或负载均衡检查使用。"},
	"/readyz":             {"就绪检查", "返回服务就绪状态，当前与健康检查保持一致。"},
	"/":                   {"首页入口", "旧 PHP 首页入口兼容路由，返回初始化或默认页面相关数据。"},
	"/index":              {"首页入口", "旧 PHP /index 入口兼容路由，返回初始化或默认页面相关数据。"},
	"/sysavatar":          {"系统头像列表", "返回内置男女系统头像资源列表。"},
	"/register":           {"用户注册", "兼容旧 PHP 注册接口，处理账号、手机号、邮箱和验证码等注册流程。"},
	"/v2/register":        {"v2 用户注册", "v2 注册接口，兼容新客户端注册请求和旧 PHP 响应结构。"},
	"/login":              {"用户登录", "兼容旧 PHP 登录接口，校验账号密码、验证码和客户端上下文并返回登录态。"},
	"/v2/login":           {"v2 用户登录", "v2 登录接口，返回新客户端期望的登录响应。"},
	"/logout":             {"用户退出登录", "清理用户登录态或 token，兼容旧 PHP 退出登录接口。"},
	"/forgot":             {"找回密码", "兼容旧 PHP 找回密码流程，处理账号校验、验证码和密码重置。"},
	"/v2/forgot":          {"v2 找回密码", "v2 找回密码接口，兼容新客户端密码重置流程。"},
	"/delete":             {"账号删除", "处理用户注销或删除账号相关请求，兼容旧 PHP 行为。"},
	"/changePhone":        {"更换手机号", "处理用户手机号换绑流程，兼容旧 PHP 请求和响应。"},
	"/captcha/req":        {"申请图形验证码", "创建图形验证码并返回验证码标识和图片地址。"},
	"/v2/captcha/req":     {"v2 申请图形验证码", "v2 图形验证码申请接口，返回新客户端使用的 picurl。"},
	"/captcha/pic":        {"获取验证码图片", "根据验证码标识输出图形验证码图片。"},
	"/v2/captcha/pic":     {"v2 获取验证码图片", "v2 验证码图片输出接口。"},
	"/captcha/picx":       {"获取备用验证码图片", "兼容旧 PHP picx 验证码图片输出路径。"},
	"/v2/captcha/picx":    {"v2 获取备用验证码图片", "兼容 v2 客户端 picx 验证码图片输出路径。"},
	"/captcha/verify":     {"校验图形验证码", "校验客户端提交的图形验证码。"},
	"/v2/captcha/verify":  {"v2 校验图形验证码", "v2 图形验证码校验接口。"},
	"/sms/sendv":          {"发送短信验证码", "发送短信验证码，常用于登录、注册或绑定前置校验。"},
	"/sms/sendu":          {"发送用户短信验证码", "向当前用户相关手机号发送短信验证码。"},
	"/email/send":         {"发送邮箱验证码", "发送邮箱验证码或验证邮件。"},
	"/getGlobalData":      {"全局配置", "返回客户端启动后需要的全局配置、资源域名和开关信息。"},
	"/getCover":           {"启动封面", "返回启动封面、弹窗或广告相关配置。"},
	"/init":               {"初始化数据", "返回客户端初始化所需的用户、配置、首页和全局数据。"},
	"/search":             {"视频搜索", "按关键词搜索长视频内容。"},
	"/minisearch":         {"小视频搜索", "按关键词搜索小视频内容。"},
	"/getLikeRows":        {"猜你喜欢", "返回与当前内容或用户偏好相关的视频推荐列表。"},
	"/getCertUuid":        {"证书 UUID", "返回客户端证书或设备绑定流程需要的 UUID。"},
	"/vod/show/:vodid":    {"视频详情", "按 vodid 查询长视频详情。"},
	"/v2/vod/show/:vodid": {"v2 视频详情", "v2 长视频详情接口，按 vodid 查询内容详情。"},
	"/vod/reqplay/:vodid": {"申请视频播放", "校验播放权限并返回长视频播放地址或权限错误。"},
	"/v2/vod/reqplay/:vodid": {"v2 申请视频播放",
		"v2 播放权限接口，校验用户、VIP、购买记录等条件后返回播放信息。"},
	"/vod/reqdown/:vodid":    {"申请视频下载", "校验下载权限并返回长视频下载地址或权限错误。"},
	"/v2/vod/reqdown/:vodid": {"v2 申请视频下载", "v2 下载权限接口，校验用户、VIP、购买记录等条件后返回下载信息。"},
	"/vod/preView/:vodid/index.m3u8": {"视频预览 m3u8",
		"输出长视频预览 m3u8 内容，属于媒体播放相关接口。"},
	"/sendfile/play/:file":     {"播放文件代理", "按 file 参数输出或跳转播放文件资源。"},
	"/sendfile/down/:file":     {"下载文件代理", "按 file 参数输出或跳转下载文件资源。"},
	"/payment/query":           {"支付订单查询", "查询用户支付订单或支付入口相关信息。"},
	"/payment/payways":         {"支付方式列表", "返回可用支付方式列表。"},
	"/payment/chpayway":        {"切换支付方式", "按客户端请求切换或选择支付方式。"},
	"/payment/unpaid":          {"未支付订单", "查询当前用户未完成支付订单。"},
	"/payment/reqpay":          {"创建支付请求", "创建支付订单并返回支付参数或跳转信息。"},
	"/payment/pay12req":        {"Pay12 支付请求", "创建 Pay12 渠道支付请求。"},
	"/respond/chan1":           {"支付回调 chan1", "处理 chan1 支付渠道回调，按旧 PHP 文本响应兼容。"},
	"/ucp/index":               {"个人中心首页", "返回个人中心首页数据，包括用户资产、资料和入口信息。"},
	"/ucp/user/profile":        {"个人资料保存", "读取或保存用户个人资料。"},
	"/ucp/user/passwd":         {"修改密码", "处理个人中心修改密码请求。"},
	"/ucp/task/sign":           {"任务签到", "处理个人中心每日签到任务。"},
	"/ucp/task/qrcodeSave":     {"保存任务二维码", "保存或生成任务分享二维码图片。"},
	"/ucp/taskbox/taskboxopen": {"打开任务宝箱", "校验任务宝箱条件并发放奖励或返回失败原因。"},
}

var moduleLabels = map[string]string{
	"activity":      "活动",
	"adstats":       "广告统计",
	"aiundress":     "AI 脱衣",
	"amazing":       "应用",
	"art":           "图文",
	"attach":        "附件",
	"bought":        "购买记录",
	"captcha":       "图形验证码",
	"comment":       "评论",
	"community":     "社区",
	"downlog":       "下载记录",
	"email":         "邮箱",
	"explore":       "探索",
	"favorite":      "收藏",
	"game":          "游戏",
	"hgame":         "H5 游戏",
	"invite":        "邀请",
	"iploc":         "IP 归属地",
	"minifavorite":  "小视频收藏",
	"miniplaylog":   "小视频播放记录",
	"minivod":       "小视频",
	"my":            "作者主页",
	"onego":         "一元购",
	"open":          "开放接口",
	"payment":       "支付",
	"playlog":       "播放记录",
	"playstats":     "播放统计",
	"respond":       "支付回调",
	"shortcutstats": "快捷方式统计",
	"sms":           "短信",
	"so":            "搜索配置",
	"special":       "专题",
	"starLive":      "星直播",
	"test":          "测试",
	"ucp":           "个人中心",
	"vod":           "长视频",
}

var sectionLabels = map[string]string{
	"account":      "账户",
	"affcenter":    "推广中心",
	"bankcard":     "银行卡",
	"beanpkg":      "金币包",
	"coinlog":      "金币日志",
	"coinpkg":      "金币套餐",
	"feedback":     "反馈",
	"lottery":      "彩票",
	"msg":          "消息",
	"notification": "通知",
	"payment":      "支付记录",
	"task":         "任务",
	"taskbox":      "任务宝箱",
	"user":         "用户资料",
	"vippkg":       "VIP 套餐",
	"vodorder":     "视频订单",
	"vodtask":      "观影任务",
	"wali":         "瓦力游戏",
	"withdraw":     "提现",
}

var actionLabels = map[string]string{
	"add":              "新增",
	"adviewClick":      "广告点击",
	"announce":         "公告",
	"attention":        "关注",
	"balance":          "余额",
	"balancelog":       "余额日志",
	"bet":              "下注",
	"bet_ranks":        "投注排行",
	"bind":             "绑定",
	"bindmobi":         "绑定手机号",
	"bonuslog":         "奖励日志",
	"breaking":         "爆料",
	"broadcasts":       "广播列表",
	"buy":              "购买",
	"categories":       "分类列表",
	"chpayway":         "切换支付方式",
	"clean":            "清理",
	"cleanread":        "清理已读",
	"clisting":         "评论列表",
	"coinorder":        "金币下单",
	"comment":          "发表评论",
	"create":           "创建",
	"current":          "当前期",
	"delete":           "删除",
	"detail":           "详情",
	"details":          "详情",
	"down":             "点踩",
	"enter":            "进入游戏",
	"errorreport":      "错误反馈",
	"exchange":         "兑换",
	"failed":           "支付失败页",
	"gameBet":          "投注",
	"gameList":         "游戏列表",
	"games":            "游戏列表",
	"gameWin":          "派奖",
	"hash":             "哈希信息",
	"history":          "历史列表",
	"historyorders":    "历史订单",
	"index":            "首页",
	"info":             "信息",
	"invite":           "邀请",
	"invitecodeInput":  "填写邀请码",
	"invitelog":        "邀请日志",
	"last":             "上一期",
	"latest":           "最新列表",
	"listing":          "分页列表",
	"luckydraw":        "抽奖",
	"luckydrawhistory": "抽奖历史",
	"lucky":            "中奖信息",
	"luckyprizes":      "奖品列表",
	"marquee":          "跑马灯",
	"modify":           "修改",
	"moduleList":       "模块列表",
	"myorders":         "我的订单",
	"mysupports":       "我的支持",
	"newyear2020":      "新年活动",
	"passwd":           "修改密码",
	"pay11":            "Pay11 支付",
	"pay12req":         "Pay12 支付请求",
	"pay7submit":       "Pay7 提交",
	"payways":          "支付方式列表",
	"pic":              "图片",
	"picx":             "备用图片",
	"placeorder":       "创建订单",
	"platforms":        "平台列表",
	"post":             "发布",
	"profile":          "资料",
	"qrcode":           "二维码",
	"qrcodeSave":       "保存二维码",
	"qrlink":           "二维码链接",
	"query":            "查询",
	"queryCoinBalance": "金币余额查询",
	"ranking":          "排行",
	"receive":          "领取",
	"recommend":        "推荐列表",
	"recommends":       "推荐列表",
	"remove":           "移除",
	"reqauth":          "授权请求",
	"reqcoin":          "金币请求",
	"reqdown":          "下载请求",
	"reqlist":          "请求列表",
	"reqlong":          "长视频请求",
	"reqpay":           "支付请求",
	"reqplay":          "播放请求",
	"resourceList":     "资源列表",
	"resourceTypeList": "资源类型列表",
	"rooms":            "房间列表",
	"rule":             "规则",
	"rules":            "规则",
	"safepaylog":       "安全支付日志",
	"search":           "搜索",
	"send":             "发送",
	"sendemail":        "发送邮箱验证",
	"setread":          "设为已读",
	"share":            "分享",
	"sharepic":         "分享图",
	"show":             "详情",
	"sign":             "签到",
	"slides":           "轮播图",
	"success":          "支付成功页",
	"support":          "支持",
	"taskboxlog":       "宝箱日志",
	"taskboxopen":      "打开宝箱",
	"test":             "测试",
	"throwcoin":        "投币",
	"topup":            "上分",
	"translate":        "翻译",
	"tryAgain":         "重试",
	"undress":          "处理",
	"unpaid":           "未支付订单",
	"up":               "点赞",
	"up_comment":       "评论点赞",
	"upavatar":         "上传头像",
	"upload":           "上传",
	"verify":           "校验",
	"verifyemail":      "验证邮箱",
	"wappay1":          "WAP 支付 1",
	"wappay2":          "WAP 支付 2",
	"withdraw":         "下分",
}

func actionLabel(segment string) string {
	cleaned := segment
	if idx := strings.Index(cleaned, "-:"); idx >= 0 {
		cleaned = cleaned[:idx]
	}
	if strings.HasPrefix(cleaned, ":") || strings.HasPrefix(cleaned, "*") {
		return "详情"
	}
	return labelFor(actionLabels, cleaned, cleaned)
}

func sectionLabel(segment string) string {
	if strings.HasPrefix(segment, ":") || strings.HasPrefix(segment, "*") {
		return ""
	}
	return labelFor(sectionLabels, segment, segment)
}

func labelFor(labels map[string]string, key string, fallback string) string {
	if value, ok := labels[key]; ok {
		return value
	}
	return fallback
}

func usesLegacyAnyMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}

func isPictureSizeModule(module string) bool {
	switch module {
	case "C1", "C2", "C3", "C4", "C5", "C6", "C7", "C8", "C9",
		"T1", "T2", "T3", "T4", "T5", "T6", "T7", "T8", "T9",
		"R1", "R2", "R3", "R4", "R5", "R6", "R7", "R8", "R9",
		"M", "N":
		return true
	default:
		return false
	}
}

// openAPIPath converts Gin path parameters to OpenAPI syntax:
// /vod/show/:vodid -> /vod/show/{vodid}
// /C1/*uri -> /C1/{uri}
// /vod/listing-:params -> /vod/listing-{params}
func openAPIPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") || strings.HasPrefix(part, "*") {
			parts[i] = "{" + strings.TrimLeft(part, ":*") + "}"
			continue
		}
		if strings.Contains(part, ":") {
			parts[i] = colonParamPattern.ReplaceAllString(part, `{$1}`)
		}
	}
	return strings.Join(parts, "/")
}

var colonParamPattern = regexp.MustCompile(`:([A-Za-z0-9_]+)`)

// pathParameters extracts every {param} token from an OpenAPI path and turns it
// into a required string path parameter.
func pathParameters(path string) []map[string]any {
	matches := openAPIParamPattern.FindAllStringSubmatch(path, -1)
	params := make([]map[string]any, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		name := match[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		params = append(params, map[string]any{
			"name":     name,
			"in":       "path",
			"required": true,
			"schema": map[string]string{
				"type": "string",
			},
		})
	}
	return params
}

var openAPIParamPattern = regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)

// routeTag groups endpoints by their first path segment. v2 routes keep the
// second segment too, which makes /v2/vod and /vod easier to distinguish.
func routeTag(path string) string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "root"
	}
	parts := strings.Split(trimmed, "/")
	if parts[0] == "v2" && len(parts) > 1 {
		return "v2/" + parts[1]
	}
	return parts[0]
}

// operationID produces stable, import-friendly operation IDs from method and
// path without depending on handler names.
func operationID(method string, path string) string {
	raw := strings.ToLower(method) + "_" + strings.Trim(path, "/")
	if raw == strings.ToLower(method)+"_" {
		raw += "root"
	}
	var b strings.Builder
	lastUnderscore := false
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}
