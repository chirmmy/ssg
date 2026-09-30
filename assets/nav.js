(function () {
  var toggle = document.querySelector('[data-nav-toggle]');
  var nav = document.querySelector('[data-nav]');

  if (toggle && nav) {
    function setOpen(open) {
      nav.setAttribute('data-open', open ? 'true' : 'false');
      toggle.setAttribute('aria-expanded', open ? 'true' : 'false');
    }

    toggle.addEventListener('click', function () {
      var open = nav.getAttribute('data-open') === 'true';
      setOpen(!open);
    });

    // 点击导航链接后关闭
    nav.addEventListener('click', function (e) {
      if (e.target.tagName === 'A') setOpen(false);
    });

    // Esc 关闭
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') setOpen(false);
    });

    // 断点切换时复位
    window.matchMedia('(min-width: 640px)').addEventListener('change', function (e) {
      if (e.matches) setOpen(false);
    });

    setOpen(false);
  }

  // “返回”按钮：来源是本站时回退到上一页，否则回退到 href 指定的父级路径
  var back = document.querySelector('[data-back]');
  if (back) {
    back.addEventListener('click', function (e) {
      try {
        var ref = document.referrer && new URL(document.referrer);
        if (ref && ref.origin === location.origin && history.length > 1) {
          e.preventDefault();
          history.back();
        }
      } catch (err) { /* referrer 解析失败则直接跟随 href */ }
    });
  }
})();
