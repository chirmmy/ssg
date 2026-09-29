---
title: Vercel Serverless笔记
description: 介绍Vercel Serverless的概念和基本用法
pubDate: 2025-08-01T21:00:00+08:00
updated: 2025-07-21T21:00:00+08:00
tags: [笔记]
draft: false
cover: https://qiniu.anburger.site/cover/Note.webp
---
## 1 环境准备

:one:nodejs

https://nodejs.org/

:two:注册Vercel

https://vercel.com/

:three:Vercel CLI

```shell
npm i -g vercel
```

:four:命令行登录vercel

```shell
vercel login
```

![image-20250801220752777](https://qiniu.anburger.site/post/image-20250801220752777.png)

## 2 Quick Start

> **云函数 (Cloud Functions)**：在 Vercel 平台上，这通常被称为 **Serverless Functions (无服务器函数)**。
>
> Vercel 的核心理念是**约定优于配置**。你不需要复杂的配置文件，只需要将代码放在正确的目录下，Vercel 就会自动将其部署为 Serverless Function。

核心理念**约定优于配置**的加持下，编码的核心便是：**使用 `api/` 目录**

1. **创建项目**: 你的项目可以是一个 Next.js 应用，也可以是一个静态网站，甚至是空项目。
2. **创建 `api` 目录**: 在你的项目根目录下，创建一个名为 `api` 的文件夹。
3. **编写函数代码**: 在 `api` 目录下创建一个 JavaScript 或 TypeScript 文件。文件名将成为 API 的路径。

:one:新建空项目`first-serverless` :arrow_right: 项目根目录下新建文件夹`api` :arrow_right: `New hello.js`

```js
// 标准 Nodejs HTTP 处理函数
export default function handler(request, response) {
    const { name } = request.query;
    response.status(200).json({
        message: `Hello, ${name || 'World'}! This function is running on the Vercel cloud.`,
    });
}
```

:two:部署

```shell
vercel --prod
```

![image-20250801221517995](https://qiniu.anburger.site/post/image-20250801221517995.png)

执行命令后：

- Vercel 会自动检测到 `api/hello.js` 文件。
- 它会将这个文件打包成一个独立的 Serverless Function。
- 它会为你生成一个可访问的 URL 端点：`https://<你的域名>.vercel.app/api/hello`。
- 当你访问 `.../api/hello` 时，这个函数就会在云端被触发执行，例如`.../api/hello`、`.../api/hello?name=anby`。

:triangular_flag_on_post:部署成功后得到一个域名，国内访问需要绑定自己的域名，加上后缀`api/hello`，成功，Congratulations​!:tada:

![image-20250801221800843](https://qiniu.anburger.site/post/image-20250801221800843.png)

![image-20250801222424603](https://qiniu.anburger.site/post/image-20250801222424603.png)

## 3 边缘函数

Vercel中编写`Edge Function`只需要在`api`中指定一个配置项即可：

```js
export const config = {
  runtime: 'edge',
};
```

## 4 案例

场景：七牛云对象存储位于海外（亚太-新加坡），通过`Edge Function`中转七牛云资源请求进行加速。

```js
// /api/proxy.js
export const config = {
  runtime: 'edge',
};

export default async function handler(request) {
  // 从原始请求中获取路径
  const { pathname, search } = new URL(request.url);

  if (pathname === '/') {
    return new Response('Qiniu Proxy is Running!', {
      status: 200,
      headers: { 'Content-Type': 'text/html' }
    });
  }

  // 构建指向七牛云存储桶的目标 URL
  const destinationUrl = `https://your.qiniu.domain${pathname}${search}`;

  // 从七牛云获取响应
  // 我们将原始请求的 headers, method, 和 body 都转发过去
  const qiniuResponse = await fetch(destinationUrl, {
    headers: request.headers,
    method: request.method,
    body: request.body,
    redirect: 'follow',
  });

  // 将七牛云的响应直接返回给客户端
  return new Response(qiniuResponse.body, {
    status: qiniuResponse.status,
    statusText: qiniuResponse.statusText,
    headers: qiniuResponse.headers,
  });
}
```

