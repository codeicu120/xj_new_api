---
name: requirement-analysis
description: >-
  梳理后端需求、API 契约、输入输出、业务规则、边界条件、非目标和验收标准。使用场景：需求不清、接口设计、迁移前定契约、拆分任务。触发词：分析需求、接口契约、验收标准、边界条件。
metadata:
  author: xj_comp
  version: 1.0.0
  mcp-server: optional-local-files
license: MIT
compatibility: Go/Gin 后端项目和 PHP 迁移项目。
---

# Requirement Analysis

## Instructions

1. 明确用户目标和非目标。
2. 定义 API 契约：path、method、输入、输出、错误、状态码。
3. 标注兼容要求：字段名、字段类型、空值、错误码、加密、header。
4. 标注依赖：DB、Redis、配置、外部服务、文件资源。
5. 输出验收标准和测试建议。

## Examples

```text
/requirement-analysis 分析 /captcha/req 迁移到 Go 的验收标准
```

## Performance Notes

- 不急着实现，先把未知项列出。
- 能从代码发现的事实不要问用户。
- 涉及生产风险时明确需要确认。

## Troubleshooting

- 如果旧 PHP 行为不明，调用 `/php-route-discovery`。
- 如果数据模型不明，调用 `/data-model-do`。
