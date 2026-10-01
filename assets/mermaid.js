/* Mermaid 图表渲染：
   构建期 ```mermaid 代码块输出为 <pre class="mermaid">源码</pre>，
   此脚本按需懒加载 mermaid 库渲染，并在主题切换时以对应主题重新渲染。 */
(function () {
  var blocks = Array.prototype.slice.call(document.querySelectorAll('pre.mermaid'));
  if (!blocks.length) return;

  // 原始源码缓存：重渲染（主题切换）时以源码为准，避免拿到已替换的 svg
  var sources = blocks.map(function (el) { return el.textContent.trim(); });
  var renderSeq = 0;

  function currentTheme() {
    return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'neutral';
  }

  /* 懒加载 mermaid（ESM）。jsdelivr 主源，unpkg 兜底；
     加载结果缓存在 window 上，多页跳转/主题重渲染只加载一次。 */
  function loadMermaid() {
    if (!window.__mermaidLoader) {
      window.__mermaidLoader = import('https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs')
        .catch(function () {
          return import('https://unpkg.com/mermaid@11/dist/mermaid.esm.min.mjs');
        });
    }
    return window.__mermaidLoader;
  }

  function renderAll() {
    var seq = ++renderSeq;
    loadMermaid().then(function (mod) {
      var mermaid = mod.default;
      mermaid.initialize({
        startOnLoad: false,
        theme: currentTheme(),
        securityLevel: 'strict'
      });
      var chain = Promise.resolve();
      blocks.forEach(function (el, i) {
        chain = chain.then(function () {
          // 主题切换后可能又触发了新一轮渲染，丢弃过期任务
          if (seq !== renderSeq) return;
          return mermaid.render('mmd-' + seq + '-' + i, sources[i]).then(function (res) {
            if (seq !== renderSeq) return;
            el.innerHTML = res.svg;
            el.classList.add('is-rendered');
            el.classList.remove('is-error');
          }).catch(function () {
            // 语法错误：保留源码展示，标记错误态便于排查
            el.textContent = sources[i];
            el.classList.add('is-error');
            el.classList.remove('is-rendered');
          });
        });
      });
    }).catch(function () {
      // CDN 全部失败：保留源码，不强隐藏内容
      blocks.forEach(function (el) { el.classList.add('is-error'); });
    });
  }

  renderAll();

  // 主题切换时以对应主题重新渲染
  var lastTheme = currentTheme();
  new MutationObserver(function () {
    var t = currentTheme();
    if (t !== lastTheme) {
      lastTheme = t;
      renderAll();
    }
  }).observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
})();
