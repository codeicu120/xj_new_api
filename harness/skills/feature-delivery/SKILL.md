---
name: feature-delivery
description: >-
  从需求分析到代码实现、测试、CI 和交付总结的完整 Go/Gin 后端功能交付。使用场景：新增功能、接口改造、跨模块后端需求、需要架构、开发、测试、审查协同。触发词：实现功能、完整交付、后端需求、接口开发。
metadata:
  author: xj_comp
  version: 1.0.0
  mcp-server: optional-multi-agent
license: MIT
compatibility: Go 1.23、Gin、本地 Codex 工作区。
---

# Feature Delivery

## Instructions

1. 先用 `/requirement-analysis` 明确目标、非目标、验收标准。
2. 涉及数据结构时，用 `/data-model-do` 和 `/database-architecture`。
3. 涉及 PHP 旧行为时，切换到 `/php-to-go-migration`。
4. 实现时遵守 `/code-standards`。
5. 测试时使用 `/unit-testing`。
6. 涉及 Docker、K8s、CI/CD 时使用 `/ci-containerization`。
7. 结束前运行聚焦测试，必要时运行 `make ci`。

## Examples

```text
/feature-delivery 新增一个读取系统配置的 Go API，并补测试
```

## Performance Notes

- 小步提交思路：契约、测试、实现、验证。
- 不把不明确需求藏在代码默认值里。
- 不做无关目录重组。

## Troubleshooting

- 如果需求边界不清，先停止实现，补 `/requirement-analysis`。
- 如果测试需要真实外部服务，先设计接口和 fake。
