(function () {
  var COPY_ICON = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/></svg>';
  var CHECK_ICON = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M20 6 9 17l-5-5"/></svg>';

  var blocks = document.querySelectorAll('.content pre');
  blocks.forEach(function (pre) {
    if (pre.closest('.code-block')) return;

    var wrapper = document.createElement('div');
    wrapper.className = 'code-block';

    pre.parentNode.insertBefore(wrapper, pre);
    wrapper.appendChild(pre);

    var btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'code-block__copy';
    btn.setAttribute('aria-label', '复制代码');
    btn.innerHTML = COPY_ICON;

    btn.addEventListener('click', function () {
      var code = pre.querySelector('code');
      var text = code ? code.innerText : pre.innerText;
      copy(text).then(function (ok) {
        if (!ok) return;
        btn.classList.add('is-copied');
        btn.innerHTML = CHECK_ICON;
        setTimeout(function () {
          btn.classList.remove('is-copied');
          btn.innerHTML = COPY_ICON;
        }, 1500);
      });
    });

    wrapper.appendChild(btn);
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