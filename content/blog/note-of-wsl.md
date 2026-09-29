---
title: WSL笔记
description: 介绍WSL的概念和基本用法
pubDate: 2025-08-03T21:00:00+08:00
updated: 2025-08-03T21:00:00+08:00
tags: [笔记, WSL, Windows]
draft: false
cover: https://qiniu.anburger.site/cover/Note.webp
---
## 1 安装

```shell
# 默认安装Unbuntu
wsl --install
```

```shell
# 安装其他分发版

## 查看其他分发版
wsl --list --online

## 安装指定分发版
wsl --install -d Debian
```

## 2 备份、卸载

```shell
# 备份
wsl --export Debian debian.tar
```

```shell
# 卸载
wsl --unregister Debian
```

```shell
# 从备份导入
wsl --import Debian D:/wsl .\debian.tar
```

```shell
# 关闭wsl子系统
wsl --shutdown
```

## 3 Docker安装

```shell
# 下载并执行Docker官方安装脚本
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh    # 执行后会推荐下载Docker Desktop版，需等待20s

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker
```

## 4 配置

在用户目录下`C:\Users\username`新建全局配置文件`.wslconfig`

### 4.1 网络模式

```
[wsl2]
networkingMode=mirrored
```

