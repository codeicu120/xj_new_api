---
name: database-architecture
description: >-
  设计 MySQL、Redis、文件缓存、fixture、迁移和回滚方案。使用场景：接口依赖数据库、Redis key、缓存兼容、索引、测试 fixture、数据迁移。触发词：数据库、Redis、缓存、fixture、迁移、回滚。
metadata:
  author: xj_comp
  version: 1.0.0
  mcp-server: optional-local-files
license: MIT
compatibility: PHP 旧项目、Go/Gin 服务、Docker/Kubernetes 环境。
---

# Database Architecture

## Instructions

1. 识别 PHP 依赖：`init.db`、`init.redis`、`init.cache`、相关模型和配置。
2. 标注读写路径、事务需求、缓存 key 和 TTL。
3. 测试中优先使用 fake、fixture 或本地容器，不连接生产。
4. 设计迁移和回滚：
   - schema 变更。
   - 数据回填。
   - 灰度读写。
   - 回滚条件。
5. 输出依赖清单和风险。

## Examples

```text
/database-architecture 分析 /game/categories 需要哪些表和 Redis key
```

## Performance Notes

- 不设计请求链路中的全表扫描。
- 不污染 PHP 仍在使用的 Redis serialize 缓存。

## Troubleshooting

- 缺失数据库配置时，记录为运行时依赖。
- 无法确定表结构时，先让主线确认 fixture 或 schema 来源。
