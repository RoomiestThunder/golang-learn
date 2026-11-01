'use strict';

const $ = (sel, root = document) => root.querySelector(sel);
const el = (tag, cls, html) => { const e = document.createElement(tag); if (cls) e.className = cls; if (html != null) e.innerHTML = html; return e; };
const state = { course: null, progress: null, cards: null };
const LEVELS = { easy: 'Easy', medium: 'Medium', hard: 'Hard', 1: 'Лёгкая', 2: 'Средняя', 3: 'Сложная' };

async function api(path, method = 'GET', body) {
  const opt = { method, headers: { 'X-Requested-With': 'golearn' } };
  if (body !== undefined) { opt.headers['Content-Type'] = 'application/json'; opt.body = JSON.stringify(body); }
  const r = await fetch('/api' + path, opt);
  if (r.status === 401) { location.href = '/login'; throw new Error('нужен вход'); }
  const data = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(data.error || r.statusText);
  return data;
}

// ---------- статистика ----------
function stats() {
  const { course, progress: p } = state;
  const lessons = course.lessons;
  const lessonsDone = lessons.filter((l) => p.lessons[l.id]).length;
  const taskIds = Object.keys(course.tasks);
  const tasksDone = taskIds.filter((t) => p.tasks[t]?.solved).length;
  const lcAll = new Set(lessons.flatMap((l) => l.lc));
  let lcDone = 0;
  lcAll.forEach((s) => { if (p.lc[s]?.done) lcDone++; });
  const frac = (a, b) => (b ? a / b : 0);
  const overall = Math.round(100 * (0.4 * frac(lessonsDone, lessons.length) + 0.4 * frac(tasksDone, taskIds.length) + 0.2 * frac(lcDone, lcAll.size)));
  // серия дней
  let streak = 0;
  const d = new Date();
  const key = (dt) => `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`;
  if (!p.days[key(d)]) d.setDate(d.getDate() - 1);
  while (p.days[key(d)]) { streak++; d.setDate(d.getDate() - 1); }
  return { lessonsDone, lessonsTotal: lessons.length, tasksDone, tasksTotal: taskIds.length, lcDone, lcTotal: lcAll.size, overall, streak };
}

function lessonState(l) {
  const p = state.progress;
  if (p.lessons[l.id]) return 'done';
  const solved = l.tasks.filter((t) => p.tasks[t]?.solved).length;
  return solved > 0 ? 'part' : 'new';
}

function nextLesson() {
  return state.course.lessons.find((l) => !state.progress.lessons[l.id]) || null;
}

