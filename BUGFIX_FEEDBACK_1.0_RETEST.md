# 接口测试反馈 1.0 修复复测报告

## 基本信息

- 反馈文件：`接口测试反馈_1.0.md`
- Go 服务：`xj-comp-api`
- 修复日期：2026-07-21
- 验证命令：`make ci`
- 验证结果：通过

## 本次已修复项

### 1. `/ucp/index` 会员剩余时间格式错误

问题：

```text
59天后22小时后0分钟后41秒后过期
```

修复：

- 修正 `internal/service/ucp/myaff.go` 中 `formatRemain` 的拼接规则。
- 现在只在最终文案中保留一个“后”，不在每个时间单位后重复添加。

复测接口：

```http
GET /ucp/index
```

复测重点：

```text
59天22小时0分钟41秒后过期
```

预期：

- `retcode=0`
- `data.user.dueday` 不再出现多个“后”

### 2. `/v2/vod/reqplay/76900` 非付费影片被错误提示购买

修复：

- 修正 `internal/service/vod/listing.go` 中 `isvip=2` 的购买判断。
- 现在只有 `view_price > 0` 且不在限免期时，才检查购买记录并提示金币购买。
- 免费影片或限免影片不会再进入购买拦截。
- 追加修复：如果当前登录用户已有 `allow.vod.vip` 权限，则 `isvip=2` 影片也不会进入购买拦截，按 VIP 权限继续走播放流程。

复测接口：

```http
GET /v2/vod/reqplay/76900
```

复测重点：

- VIP 可观看的非付费影片不应弹出金币购买提示。
- VIP 用户不应因为 `isvip=2` 被优先拦截到购买分支。
- 不应返回：

```json
{
  "retcode": 168,
  "errmsg": "此内容需要付费购买."
}
```

预期：

- 可播放时返回 `retcode=0`
- 响应中包含播放地址字段，例如 `data.httpurl` 或 `data.httpurls`

### 3. `/v2/vod/reqplay/68848` VIP 用户仍提示开通 VIP

修复：

- `vod` 播放权限判断现在会基于 `user_groups` 初始化用户 `perms`。
- 修复 VIP 用户因 `perms` 缺失被误判为无 `allow.vod.vip` 权限的问题。

复测接口：

```http
GET /v2/vod/reqplay/68848
```

复测重点：

- VIP 用户不应被提示开通 VIP。
- 不应返回：

```json
{
  "retcode": 5,
  "errmsg": "VIP独享内容，请升级"
}
```

预期：

- VIP 权限范围内返回 `retcode=0`
- 响应中包含播放地址字段，例如 `data.httpurl` 或 `data.httpurls`

### 4. `/ucp/feedback/create` 意见反馈提交失败

修复：

- 修正 `internal/repository/ucp/ucp.go` 中 `feedbacks` 插入字段。
- 兼容 MySQL 8 严格模式下 `url`、`contact`、`replytext` 为 `NOT NULL` 且无默认值的表结构。
- 插入反馈时显式写入空字符串，避免 DB 报错导致 `提交反馈失败`。

复测接口：

```http
POST /ucp/feedback/create
```

建议请求：

```json
{
  "cid": 4,
  "content": "接口复测反馈",
  "device": "h5"
}
```

复测重点：

- 不应返回：

```json
{
  "retcode": -1,
  "errmsg": "提交反馈失败"
}
```

预期：

```json
{
  "retcode": 0,
  "errmsg": "信息已反馈"
}
```

### 5. `/comment/post` 视频评论失败

修复：

- 修正 `internal/repository/comment/comment.go` 的 `UserGroups` 查询字段。
- 现在会读取 `perms`、`weight`、`scope`，让评论服务在用户没有内联 `perms` 时可以正确从用户组计算评论权限。

复测接口：

```http
POST /comment/post
```

建议请求：

```json
{
  "vodid": 76900,
  "content": "不错"
}
```

复测重点：

- 普通可评论账号不应因为权限字段缺失失败。
- 如果仍返回：

