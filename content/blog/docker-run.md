---
title: Docker常用容器部署
date: 2025-07-20T12:00:00+08:00
lastmod: 2025-07-20T12:00:00+08:00
cover: https://qiniu.anburger.site/cover/Blog.webp
---
------

nginx

```shell
docker run -d \
--name nginx \
-p 80:80 -p 443:443 \
-v ~/docker/nginx/conf/nginx.conf:/etc/nginx/nginx.conf \
-v ~/docker/nginx/conf/default.conf:/etc/nginx/conf.d/default.conf \
-v ~/docker/nginx/html:/usr/share/nginx/html \
-v ~/docker/nginx/ssl:/etc/nginx/ssl \
nginx

# 复制初始化后的nginx.conf、default.conf
```

------

mysql

```shell
docker run -d \
--name mysql \
-p 3306:3306 \
-e TZ=Asia/Shanghai \
-e MYSQL_ROOT_PASSWORD=root \
-v /root/mysql/data:/var/lib/mysql \
-v /root/mysql/conf:/etc/mysql/conf.d \
mysql
```

------

redis

```shell
docker run -d \
--name redis \
-p 6379:6379 \
-v /usr/local/src/docker/redis/data:/data \
-v /usr/local/src/docker/redis/conf/redis.conf:/etc/redis/redis.conf \
redis redis-server /etc/redis/redis.conf

# 配置文件/etc/redis/redis.conf
```

