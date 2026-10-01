(function () {
  var content = document.querySelector('[data-toc-source]');
  var inlineRoot = document.querySelector('[data-toc-inline]');
  var railRoot = document.querySelector('[data-toc-rail]');
  var countEl = document.querySelector('[data-toc-count]');
  if (!content) return;

  var headings = content.querySelectorAll('h2, h3');

  // 标题锚点：悬停出现 # 按钮，点击复制该小节链接。
  // 放在目录生成之前会污染 textContent，所以顺序很关键。
  function addAnchors() {
    var targets = content.querySelectorAll('h2, h3, h4');
    Array.prototype.forEach.call(targets, function (h) {
      if (!h.id) return;
      var a = document.createElement('a');
      a.className = 'h-anchor';
      a.href = '#' + h.id;
      a.textContent = '#';
      a.setAttribute('aria-label', '复制「' + h.textContent.trim() + '」的标题链接');
      a.addEventListener('click', function (e) {
        e.preventDefault();
        var url = location.href.split('#')[0] + '#' + h.id;
        if (history.replaceState) history.replaceState(null, '', '#' + h.id);
        copy(url, a);
      });
      h.appendChild(a);
    });
  }

  function copy(text, el) {
    var done = function () {
      if (el) {
        el.classList.add('is-copied');
        setTimeout(function () { el.classList.remove('is-copied'); }, 1400);
      }
      toast('已复制链接');
    };
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(done, done);
    } else {
      done();
    }
  }

  var toastEl = null;
  var toastTimer = null;
  function toast(msg) {
    if (!toastEl) {
      toastEl = document.createElement('div');
      toastEl.className = 'toast';
      toastEl.setAttribute('role', 'status');
      document.body.appendChild(toastEl);
    }
    toastEl.textContent = msg;
    toastEl.classList.add('is-on');
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function () { toastEl.classList.remove('is-on'); }, 1600);
  }

  var links = [];

  if (headings.length >= 2) {
    var ol = document.createElement('ol');
    Array.prototype.forEach.call(headings, function (h, i) {
      if (!h.id) h.id = 'h-' + i;
      var li = document.createElement('li');
      if (h.tagName === 'H3') li.className = 'toc__l3';
      var a = document.createElement('a');
      a.href = '#' + h.id;
      a.setAttribute('data-target', h.id);
      a.textContent = h.textContent.trim();
      li.appendChild(a);
      ol.appendChild(li);
    });

    if (inlineRoot) inlineRoot.appendChild(ol.cloneNode(true));
    if (railRoot) railRoot.appendChild(ol.cloneNode(true));
    if (countEl) countEl.textContent = headings.length + ' 节';

    var map = {};
    [inlineRoot, railRoot].forEach(function (root) {
      if (!root) return;
      Array.prototype.forEach.call(root.querySelectorAll('a'), function (a) {
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

      // 右栏目录超长时让高亮项留在可视范围内（只滚目录栏自身）
      var inner = railRoot && railRoot.closest('.rail__inner');
      if (inner && inner.scrollHeight > inner.clientHeight + 4) {
        var active = arr[0];
        inner.scrollTop += active.offsetTop - inner.clientHeight / 2 + active.offsetHeight;
      }
    };

    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) setActive(entry.target.id);
      });
    }, { rootMargin: '-96px 0px -70% 0px', threshold: 0 });
    Array.prototype.forEach.call(headings, function (h) { observer.observe(h); });

    // 目录跳转不写历史记录：否则每次点目录都会压一条 #h-x 记录，
    // 「返回」会先退回点击目录前的位置而不是上一页。
    [inlineRoot, railRoot].forEach(function (root) {
      if (!root) return;
      root.addEventListener('click', function (e) {
        var a = e.target.closest('a');
        if (!a) return;
        var id = a.getAttribute('href').slice(1);
        var target = document.getElementById(id);
        if (!target) return;
        e.preventDefault();
        target.scrollIntoView(); // 平滑滚动由全局 scroll-behavior 控制，遵循 scroll-margin-top
        if (history.replaceState) history.replaceState(null, '', '#' + id);
        setActive(id);
        // 窄屏点击目录后收起折叠面板
        var details = inlineRoot.closest('details');
        if (details && details.open && window.matchMedia('(max-width: 1099px)').matches) {
          details.open = false;
        }
      });
    });
  } else {
    // 标题太少，目录没有意义：收起折叠目录、移掉右栏目录与它的标题
    var inlineBox = inlineRoot && inlineRoot.closest('details');
    if (inlineBox) inlineBox.remove();
    if (railRoot) {
      var inner = railRoot.parentNode; // .rail__inner
      var label = inner && inner.querySelector('.rail__label');
      if (label) label.remove();
      railRoot.remove();
    }
  }

  addAnchors();

  // 顶栏与文末的「复制链接」
  Array.prototype.forEach.call(document.querySelectorAll('[data-copy-link]'), function (b) {
    b.addEventListener('click', function () { copy(location.href, b); });
  });

  // 回到顶部
  Array.prototype.forEach.call(document.querySelectorAll('[data-totop]'), function (b) {
    b.addEventListener('click', function () {
      window.scrollTo({ top: 0, behavior: 'smooth' });
    });
  });
})();
