/* 文章页背景星空：与首页 hero.js 的星星参数/绘制保持一致（无吃豆人） */
(function () {
  'use strict';

  var canvas = document.querySelector('[data-starfield]');
  var root = document.documentElement;
  if (!canvas) return;

  var ctx = canvas.getContext('2d', { alpha: true });
  if (!ctx) return;

  var reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  /* ---------- 可调参数（与首页一致） ---------- */

  var STAR_DENSITY = 30000;
  var STAR_COUNT_MIN = 14;
  var STAR_COUNT_MAX = 55;
  var STAR_R_MIN = 1.8;
  var STAR_R_MAX = 4.6;
  var STAR_BIG_RATIO = 0.22;
  var STAR_ACCENT_RATIO = 0.35;
  var STAR_TWINKLE_MIN = 0.0012;
  var STAR_TWINKLE_MAX = 0.0032;

  /* ---------- 状态 ---------- */

  var width = 0;
  var height = 0;
  var dpr = Math.min(window.devicePixelRatio || 1, 2);

  var stars = [];       // { tx, ty, r, rot, rotSpeed, phase, speed, color, big, spawnAt }
  var colors = { fg: '#111111', accent: '#2563eb' };

  var rafId = null;
  var running = false;
  var lastTime = 0;

  /* ---------- 颜色 ---------- */

  function readColors() {
    var cs = getComputedStyle(root);
    colors.fg = cs.getPropertyValue('--color-fg').trim() || '#111111';
    colors.accent = cs.getPropertyValue('--color-accent').trim() || '#2563eb';
    applyStarColors();
  }

  function applyStarColors() {
    for (var i = 0; i < stars.length; i++) {
      stars[i].color = stars[i].accent ? colors.accent : colors.fg;
    }
  }

  /* ---------- 构建星空 ---------- */

  function build() {
    var rect = canvas.getBoundingClientRect();
    width = rect.width;
    height = rect.height;
    if (width <= 0 || height <= 0) return;

    canvas.width = Math.round(width * dpr);
    canvas.height = Math.round(height * dpr);
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

    stars = [];

    var count = Math.round((width * height) / STAR_DENSITY);
    count = Math.max(STAR_COUNT_MIN, Math.min(STAR_COUNT_MAX, count));

    var margin = 24;
    for (var i = 0; i < count; i++) {
      var big = Math.random() < STAR_BIG_RATIO;

      stars.push({
        tx: margin + Math.random() * Math.max(1, width - margin * 2),
        ty: margin + Math.random() * Math.max(1, height - margin * 2),
        r: big
          ? STAR_R_MAX * (0.85 + Math.random() * 0.3)
          : STAR_R_MIN + Math.random() * (STAR_R_MAX - STAR_R_MIN) * 0.7,
        rot: Math.random() * Math.PI,
        rotSpeed: (Math.random() - 0.5) * 0.0006,
        phase: Math.random() * Math.PI * 2,
        speed: STAR_TWINKLE_MIN + Math.random() * (STAR_TWINKLE_MAX - STAR_TWINKLE_MIN),
        accent: Math.random() < STAR_ACCENT_RATIO,
        big: big,
        spawnAt: 0,
      });
    }

    applyStarColors();
  }

  /* ---------- 绘制 ---------- */

  function starPath(x, y, r, rot) {
    var inner = r * 0.36;
    ctx.beginPath();
    for (var i = 0; i < 8; i++) {
      var ang = rot + i * Math.PI / 4;
      var rad = (i % 2 === 0) ? r : inner;
      var px = x + Math.cos(ang) * rad;
      var py = y + Math.sin(ang) * rad;
      if (i === 0) ctx.moveTo(px, py);
      else ctx.lineTo(px, py);
    }
    ctx.closePath();
  }

  function drawStars(now) {
    ctx.clearRect(0, 0, width, height);

    for (var i = 0; i < stars.length; i++) {
      var s = stars[i];

      // 闪烁曲线：多数时间偏暗、偶尔骤亮
      var tw = Math.sin(now * s.speed + s.phase);
      var twinkle = 0.3 + 0.7 * Math.pow((tw + 1) / 2, 2.2);
      var alpha = twinkle;
      var r = s.r * (0.82 + 0.18 * twinkle);
      var rot = s.rot + now * s.rotSpeed;

      ctx.save();
      ctx.globalAlpha = alpha;
      ctx.fillStyle = s.color;

      if (s.big && alpha > 0.45) {
        ctx.shadowColor = s.color;
        ctx.shadowBlur = 10 * alpha;
      }

      starPath(s.tx, s.ty, r, rot);
      ctx.fill();
      ctx.restore();
    }
    ctx.globalAlpha = 1;
  }

  function drawStatic() {
    for (var i = 0; i < stars.length; i++) {
      var s = stars[i];
      ctx.globalAlpha = 0.85;
      ctx.fillStyle = s.color;
      starPath(s.tx, s.ty, s.r, s.rot);
      ctx.fill();
    }
    ctx.globalAlpha = 1;
  }

  /* ---------- 循环 ---------- */

  function loop(now) {
    if (!running) return;
    drawStars(now);
    rafId = requestAnimationFrame(loop);
  }

  function start() {
    if (running) return;
    running = true;
    lastTime = performance.now();
    rafId = requestAnimationFrame(loop);
  }

  function stop() {
    running = false;
    if (rafId) {
      cancelAnimationFrame(rafId);
      rafId = null;
    }
  }

  /* ---------- 事件 ---------- */

  var resizeTimer = null;
  function onResize() {
    if (resizeTimer) clearTimeout(resizeTimer);
    resizeTimer = setTimeout(function () {
      build();
      if (reduceMotion) {
        ctx.clearRect(0, 0, width, height);
        drawStatic();
      }
    }, 150);
  }

  function onVisibility() {
    if (document.hidden) stop();
    else if (!reduceMotion) start();
  }

  function init() {
    readColors();
    build();

    if (reduceMotion) {
      drawStatic();
    } else {
      start();
    }

    window.addEventListener('resize', onResize);
    document.addEventListener('visibilitychange', onVisibility);

    // 主题切换期间（html.theme-anim）暂停星空绘制，避免与全页颜色过渡
    // 争抢主线程造成掉帧；过渡结束后读取最终颜色并恢复。
    new MutationObserver(function () {
      var animating = root.classList.contains('theme-anim');
      if (animating) {
        stop();
        return;
      }
      readColors();
      if (reduceMotion) {
        drawStatic();
      } else {
        start();
      }
    }).observe(root, { attributes: true, attributeFilter: ['data-theme', 'class'] });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
