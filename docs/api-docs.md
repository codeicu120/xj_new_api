# 接口文档

Go 服务内置接口文档入口，启动服务后可直接访问：

- `GET /docs`：浏览器入口，提供文档链接和基础说明。
- `GET /docs/openapi.json`：OpenAPI 3.0 JSON，可导入 Apifox、Postman、Swagger Editor。
- `GET /docs/routes.json`：当前 Gin engine 实际注册的路由清单。

## 本地使用

```shell
make run
```

默认地址：

```text
http://127.0.0.1:8080/docs
http://127.0.0.1:8080/docs/openapi.json
http://127.0.0.1:8080/docs/routes.json
```

如果 `.env` 或环境变量调整了 `HTTP_HOST`、`HTTP_PORT`，按实际监听地址访问。

## 认证约定

登录态接口沿用旧 PHP 请求头：

```text
xxx_api_auth: <token>
```

OpenAPI 中已声明 `LegacyAuthToken`，但目前生成器不会自动判断每个接口是否必须登录。迁移期需要结合 `MIGRATION_ENDPOINTS.md`、`docs/auth-required-endpoints.md` 和具体 handler/service 实现确认。

## 响应约定

默认 JSON API 兼容旧 PHP 响应壳：

```json
{
  "retcode": 0,
  "errmsg": "",
  "data": {}
}
```

部分旧接口会返回空响应、纯文本、HTML、图片、m3u8 或其他二进制内容，OpenAPI 里统一以通用响应描述标注。字段级契约仍以迁移测试、PHP-Go 对比记录和具体 handler 为准。

## 生成规则说明

当前接口文档是路由级文档，生成逻辑位于 `internal/server/apidocs.go`：

- 读取 Gin engine 已注册路由，自动输出 method、path 和 handler 名称。
- 每个接口都会输出中文 `summary` 和 `description`：
  - 登录、验证码、视频播放、支付、个人中心等高频/高风险接口使用精确说明。
  - 其它接口按模块名、二级目录和 action 自动生成说明，保证新增路由不会出现空文档。
- 忽略 `HEAD`、`OPTIONS`，减少由 `router.Any` 自动展开导致的噪音。
- 将 Gin 参数转换为 OpenAPI 参数，例如 `:vodid` 转成 `{vodid}`，`*uri` 转成 `{uri}`，`listing-:params` 转成 `listing-{params}`。
- 按路径第一段自动生成 tag；`/v2/*` 会按 `v2/<module>` 分组，便于区分 v1 和 v2 接口。
- OpenAPI 目前只声明通用 PHP 兼容响应壳，不自动推断每个接口的字段、登录要求、支付副作用、上传格式或媒体返回类型。

## 维护规则

- 新增 Gin 路由后，`/docs/openapi.json` 和 `/docs/routes.json` 会自动包含该路由。
- 如果接口字段、认证、支付、上传、媒体返回等契约发生变化，需要同步更新对应迁移文档。
- PHP 接口迁移任务仍需按 harness 要求更新 `MIGRATION_ENDPOINTS.md` 和相关 `docs/*` 进度文档。
