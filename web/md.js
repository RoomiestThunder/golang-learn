// Минимальный markdown-рендерер + подсветка Go. Без внешних зависимостей (работает офлайн).
'use strict';

const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

const GO_KW = new Set('break case chan const continue default defer else fallthrough for func go goto if import interface map package range return select struct switch type var'.split(' '));
const GO_BI = new Set('bool byte rune string int int8 int16 int32 int64 uint uint8 uint16 uint32 uint64 uintptr float32 float64 complex64 complex128 error any comparable true false nil iota append cap clear close complex copy delete imag len make max min new panic print println real recover'.split(' '));

function highlightGo(code) {
  const re = /(\/\/[^\n]*|\/\*[\s\S]*?\*\/)|("(?:\\.|[^"\\\n])*"|`[^`]*`|'(?:\\.|[^'\\\n])+')|(\b0x[0-9a-fA-F_]+\b|\b\d[\d_]*(?:\.\d+)?(?:[eE][+-]?\d+)?\b)|([A-Za-z_]\w*)/g;
  let out = '', last = 0, m;
  while ((m = re.exec(code))) {
    out += esc(code.slice(last, m.index));
    last = re.lastIndex;
    const t = m[0];
    if (m[1]) out += '<span class="c">' + esc(t) + '</span>';
    else if (m[2]) out += '<span class="s">' + esc(t) + '</span>';
    else if (m[3]) out += '<span class="n">' + esc(t) + '</span>';
    else if (GO_KW.has(t)) out += '<span class="k">' + t + '</span>';
    else if (GO_BI.has(t)) out += '<span class="b">' + t + '</span>';
    else if (code[last] === '(') out += '<span class="f">' + t + '</span>';
    else if (/^[A-Z]/.test(t)) out += '<span class="t">' + t + '</span>';
    else out += t;
  }
  return out + esc(code.slice(last));
}

function inline(s) {
  const codes = [];
  s = s.replace(/`([^`]+)`/g, (m, c) => { codes.push(c); return '\u0000' + (codes.length - 1) + '\u0000'; });
  s = esc(s);
  s = s.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
       .replace(/(^|[^*\w])\*([^*\s][^*]*?)\*(?!\w)/g, '$1<em>$2</em>')
       .replace(/\[([^\]]+)\]\(((?:https?:|#)[^)\s]+)\)/g, (m, t, u) => u.startsWith('#') ? `<a href="${u}">${t}</a>` : `<a href="${u}" target="_blank" rel="noopener">${t}</a>`);
  return s.replace(/\u0000(\d+)\u0000/g, (m, i) => '<code>' + esc(codes[i]) + '</code>');
}

const isBlockStart = (l) => /^(#{1,4}\s|```|>|\s*([-*]|\d+\.)\s|\||---+$|<\/?(details|summary))/.test(l);