```json
{
  "retcode": 11,
  "errmsg": "账号异常，请联系管理员"
}
```

说明当前登录账号昵称仍触发旧 PHP 兼容规则：昵称中数字超过 5 个会被判定为账号异常。请换一个昵称正常的测试账号复测评论成功分支。

预期：

```json
{
  "retcode": 0,
  "errmsg": "发表成功"
}
```

## 本次未直接修复但已定位项

### 6. `/aiundress/upload` AI 脱衣上传接口失败

当前结论：

- 该接口不是小 bug，而是上传成功链路尚未完整迁移。
- 旧 PHP 链路包含：登录校验、本地上传、R2 上传、`ai_undress` 表新增或更新、返回 `data.file`。
- 当前 Go 仍返回：

```json
{
  "retcode": -1,
  "errmsg": "AI 上传成功分支暂未迁移"
}
```

复测结论：

- 本轮不作为已修复项复测。
- 需要单独迁移 AI 上传链路。

### 7. `/aiundress/resourceTypeList?module=1` 获取模板失败

当前结论：

- Go 逻辑已按旧 PHP 调用第三方接口。
- 如果服务器未配置以下环境变量，会返回 `请求失败`：

```text
AIUNDRESS_THIRD_HOST
AIUNDRESS_THIRD_KEY
```

复测方式：

1. 确认服务器 `.env` 或 systemd 环境里已配置上述两个值。
2. 重启服务。
3. 复测：

```http
GET /aiundress/resourceTypeList?module=1
```

预期：

- 配置正确且第三方正常时返回 `retcode=0`
- `data.data` 中包含模板分类数据

### 8. `/index` 客户端无法渲染视频封面

当前结论：

- Go 端会返回封面字段，且 `coverpic` 生成逻辑基本对齐旧 PHP `getCoverpic`。
- 需要拿一条 PHP 与 Go 的实际响应样例对比 `coverpic`、`coverx`、`play_url`、`preview_url` 等字段后精确修复。

复测建议：

- 抓取 `/index` 响应中任意一个无法渲染的视频 row。
- 同时提供 PHP 老接口和 Go 新接口的同一 row JSON。

### 9. `/v2/vod/listing-0-0-0-0-0-0-0-0-0-1` 分类页未渲染

当前结论：

- Go 接口已有数据返回，问题更可能是字段形态兼容。
- 需要对比 `vodrows[0]` 的字段。

复测建议：

- 提供 PHP/Go 两边 `vodrows[0]`。
- 重点对比：

```text
vodid
title
coverpic
coverx
play_url
down_url
preview_url
need_buy
isvip
view_price
```

### 10. `/vod/latest-4-0-0-0-0-0-0-0-0-1` 首页栏目无法渲染

当前结论：

- 查询参数解析和分类过滤逻辑已有覆盖。
- 本轮未直接修改列表字段形态。

复测建议：

- 提供 PHP/Go 两边 `vodrows[0]` 对比。
- 如果确认是 `play_url` 或 `preview_url` 需要绝对 URL，再按旧 PHP `R_URL` 规则修复。

## 本次修改文件

```text
internal/repository/comment/comment.go
internal/repository/ucp/ucp.go
internal/service/comment/listing_test.go
internal/service/ucp/myaff.go
internal/service/ucp/myaff_test.go
internal/service/vod/listing.go
internal/service/vod/listing_test.go
```

## 测试结果

已执行：

```shell
make ci
```

结果：

```text
go test ./... 通过
go vet ./... 通过
go build ./cmd/api 通过
```

## 复测优先级

建议先复测：

1. `/ucp/index`
2. `/v2/vod/reqplay/76900`
3. `/v2/vod/reqplay/68848`
4. `/ucp/feedback/create`
5. `/comment/post`

如果 2、3 仍异常，请补充对应响应 JSON 和当前测试账号的 VIP 信息。

如果 8、9、10 仍异常，请补充 PHP/Go 的 `vodrows[0]` 对比样例。
