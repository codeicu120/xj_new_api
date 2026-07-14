# PHP 路由发现 Checklist

- Legacy path 和 method 行为。
- Route group 和 middleware。
- PHP controller 文件、class、method。
- 输入来源：path、query、body、headers、cookies。
- 输出类型：JSON、图片、HTML、文本、redirect、stream。
- JSON 壳：`retcode`、`errmsg`、`data`。
- HTTP status 和 headers。
- `M('conf.*')` 配置。
- `M('init.*')` 运行时模块。
- DB 表/模型依赖。
- Redis/cache key。
- 本地资源：字体、图片、IP 库、上传目录。
- 外部服务：短信、支付、Telegram、AI、图片处理。
- 建议 Go package、handler 名称。
- 必写 contract tests。
