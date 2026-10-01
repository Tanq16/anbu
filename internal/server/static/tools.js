import { api, highlightJSON } from './app.js';
import { esc, seg, segToggle, segVal, copyBtn, codeBlock, numInput, datePicker, ymd } from './controls.js';

const enc = new TextEncoder();
const dec = new TextDecoder();

export function b64Encode(text, urlSafe = false) {
  const bin = Array.from(enc.encode(text), (b) => String.fromCharCode(b)).join('');
  const out = btoa(bin);
  return urlSafe ? out.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '') : out;
}

export function b64Decode(input) {
  let s = input.trim().replace(/-/g, '+').replace(/_/g, '/');
  s += '='.repeat((4 - (s.length % 4)) % 4);
  return dec.decode(Uint8Array.from(atob(s), (c) => c.charCodeAt(0)));
}

export const urlEncode = (s) => encodeURIComponent(s);
export const urlDecode = (s) => decodeURIComponent(s.replace(/\+/g, ' '));

export function parseJWT(token) {
  const [h, p, sig] = token.trim().split('.');
  if (!h || !p) throw new Error('not a JWT');
  const header = JSON.parse(b64Decode(h));
  const payload = JSON.parse(b64Decode(p));
  const times = {};
  for (const k of ['iat', 'nbf', 'exp']) {
    if (typeof payload[k] === 'number') times[k] = new Date(payload[k] * 1000).toISOString();
  }
  const expired = typeof payload.exp === 'number' ? Date.now() / 1000 > payload.exp : null;
  return { header, payload, times, expired, signature: sig ?? '' };
}

function words(s) {
  return s
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .replace(/([A-Z]+)([A-Z][a-z])/g, '$1 $2')
    .split(/[^A-Za-z0-9]+/)
    .filter(Boolean)
    .map((w) => w.toLowerCase());
}

const cap = (w) => w.charAt(0).toUpperCase() + w.slice(1);

export function convertCase(s) {
  const w = words(s);
  return {
    lower: s.toLowerCase(),
    upper: s.toUpperCase(),
    camel: w.map((x, i) => (i ? cap(x) : x)).join(''),
    pascal: w.map(cap).join(''),
    snake: w.join('_'),
    constant: w.join('_').toUpperCase(),
    kebab: w.join('-'),
    dot: w.join('.'),
    path: w.join('/'),
    title: w.map(cap).join(' '),
    sentence: cap(w.join(' ')),
  };
}

export function textStats(s) {
  const wordsList = s.trim() ? s.trim().split(/\s+/) : [];
  return {
    characters: [...s].length,
    characters_no_spaces: [...s.replace(/\s/g, '')].length,
    words: wordsList.length,
    lines: s ? s.split(/\r\n|\r|\n/).length : 0,
    paragraphs: s.split(/\n\s*\n/).filter((p) => p.trim()).length,
    bytes_utf8: enc.encode(s).length,
  };
}

export function eta({ total, start, perAmount, perMs }) {
  if (total <= 0 || perAmount <= 0 || perMs <= 0) throw new Error('values must be positive');
  const durationMs = (total / perAmount) * perMs;
  return { durationMs, endsAt: new Date(start.getTime() + durationMs) };
}

function formatDuration(seconds) {
  let s = Math.abs(Math.round(seconds));
  const parts = [];
  for (const [n, u] of [[86400, 'day'], [3600, 'hour'], [60, 'minute'], [1, 'second']]) {
    const q = Math.floor(s / n);
    s -= q * n;
    if (q) parts.push(`${q} ${u}${q === 1 ? '' : 's'}`);
  }
  return parts.join(', ') || '0 seconds';
}

