(function () {
  'use strict';

  var canvas = document.querySelector('[data-hero-canvas]');
  var content = document.querySelector('[data-hero-content]');
  var root = document.documentElement;
  if (!canvas || !content) return;

  var ctx = canvas.getContext('2d', { alpha: true });
  if (!ctx) return;

  var reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  var isCoarse = window.matchMedia('(hover: none) and (pointer: coarse)').matches;

  /* ---------- 可调参数 ---------- */

  var GAP = 14;                 // 豆子网格间距
  var DOT_R_TEXT = 2.6;         // 文字豆半径
  var DOT_R_BG = 1.1;           // 背景豆半径
  var DOT_A_TEXT = 1.0;         // 文字豆透明度
  var DOT_A_BG = 0.14;          // 背景豆透明度

  var PACMAN_RADIUS = 12;       // 吃豆人半径
  var PACMAN_SPEED = 210;       // px/秒
  var PACMAN_BOOST = 2.2;       // 点击冲刺倍率
  var PACMAN_TURN_RATE = 6;     // 转向速度（弧度/秒）
  var PACMAN_MOUTH_SPEED = 14;  // 嘴张合频率

  var RESPAWN_MIN = 1800;       // 豆子重生最短时间（ms）
  var RESPAWN_MAX = 3200;       // 豆子重生最长时间（ms）
  var RESPAWN_FADE = 320;       // 重生淡入时长（ms）

  var MOUSE_IDLE_MS = 2200;     // 鼠标停止多久后吃豆人开始漫游
  var WANDER_MIN_MS = 1800;     // 漫游目标最短停留
  var WANDER_MAX_MS = 3600;     // 漫游目标最长停留

  /* ---------- 状态 ---------- */

  var width = 0;
  var height = 0;
  var dpr = Math.min(window.devicePixelRatio || 1, 2);

  var dots = [];        // { tx, ty, r, alpha, alive, respawnAt, spawnAt }
  var colors = { fg: '#111', accent: '#2563eb' };

  var mouse = { x: 0, y: 0, active: false, lastMove: 0 };

  var pacman = {
    x: 0,
    y: 0,
    angle: 0,
    mouthPhase: 0,
    boost: 1,
    wanderX: 0,
    wanderY: 0,
    wanderUntil: 0,
  };

  var rafId = null;
  var running = false;
  var lastTime = 0;

  /* ---------- 颜色 ---------- */

  function readColors() {
    var cs = getComputedStyle(root);
    colors.fg = cs.getPropertyValue('--color-fg').trim() || '#111111';
    colors.accent = cs.getPropertyValue('--color-accent').trim() || '#2563eb';
  }

  /* ---------- 文字采样 ---------- */

  function buildTextMask() {
    var w = Math.max(1, Math.round(width));
    var h = Math.max(1, Math.round(height));

    var off = document.createElement('canvas');
    off.width = w;
    off.height = h;
    var octx = off.getContext('2d');
    octx.clearRect(0, 0, w, h);

    var canvasRect = canvas.getBoundingClientRect();
    var els = content.querySelectorAll('[data-hero-text]');

    els.forEach(function (el) {
      var cs = getComputedStyle(el);
      var fontSize = parseFloat(cs.fontSize);

      // 关键修复 1：优先用 cs.font（完整简写），保证 canvas 与 DOM 同源
      var fontStr = (cs.font || '').trim();
      if (!fontStr || fontStr.indexOf('px') === -1) {
        // 回退：手动拼，但用 cs.fontSize 保留原始单位
        fontStr = cs.fontWeight + ' ' + cs.fontSize + ' ' + cs.fontFamily;
      }
      octx.font = fontStr;

      var lineHeight = parseFloat(cs.lineHeight);
      if (!lineHeight) lineHeight = fontSize * 1.3;

      var elRect = el.getBoundingClientRect();

      octx.fillStyle = '#ffffff';
      octx.textBaseline = 'middle';
      octx.textAlign = 'center';

      var relX = elRect.left - canvasRect.left;
      var relY = elRect.top - canvasRect.top;

      var text = (el.textContent || '').trim();
      if (!text) return;

      var maxWidth = elRect.width + 4;
      var lines = [];
      var chars = Array.from(text);
      var currentLine = '';

      for (var i = 0; i < chars.length; i++) {
        var test = currentLine + chars[i];
        if (octx.measureText(test).width > maxWidth && currentLine) {
          lines.push(currentLine);
          currentLine = chars[i];
        } else {
          currentLine = test;
        }
      }
      if (currentLine) lines.push(currentLine);

      var totalHeight = lines.length * lineHeight;
      var startY = relY + (elRect.height - totalHeight) / 2 + lineHeight / 2;

      lines.forEach(function (line, idx) {
        octx.fillText(line, relX + elRect.width / 2, startY + idx * lineHeight);
      });
    });

    return octx.getImageData(0, 0, w, h).data;
  }

  /* ---------- 构建豆子网格 ---------- */

  function build() {
    var rect = canvas.getBoundingClientRect();
    width = rect.width;
    height = rect.height;
    if (width <= 0 || height <= 0) return;

    canvas.width = Math.round(width * dpr);
    canvas.height = Math.round(height * dpr);
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

    var mask = buildTextMask();
    var maskW = Math.round(width);
    var maskH = Math.round(height);

    dots = [];

    for (var y = GAP / 2; y < height; y += GAP) {
      for (var x = GAP / 2; x < width; x += GAP) {
        var px = Math.round(x);
        var py = Math.round(y);
        if (px < 0 || py < 0 || px >= maskW || py >= maskH) continue;
        var idx = (py * maskW + px) * 4;
        var inText = mask[idx + 3] > 128;

        dots.push({
          tx: x,
          ty: y,
          r: inText ? DOT_R_TEXT : DOT_R_BG,
          alpha: inText ? DOT_A_TEXT : DOT_A_BG,
          alive: true,
          respawnAt: 0,
          spawnAt: 0,
        });
      }
    }

    // 吃豆人初始位置：中心偏左上
    pacman.x = width * 0.5;
    pacman.y = height * 0.5;
    pacman.angle = Math.random() * Math.PI * 2;

    // 初始漫游目标
    pickWanderTarget(performance.now());

    root.classList.add('hero-ready');
  }

  /* ---------- 吃豆人 ---------- */

  function pickWanderTarget(now) {
    var margin = 60;
    pacman.wanderX = margin + Math.random() * Math.max(1, width - margin * 2);
    pacman.wanderY = margin + Math.random() * Math.max(1, height - margin * 2);
    pacman.wanderUntil = now + WANDER_MIN_MS + Math.random() * (WANDER_MAX_MS - WANDER_MIN_MS);
  }

  function updatePacman(dt, now) {
    var targetX, targetY;

    var mouseActive = mouse.active && (now - mouse.lastMove) < MOUSE_IDLE_MS;

    if (mouseActive) {
      targetX = mouse.x;
      targetY = mouse.y;
    } else {
      if (now > pacman.wanderUntil) pickWanderTarget(now);
      targetX = pacman.wanderX;
      targetY = pacman.wanderY;
    }

    var dx = targetX - pacman.x;
    var dy = targetY - pacman.y;
    var dist = Math.sqrt(dx * dx + dy * dy);

    // 转向
    if (dist > 2) {
      var desired = Math.atan2(dy, dx);
      var diff = desired - pacman.angle;
      while (diff > Math.PI) diff -= Math.PI * 2;
      while (diff < -Math.PI) diff += Math.PI * 2;

      var maxTurn = PACMAN_TURN_RATE * dt / 1000;
      if (Math.abs(diff) < maxTurn) pacman.angle = desired;
      else pacman.angle += Math.sign(diff) * maxTurn;
    }

    // 前进
    var speed = PACMAN_SPEED * pacman.boost;
    // 距目标很近时减速，避免抖动
    if (dist < 40) speed *= dist / 40;
    pacman.x += Math.cos(pacman.angle) * speed * dt / 1000;
    pacman.y += Math.sin(pacman.angle) * speed * dt / 1000;

    // 冲刺衰减
    if (pacman.boost > 1) {
      pacman.boost = Math.max(1, pacman.boost - dt / 400);
    }

    // 边界回绕
    if (pacman.x < -PACMAN_RADIUS) pacman.x = width + PACMAN_RADIUS;
    if (pacman.x > width + PACMAN_RADIUS) pacman.x = -PACMAN_RADIUS;
    if (pacman.y < -PACMAN_RADIUS) pacman.y = height + PACMAN_RADIUS;
    if (pacman.y > height + PACMAN_RADIUS) pacman.y = -PACMAN_RADIUS;

    // 嘴张合
    pacman.mouthPhase += PACMAN_MOUTH_SPEED * dt / 1000;
  }

  /* ---------- 吃豆 ---------- */

  function eatDots(now) {
    var eatR = PACMAN_RADIUS;
    var eatR2 = eatR * eatR;

    for (var i = 0; i < dots.length; i++) {
      var d = dots[i];
      if (!d.alive) continue;

      var dx = d.tx - pacman.x;
      var dy = d.ty - pacman.y;
      var dist2 = dx * dx + dy * dy;

      if (dist2 < eatR2) {
        d.alive = false;
        d.respawnAt = now + RESPAWN_MIN + Math.random() * (RESPAWN_MAX - RESPAWN_MIN);
      }
    }
  }

  function updateRespawns(now) {
    for (var i = 0; i < dots.length; i++) {
      var d = dots[i];
      if (!d.alive && now >= d.respawnAt) {
        d.alive = true;
        d.spawnAt = now;
      }
    }
  }

  /* ---------- 绘制 ---------- */

  function drawDots(now) {
    for (var i = 0; i < dots.length; i++) {
      var d = dots[i];
      if (!d.alive) continue;

      var alpha = d.alpha;
      var age = now - d.spawnAt;
      if (d.spawnAt > 0 && age < RESPAWN_FADE) {
        alpha *= age / RESPAWN_FADE;
      }

      ctx.globalAlpha = alpha;
      ctx.fillStyle = colors.fg;
      ctx.beginPath();
      ctx.arc(d.tx, d.ty, d.r, 0, Math.PI * 2);
      ctx.fill();
    }
    ctx.globalAlpha = 1;
  }

  function drawPacman() {
    var mouthOpen = (Math.sin(pacman.mouthPhase) + 1) / 2; // 0..1
    var mouthAngle = 0.12 + mouthOpen * (Math.PI / 4 - 0.12);

    ctx.save();
    ctx.translate(pacman.x, pacman.y);
    ctx.rotate(pacman.angle);

    ctx.fillStyle = colors.accent;
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.arc(0, 0, PACMAN_RADIUS, mouthAngle, Math.PI * 2 - mouthAngle);
    ctx.closePath();
    ctx.fill();

    ctx.restore();
  }

  function draw(now) {
    ctx.clearRect(0, 0, width, height);
    drawDots(now);
    drawPacman();
  }

  /* ---------- 循环 ---------- */

  function loop(now) {
    if (!running) return;

    var dt = Math.min(50, now - lastTime);
    lastTime = now;

    updatePacman(dt, now);
    eatDots(now);
    updateRespawns(now);
    draw(now);

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

  function drawStatic() {
    ctx.clearRect(0, 0, width, height);
    for (var i = 0; i < dots.length; i++) {
      var d = dots[i];
      ctx.globalAlpha = d.alpha;
      ctx.fillStyle = colors.fg;
      ctx.beginPath();
      ctx.arc(d.tx, d.ty, d.r, 0, Math.PI * 2);
      ctx.fill();
    }
    ctx.globalAlpha = 1;
  }

  /* ---------- 事件 ---------- */

  function relPos(clientX, clientY) {
    var rect = canvas.getBoundingClientRect();
    return { x: clientX - rect.left, y: clientY - rect.top };
  }

  function onMouseMove(e) {
    var p = relPos(e.clientX, e.clientY);
    mouse.x = p.x;
    mouse.y = p.y;
    mouse.active = true;
    mouse.lastMove = performance.now();
  }

  function onMouseLeave() {
    mouse.active = false;
  }

  function onMouseDown(e) {
    var p = relPos(e.clientX, e.clientY);
    mouse.x = p.x;
    mouse.y = p.y;
    mouse.active = true;
    mouse.lastMove = performance.now();
    pacman.boost = PACMAN_BOOST;
  }

  function onTouchStart(e) {
    if (!e.touches.length) return;
    var p = relPos(e.touches[0].clientX, e.touches[0].clientY);
    mouse.x = p.x;
    mouse.y = p.y;
    mouse.active = true;
    mouse.lastMove = performance.now();
    pacman.boost = PACMAN_BOOST;
  }

  function onTouchMove(e) {
    if (!e.touches.length) return;
    var p = relPos(e.touches[0].clientX, e.touches[0].clientY);
    mouse.x = p.x;
    mouse.y = p.y;
    mouse.active = true;
    mouse.lastMove = performance.now();
  }

  function onTouchEnd() {
    mouse.active = false;
  }

  /* ---------- 响应式 ---------- */

  var resizeTimer = null;
  function onResize() {
    if (resizeTimer) clearTimeout(resizeTimer);
    resizeTimer = setTimeout(function () {
      build();
      if (reduceMotion) drawStatic();
    }, 150);
  }

  function onVisibility() {
    if (document.hidden) stop();
    else if (!reduceMotion) start();
  }

  function onThemeChange() {
    readColors();
    if (reduceMotion) drawStatic();
  }

  /* ---------- 滚动淡出 ---------- */

  var ticking = false;
  function onScroll() {
    if (ticking) return;
    ticking = true;
    requestAnimationFrame(function () {
      var vh = window.innerHeight;
      var scrolled = window.scrollY;
      var progress = Math.min(1, scrolled / (vh * 0.6));
      var opacity = String(1 - progress);
      canvas.style.opacity = opacity;
      content.style.opacity = opacity;
      ticking = false;
    });
  }

  /* ---------- 初始化 ---------- */

  function init() {
    readColors();

    var ready = (document.fonts && document.fonts.ready)
      ? document.fonts.ready
      : Promise.resolve();

    ready.then(function () {
      build();

      if (reduceMotion) {
        drawStatic();
      } else {
        start();

        canvas.addEventListener('mousemove', onMouseMove);
        canvas.addEventListener('mouseleave', onMouseLeave);
        canvas.addEventListener('mousedown', onMouseDown);
        canvas.addEventListener('touchstart', onTouchStart, { passive: true });
        canvas.addEventListener('touchmove', onTouchMove, { passive: true });
        canvas.addEventListener('touchend', onTouchEnd, { passive: true });
      }

      window.addEventListener('resize', onResize);
      document.addEventListener('visibilitychange', onVisibility);
      window.addEventListener('scroll', onScroll, { passive: true });

      new MutationObserver(function (mutations) {
        for (var i = 0; i < mutations.length; i++) {
          if (mutations[i].attributeName === 'data-theme') {
            onThemeChange();
            return;
          }
        }
      }).observe(root, { attributes: true, attributeFilter: ['data-theme'] });
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();