// opts.checks: {key: bool} для интерактивных чек-листов, opts.prefix — префикс ключей
function renderMarkdown(src, opts = {}) {
  const lines = src.replace(/\r\n/g, '\n').split('\n');
  const out = [];
  let i = 0, checkIdx = 0;
  const checks = opts.checks || {};

  while (i < lines.length) {
    let line = lines[i];
    if (!line.trim()) { i++; continue; }

    let m;
    if ((m = line.match(/^```(.*)$/))) {
      const info = m[1].trim().split(/\s+/).filter(Boolean);
      const buf = [];
      i++;
      while (i < lines.length && !/^```\s*$/.test(lines[i])) buf.push(lines[i++]);
      i++;
      const lang = info[0] || '';
      const run = info.includes('run');
      out.push(`<div class="cb" data-lang="${esc(lang)}" data-run="${run}" data-code="${esc(buf.join('\n'))}"></div>`);
      continue;
    }
    if (/^<\/?(details|summary)/.test(line)) { out.push(line); i++; continue; }
    if ((m = line.match(/^(#{1,4})\s+(.*)$/))) {
      const n = m[1].length;
      const id = m[2].toLowerCase().replace(/[^\p{L}\p{N}]+/gu, '-').replace(/^-|-$/g, '');
      out.push(`<h${n} id="${id}">${inline(m[2])}</h${n}>`);
      i++; continue;
    }
    if (/^---+$/.test(line)) { out.push('<hr>'); i++; continue; }
    if (line.startsWith('|') && i + 1 < lines.length && /^\|[\s:|-]+\|?\s*$/.test(lines[i + 1])) {
      const cells = (l) => l.replace(/^\||\|\s*$/g, '').split('|').map((c) => c.trim());
      const head = cells(line);
      i += 2;
      let html = '<div class="tw"><table><thead><tr>' + head.map((c) => `<th>${inline(c)}</th>`).join('') + '</tr></thead><tbody>';
      while (i < lines.length && lines[i].startsWith('|')) {
        html += '<tr>' + cells(lines[i]).map((c) => `<td>${inline(c)}</td>`).join('') + '</tr>';
        i++;
      }
      out.push(html + '</tbody></table></div>');
      continue;
    }
    if (line.startsWith('>')) {
      const buf = [];
      while (i < lines.length && lines[i].startsWith('>')) buf.push(lines[i++].replace(/^>\s?/, ''));
      let cls = '';
      const cm = buf[0].match(/^\[!(NOTE|TIP|WARNING|IMPORTANT|INTERVIEW)\]\s*(.*)$/);
      let titleHtml = '';
      if (cm) {
        cls = ' ' + cm[1].toLowerCase();
        const names = { NOTE: 'Заметка', TIP: 'Совет', WARNING: 'Осторожно', IMPORTANT: 'Важно', INTERVIEW: 'На собеседовании' };
        titleHtml = `<div class="ct">${names[cm[1]]}</div>`;
        buf[0] = cm[2];
      }
      out.push(`<blockquote class="callout${cls}">${titleHtml}${renderMarkdown(buf.join('\n'), opts)}</blockquote>`);
      continue;
    }
    if (/^\s*([-*]|\d+\.)\s+/.test(line)) {
      const items = [];
      while (i < lines.length && (/^\s*([-*]|\d+\.)\s+/.test(lines[i]) || (/^\s+\S/.test(lines[i]) && items.length))) {
        const lm = lines[i].match(/^(\s*)([-*]|\d+\.)\s+(.*)$/);
        if (lm) items.push({ indent: Math.floor(lm[1].length / 2), ordered: /\d/.test(lm[2]), text: lm[3] });
        else items[items.length - 1].text += ' ' + lines[i].trim();
        i++;
      }
      let html = '', stack = [];
      for (const it of items) {
        while (stack.length && stack[stack.length - 1].indent > it.indent) { html += `</li></${stack.pop().tag}>`; }
        const tag = it.ordered ? 'ol' : 'ul';
        if (!stack.length || stack[stack.length - 1].indent < it.indent) {
          html += `<${tag}>`; stack.push({ indent: it.indent, tag });
        } else html += '</li>';
        let t = it.text, cb = '';
        const tm = t.match(/^\[( |x)\]\s+(.*)$/);
        if (tm) {
          const key = (opts.prefix || 'chk') + ':' + (checkIdx++);
          cb = `<input type="checkbox" data-check="${key}" ${checks[key] ? 'checked' : ''}> `;
          t = tm[2];
        }
        html += `<li class="${cb ? 'task' : ''}">${cb}${inline(t)}`;
      }
      while (stack.length) html += `</li></${stack.pop().tag}>`;
      out.push(html);
      continue;
    }
    const buf = [];
    while (i < lines.length && lines[i].trim() && !(buf.length && isBlockStart(lines[i]))) buf.push(lines[i++]);
    out.push(`<p>${inline(buf.join(' '))}</p>`);
  }
  return out.join('\n');
}
