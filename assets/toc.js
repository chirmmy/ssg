(function () {
  var content = document.querySelector('[data-toc-source]');
  var inlineRoot = document.querySelector('[data-toc]');
  var railRoot = document.querySelector('[data-toc-rail]');
  if (!content || (!inlineRoot && !railRoot)) return;

  var headings = content.querySelectorAll('h2, h3');
  if (headings.length < 2) {
    if (inlineRoot) inlineRoot.remove();
    if (railRoot) railRoot.remove();
    return;
  }

  var html = '<p class="toc__title">目录</p><ul class="toc__list">';
  var lastLevel = 2;
  var stack = [];

  headings.forEach(function (h, i) {
    var level = h.tagName === 'H2' ? 2 : 3;
    if (!h.id) h.id = 'h-' + i;

    if (level > lastLevel) {
      html += '<ul class="toc__list">';
      stack.push(true);
    } else if (level < lastLevel) {
      while (stack.length && level < lastLevel) {
        html += '</li></ul>';
        stack.pop();
        lastLevel--;
      }
    } else if (i > 0) {
      html += '</li>';
    }

    html += '<li class="toc__item"><a class="toc__link" href="#' + h.id + '">' + h.textContent + '</a>';
    lastLevel = level;
  });

  while (stack.length) {
    html += '</li></ul>';
    stack.pop();
  }
  html += '</li></ul>';

  // 同一份目录注入两个容器：画框内的行内目录（窄屏）+ 右侧空白区固定目录栏（宽屏）
  if (inlineRoot) inlineRoot.innerHTML = html;
  if (railRoot) railRoot.innerHTML = html;

  var links = [];
  var map = {};
  [inlineRoot, railRoot].forEach(function (root) {
    if (!root) return;
    root.querySelectorAll('.toc__link').forEach(function (a) {
      links.push(a);
      var id = a.getAttribute('href').slice(1);
      (map[id] = map[id] || []).push(a);
    });
  });

  var setActive = function (id) {
    links.forEach(function (l) { l.classList.remove('is-active'); });
    var arr = map[id];
    if (!arr) return;
    arr.forEach(function (a) { a.classList.add('is-active'); });

    // 宽屏目录栏超长时，让高亮项保持在可视范围内（只滚目录栏自身，不牵动页面）
    if (railRoot && railRoot.scrollHeight > railRoot.clientHeight + 4) {
      var active = arr[0];
      railRoot.scrollTop +=
        active.offsetTop - railRoot.clientHeight / 2 + active.offsetHeight;
    }
  };

  var observer = new IntersectionObserver(function (entries) {
    entries.forEach(function (entry) {
      if (entry.isIntersecting) setActive(entry.target.id);
    });
  }, { rootMargin: '-80px 0px -70% 0px', threshold: 0 });

  headings.forEach(function (h) { observer.observe(h); });

  // 目录跳转不产生历史记录：手动滚动 + replaceState 仅更新地址栏。
  // 否则每次点击目录都会向 history 压入一条 #h-x 记录，
  // 导致导航「返回」（history.back()）先退回点击目录前的位置而非上一页面。
  [inlineRoot, railRoot].forEach(function (root) {
    if (!root) return;
    root.addEventListener('click', function (e) {
      var a = e.target.closest('a.toc__link');
      if (!a) return;
      var id = a.getAttribute('href').slice(1);
      var target = document.getElementById(id);
      if (!target) return;
      e.preventDefault();
      target.scrollIntoView(); // 平滑滚动由全局 scroll-behavior 控制，遵循 scroll-margin-top
      if (history.replaceState) history.replaceState(null, '', '#' + id);
      setActive(id); // 立即高亮，不等 IntersectionObserver 异步触发
    });
  });
})();
