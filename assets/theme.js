(function () {
  var root = document.documentElement;
  var btn = document.querySelector('[data-theme-toggle]');
  if (!btn) return;

  var mq = window.matchMedia('(prefers-color-scheme: dark)');
  var order = ['auto', 'light', 'dark'];
  var labels = { auto: '跟随系统', light: '亮色', dark: '暗色' };

  function getPref() {
    try {
      return localStorage.getItem('theme') || 'auto';
    } catch (e) {
      return 'auto';
    }
  }

  function setPref(pref, originEvent) {
    try {
      localStorage.setItem('theme', pref);
    } catch (e) {}
    apply(originEvent);
  }

  function apply(originEvent) {
    var pref = getPref();
    var effective = pref === 'auto'
      ? (mq.matches ? 'dark' : 'light')
      : pref;

    var doApply = function () {
      root.setAttribute('data-theme', effective);
      root.setAttribute('data-theme-preference', pref);
      btn.setAttribute('aria-label', '切换主题（当前：' + labels[pref] + '）');
    };

    // 点击触发的主题切换：用圆形扩散
    if (
      originEvent &&
      document.startViewTransition &&
      !window.matchMedia('(prefers-reduced-motion: reduce)').matches
    ) {
      // 以点击位置为中心
      var x = originEvent.clientX;
      var y = originEvent.clientY;
      root.style.setProperty('--theme-x', x + 'px');
      root.style.setProperty('--theme-y', y + 'px');
      root.classList.add('theme-transition');

      var transition = document.startViewTransition(doApply);
      transition.finished.finally(function () {
        root.classList.remove('theme-transition');
        root.style.removeProperty('--theme-x');
        root.style.removeProperty('--theme-y');
      });
    } else {
      doApply();
    }
  }

  btn.addEventListener('click', function (e) {
    var curr = getPref();
    var i = order.indexOf(curr);
    if (i < 0) i = 0;
    setPref(order[(i + 1) % order.length], e);
  });

  // 系统主题变化时，如果当前是 auto，跟着变
  if (mq.addEventListener) {
    mq.addEventListener('change', function () {
      if (getPref() === 'auto') apply();
    });
  } else if (mq.addListener) {
    mq.addListener(function () {
      if (getPref() === 'auto') apply();
    });
  }

  apply();
})();