export const tools = [
  { id: 'hash', name: 'Hash', icon: 'hash', group: 'Crypto', where: 'server', desc: 'MD5, SHA-1, SHA-2 family' },
  { id: 'jwt', name: 'JWT decoder', icon: 'badge-check', group: 'Crypto', where: 'browser', desc: 'Header, payload, and expiry' },
  { id: 'uuid', name: 'UUID', icon: 'fingerprint', group: 'Generators', where: 'server', desc: 'v4, v7, and short forms' },
  { id: 'passphrase', name: 'Passphrase', icon: 'whole-word', group: 'Generators', where: 'server', desc: 'Word-based passwords' },
  { id: 'random', name: 'Random string', icon: 'dices', group: 'Generators', where: 'server', desc: 'Length and charset' },
  { id: 'base64', name: 'Base64', icon: 'binary', group: 'Encoding', where: 'browser', desc: 'Standard and URL-safe' },
  { id: 'url', name: 'URL encode', icon: 'link', group: 'Encoding', where: 'browser', desc: 'Percent-encoding' },
  { id: 'yaml', name: 'YAML ↔ JSON', icon: 'file-json', group: 'Encoding', where: 'server', desc: 'Convert either way' },
  { id: 'case', name: 'Case converter', icon: 'case-sensitive', group: 'Text', where: 'browser', desc: 'camel, snake, kebab, ...' },
  { id: 'stats', name: 'Text stats', icon: 'text', group: 'Text', where: 'browser', desc: 'Characters, words, bytes' },
  { id: 'time', name: 'Time', icon: 'clock', group: 'Time', where: 'server', desc: 'Epoch, ISO, until' },
  { id: 'eta', name: 'ETA', icon: 'hourglass', group: 'Time', where: 'browser', desc: 'When a batch finishes' },
];

export const charsets = ['alphanumeric', 'alpha', 'digits', 'hex', 'all'];

const field = 'w-full bg-surface0 rounded-lg px-3 py-2 text-text text-sm font-mono outline-none focus:ring-2 focus:ring-mauve pointer-coarse:text-[16px]';
const area = `${field} min-h-32 resize-y`;
const well = 'bg-base rounded-lg p-3 font-mono text-sm text-text whitespace-pre-wrap break-all overflow-auto';
const btn = 'inline-flex items-center gap-2 rounded-full bg-mauve text-crust font-medium text-sm px-4 py-2 hover:opacity-90';
const lbl = 'text-xs text-overlay1';

const outRow = (k, v) => `<li class="flex items-center gap-3 rounded-lg px-3 py-2 hover:bg-surface0"><span class="w-32 shrink-0 ${lbl}">${esc(k)}</span><code class="flex-1 min-w-0 break-all font-mono text-sm text-text">${esc(v)}</code>${copyBtn(String(v))}</li>`;
const errLine = (e) => `<p class="text-red text-sm">${esc(e.message)}</p>`;
const gen = (icon, text) => `<button type="button" data-gen class="${btn}"><i data-lucide="${icon}" class="size-4"></i>${text}</button>`;

