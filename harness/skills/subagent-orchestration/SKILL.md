---
name: subagent-orchestration
description: >-
  编排 xj_comp PHP 到 Go 重构任务中的主线 agent 和 subagent 分工。使用场景：接口迁移、PHP 重构到 Go、复杂后端需求、需要代码实现、测试、CR 审核、CI/CD 验证并行或分阶段协作。触发词：subagent、分工、多人协作、主线编排、重构流水线。
metadata:
  author: xj_comp
  version: 1.0.0
  mcp-server: optional-multi-agent
license: MIT
compatibility: 本地 Codex 工作区、Go 项目文件和可选 multi-agent/subagent 工具；没有工具时按同一角色清单顺序执行。
---

# Subagent 编排

主线 agent 负责理解需求、拆任务、选择合适的 subagent、合并结果和对用户交付；subagent 只负责各自边界内的产出。

## Instructions

1. 主线先读取本 skill，再按任务类型读取相关基础 skill：
   - PHP 接口迁移：`/php-to-go-migration`。
   - PHP 路由发现：`/php-route-discovery`。
   - 新功能交付：`/feature-delivery`。
   - 数据结构或存储变更：`/data-model-do`、`/database-architecture`。
2. 主线建立任务上下文：
   - 用户目标和非目标。
   - 旧接口或现有代码位置。
   - 新接口路径、兼容要求和验收样例。
   - 已知风险：支付、鉴权、VIP、生产写入、缓存、数据迁移。
3. 主线选择 subagent，不要求每次全部启用，但 PHP 接口迁移和公开 API 行为变更必须启用 `developer`、`tester`、`reviewer`。
4. 给每个 subagent 下发独立任务卡，使用 `harness/templates/subagent-task.md`。
5. 主线汇总 subagent 结果：
   - 冲突时以用户需求、旧接口行为、项目规范和测试证据为准。
   - 不直接采纳未验证的结论。
6. 主线完成最终验证：
   - `gofmt`。
   - 聚焦 `go test`，必要时 `make ci`。
   - PHP 迁移任务必须跑 PHP-Go 对比脚本或说明无法运行的原因。
   - PHP 接口迁移任务必须更新根目录 `MIGRATION_ENDPOINTS.md`，并同步必要的 `docs/*` 迁移记录。
7. 主线执行 CR Gate：
   - PHP 接口迁移、公开 API 行为变更、缓存/数据库/鉴权/支付/VIP 相关变更，交付前必须有 reviewer 结论。
8. 主线输出交付总结：
   - 哪些 subagent 被启用。
   - 每个 subagent 的结论。
   - CR 结论和未解决问题。
   - 修改文件。
   - `MIGRATION_ENDPOINTS.md` 和相关迁移文档的更新情况。
   - 运行命令和结果。
   - 剩余风险、灰度和回滚建议。

## Examples

```text
/subagent-orchestration 重构 PHP 的 /captcha/req 到 Go，并安排测试和 CR 审核
```

```text
/subagent-orchestration 当前 CI 失败了，请让 CI/CD 角色先排查，再让 developer 修改
```

## Performance Notes

- subagent 不是越多越好；小改动可以只启用 tester 或 reviewer。
- PHP 接口迁移不是“小改动”，即使代码很少也必须经过 tester 和 reviewer。
- developer 不负责决定需求边界；边界由主线、architect 和 requirement-analysis 决定。
- tester 不为通过测试而放宽业务断言；ignore 规则必须写明原因。
- reviewer 不做大重构，只指出阻断项、风险和必要修复。

## Troubleshooting

- 如果 subagent 输出互相矛盾，主线必须复核源代码、测试结果和旧接口响应。
- 如果测试无法运行，记录具体命令、错误和替代验证。
- 如果 CI/CD 涉及生产部署、数据库迁移、支付或鉴权，主线必须先确认灰度和回滚。
