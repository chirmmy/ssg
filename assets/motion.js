(function () {
  var reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  if (reduce) {
    document.querySelectorAll('[data-reveal]').forEach(function (el) {
      el.classList.add('is-revealed');
    });
    document.querySelectorAll('.enter').forEach(function (el) {
      el.classList.add('is-entering');
    });
    return;
  }

  // 首屏进入动画
  var enterEls = document.querySelectorAll('.enter');
  if (enterEls.length) {
    requestAnimationFrame(function () {
      enterEls.forEach(function (el) {
        el.classList.add('is-entering');
      });
    });
  }

  // 滚动进入动画
  var revealEls = document.querySelectorAll('[data-reveal]');
  if (!revealEls.length) return;

  if (!('IntersectionObserver' in window)) {
    revealEls.forEach(function (el) { el.classList.add('is-revealed'); });
    return;
  }

  var observer = new IntersectionObserver(
    function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          entry.target.classList.add('is-revealed');
          observer.unobserve(entry.target);
        }
      });
    },
    {
      rootMargin: '0px 0px -10% 0px',
      threshold: 0.05,
    }
  );

  revealEls.forEach(function (el) {
    observer.observe(el);
  });
})();