const views = {
  hash: () => `<textarea data-in class="${area}" placeholder="Text to hash"></textarea><ul data-out class="mt-4 flex flex-col gap-1"></ul>`,
  jwt: () => `<textarea data-in class="${area}" placeholder="Paste a JWT"></textarea><div data-out class="mt-4 grid gap-4 lg:grid-cols-2"></div>`,
  uuid: (s) => `<div class="flex flex-wrap items-center gap-3">${seg('version', [[4, 'v4'], [7, 'v7']], s.uuid_version)}${seg('short', [['false', 'Full'], ['true', 'Short']], 'false')}<label class="flex items-center gap-2 ${lbl}">Count${numInput('data-count min="1" max="100" value="5"', 'w-24')}</label>${gen('refresh-cw', 'Generate')}</div><ul data-out class="mt-4 flex flex-col gap-1"></ul>`,
  passphrase: (s) => `<div class="flex flex-wrap items-center gap-3"><label class="flex items-center gap-2 ${lbl}">Words${numInput(`data-count min="1" max="50" value="${esc(s.passphrase_words)}"`, 'w-24')}</label>${seg('simple', [['false', 'Capital + digit'], ['true', 'Simple']], String(s.passphrase_simple))}${gen('refresh-cw', 'Generate')}</div><ul data-out class="mt-4 flex flex-col gap-1"></ul>`,
  random: (s) => `<div class="flex flex-wrap items-center gap-3"><label class="flex items-center gap-2 ${lbl}">Length${numInput(`data-count min="1" max="512" value="${esc(s.random_length)}"`, 'w-28')}</label>${seg('charset', charsets.map((c) => [c, c]), s.random_charset)}${gen('refresh-cw', 'Generate')}</div><ul data-out class="mt-4 flex flex-col gap-1"></ul>`,
  base64: () => `<div class="flex flex-wrap gap-3 mb-3">${seg('mode', [['enc', 'Encode'], ['dec', 'Decode']], 'enc')}${seg('urlsafe', [['false', 'Standard'], ['true', 'URL-safe']], 'false')}</div><div class="grid gap-4 lg:grid-cols-2"><textarea data-in class="${area}" placeholder="Text to encode or decode"></textarea>${codeBlock(`${well} min-h-32`, '', 'data-out')}</div>`,
  url: () => `<div class="mb-3">${seg('mode', [['enc', 'Encode'], ['dec', 'Decode']], 'enc')}</div><div class="grid gap-4 lg:grid-cols-2"><textarea data-in class="${area}" placeholder="Text to encode or decode"></textarea>${codeBlock(`${well} min-h-32`, '', 'data-out')}</div>`,
  yaml: () => `<div class="mb-3">${seg('to', [['json', 'YAML → JSON'], ['yaml', 'JSON → YAML']], 'json')}</div><div class="grid gap-4 lg:grid-cols-2"><textarea data-in class="${area} min-h-64" placeholder="YAML or JSON"></textarea>${codeBlock(`${well} min-h-64`, '', 'data-out')}</div>`,
  case: () => `<input data-in class="${field}" placeholder="Text to convert"><ul data-out class="mt-4 grid gap-1 xl:grid-cols-2"></ul>`,
  stats: () => `<textarea data-in class="${area} min-h-48" placeholder="Text to measure"></textarea><dl data-out class="mt-4 grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-6 gap-3"></dl>`,
  time: () => `<div class="flex flex-wrap items-center gap-3"><input data-in class="${field} flex-1 min-w-60" placeholder="Epoch, RFC 3339, date, or blank for now">${gen('clock', 'Now')}</div><ul data-out class="mt-4 flex flex-col gap-1"></ul><p class="mt-6 mb-2 ${lbl}">Difference between two epochs</p><div class="flex flex-wrap gap-3"><input data-a class="${field} flex-1 min-w-40" placeholder="first epoch"><input data-b class="${field} flex-1 min-w-40" placeholder="blank for now"></div><p data-diff class="mt-3 text-sm text-text"></p>`,
  eta: () => `<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"><label class="flex flex-col gap-1 ${lbl}">Total items${numInput('data-total min="1" value="1200"')}</label><label class="flex flex-col gap-1 ${lbl}">Items per step${numInput('data-per min="1" value="3"')}</label><label class="flex flex-col gap-1 ${lbl}">Step every (seconds)${numInput('data-ms min="1" value="5"')}</label><div class="flex flex-col gap-1 ${lbl}">Started at<div class="flex gap-2">${datePicker({ attrs: 'data-start', value: ymd(new Date()) })}<input data-start-time type="time" value="${new Date().toTimeString().slice(0, 5)}" class="${field} w-28"></div></div></div><ul data-out class="mt-4 flex flex-col gap-1"></ul>`,
};

export function toolView(id, settings) {
  return views[id](settings);
}

