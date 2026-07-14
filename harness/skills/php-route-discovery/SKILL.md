---
name: php-route-discovery
description: >-
  发现旧 Swoole PHP 项目的入口、路由、controller、middleware、依赖和迁移优先级。使用场景：阅读 PHP 项目、建立接口清单、寻找低风险迁移目标、为 Go 迁移准备契约。触发词：读取 PHP、路由发现、接口清单、旧接口依赖。
metadata:
  author: xj_comp
  version: 1.0.0
  mcp-server: optional-local-files
license: MIT
compatibility: 本地 PHP 项目文件和 Codex 工作区。
---

# PHP 路由发现

## Instructions

1. 建立 legacy scope：
   - 默认 PHP 根目录：`/Users/canavs/xjProj/XJBackend/api`。
   - 优先读取 `api.php`、`kernel/Route.php`、`kernel/Json.php` 和目标 controller。
2. 记录路由：
   - group、path pattern、method、middleware、controller、method、path 参数。
   - 动态 action 路由必须展开成实际 endpoint 清单；不要只记录原始正则 pattern。
   - 对 `Route::any('/vod/:action=(listing|recommend|hot|latest)(-:params=([A-Za-z0-9\-\s]+))?', ...)` 这类写法：
     - group 为 `/v2` 时，实际公共入口是 `/v2/vod/listing`、`/v2/vod/recommend`、`/v2/vod/hot`、`/v2/vod/latest`。
     - 不是 `/vod/list`；`listing` 才是旧 PHP controller 收到的 action 值。
     - 可选 `-:params` 属于同一路由，例如 `/v2/vod/listing-page-2`，会传入 `$params='page-2'`。
     - 四个 action 都调用 `M('c.apiv2.vod')->listing($context, $action, $params)`，Go 迁移时应复用同一 handler/service，并按 action 选择排序或过滤逻辑。
     - 该路由依赖 `c.api.__init__` 中间件，迁移前需要确认游客、setting、缓存和分页行为。
3. 记录输出：
   - JSON、图片、HTML、纯文本、redirect、支付回调文本。
4. 记录依赖：
   - `M('conf.*')` 配置。
   - `M('init.*')` 运行时组件。
   - DB 表/模型、Redis key、缓存、文件资源、外部服务。
5. 推荐迁移顺序：
   - 优先本地配置、只读、无鉴权副作用、无支付和无外部服务接口。
6. 输出迁移笔记：
   - 文件路径、行号、输入输出、风险、测试建议。

详细 checklist 见 `references/discovery-checklist.md`。

## Examples

```text
/php-route-discovery 分析 /captcha/req 的 PHP 路由和依赖
```

```text
/php-route-discovery 找出 5 个适合先迁移的低风险接口
```

## Performance Notes

- 使用 `rg` 和定向 `sed`，避免扫日志、vendor、runtime cache 和大二进制文件。
- 动态路由如 `Route::any('/:action')` 需要同时确认实际 controller 是否存在。
- 输出要可直接交给 `/php-to-go-migration` 使用。

## Troubleshooting

- 缺失 `conf/db*.php` 时记录为运行时注入依赖，不要猜生产配置。
- 中间件有副作用时，标注需要 fake 的上下文值。
- 旧接口加密时，记录 `conf.base['aes_encrypt']` 和 `x-cookie-auth` 规则。
