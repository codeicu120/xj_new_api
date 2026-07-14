# 本地 MySQL Docker 环境

本配置只用于本地开发和导入旧库数据，不要把这些密码用于生产。

## 默认连接信息

- Host: `127.0.0.1`
- Port: `3306`
- Root user: `root`
- Root password: `xj_root_123456`
- App database: `xj_comp`
- App user: `xj_app`
- App password: `xj_app_123456`

## 启动

方式一：使用 compose。

```sh
docker compose -f docker-compose.mysql.yml --env-file .env.mysql.example up -d --build
```

如果你的本机没有 Docker Compose v2 插件，可以直接用 Dockerfile 启动。

方式二：只使用 `docker build` 和 `docker run`。

```sh
docker build -f deploy/mysql/Dockerfile -t xj-comp-mysql:8-local .
mkdir -p .docker/mysql/data
docker run -d \
  --name xj-comp-mysql \
  --restart unless-stopped \
  -p 3306:3306 \
  -v "$PWD/.docker/mysql/data:/var/lib/mysql" \
  -v "$PWD/deploy/mysql/init:/docker-entrypoint-initdb.d:ro" \
  xj-comp-mysql:8-local
```

查看状态：

```sh
docker compose -f docker-compose.mysql.yml ps
docker logs -f xj-comp-mysql
```

## 导入数据库

如果是第一次启动，把数据库 dump 放到：

```text
deploy/mysql/init/newxx.sql
```

MySQL 容器首次初始化空数据目录时会自动导入 `deploy/mysql/init/` 下的 `.sql` 或 `.sql.gz` 文件。因此只要 `.docker/mysql/data` 还不存在，启动容器时会自动导入 `newxx.sql`。

如果容器已经启动并已有数据目录，使用手动导入：

```sh
docker exec -i xj-comp-mysql mysql -uroot -pxj_root_123456 xj_comp < /path/to/dump.sql
```

导入 gzip：

```sh
gunzip -c /path/to/dump.sql.gz | docker exec -i xj-comp-mysql mysql -uroot -pxj_root_123456 xj_comp
```

## 重置本地数据库

这会删除本地 MySQL 数据：

```sh
docker compose -f docker-compose.mysql.yml down
rm -rf .docker/mysql/data
docker compose -f docker-compose.mysql.yml --env-file .env.mysql.example up -d --build
```

如果用 `docker run` 启动：

```sh
docker rm -f xj-comp-mysql
rm -rf .docker/mysql/data
docker run -d \
  --name xj-comp-mysql \
  --restart unless-stopped \
  -p 3306:3306 \
  -v "$PWD/.docker/mysql/data:/var/lib/mysql" \
  -v "$PWD/deploy/mysql/init:/docker-entrypoint-initdb.d:ro" \
  xj-comp-mysql:8-local
```

## 连接测试

```sh
docker exec -it xj-comp-mysql mysql -uroot -pxj_root_123456 -e "SELECT VERSION();"
```
