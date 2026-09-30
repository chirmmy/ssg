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

  var STAR_DENSITY = 30000;     // 平均每多少 px² 一颗星星
  var STAR_COUNT_MIN = 14;      // 星星数量下限
  var STAR_COUNT_MAX = 55;      // 星星数量上限
  var STAR_R_MIN = 1.8;         // 星星最小半径
  var STAR_R_MAX = 4.6;         // 星星最大半径
  var STAR_BIG_RATIO = 0.22;    // 大星星（带光晕）比例
  var STAR_ACCENT_RATIO = 0.35; // accent 色星星比例
  var STAR_TWINKLE_MIN = 0.0012; // 闪烁最快频率系数
  var STAR_TWINKLE_MAX = 0.0032; // 闪烁最慢频率系数

  var PACMAN_RADIUS = 12;       // 吃豆人半径
  var PACMAN_SPEED = 210;       // px/秒
  var PACMAN_BOOST = 2.2;       // 点击冲刺倍率
  var PACMAN_TURN_RATE = 6;     // 转向速度（弧度/秒）
  var PACMAN_MOUTH_SPEED = 14;  // 嘴张合频率

  var RESPAWN_MIN = 2600;       // 星星重生最短时间（ms）
  var RESPAWN_MAX = 5200;       // 星星重生最长时间（ms）
  var RESPAWN_FADE = 480;       // 重生淡入时长（ms）

  var MOUSE_IDLE_MS = 2200;     // 鼠标停止多久后吃豆人开始漫游
  var WANDER_MIN_MS = 1800;     // 漫游目标最短停留
  var WANDER_MAX_MS = 3600;     // 漫游目标最长停留

  /* ---------- 状态 ---------- */

  var width = 0;
  var height = 0;
  var dpr = Math.min(window.devicePixelRatio || 1, 2);

  var stars = [];       // { tx, ty, r, rot, rotSpeed, phase, speed, color, big, alive, respawnAt, spawnAt }
  var colors = { fg: '#111111', accent: '#2563eb', bg: '#ffffff' };
  var colorsRgb = { fg: [17, 17, 17], accent: [37, 99, 235], bg: [255, 255, 255] };

  // 吃星星时的闪光粒子
  var particles = [];
  var PARTICLE_LIFE = 450;
  var MAX_PARTICLES = 60;

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

  function hexToRgb(hex) {
    var m = /^#?([0-9a-f]{3}|[0-9a-f]{6})$/i.exec((hex || '').trim());
    if (!m) return null;
    var h = m[1];
    if (h.length === 3) h = h[0] + h[0] + h[1] + h[1] + h[2] + h[2];
    var n = parseInt(h, 16);
    if (isNaN(n)) return null;
    return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
  }

  function readColors() {
    var cs = getComputedStyle(root);
    colors.fg = cs.getPropertyValue('--color-fg').trim() || '#111111';
    colors.accent = cs.getPropertyValue('--color-accent').trim() || '#2563eb';
    colors.bg = cs.getPropertyValue('--color-bg').trim() || '#ffffff';
    colorsRgb.fg = hexToRgb(colors.fg) || [17, 17, 17];
    colorsRgb.accent = hexToRgb(colors.accent) || [37, 99, 235];
    colorsRgb.bg = hexToRgb(colors.bg) || [255, 255, 255];
  }

  /* ---------- 星星配色 ---------- */

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
        rotSpeed: (Math.random() - 0.5) * 0.0006,  // 大部分近乎静止，少数缓慢自转
        phase: Math.random() * Math.PI * 2,
        speed: STAR_TWINKLE_MIN + Math.random() * (STAR_TWINKLE_MAX - STAR_TWINKLE_MIN),
        accent: Math.random() < STAR_ACCENT_RATIO,
        big: big,
        alive: true,
        respawnAt: 0,
        spawnAt: 0,
      });
    }

    applyStarColors();

    // 吃豆人初始位置：中心偏左上
    pacman.x = width * 0.5;
    pacman.y = height * 0.5;
    pacman.angle = Math.random() * Math.PI * 2;

    // 初始漫游目标
    pickWanderTarget(performance.now());
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

  /* ---------- 吃星星 ---------- */

  function eatStars(now) {
    var eatR = PACMAN_RADIUS + 4;
    var eatR2 = eatR * eatR;

    for (var i = 0; i < stars.length; i++) {
      var s = stars[i];
      if (!s.alive) continue;

      var dx = s.tx - pacman.x;
      var dy = s.ty - pacman.y;
      var dist2 = dx * dx + dy * dy;

      if (dist2 < eatR2) {
        s.alive = false;
        s.respawnAt = now + RESPAWN_MIN + Math.random() * (RESPAWN_MAX - RESPAWN_MIN);
        spawnParticles(s.tx, s.ty, s.color);
      }
    }
  }

  /* ---------- 吃豆粒子 ---------- */

  function spawnParticles(x, y, color) {
    particles.push({ x: x, y: y, age: 0, life: PARTICLE_LIFE, color: color });
    if (particles.length > MAX_PARTICLES) particles.shift();
  }

  function updateParticles(dt) {
    for (var i = particles.length - 1; i >= 0; i--) {
      particles[i].age += dt;
      if (particles[i].age >= particles[i].life) particles.splice(i, 1);
    }
  }

  function drawParticles() {
    for (var i = 0; i < particles.length; i++) {
      var p = particles[i];
      var t = p.age / p.life;
      ctx.globalAlpha = (1 - t) * 0.55;
      ctx.strokeStyle = p.color;
      ctx.lineWidth = 1.5;
      ctx.beginPath();
      ctx.arc(p.x, p.y, 3 + 16 * t, 0, Math.PI * 2);
      ctx.stroke();
    }
    ctx.globalAlpha = 1;
  }

  function updateRespawns(now) {
    for (var i = 0; i < stars.length; i++) {
      var s = stars[i];
      if (!s.alive && now >= s.respawnAt) {
        s.alive = true;
        s.spawnAt = now;
      }
    }
  }

  /* ---------- 绘制 ---------- */

  // 四角星光路径（外尖内凹的菱形星芒）
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
    for (var i = 0; i < stars.length; i++) {
      var s = stars[i];
      if (!s.alive) continue;

      var age = now - s.spawnAt;
      var fadeIn = (s.spawnAt > 0 && age < RESPAWN_FADE) ? age / RESPAWN_FADE : 1;

      // 闪烁曲线：pow 让星星多数时间偏暗、偶尔骤亮，更像真实星光
      var tw = Math.sin(now * s.speed + s.phase);
      var twinkle = 0.3 + 0.7 * Math.pow((tw + 1) / 2, 2.2);

      var alpha = fadeIn * twinkle;
      var r = s.r * (0.82 + 0.18 * twinkle);
      var rot = s.rot + now * s.rotSpeed;

      ctx.save();
      ctx.globalAlpha = alpha;
      ctx.fillStyle = s.color;

      // 大星星带柔光光晕
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

  function drawPacman() {
    var mouthOpen = (Math.sin(pacman.mouthPhase) + 1) / 2; // 0..1
    var mouthAngle = 0.12 + mouthOpen * (Math.PI / 4 - 0.12);

    ctx.save();
    ctx.translate(pacman.x, pacman.y);
    ctx.rotate(pacman.angle);

    // 光晕
    ctx.shadowColor = colors.accent;
    ctx.shadowBlur = 18;

    ctx.fillStyle = colors.accent;
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.arc(0, 0, PACMAN_RADIUS, mouthAngle, Math.PI * 2 - mouthAngle);
    ctx.closePath();
    ctx.fill();

    ctx.shadowBlur = 0;

    // 眼睛
    ctx.fillStyle = colors.bg;
    ctx.beginPath();
    ctx.arc(PACMAN_RADIUS * 0.18, -PACMAN_RADIUS * 0.45, PACMAN_RADIUS * 0.16, 0, Math.PI * 2);
    ctx.fill();

    ctx.restore();
  }

  function draw(now) {
    ctx.clearRect(0, 0, width, height);
    drawStars(now);
    drawParticles();
    drawPacman();
  }

  /* ---------- 循环 ---------- */

  function loop(now) {
    if (!running) return;

    var dt = Math.min(50, now - lastTime);
    lastTime = now;

    updatePacman(dt, now);
    eatStars(now);
    updateRespawns(now);
    updateParticles(dt);
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
    for (var i = 0; i < stars.length; i++) {
      var s = stars[i];
      ctx.globalAlpha = 0.85;
      ctx.fillStyle = s.color;
      starPath(s.tx, s.ty, s.r, s.rot);
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
    applyStarColors();
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

  /* ---------- 描边标题（EMS Allure 单线手写笔迹，逐字书写） ---------- */

  var LETTER_STAGGER_MS = 150;   // 相邻字母起笔间隔
  var TITLE_TARGET_WIDTH = 1000; // 标题在 viewBox 中的目标宽度
  var TITLE_MAX_SCALE = 0.185;   // 缩放上限（防止过短标题被放得过大）
  var SCRIPT_BASELINE = 170;     // viewBox 中的基线位置

  function setupTitleStroke() {
    var svg = document.querySelector('[data-title-stroke]');
    var font = window.HERSHEY_SCRIPT;
    if (!svg || !font) return;
    var textEl = svg.querySelector('[data-stroke-text]');
    if (!textEl) return;

    var title = (textEl.textContent || '').trim();
    if (!title) return;

    // 总宽（原始单位）→ 求缩放比
    var total = 0;
    for (var i = 0; i < title.length; i++) {
      var gc = font.glyphs[title[i]];
      total += (gc && gc.d) ? gc.w : font.space;
    }
    var S = Math.min(TITLE_TARGET_WIDTH / total, TITLE_MAX_SCALE);

    var x = 40;
    var idx = 0;
    var frag = document.createDocumentFragment();
    var svgNS = 'http://www.w3.org/2000/svg';

    for (var j = 0; j < title.length; j++) {
      var ch = title[j];
      var g = font.glyphs[ch];
      if (!g || !g.d) { x += font.space * S; continue; }

      var p = document.createElementNS(svgNS, 'path');
      p.setAttribute('d', g.d);
      // 字体坐标 y 向上，翻转 y 使基线落在 SCRIPT_BASELINE
      p.setAttribute('transform',
        'translate(' + x.toFixed(2) + ',' + SCRIPT_BASELINE + ') scale(' + S + ',-' + S + ')');
      p.setAttribute('class', 'hero-title-stroke__path');
      p.setAttribute('pathLength', '1');
      p.style.animationDelay = (300 + idx * LETTER_STAGGER_MS) + 'ms';
      frag.appendChild(p);

      x += g.w * S;
      idx++;
    }

    // 按内容宽度收紧 viewBox
    var width = Math.ceil(x + 40);
    svg.setAttribute('viewBox', '0 0 ' + width + ' 220');

    // 渐变随内容宽度横向铺满整句
    var grad = svg.querySelector('#title-grad');
    if (grad) {
      grad.setAttribute('gradientUnits', 'userSpaceOnUse');
      grad.setAttribute('x2', width);
    }

    textEl.style.display = 'none';
    svg.appendChild(frag);
  }

  /* ---------- 开屏时隐藏 header ---------- */

  // Hero 占视口一半以上时隐藏 header，进入内容区后滑出显示。
  // 用 IntersectionObserver 而非 scroll 阈值：对 snap 吸附动画、键盘翻页同样有效
  function setupHeaderAutohide() {
    var header = document.querySelector('.site-header');
    var heroSection = canvas.closest('.hero-fullscreen');
    if (!header || !heroSection || !('IntersectionObserver' in window)) return;

    var io = new IntersectionObserver(function (entries) {
      var visible = entries[0].intersectionRatio >= 0.5;
      header.classList.toggle('site-header--hidden', visible);
    }, { threshold: [0, 0.25, 0.5, 0.75, 1] });

    io.observe(heroSection);
  }

  /* ---------- 初始化 ---------- */

  function init() {
    readColors();

    var ready = (document.fonts && document.fonts.ready)
      ? document.fonts.ready
      : Promise.resolve();

    ready.then(function () {
      setupTitleStroke();
      build();
      setupHeaderAutohide();

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

      // 主题切换期间（html.theme-anim）暂停星空绘制，避免与全页颜色过渡
      // 争抢主线程造成掉帧；过渡结束后读取最终颜色并恢复。
      new MutationObserver(function () {
        var animating = root.classList.contains('theme-anim');
        if (animating) {
          stop();
          return;
        }
        onThemeChange();
        start();
      }).observe(root, { attributes: true, attributeFilter: ['data-theme', 'class'] });
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();