(function () {
  var COPY_ICON = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/></svg>';
  var CHECK_ICON = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M20 6 9 17l-5-5"/></svg>';
  var CHEVRON_ICON = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M6 9l6 6 6-6"/></svg>';

  /* 构建期（goldmark-highlighting wrapper）已输出 mac 风格头部；
     此处兜底处理未经构建期包装的代码块（如无语言标注的缩进代码块）。 */
  document.querySelectorAll('.content pre').forEach(function (pre) {
    if (pre.closest('.code-block')) return;

    var wrapper = document.createElement('div');
    wrapper.className = 'code-block';
    wrapper.setAttribute('data-lang', 'code');
    wrapper.innerHTML =
      '<div class="code-block__header">' +
        '<span class="code-block__dots" aria-hidden="true"><i></i><i></i><i></i></span>' +
        '<span class="code-block__lang">CODE</span>' +
        '<div class="code-block__actions">' +
          '<button type="button" class="code-block__collapse" aria-label="折叠代码" aria-expanded="true">' + CHEVRON_ICON + '</button>' +
          '<button type="button" class="code-block__copy" aria-label="复制代码">' + COPY_ICON + '</button>' +
        '</div>' +
      '</div>' +
      '<div class="code-block__body"></div>';
    pre.parentNode.insertBefore(wrapper, pre);
    wrapper.querySelector('.code-block__body').appendChild(pre);
  });

  /* 复制 */
  document.querySelectorAll('.code-block__copy').forEach(function (btn) {
    btn.addEventListener('click', function () {
      var block = btn.closest('.code-block');
      if (!block) return;
      var code = block.querySelector('pre code') || block.querySelector('pre');
      if (!code) return;
      copy(code.innerText).then(function (ok) {
        if (!ok) return;
        btn.classList.add('is-copied');
        btn.innerHTML = CHECK_ICON;
        setTimeout(function () {
          btn.classList.remove('is-copied');
          btn.innerHTML = COPY_ICON;
        }, 1500);
      });
    });
  });

  /* 折叠 / 展开 */
  document.querySelectorAll('.code-block__collapse').forEach(function (btn) {
    btn.addEventListener('click', function () {
      var block = btn.closest('.code-block');
      if (!block) return;
      var collapsed = block.classList.toggle('is-collapsed');
      btn.setAttribute('aria-expanded', collapsed ? 'false' : 'true');
      btn.setAttribute('aria-label', collapsed ? '展开代码' : '折叠代码');
    });
  });

  function copy(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text).then(function () {
        return true;
      }).catch(function () {
        return fallback(text);
      });
    }
    return Promise.resolve(fallback(text));
  }

  function fallback(text) {
    try {
      var ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed';
      ta.style.opacity = '0';
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
      return true;
    } catch (e) {
      return false;
    }
  }
})();
