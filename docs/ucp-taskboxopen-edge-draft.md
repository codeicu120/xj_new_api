# UCP taskboxopen migration note

Scope: `/ucp/taskbox/taskboxopen` is now migrated through the PHP-compatible success branch.

Migrated branches:

- Login required: `retcode=-9999`, `errmsg=您还没有登录`.
- Missing, disabled, or non-public task box: `retcode=-1`, `errmsg=任务不存在或已停用`.
- Zero coin task box: `retcode=-1`, `errmsg=宝箱赠送金币为0`.
- Daily mystery box `taskid=1022`: not started, ended, and already claimed messages.
- Weekly mystery box `taskid=1622`: non-Saturday, not started, ended, and already claimed messages.
- Promotion task boxes: already claimed and insufficient recommendation count messages.
- Success: writes `promotion_taskboxlogs`, updates `users_quota.goldcoin`, inserts `user_coinlogs(cointype=19)`, and returns `errmsg=宝箱成功开启` with `data.taskdone`.

Transaction boundary:

- The repository locks `promotion_taskboxlogs(uid,taskid,daykey)` before insert to preserve duplicate-claim behavior.
- The same transaction locks `users_quota` before increasing `goldcoin`.
- Coin log remark remains `宝箱开启收入[taskid]`.
