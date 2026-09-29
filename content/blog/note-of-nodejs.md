---
title: Nodejs笔记
description: Nodejs学习笔记
pubDate: 2025-08-07T19:00:00+08:00
updated: 2025-08-07T19:00:00+08:00
tags: [笔记, Nodejs]
draft: false
cover: https://qiniu.anburger.site/cover/Note.webp
---
# Nodejs学习笔记

## 1 简介

官方文档：

{{<externalLinkCard title="Nodejs官方文档" link="https://nodejs.org/docs/latest/api/" cover="auto">}}

> Node.js is a JavaScript runtime built on the [V8 JavaScript engine](https://v8.dev/).

下载&安装：

{{<externalLinkCard title="Nodejs下载" link="https://nodejs.org/en/download" cover="auto">}}

## 2 Quick Start

:one:新建空项目`first-node`

:two:项目下新建文件`hello.js`

```js
const http = require('node:http')

const hostname = 'localhost';
const port = 8080;

const server = http.createServer((req, res) => {
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    res.end('Hello, World!\n');  
});

server.listen(port, hostname, () => {
    console.log(`Server running at http://${hostname}:${port}`);
    
})
```

:three:运行

```shell
node hello.js
```

运行结果如下，浏览器访问该url可以得到返回结果`Hello, World!`：

<div style="display: flex;">
    <img src="https://qiniu.anburger.site/post/image-20250807191737364.png" alt="image-20250807191737364" />
    <img src="https://qiniu.anburger.site/post/image-20250807191905346.png" alt="image-20250807191905346" />
</div>

:tada:Congratulations！



## X 数据库操作

### X.1 SQLite

