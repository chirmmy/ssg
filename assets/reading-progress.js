(function () {
  var bar = document.querySelector('[data-reading-progress]');
  var article = document.querySelector('[data-reading-source]');
  if (!bar || !article) return;

  var ticking = false;

  function update() {
    var rect = article.getBoundingClientRect();
    var top = window.scrollY + rect.top;
    var height = article.offsetHeight;
    var vh = window.innerHeight;
    var scrolled = window.scrollY - top + vh * 0.3;
    var ratio = Math.min(1, Math.max(0, scrolled / (height - vh * 0.3)));
    bar.style.width = (ratio * 100).toFixed(2) + '%';
    ticking = false;
  }

  window.addEventListener('scroll', function () {
    if (!ticking) {
      requestAnimationFrame(update);
      ticking = true;
    }
  }, { passive: true });

  update();
})();