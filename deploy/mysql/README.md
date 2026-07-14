连接命令 

```http request
docker exec -it xj-comp-mysql mysql -uroot -pxj_root_123456 xj_comp
```

初始化启动
```http request
docker build -f Dockerfile -t xj-comp-mysql:8-local ../..

mkdir -p ../../.docker/mysql/data

docker run -d \
  --name xj-comp-mysql \
  --restart unless-stopped \
  -p 3306:3306 \
  -v "/Users/canavs/xjProj/xj_comp/.docker/mysql/data:/var/lib/mysql" \
  -v "/Users/canavs/xjProj/xj_comp/deploy/mysql/init:/docker-entrypoint-initdb.d:ro" \
  xj-comp-mysql:8-local
```