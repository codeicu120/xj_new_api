---
name: data-model-do
description: >-
  设计接口 request/response DO、业务 domain DO、database DO、cache DO 和转换规则。使用场景：字段映射、响应结构、数据库对象、缓存对象、PHP-Go 类型兼容。触发词：DO、DTO、字段映射、数据模型。
metadata:
  author: xj_comp
  version: 1.0.0
  mcp-server: optional-local-files
license: MIT
compatibility: Go/Gin、PHP 迁移和数据库建模。
---

# Data Model DO

## Instructions

1. 区分 API DO、domain DO、database DO、cache DO。
2. 对 PHP 兼容字段，记录字段名、类型、空值和默认值。
3. 对列表响应，明确 `rows`、分页、排序和总数字段。
4. 对时间、金额、ID、布尔值，确认 PHP 是否用字符串数字。
5. 转换规则写入测试。

## Examples

```text
/data-model-do 设计 /v2/amazing/categories 的响应 DO
```

## Performance Notes

- 不让数据库对象直接泄漏到 API。
- 保持 PHP 兼容优先于理想化重命名。

## Troubleshooting

- 如果字段来源不明，回到 PHP controller 和 model 查。
- 如果空值行为不明，用 contract test 固定。
