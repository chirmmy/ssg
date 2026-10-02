---
title: 单服务器多网站部署
description: Docker + Nginx 反向代理实践
pubDate: 2026-09-26
updated: 2026-10-02
tags: [docker, nginx]
draft: false
cover: https://qiniu.anburger.site/cover/Tech.webp
---

# 一、场景

对于个人开发者，往往只拥有少量资源，例如仅有一台服务器、一个域名(`example.com`)，希望部署多个独立的小网站：如个人网盘、静态博客、工具站等。要求各站点隔离、互不干扰，统一走 HTTPS，且尽量自动化运维。

在这种场景下，`Docker + Nginx 反向代理 + 泛域名证书` 的架构比较灵活，所有网站跑在独立的 Docker 容器里，通过一个共享的 proxy-net 网络与入口 Nginx 通信，外部只暴露 80/443 端口

```text
                用户浏览器
                    │
                    ▼
┌─────────────────────────────────────────┐
│ 云服务器                                 │
│                                         │
│  ┌─────────────────────────────────┐    │
│  │ nginx-proxy (入口 Nginx 容器)    │    │
│  │ 监听 80/443，按域名分发           │    │
│  └────────────┬────────────────────┘    │
│               │ proxy-net (Docker 网络)  │
│     ┌─────────┼─────────┬──────────┐    │
│     ▼         ▼         ▼          ▼    │
│  openlist   static-   blog-     ...     │
│  -app       web       app               │
│  (网盘)     (静态站)  (博客)              │
└─────────────────────────────────────────┘
```

# 二、搭建步骤

## 2.1 创建共享网络
```bash
docker network create proxy-net
```

## 2.2 入口 Nginx 部署

1. 按照目录结构创建基础配置
```text
/opt/nginx/
├── docker-compose.yml
└── conf.d/          # 各站点配置文件
```

2. 编写入口nginx的`docker-compose.yml`

```yml
services:
  nginx:
    image: nginx:alpine
    container_name: nginx-proxy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./conf.d:/etc/nginx/conf.d:ro
      - ./ssl:/etc/nginx/ssl:ro
    networks:
      - proxy-net

networks:
  proxy-net:
    external: true
```

3. 启动入口Nginx
```bash
cd /opt/nginx

sudo docker compose up -d
```

该入口Nginx的作用：把服务器的 80 和 443 端口交给这个容器，外部所有 HTTP/HTTPS 请求都先到这个 Nginx

## 2.3 配置 

2.1~2.2实现了`Users->Server`统一由`nginx-proxy`容器处理，接下来要实现将对应的请求转发到具体的独立网站，即配置**反向代理规则**（例如 pan.example.com → 某个后端容器）

在此之前，我们需要预先进行SSL证书配置，以实现HTTPS协议。SSL证书申请详见：

{{<postLinkCard path="ssl-self" cover="https://qiniu.anburger.site/cover/Tech.webp">}}

**配置一张 `*.example.com` 泛域名证书**
确保生成的 `privkey.pem` 和 `fullchain.pem` 在 `/opt/nginx/ssl/example.com/`下


## 2.4 独立网站部署

以部署两个独立网站为例：**个人网盘(`pan.example.com`)** + **静态主页(`home.example.com`)**

### 2.4.1 个人网盘

1. 创建OpenList容器
```bash
vim /opt/openlist/docker-compose.yml
```
```yml
services:
  openlist:
    image: openlistteam/openlist:latest
    container_name: openlist-app
    restart: unless-stopped
    user: '0:0'
    volumes:
      - ./data:/opt/openlist/data
      - /root/local/video:/mnt/local
    environment:
      - TZ=Asia/Shanghai
      - UMASK=022
    networks:
      - proxy-net

networks:
  proxy-net:
    external: true
```
启动容器：
```bash
cd /opt/openlist/
sudo docker copose up -d
```

2. 新增反向代理规则配置
```
server {
    listen 443 ssl;
    http2 on;
    server_name pan.example.com;

    ssl_certificate     /etc/nginx/ssl/example.com/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/example.com/privkey.pem;

    location / {
        proxy_pass http://openlist-app:5244;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 下面这几行对 OpenList 的大文件上传/下载很重要
        proxy_set_header Range $http_range;
        proxy_set_header If-Range $http_if_range;
        client_max_body_size 20000m;
    }
}

server {
    listen 80;
    server_name pan.example.com;
    return 301 https://$host$request_uri;
}
```
效果：`https://pan.example.com -> http://openlist-app:5244 (openlist容器)`

### 2.4.2 静态主页部署

方案：**独立 Nginx 容器托管静态文件**

1. 静态文件准备
上传`dist/`到`/opt/static/homepage/dist/`

2. 创建容器
`vim /opt/static/homepage/docker-compose.yml` 
```yml
services:
  static-web:
    image: nginx:alpine
    container_name: homepage
    restart: unless-stopped
    volumes:
      - ./dist:/usr/share/nginx/html:ro
    networks:
      - proxy-net

networks:
  proxy-net:
    external: true
```
`sudo docker compose up -d`

3. 在 /opt/nginx/conf.d/homepage.conf 添加反代：
```
server {
    listen 443 ssl;
    http2 on;
    server_name home.example.com;

    ssl_certificate     /etc/nginx/ssl/example.com/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/example.com/privkey.pem;

    location / {
        proxy_pass http://homepage:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name home.example.com;
    return 301 https://$host$request_uri;
}
```

效果：`https://home.example.com -> http://homepage:80 (静态文件托管容器)`


## 2.5 DNS解析

在域名托管平台(如Cloudflare)为每个子域名添加 A 记录，指向服务器公网 IP：
| 类型 | 名称 | 内容 |
|-----|----|-------|
|  A  | pan  | 服务器 IP |
|  A  | home | 服务器 IP |

# 三、总结

从全局来看，入口Nginx决定了整理架构
```text
/opt/nginx/
├── docker-compose.yml
├── conf.d/   # 各站点配置文件，一个配置文件对应一个网站
│   ├── openlist.conf    # -> pan.example.com
│   └── homepage.conf    # -> home.example.com
└── ssl/      # SSL证书
    └── example.com/
        ├── fullchain.pem
        └── privkey.pem
```

核心原则：

- **每个网站一个容器**，独立运行、独立更新，互不影响。

- **只有 Nginx 容器映射端口到宿主机**，其他容器不暴露端口，只加入共享 proxy-net 网络。

- **入口 Nginx 按 `server_name` 匹配域名**，将请求转发给对应容器。

- **一张泛域名证书** *.example.com 覆盖所有子域名，新站无需重新申请证书。