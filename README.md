# ssg

一个从零自研的轻量静态站点生成器，用于个人主页与博客。

目标不是替代 Hugo，而是通过自研理解静态站点生成器的核心管线：内容加载、Markdown 渲染、模板引擎、并发构建、增量构建与开发体验。

## 目标

- 学习 Go 并发、模板、构建管线设计
- 单二进制，零运行时依赖
- 构建时生成纯静态 HTML，默认零 JS
- 性能与开发者体验优先
- 架构可演进：内容源、渲染器、部署目标可替换

## 当前状态

Phase 1：核心管线

- [x] 配置加载（TOML）
- [x] Markdown + Frontmatter 解析
- [x] 并发内容加载
- [x] `html/template` 模板引擎
- [x] 路由与并发渲染
- [x] 首页 / 关于 / 博客列表 / 文章页
- [x] 静态资源拷贝
- [x] 开发服务器与热重载
- [ ] RSS / sitemap / robots
- [x] CSS / JS 处理
- [ ] 增量构建
- [ ] 搜索索引

## 技术栈

- Go 1.22+
- goldmark：Markdown 解析
- chroma：构建时语法高亮
- html/template：模板渲染
- cobra：CLI
- errgroup：并发控制
- BurntSushi/toml：配置解析
- fsnotify：开发服务器文件监听（计划）

## 目录结构

```txt
ssg/
  cmd/
    ssg/
      main.go          # CLI 入口
  internal/
    config/            # 配置加载
    content/           # 内容模型、加载、Markdown 渲染
    template/          # 模板引擎
    render/            # 路由与渲染
    build/             # 构建器
    site/              # 站点聚合
  templates/
    layouts/
    partials/
    home.html
    blog-index.html
    post.html
    page.html
  content/
    pages/
      about.md
    blog/
      hello-world.md
  static/
    style.css
  site.toml