export function wireTool(root, id) {
  const $ = (s) => root.querySelector(s);
  const out = $('[data-out]');
  const input = $('[data-in]');
  let seq = 0;
  const list = (rows) => { out.innerHTML = rows.map(([k, v]) => outRow(k, v)).join(''); lucide.createIcons(); };
  const latest = (fn) => async () => {
    const n = ++seq;
    try {
      const paint = await fn();
      if (n === seq) paint();
    } catch (e) {
      if (n === seq) { out.innerHTML = errLine(e); }
    }
  };
  const count = () => Math.max(1, +$('[data-count]').value || 1);
  const run = {
    hash: latest(async () => {
      if (!input.value) return () => { out.innerHTML = ''; };
      const h = await api('POST', '/api/tools/hash', { text: input.value });
      return () => list(['md5', 'sha1', 'sha224', 'sha256', 'sha384', 'sha512'].map((k) => [k, h[k]]));
    }),
    jwt: () => {
      if (!input.value.trim()) { out.innerHTML = ''; return; }
      try {
        const j = parseJWT(input.value);
        out.innerHTML = `<div><p class="${lbl} mb-2">Header</p>${codeBlock(well, highlightJSON(j.header))}</div><div><p class="${lbl} mb-2">Payload</p>${codeBlock(well, highlightJSON(j.payload))}</div><ul class="lg:col-span-2 flex flex-col gap-1">${Object.entries(j.times).map(([k, v]) => outRow(k, v)).join('')}${j.expired === null ? '' : `<li class="px-3 py-2 text-sm ${j.expired ? 'text-red' : 'text-green'}">${j.expired ? 'Expired' : 'Not expired'}</li>`}</ul>`;
        lucide.createIcons();
      } catch (e) { out.innerHTML = errLine(e); }
    },
    uuid: latest(async () => {
      const q = new URLSearchParams({ version: segVal(root, 'version'), count: Math.min(100, count()), short: segVal(root, 'short') });
      const { values } = await api('GET', `/api/tools/uuid?${q}`);
      return () => list(values.map((v, i) => [`#${i + 1}`, v]));
    }),
    passphrase: latest(async () => {
      const q = new URLSearchParams({ words: count(), simple: segVal(root, 'simple') });
      const values = await Promise.all(Array.from({ length: 4 }, () => api('GET', `/api/tools/passphrase?${q}`)));
      return () => list(values.map((v, i) => [`#${i + 1}`, v.value]));
    }),
    random: latest(async () => {
      const q = new URLSearchParams({ length: count(), charset: segVal(root, 'charset') });
      const values = await Promise.all(Array.from({ length: 3 }, () => api('GET', `/api/tools/random?${q}`)));
      return () => list(values.map((v, i) => [`#${i + 1}`, v.value]));
    }),
    base64: () => {
      try { out.textContent = segVal(root, 'mode') === 'enc' ? b64Encode(input.value, segVal(root, 'urlsafe') === 'true') : b64Decode(input.value); } catch (e) { out.textContent = e.message; }
    },
    url: () => { try { out.textContent = segVal(root, 'mode') === 'enc' ? urlEncode(input.value) : urlDecode(input.value); } catch (e) { out.textContent = e.message; } },
    yaml: latest(async () => {
      if (!input.value.trim()) return () => { out.textContent = ''; };
      const to = segVal(root, 'to');
      const { output } = await api('POST', '/api/tools/yaml', { input: input.value, to });
      return () => { out.innerHTML = to === 'json' ? highlightJSON(output) : esc(output); };
    }),
    case: () => list(Object.entries(convertCase(input.value))),
    stats: () => { out.innerHTML = Object.entries(textStats(input.value)).map(([k, v]) => `<div class="bg-surface0 rounded-xl px-4 py-3"><dt class="${lbl}">${k.replace(/_/g, ' ')}</dt><dd class="text-2xl font-display text-text">${v}</dd></div>`).join(''); },
    time: latest(async () => {
      const t = input.value.trim();
      const [info, until] = await Promise.all([
        api('POST', '/api/tools/time', t ? { op: 'parse', input: t } : { op: 'now' }),
        t ? api('POST', '/api/tools/time', { op: 'until', input: t }) : null,
      ]);
      const rows = ['epoch', 'epoch_ms', 'rfc822_local', 'iso8601_local', 'iso8601_utc', 'human_utc'].map((k) => [k, info[k]]);
      if (until) rows.push([until.future ? 'until' : 'since', until.human]);
      return () => list(rows);
    }),
    eta: () => {
      try {
        const start = new Date(`${$('[data-start]').value || ymd(new Date())}T${$('[data-start-time]').value || '00:00'}`);
        const r = eta({ total: +$('[data-total]').value, start, perAmount: +$('[data-per]').value, perMs: +$('[data-ms]').value * 1000 });
        list([['duration', formatDuration(r.durationMs / 1000)], ['ends at', r.endsAt.toLocaleString()]]);
      } catch (e) { out.innerHTML = errLine(e); }
    },
  }[id];

  let diffSeq = 0;
  const diff = async () => {
    const p = $('[data-diff]');
    const a = $('[data-a]').value.trim();
    const b = $('[data-b]').value.trim();
    const n = ++diffSeq;
    if (!a) { p.textContent = ''; return; }
    const nums = [a, b].filter(Boolean).map(Number);
    if (nums.some((x) => !Number.isInteger(x))) { p.textContent = 'epochs must be whole seconds'; return; }
    try {
      const r = await api('POST', '/api/tools/time', { op: 'diff', epochs: nums });
      if (n === diffSeq) p.textContent = r.human;
    } catch (e) { if (n === diffSeq) p.textContent = e.message; }
  };

  const debounced = ['hash', 'yaml', 'time'].includes(id);
  let timer;
  let diffTimer;
  root.addEventListener('input', (e) => {
    if (e.target.matches('[data-a], [data-b]')) { clearTimeout(diffTimer); diffTimer = setTimeout(diff, 250); return; }
    if (!debounced) { run(); return; }
    clearTimeout(timer);
    timer = setTimeout(run, 250);
  });
  root.addEventListener('click', (e) => {
    const b = e.target.closest('[data-seg] button');
    if (b) { segToggle(b); run(); }
    if (e.target.closest('[data-gen]')) { if (id === 'time') input.value = ''; run(); }
  });
  run();
}
