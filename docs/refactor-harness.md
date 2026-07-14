# PHP to Go Refactor Harness

## Scope

This repository is the Go/gin replacement for the Swoole PHP API at `/Users/canavs/xjProj/XJBackend/api`.

The harness now follows the `newjp` / `harness-skills` style: root agent instructions, `.codex` agent definitions, `harness/skills`, subagent playbooks, references, scripts, and task templates.

The harness starts with:

- A production-shaped HTTP entrypoint in `cmd/api`.
- Config through environment variables for container and Kubernetes use.
- Health probes at `/healthz` and `/readyz`.
- A legacy JSON response package matching PHP `kernel\Json`.
- Placeholder routes for the first PHP API migration targets.
- Unit tests that must pass before each interface migration is considered done.
- CI checks for tests, vet, build, and Docker image build.
- CD scaffold that publishes a container image to GHCR on `v*` tags.
- Kubernetes Deployment and Service manifests in `deploy/k8s`.
- Skills-driven execution through `AGENTS.md`, `.codex/agents/*.toml`, and `harness/skills/*/SKILL.md`.

## Go Project Structure

```text
cmd/api/
  main.go
internal/config/
  config.go
internal/server/
  router.go              # Gin engine, middleware, route wiring only
internal/handler/
  *.go                   # HTTP adapter layer
internal/service/
  <domain>/*.go          # business logic without Gin dependency
internal/domain/
  *.go                   # API/domain DTO/DO
internal/legacyjson/
  response.go            # PHP JSON shell compatibility
```

Future dependency layers:

```text
internal/repository/<domain>/   # MySQL/Redis persistence
internal/client/<provider>/     # external providers
```

Dependency direction:

```text
server -> handler -> service -> repository/client
handler -> legacyjson
service -> domain
```

## Skill Architecture

```text
AGENTS.md
.codex/
  config.toml
  agents/
    harness.toml
    architect.toml
    developer.toml
    tester.toml
    reviewer.toml
    ci-cd.toml
harness/
  AGENTS.md
  skills/
    subagent-orchestration/
      SKILL.md
    php-to-go-migration/
      SKILL.md
    php-route-discovery/
      SKILL.md
      references/discovery-checklist.md
    feature-delivery/
      SKILL.md
    requirement-analysis/
      SKILL.md
    code-standards/
      SKILL.md
    data-model-do/
      SKILL.md
    database-architecture/
      SKILL.md
    unit-testing/
      SKILL.md
    ci-containerization/
      SKILL.md
  subagents/
    main-orchestrator.md
    coder.md
    tester.md
    reviewer.md
    ci-cd.md
  templates/
    subagent-task.md
    operation-summary.md
  scripts/
    validate-skills.sh
```

Run `harness/scripts/validate-skills.sh` after editing skill files.

## Development Loop

For each migrated endpoint:

1. Read the matching PHP route and controller.
2. Write or update the Go handler contract test first.
3. Implement the Go handler.
4. Run `make ci`.
5. If tests fail, fix the code and rerun `make ci`.
6. Only then replace the placeholder route with the real handler.

## Commands

```sh
make test
make lint
make build
make ci
make run
make docker-build
```

## PHP Route Notes

The PHP project uses `kernel\Route`:

- Main API entrypoint: `/Users/canavs/xjProj/XJBackend/api/api.php`, listening on `0.0.0.0:18765`.
- `Route::group('/respond')` maps payment/provider callbacks.
- `Route::group('/v2')` maps newer API endpoints and uses `c.api.__init__` middleware.
- `Route::group('*')` maps original API endpoints and also uses `c.api.__init__`.
- Most routes use `Route::any`, so contract tests should include method compatibility where legacy clients depend on it.

The PHP JSON response shell is:

```json
{"retcode":0,"errmsg":"","data":{}}
```

Use `internal/legacyjson` for migrated API handlers unless an endpoint intentionally returns binary data, HTML, or provider-specific text.

The first Go placeholders are intentionally narrow:

- `/v2/register` -> `c.apiv2.user.register`
- `/v2/login` -> `c.apiv2.user.login`
- `/v2/forgot` -> `c.apiv2.user.forgot`
- `/v2/vod/*path` -> `c.apiv2.vod`

Suggested low-risk migration order:

1. `/captcha/req` -> `c.api.captcha.req`
2. `/sysavatar` -> `c.api.user.sysavatar`
3. `/game/platforms` -> `c.api.game.index.platforms`
4. `/game/categories` -> `c.api.game.index.categories`
5. `/v2/amazing/categories` -> `c.apiv2.amazing.categories`

## Dependency Inventory

Use local fakes or container services for contract/integration tests:

- MySQL configs are loaded from `conf.db`, `conf.db2`, and `conf.db3`; the PHP tree currently does not include `conf/db*.php`, so CI needs injected config and fixture data.
- Redis is required by middleware for maintenance flags such as `isclosed` and `closetips`.
- File cache defaults to `var/cache`.
- Local assets include `data/ipipfree.ipdb`, captcha fonts, and static avatar files.
- External services to stub: SMS, Telegram, payment gateways, game providers, image proxy/cover URL services, and AI/image processing endpoints.

## Subagent Workflow

Use `.codex/agents/harness.toml` as the main orchestrator. It selects bounded subagents while the main agent owns integration and final validation.

- `architect`: requirement analysis, DO, database/cache/migration/rollback design.
- `developer`: Go/Gin handler, helper, route, and compatibility implementation.
- `tester`: unit tests, handler tests, regression tests, and PHP-Go comparison evidence.
- `reviewer`: CR review, compatibility regressions, security risk, and test gaps.
- `ci-cd`: CI/CD, Docker, Kubernetes, rollout, rollback, and pipeline failures.

Rules:

- Each subagent gets a task card based on `harness/templates/subagent-task.md`.
- Agents must not revert changes made by others.
- PHP endpoint migration and public API behavior changes must include `developer`, `tester`, and `reviewer`.
- Main agent integrates results and owns final `make ci`.
- If real subagent tooling is unavailable, execute the same role checklist sequentially and say so in the final summary.

## Container Contract

The binary listens on `HTTP_HOST:HTTP_PORT`, defaulting to `0.0.0.0:8080`.

Kubernetes probes should call:

- liveness: `GET /healthz`
- readiness: `GET /readyz`
