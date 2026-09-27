---
title: Github Actions笔记
date: 2025-08-08T23:00:00+08:00
lastmod: 2025-08-08T23:00:00+08:00
categories: [笔记,Github]
cover: https://qiniu.anburger.site/cover/Note.webp
---


## 1 简介

### 1.1 Actions组成

> An **action** is a pre-defined, reusable set of jobs or code that performs specific tasks within a **workflow**.

:one: **Workflows**

> A **workflow** is a configurable automated process that will run one or more jobs.

:two: **Events**

> An **event** is a specific activity in a repository that triggers a **workflow** run.

:three: **Jobs**

> A **job** is a set of **steps** in a workflow that is executed on the same **runner**.

:four: **Runners**

> A **runner** is a server that runs your workflows when they're triggered. Each runner can run a single **job** at a time.

### 1.2 Starter Workflows

{{<externalLinkCard title="Starter Workflows" link="https://github.com/actions/starter-workflows" cover="auto">}}

## 2 Quick Start

:one: `new workflow`

`进入仓库-->Actions-->new workflow`，可以选择Github推荐的workflow，或者`set up  a workflow yourself`

![image-20250809130637620](https://qiniu.anburger.site/post/image-20250809130637620.png)

创建`action-demo.yml`（也可以自己新建`.github/workflows`文件夹，在该文件夹下新建`.yml`文件配置新的workflow）：

```yml
name: GitHub Actions Demo
run-name: ${{ github.actor }} is testing out GitHub Actions 🚀
on: [push]
jobs:
  Explore-GitHub-Actions:
    runs-on: ubuntu-latest
    steps:
      - run: echo "🎉 The job was automatically triggered by a ${{ github.event_name }} event."
      - run: echo "🐧 This job is now running on a ${{ runner.os }} server hosted by GitHub!"
      - run: echo "🔎 The name of your branch is ${{ github.ref }} and your repository is ${{ github.repository }}."
      - name: Check out repository code
        uses: actions/checkout@v4
      - run: echo "💡 The ${{ github.repository }} repository has been cloned to the runner."
      - run: echo "🖥️ The workflow is now ready to test your code on the runner."
      - name: List files in the repository
        run: |
          ls ${{ github.workspace }}
      - run: echo "🍏 This job's status is ${{ job.status }}."
```

![image-20250809131408013](https://qiniu.anburger.site/post/image-20250809131408013.png)

该workflow被设置为当仓库`push`事件时触发，主要进行一些打印和将该仓库代码clone到我们的`Runner`中。点击`commit changes`后选择一种提交方式，则会push我们本次修改到仓库，同时也触发了新建的workflow。进入`Action`可以查看`workflow`执行结果：

![image-20250809132725794](https://qiniu.anburger.site/post/image-20250809132725794.png)

:tada:Congratulations! Actions的基本使用已经完成。在此基础上，可以通过更详细的配置自由控制你想执行Jobs。

## 3 CI/CD

### 3.1 Continuous Integration (CI)

> Continuous integration (CI) is a software practice that requires frequently committing code to a shared repository. 

> When you commit code to your repository, you can continuously build and test the code to make sure that the commit doesn't introduce errors. 

{{<externalLinkCard title="build-and-test-code" link="https://docs.github.com/en/actions/tutorials/build-and-test-code" cover="auto">}}

:triangular_flag_on_post:案例：`Java with Maven`

`进入仓库-->Actions-->new workflow`，搜索并选择`Java with Maven`。适当调整配置，例如修改Java版本：

```yml
name: Java CI with Maven

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  build:

    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4
    - name: Set up JDK 11
      uses: actions/setup-java@v4
      with:
        java-version: '11'
        distribution: 'temurin'
        cache: maven
    - name: Build with Maven
      run: mvn -B package --file pom.xml
```

提交该workflow后，会因为push触发一次，若此时仓库中没有代码（pom.xml），将会运行失败。

两种触发事件：

:one: push

push代码到主分支（我们提交一个可以正常运行的maven项目测试），等待执行完成后查看结果。

```shell
git pull origin main    # 拉取远程最新代码

git add .	# 添加所有修改

git commit -m "first commit"	# 提交到本地仓库

git push -u origin main		# 推送到远程main分支
```

<div style="display: flex;">
	<img src="https://qiniu.anburger.site/post/image-20250809174945421.png" alt="image-20250809174945421"/>
    <img src="https://qiniu.anburger.site/post/image-20250809175308575.png" alt="image-20250809175308575"/>
</div>

:two: pull_request

我们设置一个Error，然后提交代码到新分支。

```shell
# 如果是最新则不执行也行
git checkout main       # 切换到主分支
git pull origin main    # 拉取远程最新代码
```

```shell
# 创建并切换到新分支
git checkout -b your-branch-name 

git add .	# 添加所有修改

git commit -m "testing ci error"	# 提交到本地仓库

git push -u origin your-branch-name 	# 推送到远程
```

进入github仓库，可以看到pull request提示：点击`Compare & pull request`，并提交。

![image-20250809181117509](https://qiniu.anburger.site/post/image-20250809181117509.png)

可以看到，我们的CI Action检测出了错误，我们就需要谨慎merge啦：

![image-20250809181332473](https://qiniu.anburger.site/post/image-20250809181332473.png)

![image-20250809181515738](https://qiniu.anburger.site/post/image-20250809181515738.png)

### 3.2 Continuous deployment(CD)

> *Continuous deployment* (CD) is the practice of using automation to publish and deploy software updates.

## 4 常用语法

### on

`on`：use `on` to define which events can cause the workflow to run

:one: 基本使用

```yml
on: push    # 单一事件

on: [push, fork]	# 多个事件。任意事件发送均会触发一次workflow
```

:two: active types

> Some events have activity types. Use `on.<event_name>.types` to define the type of event activity that will trigger a workflow run.

```yml
on:
  issues:
    types:
      - opened
      - labeled
```

:three: filter

```yml
on:
  push:
    branches:
      - main
      - 'releases/**'
```

