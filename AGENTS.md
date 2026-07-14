# AGENTS.md

## 项目概览

这是一个把 `/Users/canavs/xjProj/XJBackend/api` 中 Swoole PHP API 重构到 Go + Gin 的项目，模块名为 `xj_comp`。当前 Go 服务入口在 `cmd/api/main.go`，HTTP 路由在 `internal/server/router.go`，兼容 PHP JSON 响应壳在 `internal/legacyjson`。

目标是先完成接口级重构，后续支持服务拆分、Docker 镜像构建和 Kubernetes 部署。

主要目录：

- `cmd/api/`: Go API 服务入口。
- `internal/server/`: Gin engine、中间件和路由装配，不写业务逻辑。
- `internal/handler/`: Gin handler，负责输入解析、状态码/header 和 `legacyjson` 输出。
- `internal/service/`: 业务逻辑，不能依赖 Gin，后续服务拆分优先从这里切。
- `internal/domain/`: API/domain 数据结构和跨层共享 DTO/DO。
- `internal/config/`: 环境变量配置。
- `internal/legacyjson/`: PHP `kernel\Json` 兼容响应结构。
- `deploy/k8s/`: Kubernetes 部署模板。
- `.codex/`: 项目级 Codex 配置和 agent TOML 定义。
- `harness/`: skills、subagent 手册、共享 playbook、模板和验证脚本，不直接参与 Go 编译。

## Harness Skills Index

本项目的可复用 agent 工作流放在 `harness/skills/*/SKILL.md`。处理需求时，如果用户语义匹配下列 skill，应先读取对应 `SKILL.md`，再按其中流程执行。

| Skill | 文件 | 触发场景 |
| --- | --- | --- |
| `/subagent-orchestration` | `harness/skills/subagent-orchestration/SKILL.md` | 使用主线 agent 拆分任务，并按需分配 architect、developer、tester、reviewer、ci-cd subagent。 |
| `/php-to-go-migration` | `harness/skills/php-to-go-migration/SKILL.md` | 将旧 PHP 接口迁移或重构到 Go，保持旧行为、响应结构、数据读写和灰度回滚可控。 |
| `/php-route-discovery` | `harness/skills/php-route-discovery/SKILL.md` | 阅读 PHP 入口、Route、middleware、controller，建立迁移清单和依赖清单。 |
| `/feature-delivery` | `harness/skills/feature-delivery/SKILL.md` | 从需求分析到代码、测试、CI 和交付总结的完整后端交付。 |
| `/requirement-analysis` | `harness/skills/requirement-analysis/SKILL.md` | 梳理需求、API 契约、输入输出、业务规则、边界条件和非目标。 |
| `/code-standards` | `harness/skills/code-standards/SKILL.md` | 编写或审查 Go/Gin 代码规范、配置、安全、日志、错误处理。 |
| `/data-model-do` | `harness/skills/data-model-do/SKILL.md` | 设计 request/response DO、domain DO、database DO、cache DO 和转换规则。 |
| `/database-architecture` | `harness/skills/database-architecture/SKILL.md` | 设计 MySQL、Redis、缓存、索引、迁移、回滚和 fixture 策略。 |
| `/unit-testing` | `harness/skills/unit-testing/SKILL.md` | 编写 handler 测试、helper 测试、兼容响应测试和 PHP-Go 对比测试。 |
| `/ci-containerization` | `harness/skills/ci-containerization/SKILL.md` | Docker、CI/CD、Kubernetes、探针、镜像发布和部署就绪检查。 |

### Skill 使用规则

- 用户明确输入 `/skill-name` 时，必须读取对应 `SKILL.md`。
- 用户没有显式输入 skill，但需求明显匹配时，也应主动使用对应 skill。
- 任何“PHP 代码重构到 Go”“PHP 接口迁移”“对齐旧接口行为”的任务，必须优先读取 `/php-to-go-migration`。
- 任何需要多人分工、代码实现、测试、CR 审核或 CI/CD 验证的任务，必须优先读取 `/subagent-orchestration`；不一定每次启用全部 subagent。
- 共享 playbook 放在 `harness/references/`，交付总结模板放在 `harness/templates/operation-summary.md`。
- 原生 subagent 配置以 `.codex/agents/*.toml` 为准；`harness/subagents/*.md` 只作为长文档角色手册。
- 任务卡模板为 `harness/templates/subagent-task.md`。
- 新增或修改 skill 后运行：

```shell
./harness/scripts/validate-skills.sh
```

## 常用命令

```shell
make ci
make test
make lint
make build
make docker-build
go test ./...
go run ./cmd/api
```

## Subagents

项目级 subagent 配置位于：

```text
.codex/config.toml
.codex/agents/
```

当前 agent：

- `harness`: 主线编排，负责理解需求、拆任务、选择 subagent、合并结果和交付。
- `architect`: 需求分析、DO、数据库架构、缓存、迁移和回滚方案。
- `developer`: Go/Gin 代码实现。
- `tester`: 单元测试、回归测试和 PHP-Go 对比。
- `reviewer`: CR 审核、兼容性、安全和测试缺口检查。
- `ci-cd`: CI/CD、Docker、Kubernetes、灰度、回滚和流水线失败排查。

默认由 `harness` 作为主线；按任务风险选择需要的 subagent，不要求每次全部启用。

## 开发约定

- 新增接口时，按 `server -> handler -> service -> repository/client` 分层实现，再在 `internal/server/router.go` 注册路由。
- `internal/server` 只做路由装配，不能写业务逻辑。
- `internal/handler` 只做 HTTP 适配，复杂逻辑必须下沉到 `internal/service/<domain>`。
- `internal/service` 不能依赖 `gin.Context`。
- 对外 JSON API 默认使用 `internal/legacyjson`，除非旧 PHP 明确返回图片、HTML、纯文本或支付回调特定响应。
- PHP 兼容行为要保留响应字段、字段类型、错误码、空值表现、HTTP method 兼容和必要 header。
- 不要在 handler 中硬编码密钥、域名、连接串、token 等敏感配置；应通过环境变量或配置注入。
- 单元测试不连接真实 MySQL、Redis、支付网关、短信网关、Telegram、AI 或图片处理服务；需要接口或 fake 隔离。
- 保持 `gofmt`。提交前至少运行涉及包的 `go test`，改动较大时运行 `make ci`。

## 安全与数据注意事项

- 不要提交真实数据库配置、私钥、支付密钥、短信密钥、OpenAI API key、生产连接串。
- 支付、VIP、鉴权、用户账号、生产写入、数据库迁移相关改动必须启用 reviewer，并说明灰度和回滚。
- PHP 旧项目中的缺失配置（如 `conf/db*.php`）视为运行时注入依赖，不要在仓库中伪造生产值。
