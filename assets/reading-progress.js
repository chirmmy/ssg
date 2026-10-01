(function () {
  var bars = document.querySelectorAll('[data-progress-bar]');
  var nums = document.querySelectorAll('[data-progress-num]');
  var src = document.querySelector('[data-reading-source]');
  if (!bars.length && !nums.length) return;

  var ticking = false;

  // 阅读进度按「正文」而不是整篇文档计算：从正文顶部开始，到正文结束为 100%。
  function ratio() {
    var vh = window.innerHeight;
    var y = window.pageYOffset || document.documentElement.scrollTop || 0;

    if (src) {
      var rect = src.getBoundingClientRect();
      var top = y + rect.top;
      var span = src.offsetHeight - vh * 0.3;
      if (span > 40) {
        return Math.min(1, Math.max(0, (y - top + vh * 0.3) / span));
      }
    }

    var total = document.documentElement.scrollHeight - vh;
    return total > 8 ? Math.min(1, Math.max(0, y / total)) : 0;
  }

  function update() {
    var p = ratio();
    var pct = (p * 100).toFixed(2) + '%';
    Array.prototype.forEach.call(bars, function (el) { el.style.width = pct; });
    Array.prototype.forEach.call(nums, function (el) { el.textContent = Math.round(p * 100) + '%'; });
  }

  function onScroll() {
    if (ticking) return;
    ticking = true;
    requestAnimationFrame(function () { update(); ticking = false; });
  }

  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('resize', onScroll);
  update();
})();
