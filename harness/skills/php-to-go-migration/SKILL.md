---
name: php-to-go-migration
description: >-
  将旧 Swoole PHP 接口迁移到 Go/Gin，并保持响应结构、错误码、字段类型、HTTP method、依赖和灰度回滚可控。使用场景：PHP 接口重构、旧接口迁移、对齐 PHP 行为、编写 PHP-Go 对比测试。触发词：PHP 到 Go、迁移接口、重构接口、兼容旧接口。
metadata:
  author: xj_comp
  version: 1.0.0
  mcp-server: optional-multi-agent
license: MIT
compatibility: Go 1.23、Gin、Swoole PHP 旧项目、本地 Codex 工作区。
---

# PHP 到 Go 接口迁移

## Instructions

1. 读取旧 PHP 行为：
   - 默认 PHP 根目录：`/Users/canavs/xjProj/XJBackend/api`。
   - 先读入口 `api.php`，再读 `kernel/Route.php`、middleware、controller 和依赖模块。
   - 必要时串联 `/php-route-discovery`。
2. 建立兼容契约：
   - 路径、method、path/query/body/header/cookie 输入。
   - HTTP status、header、JSON 壳、错误码、字段类型、空值表现。
   - 是否可能 AES 加密 `data`。
3. 先写测试：
   - handler 测试使用 `httptest`。
   - JSON API 默认断言 `retcode/errmsg/data`。
   - 旧接口支持 `Route::any` 时，至少覆盖真实客户端依赖的方法。
4. 实现 Go：
   - 按 `code-standards` 的项目结构落代码。
   - `internal/server` 只装配路由。
   - `internal/handler` 放 Gin handler。
   - `internal/service/<domain>` 放业务逻辑，不依赖 Gin。
   - `internal/domain` 放响应/领域结构。
   - 需要 DB/Redis 时再增加 `internal/repository/<domain>`。
   - 需要外部服务时再增加 `internal/client/<provider>`。
   - 保持 handler 薄。
   - 复用 `internal/legacyjson`。
   - DB、Redis、短信、支付、Telegram、AI、图片处理等外部依赖必须接口化或 fake。
5. 注册路由：
   - 在 `internal/server/router.go` 或拆出的 router 模块中替换 placeholder。
6. 验证：
   - `go test` 聚焦包。
   - 改动较大时 `make ci`。
   - 可运行旧 PHP 服务时做 PHP-Go 对比；不可运行时说明缺失依赖。
7. 交付：
   - 每次完成接口重构后，必须更新根目录 `MIGRATION_ENDPOINTS.md`：
     - 将新接口加入“已重构接口”对应分组。
     - 如果 Go 中新增或移除占位路由，同步更新“Go 已注册但仍是占位”。
     - 将已重构接口从“未重构接口”或暂缓清单中移除或改写状态。
     - 若新增登录、支付、资产、外部服务等风险说明，同步补充备注。
   - 同步更新必要的 `docs/public-endpoints.md`、`docs/auth-required-endpoints.md` 或 `docs/migration-state.md` 压缩记录。
   - 说明改动文件、测试证据、兼容差异、剩余风险和回滚建议。

## Examples

```text
/php-to-go-migration 迁移 /sysavatar，保持 PHP 的 retcode/errmsg/data 响应结构
```

```text
/php-to-go-migration 先发现 /v2/amazing/categories 的旧行为，再写 Go handler 和测试
```

## Performance Notes

- 默认一次只迁移一个 endpoint：每个 endpoint 都要独立读取 PHP 行为、写测试、实现、对比验证并更新文档。
- 如果用户明确要求“重构剩余所有接口”“全部重构”“继续直到完成”或同等语义，则进入批量迁移模式：
  - 仍按一个 endpoint 一个 endpoint 串行落地，避免混淆兼容契约。
  - 每完成一个 endpoint，立即更新 `MIGRATION_ENDPOINTS.md` 和相关 `docs/*`，并记录测试/对比证据。
  - 不要在单个 endpoint 完成后停止交付；应继续从未重构清单选择下一个风险最低、边界最清楚的 endpoint。
  - 仅在遇到阻断条件时停止：旧 PHP 行为无法确定、依赖缺失导致无法建立契约、测试环境不可用、需要用户确认高风险写入/支付/资产/鉴权策略，或当前上下文/工具限制要求压缩交接。
  - 如果需要压缩上下文，只输出压缩交接摘要并明确“下一接口继续”，不要把批量任务视为完成。
- 优先迁移低风险接口：本地配置、只读、无支付和无鉴权副作用。
- 遇到运行时缺失配置，不要伪造生产值；用 fake 或测试 fixture。
- 对支付、鉴权、VIP、用户资产相关接口，必须启用 reviewer 和 ci-cd 风险检查。
- 每个接口迁移都要保持 `server -> handler -> service` 分层，不为了快把逻辑塞回 router。

## Troubleshooting

- PHP 源码和线上响应不一致时，先判断缓存、配置、灰度或加密开关。
- JSON `data` 为空时，PHP 可能省略字段；测试要明确断言。
- 图片或文件接口不要按 JSON 断言，应断言 `Content-Type`、status 和错误路径。