// ---------- боковая панель ----------
function renderSide() {
  const side = $('#side');
  const cur = (location.hash.match(/^#\/lesson\/([^/]+)/) || [])[1];
  side.innerHTML = '';
  for (const w of state.course.weeks) {
    const ls = state.course.lessons.filter((l) => l.week === w.week);
    const done = ls.filter((l) => state.progress.lessons[l.id]).length;
    const det = el('details', 'wk');
    det.open = ls.some((l) => l.id === cur) || (!cur && ls.some((l) => !state.progress.lessons[l.id]) && state.course.lessons.find((l) => !state.progress.lessons[l.id])?.week === w.week);
    det.innerHTML = `<summary><b>Неделя ${w.week}</b> <span class="muted">${done}/${ls.length}</span><div class="wt">${esc(w.title)}</div></summary>`;
    for (const l of ls) {
      const st = lessonState(l);
      const a = el('a', 'ls' + (l.id === cur ? ' cur' : ''), `<i class="dot ${st}"></i><span>${esc(l.title)}</span>`);
      a.href = `#/lesson/${l.id}`;
      det.appendChild(a);
    }
    side.appendChild(det);
  }
  $('#overallBar').style.width = stats().overall + '%';
  $('#overallPct').textContent = stats().overall + '%';
}

// ---------- роутер ----------
async function route() {
  document.body.classList.remove('menu-open');
  const [, kind, a, b] = location.hash.replace(/^#/, '').split('/');
  document.querySelectorAll('.topnav a').forEach((x) => x.classList.remove('active'));
  const main = $('#main');
  window.scrollTo(0, 0);
  try {
    if (!kind) { $('[data-nav=home]').classList.add('active'); viewHome(main); }
    else if (kind === 'lesson') await viewLesson(main, a, b || 'theory');
    else if (kind === 'cards') { $('[data-nav=cards]').classList.add('active'); await viewCards(main); }
    else if (kind === 'page') { $(`[data-nav=${a}]`)?.classList.add('active'); await viewPage(main, a); }
    else main.innerHTML = '<p>Страница не найдена</p>';
  } catch (e) {
    main.innerHTML = `<div class="callout warning"><div class="ct">Ошибка</div><p>${esc(e.message)}</p></div>`;
  }
  renderSide();
}

// ---------- главная ----------
function viewHome(main) {
  const s = stats();
  const nl = nextLesson();
  const p = state.progress;
  const started = new Date(p.started + 'T00:00:00');
  const elapsed = Math.max(0, Math.floor((Date.now() - started) / 864e5));
  const planned = Math.min(s.lessonsTotal, Math.floor((elapsed * 5) / 7));
  const diff = s.lessonsDone - planned;
  const pace = diff > 0 ? `Опережение на ${diff} ур. — отлично.` : diff === 0 ? 'Идёшь по плану: 5 уроков в неделю.' : `Отставание на ${-diff} ур. от плана «5 уроков в неделю». Это нормально — главное не прерывать серию.`;

  main.innerHTML = `
  <section class="hero">
    <div>
      <h1>Путь от повторения до Middle Go-разработчика</h1>
      <p class="lead">8 недель · ~2 часа в день · ${s.lessonsTotal} уроков. Каждый день: теория → практика с автопроверкой → задачи LeetCode. Всё открывается прямо здесь.</p>
      ${nl ? `<a class="btn primary big" href="#/lesson/${nl.id}">${s.lessonsDone ? 'Продолжить' : 'Начать'}: ${esc(nl.title)} →</a>` : '<div class="callout tip"><div class="ct">Курс пройден 🎉</div><p>Закрепляй карточки и повторяй слабые темы.</p></div>'}
    </div>
    <div class="ring" style="--p:${s.overall}"><b>${s.overall}%</b><span>курс</span></div>
  </section>
  <section class="stats">
    <div class="stat"><b>${s.lessonsDone}/${s.lessonsTotal}</b><span>уроков</span></div>
    <div class="stat"><b>${s.tasksDone}/${s.tasksTotal}</b><span>заданий с автопроверкой</span></div>
    <div class="stat"><b>${s.lcDone}/${s.lcTotal}</b><span>задач LeetCode</span></div>
    <div class="stat"><b>${s.streak}</b><span>дней подряд</span></div>
  </section>
  <p class="muted pace">${pace}</p>
  <section class="card" id="lcbox"></section>
  <h2>Программа</h2>
  <div id="weeks"></div>
  <details class="card howto"><summary><b>Как устроен день</b></summary>
    <ol>
      <li><b>Теория (30–40 мин)</b> — читай, правь и запускай примеры прямо в тексте (кнопка ▶).</li>
      <li><b>Практика (45–60 мин)</b> — задания с тестами. Пиши решение, жми «Проверить» (Ctrl/⌘+Enter). Решение смотри в последнюю очередь.</li>
      <li><b>LeetCode (20–30 мин)</b> — 2–4 задачи по теме урока. Решай на leetcode.com на Go; платформа сама подхватит принятые решения по нику.</li>
      <li><b>Завершить урок</b> и повторить 5–10 карточек для собеседования.</li>
    </ol>
  </details>`;
  renderLCBox($('#lcbox'));
  const wrap = $('#weeks');
  for (const w of state.course.weeks) {
    const ls = state.course.lessons.filter((l) => l.week === w.week);
    const done = ls.filter((l) => state.progress.lessons[l.id]).length;
    const c = el('div', 'card week');
    c.innerHTML = `<div class="wh"><div><h3>Неделя ${w.week}. ${esc(w.title)}</h3><p class="muted">${esc(w.goal)}</p></div><div class="wp"><div class="bar"><i style="width:${(100 * done) / ls.length}%"></i></div><span>${done}/${ls.length}</span></div></div>`;
    const list = el('div', 'lessons');
    for (const l of ls) {
      const a = el('a', 'lrow', `<i class="dot ${lessonState(l)}"></i><span class="d">День ${l.day}</span><span class="t">${esc(l.title)}</span><span class="muted m">${l.tasks.length} зад. · ${l.lc.length} LC · ${l.minutes} мин</span>`);
      a.href = `#/lesson/${l.id}`;
      list.appendChild(a);
    }
    c.appendChild(list);
    wrap.appendChild(c);
  }
}

function renderLCBox(box) {
  const p = state.progress;
  const s = stats();
  box.innerHTML = `<h3>Синхронизация с LeetCode</h3>
    <p class="muted">Реши задачу на leetcode.com (аккаунт и история сохраняются там). Укажи свой ник — платформа читает публичный профиль и отмечает принятые решения (раз в 3 мин, пока открыта). Можно отмечать и вручную.</p>
    <form class="row" id="lcform"><input id="lcuser" placeholder="ник на LeetCode" value="${esc(p.lcUser || '')}" autocomplete="off"><button class="btn" type="submit">Синхронизировать</button><span id="lcmsg" class="muted">${p.lcSynced ? 'Последний синк: ' + new Date(p.lcSynced).toLocaleString('ru') : ''}</span></form>
    <p class="muted small">Ограничение LeetCode: публично отдаются только последние ~20–50 принятых решений, поэтому синхронизируйся чаще, чем решаешь 20 задач. Решено в курсе: ${s.lcDone} из ${s.lcTotal}.</p>`;
  $('#lcform', box).onsubmit = async (e) => {
    e.preventDefault();
    const msg = $('#lcmsg', box);
    msg.textContent = 'Запрашиваю…';
    try {
      const r = await api('/lc/sync', 'POST', { username: $('#lcuser', box).value });
      state.progress = r.progress;
      renderSide();
      msg.textContent = r.added ? `Готово: +${r.added} новых решений` : 'Готово: новых решений нет';
      setTimeout(route, 900);
    } catch (err) { msg.textContent = '⚠ ' + err.message; }
  };
}

// ---------- урок ----------
async function viewLesson(main, id, tab) {
  const data = await api('/lesson/' + id);
  const l = data.lesson;
  const p = state.progress;
  const lessons = state.course.lessons;
  const idx = lessons.findIndex((x) => x.id === id);
  const prev = lessons[idx - 1], next = lessons[idx + 1];
  const week = state.course.weeks.find((w) => w.week === l.week);
  const done = !!p.lessons[id];
  const solved = data.tasks.filter((t) => p.tasks[t.id]?.solved).length;
  const lcDone = l.lc.filter((x) => p.lc[x.slug]?.done).length;

  main.innerHTML = `
  <div class="crumbs"><a href="#/">Путь</a> / Неделя ${l.week}. ${esc(week?.title || '')} / День ${l.day}</div>
  <h1>${esc(l.title)}</h1>
  <p class="muted">⏱ ~${l.minutes} мин · ${esc(l.goals)}</p>
  <div class="tabs">
    <a class="${tab === 'theory' ? 'on' : ''}" href="#/lesson/${id}/theory">📖 Теория</a>
    <a class="${tab === 'practice' ? 'on' : ''}" href="#/lesson/${id}/practice">⌨ Практика <span class="pill">${solved}/${data.tasks.length}</span></a>
    <a class="${tab === 'lc' ? 'on' : ''}" href="#/lesson/${id}/lc">🧩 LeetCode <span class="pill">${lcDone}/${l.lc.length}</span></a>
  </div>
  <div id="tabbody"></div>
  <div class="lfoot">
    ${prev ? `<a class="btn" href="#/lesson/${prev.id}">← ${esc(prev.title)}</a>` : '<span></span>'}
    <button class="btn ${done ? '' : 'primary'}" id="donebtn">${done ? '✓ Урок завершён (снять)' : 'Завершить урок'}</button>
    ${next ? `<a class="btn" href="#/lesson/${next.id}">${esc(next.title)} →</a>` : '<span></span>'}
  </div>`;
  const body = $('#tabbody');
  if (tab === 'theory') {
    body.className = 'prose';
    body.innerHTML = renderMarkdown(l.body);
    hydrate(body);
    const go = el('div', 'nextstep', `<b>Дальше:</b> <a href="#/lesson/${id}/practice">перейти к практике (${data.tasks.length} зад.) →</a>`);
    body.appendChild(go);
  } else if (tab === 'practice') {
    if (!data.tasks.length) body.innerHTML = '<p class="muted">В этом уроке нет заданий с автопроверкой — переходи к LeetCode.</p>';
    data.tasks.forEach((t, i) => body.appendChild(taskCard(t, i + 1)));
  } else if (tab === 'lc') {
    body.appendChild(lcTab(l));
  }
  $('#donebtn').onclick = async () => {
    state.progress = await api(`/lesson/${id}/done`, 'POST', { done: !done });
    if (!done && next) location.hash = `#/lesson/${next.id}`; else route();
  };
}

function lcTab(l) {
  const wrap = el('div');
  const info = el('div', 'card');
  renderLCBox(info);
  wrap.appendChild(info);
  if (!l.lc.length) { wrap.appendChild(el('p', 'muted', 'Для этого урока задач нет.')); return wrap; }
  const t = el('div', 'lctable');
  for (const x of l.lc) {
    const st = state.progress.lc[x.slug];
    const row = el('div', 'lcrow' + (st?.done ? ' done' : ''));
    row.innerHTML = `<label class="chk"><input type="checkbox" ${st?.done ? 'checked' : ''}><span></span></label>
      <div class="lci"><a href="https://leetcode.com/problems/${x.slug}/" target="_blank" rel="noopener">${esc(x.title)} ↗</a>
      <span class="lvl ${x.level.toLowerCase()}">${esc(x.level)}</span>${st?.done ? `<span class="src">${st.source === 'auto' ? 'засчитано автоматически' : 'отмечено вручную'}</span>` : ''}
      <div class="muted">${esc(x.hint)}</div></div>`;
    $('input', row).onchange = async (e) => {
      state.progress = await api('/lc/toggle', 'POST', { slug: x.slug, done: e.target.checked });
      route();
    };
    t.appendChild(row);
  }
  wrap.appendChild(t);
  wrap.appendChild(el('p', 'muted small', 'Решай на Go. После статуса Accepted вернись сюда — платформа подхватит задачу по нику (или отметь вручную). Если задача «Premium», пропусти её или возьми похожую из программы.'));
  return wrap;
}

// ---------- редактор ----------
function createEditor(code, { onInput, onRun }) {
  const wrap = el('div', 'editor');
  const pre = el('pre', 'hl');
  const ta = el('textarea');
  ta.spellcheck = false; ta.wrap = 'off'; ta.autocapitalize = 'off'; ta.setAttribute('autocorrect', 'off');
  ta.value = code;
  const sync = () => { pre.innerHTML = highlightGo(ta.value) + '\n'; };
  ta.addEventListener('input', () => { sync(); onInput && onInput(ta.value); });
  ta.addEventListener('keydown', (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') { e.preventDefault(); onRun && onRun(); return; }
    const ins = (txt) => { if (!document.execCommand('insertText', false, txt)) { ta.setRangeText(txt, ta.selectionStart, ta.selectionEnd, 'end'); ta.dispatchEvent(new Event('input')); } };
    if (e.key === 'Tab' && !e.shiftKey) { e.preventDefault(); ins('\t'); }
    else if (e.key === 'Enter' && !e.shiftKey) {
      const before = ta.value.slice(0, ta.selectionStart);
      const line = before.slice(before.lastIndexOf('\n') + 1);
      let indent = line.match(/^[\t ]*/)[0];
      if (/[{(]\s*$/.test(line)) indent += '\t';
      e.preventDefault(); ins('\n' + indent);
    }
  });
  wrap.append(pre, ta);
  sync();
  return { el: wrap, get value() { return ta.value; }, set value(v) { ta.value = v; sync(); onInput && onInput(v); }, focus: () => ta.focus() };
}

function showResult(out, res, kind) {
  out.className = 'out ' + (res.ok ? 'ok' : 'bad');
  let head;
  if (kind === 'check') head = res.ok ? '✅ Все тесты пройдены' : (res.timedOut ? '⏱ Время вышло' : '❌ Есть ошибки');
  else head = res.ok ? '✓ Программа завершилась' : (res.timedOut ? '⏱ Время вышло' : '✗ Ошибка');
  out.innerHTML = `<div class="oh">${head} <span class="muted">${res.ms} мс</span></div><pre>${esc(res.output || '(пустой вывод)')}</pre>`;
}

function hydrate(root) {
  root.querySelectorAll('.cb').forEach((div) => {
    const code = div.dataset.code, lang = div.dataset.lang;
    div.removeAttribute('data-code');
    if (div.dataset.run === 'true') {
      div.classList.add('runnable');
      const bar = el('div', 'cbar', `<span>${esc(lang || 'go')} · пример: можно править и запускать</span>`);
      const run = el('button', 'btn small primary', '▶ Запустить');
      const reset = el('button', 'btn small', '↺');
      reset.title = 'Вернуть исходный код';
      bar.append(reset, run);
      const out = el('div', 'out hidden');
      const doRun = async () => {
        run.disabled = true; run.textContent = 'Выполняется…';
        out.className = 'out'; out.innerHTML = '<div class="oh">Компиляция и запуск…</div>';
        try { showResult(out, await api('/run', 'POST', { code: ed.value }), 'run'); } catch (e) { out.className = 'out bad'; out.innerHTML = `<pre>${esc(e.message)}</pre>`; }
        run.disabled = false; run.textContent = '▶ Запустить';
      };
      const ed = createEditor(code, { onRun: doRun });
      run.onclick = doRun; reset.onclick = () => { ed.value = code; };
      div.append(bar, ed.el, out);
    } else {
      const pre = el('pre', 'static');
      const c = el('code');
      c.innerHTML = lang === 'go' || lang === '' ? highlightGo(code) : esc(code);
      pre.appendChild(c);
      const btn = el('button', 'copy', 'копировать');
      btn.onclick = () => { navigator.clipboard?.writeText(code); btn.textContent = 'скопировано'; setTimeout(() => (btn.textContent = 'копировать'), 1200); };
      div.append(pre, btn);
    }
  });
  root.querySelectorAll('input[data-check]').forEach((cb) => {
    cb.onchange = async () => { state.progress = await api('/check-item', 'POST', { key: cb.dataset.check, on: cb.checked }); };
  });
}

// ---------- задание ----------
function taskCard(t, n) {
  const st = state.progress.tasks[t.id] || {};
  const card = el('section', 'card task' + (st.solved ? ' solved' : ''));
  card.id = 'task-' + t.id;
  card.innerHTML = `<div class="th"><h2>Задание ${n}. ${esc(t.title)}</h2><span class="lvl ${t.level}">${LEVELS[t.level] || ''}</span><span class="badge" ${st.solved ? '' : 'hidden'}>решено ✓</span></div><div class="prose st"></div>`;
  const stEl = $('.st', card);
  stEl.innerHTML = renderMarkdown(t.statement);
  hydrate(stEl);

  const out = el('div', 'out hidden');
  let timer;
  const check = async () => {
    btnCheck.disabled = true; btnCheck.textContent = 'Проверяю…';
    out.className = 'out'; out.innerHTML = '<div class="oh">Компиляция и запуск тестов…</div>';
    try {
      const r = await api('/check/' + t.id, 'POST', { code: ed.value });
      state.progress = r.progress;
      showResult(out, r.result, 'check');
      if (r.result.ok) {
        card.classList.add('solved'); $('.badge', card).hidden = false; renderSide();
        const pill = document.querySelector('.tabs a[href$="/practice"] .pill');
        if (pill) pill.textContent = `${document.querySelectorAll('.task.solved').length}/${document.querySelectorAll('.task').length}`;
      }
    } catch (e) { out.className = 'out bad'; out.innerHTML = `<pre>${esc(e.message)}</pre>`; }
    btnCheck.disabled = false; btnCheck.textContent = '▶ Проверить';
  };
  const ed = createEditor(st.code || t.starter, {
    onRun: check,
    onInput: (v) => { clearTimeout(timer); timer = setTimeout(() => api('/draft/' + t.id, 'POST', { code: v }).catch(() => {}), 800); },
  });
  const bar = el('div', 'tbar');
  const btnCheck = el('button', 'btn primary', '▶ Проверить');
  btnCheck.onclick = check;
  const btnReset = el('button', 'btn', '↺ Сбросить');
  btnReset.onclick = () => { if (confirm('Вернуть исходный шаблон? Твой код будет потерян.')) ed.value = t.starter; };
  bar.append(btnCheck, el('span', 'muted small', 'Ctrl/⌘ + Enter'), btnReset);
  const extra = el('div', 'extra');
  const mk = (label, fn) => { const b = el('button', 'btn small ghost', label); b.onclick = fn; extra.appendChild(b); return b; };
  if (t.hint) mk('💡 Подсказка', () => { hintBox.classList.toggle('hidden'); });
  mk('🧪 Тесты', () => { testBox.classList.toggle('hidden'); });
  mk('👁 Решение', async () => {
    if (solBox.dataset.loaded) { solBox.classList.toggle('hidden'); return; }
    if (!st.solved && !confirm('Попробуй сначала сам: поиск решения — это и есть обучение. Показать готовое решение?')) return;
    const r = await api('/solution/' + t.id);
    solBox.innerHTML = '<div class="cbar"><span>Эталонное решение</span></div>';
    const pre = el('pre', 'static'); pre.innerHTML = '<code>' + highlightGo(r.solution) + '</code>';
    solBox.appendChild(pre); solBox.dataset.loaded = '1'; solBox.classList.remove('hidden');
  });
  const hintBox = el('div', 'callout tip hidden', '<div class="ct">Подсказка</div>' + renderMarkdown(t.hint || ''));
  const testBox = el('div', 'cb hidden'); testBox.innerHTML = '<div class="cbar"><span>Тесты, которыми проверяется решение</span></div>';
  const tp = el('pre', 'static'); tp.innerHTML = '<code>' + highlightGo(t.tests) + '</code>'; testBox.appendChild(tp);
  const solBox = el('div', 'cb hidden');
  card.append(ed.el, bar, extra, out, hintBox, testBox, solBox);
  return card;
}

// ---------- карточки ----------
async function viewCards(main) {
  if (!state.cards) state.cards = await api('/cards');
  const all = state.cards;
  const tags = [...new Set(all.map((c) => c.tag).filter(Boolean))];
  let tag = '';
  let queue = [], pos = 0, shown = false;
  const build = () => {
    const known = (c) => (state.progress.cards[c.id] || 0) >= 3;
    queue = all.filter((c) => (!tag || c.tag === tag) && !known(c));
    for (let i = queue.length - 1; i > 0; i--) { const j = Math.floor(Math.random() * (i + 1)); [queue[i], queue[j]] = [queue[j], queue[i]]; }
    pos = 0; shown = false;
  };
  const draw = () => {
    const total = all.filter((c) => !tag || c.tag === tag).length;
    const learned = total - queue.length;
    const head = `<h1>Карточки для собеседования</h1><p class="muted">Вопрос → вспомни ответ вслух → открой → оцени себя. «Знаю» убирает карточку из очереди, остальные вернутся.</p>
      <div class="row"><select id="tagsel"><option value="">Все темы</option>${tags.map((t) => `<option ${t === tag ? 'selected' : ''}>${esc(t)}</option>`).join('')}</select><span class="muted">Выучено ${learned} из ${total}</span></div>`;
    if (!queue.length) { main.innerHTML = head + '<div class="callout tip"><div class="ct">Очередь пуста 🎉</div><p>Все карточки по выбранной теме выучены.</p></div>'; }
    else {
      const c = queue[pos % queue.length];
      main.innerHTML = head + `<div class="fc"><div class="fq"><span class="tag">${esc(c.tag)}</span>${inline(c.q)}</div>
        <div class="fa prose ${shown ? '' : 'hidden'}"></div>
        <div class="fbtn">${shown ? '<button class="btn" data-b="0">Не знаю</button><button class="btn" data-b="1">Сомневаюсь</button><button class="btn primary" data-b="3">Знаю</button>' : '<button class="btn primary" data-show>Показать ответ</button>'}</div></div>`;
      if (shown) { const a = $('.fa', main); a.innerHTML = renderMarkdown(c.a); hydrate(a); }
      const sb = $('[data-show]', main);
      if (sb) sb.onclick = () => { shown = true; draw(); };
      main.querySelectorAll('[data-b]').forEach((b) => (b.onclick = async () => {
        const box = +b.dataset.b;
        state.progress = await api('/card', 'POST', { id: c.id, box });
        if (box >= 3) queue.splice(pos % queue.length, 1); else pos++;
        shown = false; draw(); renderSide();
      }));
    }
    $('#tagsel', main).onchange = (e) => { tag = e.target.value; build(); draw(); };
  };
  build(); draw();
}

async function viewPage(main, name) {
  const r = await api('/page/' + name);
  main.innerHTML = '<div class="prose" id="pg"></div>';
  const pg = $('#pg');
  const checks = {};
  Object.entries(state.progress.checks).forEach(([k, v]) => { checks[k] = v; });
  pg.innerHTML = renderMarkdown(r.body, { prefix: name, checks });
  hydrate(pg);
}

// ---------- запуск ----------
function initTheme() {
  const saved = localStorage.getItem('theme');
  if (saved) document.documentElement.dataset.theme = saved;
  $('#theme').onclick = () => {
    const dark = document.documentElement.dataset.theme === 'dark' || (!document.documentElement.dataset.theme && matchMedia('(prefers-color-scheme: dark)').matches);
    const nt = dark ? 'light' : 'dark';
    document.documentElement.dataset.theme = nt;
    localStorage.setItem('theme', nt);
  };
}

(async function init() {
  initTheme();
  $('#menu').onclick = () => document.body.classList.toggle('menu-open');
  try {
    [state.course, state.progress] = await Promise.all([api('/course'), api('/progress')]);
  } catch (e) {
    $('#main').innerHTML = `<div class="callout warning"><div class="ct">Не удалось загрузить курс</div><p>${esc(e.message)}</p></div>`;
    return;
  }
  window.addEventListener('hashchange', route);
  route();
})();
