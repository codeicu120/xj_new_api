---
name: unit-testing
description: >-
  为 xj_comp 编写单元测试、handler 测试、兼容响应测试和 PHP-Go 对比测试。使用场景：新增接口测试、修复 bug 回归、PHP 迁移验证、CI 失败定位。触发词：单元测试、handler 测试、对比测试、go test。
metadata:
  author: xj_comp
  version: 1.0.0
  mcp-server: optional-local-files
license: MIT
compatibility: Go 1.23、Gin、httptest、本地 Codex 工作区。
---

# Unit Testing

## Instructions

1. handler 测试优先使用 `httptest`。
2. 不连接真实 MySQL、Redis、支付、短信、Telegram、AI 或图片服务。
3. 对 PHP JSON 兼容接口，断言：
   - `retcode`
   - `errmsg`
   - `data`
   - status code
   - 必要 header
4. 对 helper 和转换函数，写表驱动测试。
5. 运行聚焦 `go test`，改动较大运行 `make ci`。
6. PHP-Go 对比无法运行时，记录阻塞依赖。

## Examples

```text
/unit-testing 为 /healthz 和 legacyjson 响应补测试
```

## Performance Notes

- 测试断言要聚焦稳定契约。
- 非稳定字段如时间、随机 secret 要做格式断言或 ignore 说明。

## Troubleshooting

- 测试失败先缩小差异，不直接放宽断言。
- 若完整 router 误连外部依赖，改成只注册待测路由。
