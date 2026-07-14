# xj_comp Harness Skills

本 harness 遵循 `harness-skills` 仓库模式：共享行为放在 harness 根目录，可执行工作流放在 `skills/*/SKILL.md`。

## Operating Model

把 skills 当作工作流，不当作松散 prompt：

1. 建立 scope。
2. 验证依赖。
3. 先读旧 PHP、新 Go 代码和数据契约，再开始写。
4. 先产出最小安全改动。
5. 用测试或明确的人工检查验证。
6. 总结修改文件、运行命令和剩余风险。

## Skill Directory

Skills 位于：

```text
harness/skills/<skill-name>/SKILL.md
```

每个 skill 可包含：

- `references/`: skill-local 支持资料。
- `templates/`: skill-local 输出模板。
- `assets/`: 静态支持资源。

共享资产：

- `harness/references/`: 复用 playbook。
- `harness/templates/`: 输出契约和任务卡。
- `harness/scripts/`: 验证和维护脚本。
- `.codex/agents/`: 原生 TOML subagent 定义。
- `harness/subagents/`: 长文档角色手册。

## Available Skills

| Skill | Use when |
| --- | --- |
| `/subagent-orchestration` | 拆分 PHP 重构、Go 实现、测试、CR 和 CI/CD 协作。 |
| `/php-to-go-migration` | 迁移 PHP 接口到 Go/Gin，同时保持旧行为兼容。 |
| `/php-route-discovery` | 发现 PHP 路由、controller、中间件和依赖。 |
| `/feature-delivery` | 完整交付一个后端功能。 |
| `/requirement-analysis` | 澄清需求、契约、边界和非目标。 |
| `/code-standards` | Go/Gin 代码规范、配置、安全、日志和错误处理。 |
| `/data-model-do` | DO、DTO、数据库对象、缓存对象和转换规则。 |
| `/database-architecture` | MySQL、Redis、缓存、fixture、迁移和回滚。 |
| `/unit-testing` | 单元测试、handler 测试、PHP-Go 对比测试。 |
| `/ci-containerization` | CI、Docker、Kubernetes、镜像发布和部署就绪。 |

## Required Skill Shape

每个 `SKILL.md` 必须有：

- 第一行开始 YAML frontmatter。
- `name` 与目录名一致。
- `description` 小于 1024 字符，并包含“使用场景”或“触发词”。
- `metadata.author`、`metadata.version`、`metadata.mcp-server`。
- `license` 和 `compatibility`。
- 正文段落：`## Instructions`、`## Examples`、`## Performance Notes`、`## Troubleshooting` 或 `## Error Handling`。

运行：

```shell
./harness/scripts/validate-skills.sh
```

## Subagent Operating Rules

- 主线 orchestrator 负责 scope、任务分配、结果合并和最终交付。
- subagents 按任务风险选择，不要求每次全部启用。
- `.codex/agents/harness.toml` 是默认主线编排。
- `architect` 负责需求、DO、数据库、缓存和迁移方案。
- `developer` 负责 Go/Gin 实现。
- `tester` 负责单元测试、回归测试和 PHP-Go 对比证据。
- `reviewer` 负责 CR 风险、兼容性、安全和测试缺口。
- `ci-cd` 负责 CI/CD、Docker、Kubernetes、灰度、回滚和流水线风险。
- 如果真实 subagent 工具不可用，按同一角色清单顺序执行，并报告“按 subagent 角色清单模拟执行”。
