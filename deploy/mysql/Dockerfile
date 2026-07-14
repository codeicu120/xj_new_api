FROM mysql:8.0

ENV MYSQL_ROOT_PASSWORD=xj_root_123456 \
    MYSQL_DATABASE=xj_comp \
    MYSQL_USER=xj_app \
    MYSQL_PASSWORD=xj_app_123456 \
    TZ=Asia/Shanghai

COPY my.cnf /etc/mysql/conf.d/xj-comp.cnf

EXPOSE 3306
