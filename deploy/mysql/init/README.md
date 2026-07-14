# MySQL 初始化导入

把本地数据库 dump 放在这个目录，文件名使用：

```text
newxx.sql
```

首次启动 MySQL 且 `.docker/mysql/data` 为空时，官方 MySQL Docker entrypoint 会自动导入本目录下的 `.sql` 或 `.sql.gz` 文件。

如果已经启动过容器并生成了数据目录，需要先重置数据目录，或者使用手动导入命令。
