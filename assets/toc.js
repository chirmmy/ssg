(function () {
  var content = document.querySelector('[data-toc-source]');
  var tocRoot = document.querySelector('[data-toc]');
  if (!content || !tocRoot) return;

  var headings = content.querySelectorAll('h2, h3');
  if (headings.length < 3) {
    tocRoot.remove();
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

  tocRoot.innerHTML = html;

  // 滚动高亮
  var links = tocRoot.querySelectorAll('.toc__link');
  var map = {};
  links.forEach(function (a) {
    map[a.getAttribute('href').slice(1)] = a;
  });

  var observer = new IntersectionObserver(function (entries) {
    entries.forEach(function (entry) {
      var link = map[entry.target.id];
      if (!link) return;
      if (entry.isIntersecting) {
        links.forEach(function (l) { l.classList.remove('is-active'); });
        link.classList.add('is-active');
      }
    });
  }, { rootMargin: '-80px 0px -70% 0px', threshold: 0 });

  headings.forEach(function (h) { observer.observe(h); });
})();