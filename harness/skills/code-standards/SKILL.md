---
name: code-standards
description: >-
  编写或审查 xj_comp Go/Gin 代码规范，包括路由、handler、配置、安全、日志、错误处理和 PHP 兼容响应。使用场景：实现代码、代码审查、统一风格、修复质量问题。触发词：代码规范、review、Go 风格、Gin handler。
metadata:
  author: xj_comp
  version: 1.0.0
  mcp-server: optional-local-files
license: MIT
compatibility: Go 1.23、Gin、本地 Codex 工作区。
---

# Code Standards

## Instructions

1. 先读相邻代码，沿用现有包结构和命名。
2. 按项目分层放代码，不把 router、handler、service 混在一起。
3. handler 保持薄：解析、校验、调用业务逻辑、组装响应。
4. service 不依赖 Gin，方便后续服务拆分和独立测试。
5. JSON API 默认使用 `internal/legacyjson`。
6. 配置通过环境变量或配置对象注入，不硬编码敏感值。
7. 外部依赖接口化，测试用 fake。
8. Go 文件修改后运行 `gofmt`。
9. 公共行为变更必须补测试。

## Project Structure

当前 Go 分层约定：

```text
cmd/api/
  main.go                 # 进程入口：加载 config、logger、router、HTTP server

internal/config/
  config.go               # 环境变量、启动配置、资源域名等

internal/server/
  router.go               # Gin engine、中间件、路由装配；不写业务逻辑
  router_test.go          # HTTP 路由/handler 行为测试

internal/handler/
  *.go                    # Gin handler：输入解析、状态码/header、legacyjson 输出

internal/service/<domain>/
  *.go                    # 业务逻辑：不依赖 Gin，可注入 repository/client/fake
  *_test.go               # service 单元测试

internal/domain/
  *.go                    # API/domain 数据结构和跨层共享 DTO/DO

internal/legacyjson/
  response.go             # PHP kernel\Json 兼容响应壳
```

后续出现数据库或外部依赖时，优先增加：

```text
internal/repository/<domain>/   # MySQL/Redis 查询实现
internal/client/<provider>/     # 短信、支付、Telegram、AI、图片处理等外部服务
```

依赖方向必须保持：

```text
server -> handler -> service -> repository/client
handler -> legacyjson
service -> domain
```

禁止：

- `service` 依赖 `gin.Context`。
- `router.go` 直接写业务逻辑。
- handler 直接访问 MySQL/Redis/外部服务。
- domain 反向依赖 handler/server。

## Examples

```text
/code-standards 审查新增 handler 是否符合 PHP 兼容要求
```

## Performance Notes

- 最小改动优先。
- 不做无关重构。
- 兼容 PHP 的“不优雅”行为可以保留，但要有测试保护。
- 为后续服务拆分预留边界：service/domain 应尽量独立于 HTTP 框架。

## Troubleshooting

- 如果 handler 变复杂，抽 helper 或 service。
- 如果需要真实外部服务才能测试，先拆接口。
- 如果一个接口需要 DB/Redis，先定义 service 依赖的 interface，再落 repository 实现。
