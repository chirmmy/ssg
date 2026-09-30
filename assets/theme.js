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

    // 点击触发的主题切换：全局颜色属性并行动画（无 View Transitions 快照开销）
    var reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (originEvent && !reduceMotion) {
      // 先加过渡类，等浏览器提交一次带过渡的样式（双重 rAF）后再换主题，
      // 避免加类/重排/换主题挤在同一帧导致首帧过长、动画起步卡顿。
      root.classList.add('theme-anim');
      requestAnimationFrame(function () {
        requestAnimationFrame(function () {
          doApply();
        });
      });

      // 过渡结束后移除类，避免影响 hover 等常规过渡（时长与 CSS 中过渡时长一致并留余量）
      window.setTimeout(function () {
        root.classList.remove('theme-anim');
      }, 380);
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