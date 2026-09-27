---
title: Webhook实现Github自动推送到Linux自动部署
description: "Usage of Github Wehook"
pubDate: 2025-07-23
Updated: 2025-07-23
tags: [自动化工具, Github]
draft: false
---

:sunny: 场景：`Github`中代码更新后自动推送到服务器并部署

:wrench: 工具：`Webhook`​

## 1. Github配置Webhook

:round_pushpin:例如我在[anbydemara.github.io](https://github.com/anbydemara/anbydemara.github.io)该仓库中存储了我的静态博客网站，我希望当我`push`到该仓库时，`Github`自动通知我的服务器并执行自定义指令。

:one: ​进入`Github`仓库，`Settings->Webhooks->Add webhook`

![image-20250723111230291](https://qiniu.anburger.site/post/image-20250723111230291.png)

:two: 填写配置信息

- `PayLoad URL`：`http://your_servier_ip:port/path/{id}`，其中`port:9000，path:hooks`都是默认值，我们保存不变。修改修改为自己的ip地址，并且为该`Payload`自定义一个`id`

- `Content type`：选择`application/json`格式

- `Secret`：设置自己的`Secret token`用于验证，可以使用一个随机字符串

- `webhook`触发条件：`Just push event`表示push代码到仓库时触发，可以自定义其他事件

  其他保持默认，最后`Update webhook`：

![image-20250723111759923](https://qiniu.anburger.site/post/image-20250723111759923.png)

## 2. 服务器创建hook

:round_pushpin:我们使用[webhook](https://github.com/adnanh/webhook)工具：

> [webhook](https://github.com/adnanh/webhook) is a lightweight configurable tool written in Go, that allows you to easily create HTTP endpoints (hooks) on your server, which you can use to execute configured commands. You can also pass data from the HTTP request (such as headers, payload or query variables) to your commands. [webhook](https://github.com/adnanh/webhook) also allows you to specify rules which have to be satisfied in order for the hook to be triggered.

### 2.1 安装

See：[webhook/README.md at master · adnanh/webhook](https://github.com/adnanh/webhook/blob/master/README.md)

我的安装目录为`/usr/local/src/webhook/`，安装成功后，该文件夹下有一个名为`webhook`的文件，我们为其添加执行权限`chmod +x webhook`

### 2.2 配置hook

:one: 新建`hook.json`（自定义路径，这里存放在webhook安装目录下新建的`my-scripts`文件夹中）

该文件配置了一些规则，主要用于请求校验。在`Github`中我们配置了仓库`push`事件时，将使用`webhook`功能发送一个`POST`请求。在服务器上安装了`webhook`后，就可以监听对应的端口处理该请求。为了确保接受到的请求是`Github`触发指定事件时发送的，需进行校验：

- `id`：需要与`Github`中配置`PayLoad URL`时指定的`id`一致
- `execute-command`：自定义执行脚本文件路径，请求校验成功后要执行的脚本，这里我在`/usr/local/src/webhook/my-scripts`新建了一个`rebuildblog.sh`脚本
- `secret`：需与`Github`中配置的一致

```json
[
  {
    "id": "anburger-blog-rebuild",
    "execute-command": "/usr/local/src/webhook/my-scripts/rebuildblog.sh",
    "trigger-rule": {
      
      "and": [
        {
          "match": {
            "type": "payload-hmac-sha1",
            "secret": "yoursecret",  
            "parameter": {
              "source": "header",
              "name": "X-Hub-Signature"
            }
          }
        },
        {
          "match": {
            "type": "value",
            "value": "refs/heads/main",
            "parameter": {
              "source": "payload",
              "name": "ref"
            }
          }
        }
      ]
    }
  }
]
```

:two: 新建执行脚本`rebuildblog.sh`（自定义路径，这里存放在webhook安装目录下新建的`my-scripts`文件夹中）

通过上一步`hook.json`校验后，将自动执行该脚本，我们可以在这里执行部署任务

```shell
#!/bin/bash
cd /root/docker/nginx/html/
# 执行你的指令，简单示例
git pull origin main  # 从github拉取新代码
docker restart nginx  # 重启nginx    
```

给脚本增加执行权限`chmod +x rebuildblog.sh`

:three: 启动`webhook`

```shell
# 进入webhook安装路径
cd /usr/local/src/webhook

# 启动webhook，只当hook.json，默认监听9000端口
./webhook -hooks ./my-scripts/hooks.json -verbose
```

![image-20250726212413583](https://qiniu.anburger.site/post/image-20250726212413583.png)

测试：`push`代码到仓库后，成功执行脚本

![image-20250726212624165](https://qiniu.anburger.site/post/image-20250726212624165.png)

到这里，我们便实现了`push`代码到`Github`，自动通知服务器并执行相应指令的基本过程。接下来设置webhook为自启动服务。

## 3. 设置开机自启动

:one: 创建服务：

```shell
vi /etc/systemd/system/webhook.service
```

```shell
[Unit]
Description=Webhook Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/usr/local/src/webhook
ExecStart=/usr/local/src/webhook/webhook -hooks ./my-scripts/hooks.json -verbose
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

:two: 启动服务：

```shell
# 重新加载systemd
systemctl daemon-reload

# 启动服务
systemctl start webhook

# 设置开机自启动
systemctl enable webhook

# 查看服务状态
systemctl status webhook
```

:tada: Congratulations! 我们已经实现基本的自动化部署了，你可以尽情的在此基础上进行扩展，例如​完善请求校验处理、按需求撰写你的执行脚本等。

