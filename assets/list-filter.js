(function () {
  var list = document.querySelector('[data-list]');
  if (!list) return;

  var cards = Array.prototype.slice.call(list.querySelectorAll('.bcard'));
  var statusText = document.querySelector('[data-status-text]');
  var clearBtns = document.querySelectorAll('[data-clear]');
  var emptyBox = document.querySelector('[data-empty]');
  var searchInput = document.querySelector('[data-search]');
  var moreBtn = document.querySelector('[data-chips-more]');
  var extraWrap = document.querySelector('[data-chips-extra]');
  var state = { tag: '', q: '' };

  function tagsOf(card) {
    return (card.getAttribute('data-tags') || '').split('|');
  }

  function apply() {
    var q = state.q.trim().toLowerCase();
    var shown = 0;

    cards.forEach(function (c) {
      var okTag = !state.tag || tagsOf(c).indexOf(state.tag) >= 0;
      var okQ = !q || (c.getAttribute('data-search') || '').indexOf(q) >= 0;
      var ok = okTag && okQ;
      c.hidden = !ok;
      if (ok) shown++;
    });

    // 分组篇数随筛选实时更新，没有命中的分组整体隐藏
    Array.prototype.forEach.call(list.querySelectorAll('.mgroup'), function (g) {
      var vis = g.querySelectorAll('.bcard:not([hidden])').length;
      g.hidden = vis === 0;
      var n = g.querySelector('[data-mgroup-n]');
      if (n) n.textContent = vis + ' 篇';
    });

    var parts = ['显示 ' + shown + ' / ' + cards.length + ' 篇'];
    if (state.tag) parts.push('标签：' + state.tag);
    if (q) parts.push('关键词：' + state.q.trim());
    if (statusText) statusText.textContent = parts.join(' · ');

    if (emptyBox) emptyBox.hidden = shown > 0;
    Array.prototype.forEach.call(clearBtns, function (b) { b.hidden = !state.tag && !q; });

    // 精选大卡不参与匹配，筛选时收起，避免误读
    var feat = document.querySelector('.feat');
    if (feat) feat.hidden = !!(state.tag || q);

    Array.prototype.forEach.call(document.querySelectorAll('[data-chip]'), function (b) {
      b.classList.toggle('is-on', (b.getAttribute('data-chip') || '') === state.tag);
    });
    Array.prototype.forEach.call(document.querySelectorAll('[data-tag-item]'), function (b) {
      b.classList.toggle('is-on', (b.getAttribute('data-tag-item') || '') === state.tag);
    });
  }

  function setTag(tag) {
    state.tag = state.tag === tag ? '' : tag;
    apply();
  }

  Array.prototype.forEach.call(document.querySelectorAll('[data-chip]'), function (b) {
    b.addEventListener('click', function () { setTag(b.getAttribute('data-chip')); });
  });
  Array.prototype.forEach.call(document.querySelectorAll('[data-tag-item]'), function (b) {
    b.addEventListener('click', function () { setTag(b.getAttribute('data-tag-item')); });
  });

  if (searchInput) {
    searchInput.addEventListener('input', function () {
      state.q = searchInput.value;
      apply();
    });
  }

  Array.prototype.forEach.call(clearBtns, function (b) {
    b.addEventListener('click', function () {
      state.tag = '';
      state.q = '';
      if (searchInput) searchInput.value = '';
      apply();
    });
  });

  if (moreBtn && extraWrap) {
    moreBtn.addEventListener('click', function () {
      var open = extraWrap.hidden;
      extraWrap.hidden = !open;
      moreBtn.setAttribute('aria-expanded', open ? 'true' : 'false');
      moreBtn.firstChild.nodeValue = open ? '收起 ' : '更多 ';
    });
  }

  // 密度切换：同一份 DOM，靠 html 上的 mode-card / mode-row 变形
  Array.prototype.forEach.call(document.querySelectorAll('[data-set-mode]'), function (b) {
    b.addEventListener('click', function () {
      var mode = b.getAttribute('data-set-mode');
      document.documentElement.classList.toggle('mode-row', mode === 'row');
      document.documentElement.classList.toggle('mode-card', mode === 'card');
      Array.prototype.forEach.call(document.querySelectorAll('[data-set-mode]'), function (x) {
        var on = x === b;
        x.classList.toggle('is-on', on);
        x.setAttribute('aria-pressed', on ? 'true' : 'false');
      });
      try { localStorage.setItem('blog-mode', mode); } catch (e) {}
    });
  });

  // 记住上次的密度选择
  try {
    var saved = localStorage.getItem('blog-mode');
    if (saved === 'row') {
      var rowBtn = document.querySelector('[data-set-mode="row"]');
      if (rowBtn) rowBtn.click();
    }
  } catch (e) {}

  // 支持从文章页的标签链接过来：/blog/?tag=xxx（也支持 ?q=）
  try {
    var params = new URLSearchParams(location.search);
    var tag = params.get('tag');
    var q = params.get('q');
    if (tag) state.tag = tag;
    if (q) {
      state.q = q;
      if (searchInput) searchInput.value = q;
    }
    if (tag || q) {
      apply();
      // 带着筛选进来时不要停在精选大卡上
      var grid = document.querySelector('.filters');
      if (grid) grid.scrollIntoView({ block: 'start' });
    }
  } catch (e) {}

  apply();
})();
