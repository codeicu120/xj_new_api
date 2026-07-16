# UCP taskboxopen edge migration draft

Scope: `/ucp/taskbox/taskboxopen` currently migrates only PHP-compatible pre-success failure branches.

Implemented failure branches:

- Login required: `retcode=-9999`, `errmsg=您还没有登录`.
- Missing, disabled, or non-public task box: `retcode=-1`, `errmsg=任务不存在或已停用`.
- Zero coin task box: `retcode=-1`, `errmsg=宝箱赠送金币为0`.
- Daily mystery box `taskid=1022`: not started, ended, and already claimed messages.
- Weekly mystery box `taskid=1622`: non-Saturday, not started, ended, and already claimed messages.
- Promotion task boxes: already claimed and insufficient recommendation count messages.

Intentionally not migrated yet:

- Writing `promotion_taskboxlogs`.
- Calling coin balance mutation through `coinlog/change`.
- Returning successful `taskdone` data or `宝箱成功开启`.

Current success-precheck placeholder remains `retcode=-1`, `errmsg=任务宝箱开启成功分支暂未迁移`.
