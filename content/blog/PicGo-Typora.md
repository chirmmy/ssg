---
title: PicGo实现Typora插入图片上传服务
description: 一种PicGo自定义工具，实现图片插入后上传对象存储
pubDate: 2025-07-22T12:00:00+08:00
updated: 2025-07-22T12:00:00+08:00
tags: [PicGo, Typora, 自动化工具]
draft: false
cover: https://qiniu.anburger.site/cover/Tech.webp
---


现在已经有非常多的快速搭建**博客站**的工具，例如`JeckII、Hexo、Hugo`等，它们都具有非常优秀的使用体验，并且可以通过`GitHub`、`CloudFlare`等工具进行免费部署。利用它们成功搭建出自己的博客站后，往往使用`Markdown`进行博客内容撰写。在撰写过程中，我们会发现处理图片时会面临一些窘境：

- 需要将图片放到特定资源文件夹（进行上传），操作不方便
- 图片过多后占用空间
- 图片太大时加载速度慢等

我们可以采用将图片资源存放到图床的方案解决图片存储、路径问题。此外，在写作过程中通过安装`PicGo工具+压缩等插件`简化撰写过程，接下来跟我一起让写博客变得更加优雅吧！

## 1. 安装

**PicGo: 一个用于快速上传图片并获取图片 URL 链接的工具**，官方文档：[PicGo](https://picgo.github.io/PicGo-Doc/)有详细的使用教程。

1. 下载：[PicGo-release](https://github.com/Molunerfinn/PicGo/releases)，`PicGo`提供众多`beta`版本，若希望体验新功能可安装最新`beta`版本。推荐安装正式版本。下载对应系统安装包进行安装，这里以`windows`系统为例，下载`2.3.1`版本安装程序：

![image-20250722151412072](https://qiniu.anburger.site/post/image-20250722151412072.png)

2. 运行安装程序进行安装，安装成功后开始使用`PicGo`：

![image-20250722151741990](https://qiniu.anburger.site/post/image-20250722151741990.png)

## 2. 配置图床

`PicGo` 本体支持如下图床：

- `七牛图床` v1.0
- `腾讯云 COS v4\v5 版本` v1.1 & v1.5.0
- `又拍云` v1.2.0
- `GitHub` v1.5.0
- `SM.MS V2` v2.3.0-beta.0
- `阿里云 OSS` v1.6.0
- `Imgur` v1.6.0

本文使用**七牛云**对象存储。

### 2.1 七牛云对象存储申请

1. 进入[七牛云 | 一站式中立音视频云 + AI](https://www.qiniu.com/)，完成账号设置（注册、登录、认证等）。
2. 完成后进入控制台（入口在网页右上角），选择新建存储空间：

![image-20250722152658046](https://qiniu.anburger.site/post/image-20250722152658046.png)

3. 在新建存储空间面板中填写信息：
   - 存储空间名称：自定义
   - 存储区域：如果你拥有已备案的域名或者仅做测试使用，选择国内区域。如果你的域名无法备案，可选择海外区域（不需要备案，但访问可能需要梯子）
   - 访问控制：公开（更方便）

​    填写完成后点击确认。

![image-20250722153029092](https://qiniu.anburger.site/post/image-20250722153029092.png)

4. 绑定域名

   存储空间创建后，会分配一个测试域名，可以直接使用，但是会被回收，因此不推荐实际使用。

   如果没有域名，可以注册免费域名，详见：。

   拥有一个域名后，进入新建的存储空间进行域名管理：

   ![image-20250722170800583](https://qiniu.anburger.site/post/image-20250722170800583.png)

   因为CDN加速域名需要进行域名所有权认证，所以我这里选择`自定义源站域名`，点击绑定域名后，填写：

   

   从绑定的域名列表中，复制`CNAME`

   ![image-20250722171225178](https://qiniu.anburger.site/post/image-20250722171225178.png)

   然后使用`CloudFlare`配置CNAME并开启DNS减速服务(:rage:)，

5. 





### 2.2 配置PicGo图床

进入`PicGo->图床设置->七牛云`，填写配置信息：

- `AccessKey`和`SecretKey`：进入`七牛云->个人中心->密匙管理`，开启密匙访问并将对应的`Key`填入

  ![image-20250722164135467](https://qiniu.anburger.site/post/image-20250722164135467.png)

- `Bucket`：存储空间名称（如2.1中创建的`anburgerblog`）

- 访问网址

- 存储区域

- 存储路径

![image-20250722153925421](https://qiniu.anburger.site/post/image-20250722153925421.png)

设置完成后点击**确定**进行保存，同时可以点击**设为默认图床**将七牛云作为默认图床。

在上传区进行测试。

## 3. 配置Typora

在Typora中，进入`文件->偏好设置`（`Windows`快捷键`Ctrl+逗号`），进入`图片`进行设置：

- 插入图片时...：看个人需求设置，这里设置为无特殊操作。如果希望插入后直接上传到图床，选择`上传图片`。
- 上传服务：选择`PicGo（app）`，然后将`PicGo路径`设置为你对应的安装路径。

![image-20250722154539727](https://qiniu.anburger.site/post/image-20250722154539727.png)

现在就完成了所有配置，可以开始使用啦！

示例：

1. 复制图像（截图、Ctrl+C）
2. 插入图像到对应的位置（Ctrl+V），此时，图片还是本地图片：

![image-20250722155616421](https://qiniu.anburger.site/post/image-20250722155616421.png)

3. 接下来`右击图片->上传图片`：

![image-20250722155948431](https://qiniu.anburger.site/post/image-20250722155948431.png)

4. 上传完成后会自动将图片替换为图床中的链接：

![image-20250722160209444](https://qiniu.anburger.site/post/image-20250722160209444.png)
