---
name: ci-containerization
description: >-
  维护 xj_comp 的 CI、Docker、Kubernetes、镜像发布和部署就绪脚手架。使用场景：Dockerfile、GitHub Actions、Harness pipeline、K8s manifest、探针、灰度、回滚、CI 失败排查。触发词：CI/CD、Docker、Kubernetes、部署、流水线。
metadata:
  author: xj_comp
  version: 1.0.0
  mcp-server: optional-ci
license: MIT
compatibility: Go 1.23、Docker、GitHub Actions、Kubernetes、Harness CI/CD。
---

# CI Containerization

## Instructions

1. 识别构建契约：
   - Go module: `go.mod`。
   - binary entrypoint: `./cmd/api`。
   - local validation: `make ci`。
2. 保持容器行为：
   - `HTTP_HOST:HTTP_PORT`。
   - 默认 `0.0.0.0:8080`。
   - probes: `/healthz`、`/readyz`。
3. CI 至少运行：
   - `go test ./...`
   - `go vet ./...`
   - `go build ./cmd/api`
   - Docker build validation。
4. CD 应从 tag 发布不可变镜像。
5. 生成 Harness pipeline 时，缺少 org/project/connector/cluster/registry 信息必须询问，不写假值。

## Examples

```text
/ci-containerization 检查 Dockerfile 和 K8s readiness 是否满足当前服务
```

## Performance Notes

- 优先复用 `Makefile` 命令作为本地与 CI 的共同契约。
- runtime image 使用非 root。
- Docker 本地失败时区分 daemon/buildx 环境问题和代码问题。

## Troubleshooting

- Docker daemon 未启动时，记录环境问题，不误判代码失败。
- 部署涉及生产时，必须给出灰度、监控和回滚条件。
