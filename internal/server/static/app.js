import { esc, ymd, copy, seg, segToggle, segVal, copyBtn, codeBlock, numInput, dropdown, datePicker, setDate, installControls } from './controls.js';
import { tools, toolView, wireTool, charsets } from './tools.js';
import { openTerminal } from './terminal.js';

export async function api(method, path, body) {
  const res = await fetch(path, {
    method,
    headers: body === undefined ? {} : { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { if (res.ok) throw new Error('response is not JSON'); }
  if (!res.ok) throw Object.assign(new Error(data?.error ?? `${res.status} ${res.statusText}`), { status: res.status, data });
  return data;
}

export function watchTOTP(ref, onTick) {
  let code = '', remaining = 0, period = 30, stopped = false;
  async function refresh() {
    ({ code, remaining, period } = await api('GET', `/api/secrets/${encodeURIComponent(ref)}/totp`));
  }
  refresh().then(() => onTick({ code, remaining, period }), () => {});
  const timer = setInterval(async () => {
    if (stopped) return;
    remaining -= 1;
    if (remaining <= 0) await refresh().catch(() => {});
    if (!stopped) onTick({ code, remaining, period });
  }, 1000);
  return () => { stopped = true; clearInterval(timer); };
}

export function highlightJSON(value) {
  const json = typeof value === 'string' ? value : JSON.stringify(value, null, 2);
  const e = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  return e.replace(
    /("(\\u[a-fA-F0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
    (m) => {
      let cls = 'json-number';
      if (m.startsWith('"')) cls = m.endsWith(':') ? 'json-key' : 'json-string';
      else if (m === 'true' || m === 'false') cls = 'json-bool';
      else if (m === 'null') cls = 'json-null';
      return `<span class="${cls}">${m}</span>`;
    },
  );
}

export function jobSummary({ running, queued }) {
  const parts = [];
  if (running) parts.push('1 processing');
  if (queued.length) parts.push(`${queued.length} upcoming`);
  return parts.join(', ');
}

export function isOverdue(task, now = new Date()) {
  if (task.done || !task.due) return false;
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const [y, m, d] = task.due.split('-').map(Number);
  return new Date(y, m - 1, d) < today;
}

function dayDiff(due) {
  const [y, m, d] = due.split('-').map(Number);
  const now = new Date();
  return Math.round((new Date(y, m - 1, d) - new Date(now.getFullYear(), now.getMonth(), now.getDate())) / 86400000);
}

function dueLabel(due) {
  if (!due) return '';
  const diff = dayDiff(due);
  if (diff === 0) return 'today';
  if (diff === 1) return 'tomorrow';
  if (diff === -1) return 'yesterday';
  if (diff < 0) return `${-diff}d overdue`;
  if (diff < 7) return `in ${diff}d`;
  const [y, m, d] = due.split('-').map(Number);
  return new Date(y, m - 1, d).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

const dueThisWeek = (t) => !t.done && !!t.due && !isOverdue(t) && dayDiff(t.due) < 7;

function relTime(iso) {
  const s = Math.round((Date.now() - new Date(iso)) / 1000);
  const f = Math.abs(s);
  const v = f < 60 ? [f, 's'] : f < 3600 ? [Math.round(f / 60), 'm'] : f < 86400 ? [Math.round(f / 3600), 'h'] : [Math.round(f / 86400), 'd'];
  return s >= 0 ? `${v[0]}${v[1]} ago` : `in ${v[0]}${v[1]}`;
}

const rel = (iso, fmt = '') => `<span data-rel="${esc(iso)}" data-fmt="${fmt}"></span>`;

const groupCode = (c) => c.length === 8 ? `${c.slice(0, 4)} ${c.slice(4)}` : `${c.slice(0, 3)} ${c.slice(3)}`;
const money = (n) => `$${(n ?? 0).toFixed(2)}`;
const enc = encodeURIComponent;
const segEnc = (s) => enc(s).replace(/%3A/gi, ':').replace(/%40/g, '@');
const profileQ = (ref) => `profile=${enc(ref)}`;

const field = 'w-full bg-surface0 rounded-lg px-3 py-2 text-text text-sm outline-none focus:ring-2 focus:ring-mauve pointer-coarse:text-[16px]';
const pillInput = 'w-full bg-surface0 rounded-full px-4 py-2 text-text text-sm outline-none focus:ring-2 focus:ring-mauve pointer-coarse:text-[16px]';
const primary = 'inline-flex items-center gap-2 rounded-full bg-mauve text-crust font-medium text-sm px-4 py-2 hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed';
const ghost = 'inline-flex items-center gap-2 rounded-full bg-surface0 text-subtext1 text-sm px-4 py-2 hover:bg-surface1 hover:text-text disabled:opacity-40 disabled:cursor-not-allowed';
const danger = 'inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm text-red hover:bg-red/10';
const lbl = 'text-xs text-overlay1';
const well = 'bg-base rounded-xl p-4 font-mono text-sm text-text whitespace-pre-wrap break-all overflow-auto';
const h2 = 'text-2xl font-display text-text';
const h3 = 'text-sm font-medium text-subtext1 mb-3';

const typeMeta = {
  login: { icon: 'user-round', color: 'blue' },
  'ssh-key': { icon: 'key-round', color: 'green' },
  'aws-static': { icon: 'cloud', color: 'peach' },
  'aws-sso': { icon: 'cloud-cog', color: 'yellow' },
  'github-pat': { icon: 'github', color: 'mauve' },
  generic: { icon: 'box', color: 'teal' },
};
const meta = (t) => typeMeta[t] ?? { icon: 'key-round', color: 'overlay1' };

function typeIcon(type, cls = 'size-4') {
  if (type === 'aws-static' || type === 'aws-sso') return `<i class="fa-brands fa-aws ${cls} leading-none"></i>`;
  if (type === 'github-pat') return `<i class="fa-brands fa-github ${cls} leading-none"></i>`;
  return `<i data-lucide="${meta(type).icon}" class="${cls}"></i>`;
}

const modules = [
  ['vault', 'Vault', 'key-round'], ['tasks', 'Tasks', 'list-checks'], ['ssh', 'SSH', 'terminal'],
  ['aws', 'AWS', 'cloud'], ['machines', 'Machines', 'server'], ['tools', 'Tools', 'wrench'], ['settings', 'Settings', 'settings'],
];
const viewNames = { vault: ['new'], ssh: ['new'], aws: ['adhoc'], machines: ['jobs', 'scaffold', 'new'] };
const taskFilters = [['open', 'Open', 'circle'], ['overdue', 'Overdue', 'alarm-clock'], ['week', 'Due this week', 'calendar'], ['nodate', 'No date', 'calendar-off'], ['done', 'Done', 'circle-check']];
const priorities = ['critical', 'high', 'default'];
const settingsSections = [['generators', 'Generators', 'dices'], ['timeouts', 'Timeouts', 'timer'], ['machines', 'Machines', 'server'], ['password', 'Vault password', 'lock'], ['backup', 'Export and import', 'archive'], ['api', 'API and CLI', 'code']];

const transient = () => ({ confirm: null, editing: false, modify: null, remove: false, probe: null, keep: {}, importResult: null });
const st = {
  mod: 'vault', detail: false, q: '', push: false, profile: '',
  sel: { vault: '', tasks: 'open', ssh: '', aws: '', machines: '', tools: tools[0].id, settings: 'generators' },
  view: { vault: '', ssh: '', aws: '', machines: '' },
  newType: 'login', newPrio: 'default', sshTab: 'terminal', openJob: null,
  awsCmd: 'ec2 describe-instances --filters Name=tag:ManagedBy,Values=sharingan', run: null,
  mShape: '4vcpu-16gb', mClass: 'dedicated', mArch: 'arm64', mDisk: 120,
  ...transient(),
};
try { st.profile = localStorage.getItem('anbu.profile') ?? ''; } catch { }

const cache = new Map();
function load(key, fn) {
  let e = cache.get(key);
  if (!e) { e = { key }; cache.set(key, e); }
  if (!e.loading && (e.stale || (!('data' in e) && !e.error))) {
    e.loading = true;
    e.stale = false;
    fn().then((d) => { e.data = d; e.error = null; }, (err) => { delete e.data; e.error = err; }).finally(() => { e.loading = false; render(); });
  }
  return e;
}

function invalidate(...names) {
  for (const [k, e] of cache) if (names.some((n) => k === n || k.startsWith(`${n}:`))) e.stale = true;
}

const loaders = {
  types: () => api('GET', '/api/secret-types'),
  secrets: () => api('GET', '/api/secrets'),
  secret: (n) => api('GET', `/api/secrets/${enc(n)}`),
  tasks: () => api('GET', '/api/tasks'),
  settings: () => api('GET', '/api/settings'),
  targets: () => api('GET', '/api/ssh/targets'),
  sso: (s) => api('GET', `/api/aws/sso/${enc(s)}/status`),
  whoami: (r) => api('GET', `/api/aws/whoami?${profileQ(r)}`),
  machines: (r) => api('GET', `/api/aws/machines?${profileQ(r)}`),
  options: (r) => api('GET', `/api/aws/machines/options?${profileQ(r)}`),
  scaffold: (r) => api('GET', `/api/aws/scaffold?${profileQ(r)}`),
};
const get = (name, arg) => load(arg === undefined ? name : `${name}:${arg}`, () => loaders[name](arg));
const peek = (name, arg) => cache.get(arg === undefined ? name : `${name}:${arg}`)?.data;

const loadingLine = (text = 'Loading') => `<p class="px-4 py-10 text-center ${lbl}"><i data-lucide="loader-circle" class="inline size-4 animate-spin mr-1.5 align-[-3px]"></i>${text}</p>`;
const failure = (err, key) => `<div class="rounded-2xl bg-red/10 p-5 text-sm"><p class="text-red font-medium mb-1">${err.status ? `${err.status} · ` : ''}request failed</p><p class="mb-4 break-words">${esc(err.message)}</p>${key ? `<button type="button" data-retry="${esc(key)}" class="${ghost} !bg-surface1"><i data-lucide="refresh-cw" class="size-4"></i>Retry</button>` : ''}</div>`;
const listFailure = (err, key) => `<div class="px-3 py-2 text-xs"><p class="text-red break-words">${err.status ? `${err.status} · ` : ''}${esc(err.message)}</p><button type="button" data-retry="${esc(key)}" class="mt-1 text-overlay1 hover:text-text">Retry</button></div>`;
const need = (e, fn, text) => ('data' in e ? fn(e.data) : e.error ? listFailure(e.error, e.key) : loadingLine(text));
const needD = (e, fn, text) => ('data' in e ? fn(e.data) : article(e.error ? failure(e.error, e.key) : loadingLine(text)));

function toast(msg, tone = 'red') {
  const t = document.createElement('div');
  t.className = `rounded-xl bg-surface0 px-4 py-3 text-sm shadow-xl shadow-crust/60 text-${tone} pointer-events-auto`;
  t.textContent = msg;
  document.getElementById('toast').append(t);
  setTimeout(() => t.remove(), 4000);
}

const calm = matchMedia('(prefers-reduced-motion: reduce)');
const wide = matchMedia('(min-width: 48rem)');

function leave(li, delay, then) {
  if (li.dataset.leaving) return;
  li.dataset.leaving = '1';
  if (calm.matches) return then();
  li.style.overflow = 'hidden';
  li.animate([{ opacity: 1, height: `${li.offsetHeight}px` }, { opacity: 0, height: '0px', paddingTop: '0px', paddingBottom: '0px' }], { duration: 220, delay, easing: 'ease-in', fill: 'forwards' }).onfinish = then;
}

function route() {
  const m = st.mod;
  const parts = [m];
  if (m === 'machines' && st.profile) parts.push(st.profile);
  const view = st.view[m];
  const show = wide.matches || st.detail;
  if (show && !view && st.sel[m]) parts.push(...(m === 'ssh' ? st.sel[m].split('/') : [st.sel[m]]));
  return `/${parts.map(segEnc).join('/')}${show && view ? `?view=${view}` : ''}`;
}

function applyRoute() {
  let [mod, ...rest] = location.pathname.split('/').filter(Boolean).map((s) => { try { return decodeURIComponent(s); } catch { return s; } });
  if (!modules.some(([id]) => id === mod)) { mod = 'vault'; rest = []; }
  st.mod = mod;
  if (mod === 'machines' && rest.length) st.profile = rest.shift();
  const view = new URLSearchParams(location.search).get('view') ?? '';
  if (mod in st.view) st.view[mod] = viewNames[mod].includes(view) ? view : '';
  if (rest.length) st.sel[mod] = rest.join('/');
  st.detail = rest.length > 0 || !!st.view[mod];
}

const panes = { rail: { min: 160, max: 288, closed: 64, snap: 112 }, list: { min: 240, max: 560, closed: 0, snap: 160 } };
const ui = { rail: { w: 232, open: false }, list: { w: 320, open: true } };
try { Object.assign(ui, JSON.parse(localStorage.getItem('anbu.layout')) ?? {}); } catch { }
const saveLayout = () => { try { localStorage.setItem('anbu.layout', JSON.stringify(ui)); } catch { } };
const shown = (k) => (ui[k].open ? ui[k].w : panes[k].closed);

function applyLayout() {
  document.body.style.setProperty('--rail-w', `${shown('rail')}px`);
  document.body.style.setProperty('--list-w', `${shown('list')}px`);
  document.getElementById('list').classList.toggle('md:hidden', !ui.list.open);
}

function togglePane(k) {
  ui[k].open = !ui[k].open;
  applyLayout();
  saveLayout();
  renderRail();
  lucide.createIcons();
}

for (const h of document.querySelectorAll('[data-resize]')) {
  const k = h.dataset.resize;
  h.addEventListener('pointerdown', (e) => {
    const origin = h.getBoundingClientRect().left - shown(k);
    h.setPointerCapture(e.pointerId);
    const move = (ev) => {
      const x = ev.clientX - origin;
      const was = ui[k].open;
      ui[k].open = x >= panes[k].snap;
      if (ui[k].open) ui[k].w = Math.min(panes[k].max, Math.max(panes[k].min, x));
      applyLayout();
      if (k === 'rail' && was !== ui[k].open) { renderRail(); lucide.createIcons(); }
    };
    const up = () => {
      h.removeEventListener('pointermove', move);
      h.removeEventListener('pointerup', up);
      h.removeEventListener('pointercancel', up);
      saveLayout();
    };
    h.addEventListener('pointermove', move);
    h.addEventListener('pointerup', up);
    h.addEventListener('pointercancel', up);
  });
  h.addEventListener('dblclick', () => togglePane(k));
}

const stateColor = { running: 'green', stopped: 'overlay1', pending: 'yellow', stopping: 'peach', 'shutting-down': 'peach', ready: 'green' };
const dot = (s) => `<span class="size-2 rounded-full shrink-0 bg-${stateColor[s] ?? 'overlay1'} ${s === 'pending' ? 'animate-pulse' : ''}"></span>`;
const chip = (text, color = 'overlay1') => `<span class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs bg-${color}/15 text-${color}">${text}</span>`;
const item = (attrs, active, inner) => `<button type="button" ${attrs} class="w-full flex items-center gap-3 rounded-xl px-3 py-2.5 text-left ${active ? 'bg-surface0' : 'hover:bg-surface0/60'}">${inner}</button>`;
const selItem = (mod, id, inner) => item(`data-sel="${esc(id)}"`, !st.view[mod] && st.sel[mod] === id, inner);
const viewItem = (mod, view, inner) => item(`data-view="${view}"`, st.view[mod] === view, inner);
const twoLine = (a, b) => `<span class="flex-1 min-w-0"><span class="block truncate text-sm text-text">${a}</span><span class="block truncate ${lbl}">${b}</span></span>`;
const group = (t) => `<p class="px-3 pt-4 pb-1 text-[11px] uppercase tracking-wider text-overlay0">${t}</p>`;
const back = `<button type="button" data-act="back" class="md:hidden mb-4 inline-flex items-center gap-1 text-sm text-overlay1"><i data-lucide="chevron-left" class="size-4"></i>Back</button>`;
const article = (inner, cls = '') => `<article class="p-5 lg:p-8 ${cls}">${back}${inner}</article>`;
const empty = (text, action = '') => article(`<div class="grid place-items-center py-20 text-center"><div><p class="text-text mb-1">${text}</p>${action ? `<div class="mt-4">${action}</div>` : ''}</div></div>`);
const addBtn = (act, title) => `<button type="button" data-act="${act}" title="${title}" class="grid place-items-center size-9 rounded-full bg-surface0 text-text hover:bg-surface1"><i data-lucide="plus" class="size-4"></i></button>`;

function listShell(title, action, body, sub = '') {
  return `<header class="px-4 pt-5 pb-3 flex items-center gap-2"><h1 class="flex-1 text-xl font-display text-text">${title}</h1>${action}</header>${sub}<div data-scroll class="flex-1 min-h-0 overflow-y-auto px-2 pb-4">${body}</div>`;
}

let jobs = { running: null, queued: [], finished: [] };
let jobsTimer = null;
let seenFinished = null;

async function refreshJobs() {
  clearTimeout(jobsTimer);
  jobsTimer = null;
  try { jobs = await api('GET', '/api/aws/jobs'); } catch { }
  const ids = jobs.finished.map((j) => j.id).join();
  if (seenFinished !== null && ids !== seenFinished) invalidate('machines', 'scaffold', 'targets', 'secrets', 'secret');
  seenFinished = ids;
  if (jobs.running || jobs.queued.length) jobsTimer = setTimeout(refreshJobs, 3000);
  render();
}

function profiles() {
  const out = [];
  for (const s of peek('secrets') ?? []) {
    if (s.type === 'aws-static') out.push({ ref: s.name, kind: 'static', secret: s.name, region: s.fields.region ?? '', account_id: peek('whoami', s.name)?.account ?? '' });
    if (s.type === 'aws-sso') for (const p of s.profiles ?? []) out.push({ ...p, kind: 'sso', secret: s.name });
  }
  return out;
}

function renderRail() {
  const js = jobSummary(jobs);
  const open = ui.rail.open;
  const shape = open ? 'md:flex-row md:gap-3 md:px-3 md:py-2.5 md:text-sm' : 'md:self-center md:size-11 md:justify-center md:p-0';
  const label = open ? 'md:truncate' : 'md:hidden';
  const tool = (attrs, icon, title) => `<button type="button" ${attrs} title="${title}" class="hidden md:flex items-center gap-3 rounded-xl text-overlay1 hover:text-text hover:bg-surface0/60 ${shape}"><i data-lucide="${icon}" class="size-4 shrink-0"></i><span class="${label}">${title}</span></button>`;
  const html = `
    <span class="hidden md:flex items-center gap-3 mb-4 ${open ? 'px-1.5' : 'self-center'}"><img src="/static/icons/logo.png" alt="anbu" class="size-11 shrink-0">${open ? '<span class="font-display text-lg text-text">anbu</span>' : ''}</span>
    ${modules.map(([id, name, icon]) => `<button type="button" data-mod="${id}" title="${id === 'machines' && js ? `${name} · ${js}` : name}" class="flex-1 md:flex-none min-w-11 flex flex-col items-center gap-1 rounded-xl py-2 text-[10px] ${shape} ${st.mod === id ? 'bg-surface0 text-mauve' : 'text-overlay1 hover:text-text'}"><span class="relative shrink-0"><i data-lucide="${icon}" class="size-5"></i>${id === 'machines' && js ? '<span class="absolute -top-0.5 -right-1 size-2 rounded-full bg-yellow"></span>' : ''}</span><span class="${label}">${name}</span></button>`).join('')}
    <span class="hidden md:block flex-1"></span>
    ${tool('data-pane="list"', ui.list.open ? 'panel-left-close' : 'panel-left-open', ui.list.open ? 'Hide list' : 'Show list')}
    ${tool('data-pane="rail"', open ? 'chevrons-left' : 'chevrons-right', open ? 'Collapse sidebar' : 'Expand sidebar')}`;
  if (html !== last.rail) {
    document.getElementById('rail').innerHTML = html;
    last.rail = html;
  }
}

function secretSub(s) {
  const f = s.fields ?? {};
  if (s.type === 'login') return f.username || f.url || 'Login';
  if (s.type === 'aws-sso') return `${s.profiles?.length ?? 0} profiles · ${f.sso_region ?? ''}`;
  if (s.type === 'aws-static') return `${(f.access_key_id ?? '').slice(0, 8)}… · ${f.region ?? ''}`;
  if (s.type === 'github-pat') return [f.username ? `@${f.username}` : 'GitHub PAT', f.expires_at ? `expires ${dueLabel(f.expires_at)}` : ''].filter(Boolean).join(' · ');
  return peek('types')?.[s.type]?.label ?? s.type;
}

function accountPicker() {
  const ps = profiles();
  const p = ps.find((x) => x.ref === st.profile);
  const sc = peek('scaffold', st.profile);
  const place = (x) => [x.account_id, x.region].filter(Boolean).join(' · ');
  const options = ps.map((x) => [x.ref, x.ref, place(x)]);
  const sub = sc ? `${sc.account} · ${sc.region}` : p ? place(p) : '';
  return `<div class="${lbl}"><p class="mb-1.5">Active account</p>${dropdown({ attrs: 'data-profile', options, value: st.profile, cls: 'w-full bg-surface0 rounded-xl px-3.5 py-2.5 font-mono text-sm text-text outline-none focus:ring-2 focus:ring-mauve', lead: '<i class="fa-brands fa-aws text-yellow w-5"></i>' })}<p class="mt-1.5 font-mono">${esc(sub)}</p></div>`;
}

function ssoBadge(secret) {
  const s = peek('sso', secret);
  if (!s) return '';
  if (s.state === 'active') return `<span class="normal-case tracking-normal text-green ml-1">● ${rel(s.expires_at, 'left')}</span>`;
  if (s.state === 'pending') return '<span class="normal-case tracking-normal text-yellow ml-1">● pending</span>';
  if (s.state === 'error') return '<span class="normal-case tracking-normal text-red ml-1">● error</span>';
  return '<span class="normal-case tracking-normal text-overlay1 ml-1">○ logged out</span>';
}

const lists = {
  vault() {
    const all = get('secrets');
    const sub = `<div class="px-4 pb-3"><input data-q class="${pillInput}" placeholder="Search ${all.data?.length ?? ''} secrets" value="${esc(st.q)}"></div>`;
    const body = need(all, (secrets) => {
      const q = st.q.toLowerCase();
      const f = secrets.filter((s) => s.name.toLowerCase().includes(q));
      if (!f.length) return `<p class="px-3 py-6 ${lbl}">${secrets.length ? 'No match' : 'No secrets yet'}</p>`;
      return f.map((s) => selItem('vault', s.name, `<span class="grid place-items-center w-5 shrink-0 text-${meta(s.type).color}">${typeIcon(s.type)}</span>${twoLine(esc(s.name), esc(secretSub(s)))}`)).join('');
    });
    return listShell('Vault', addBtn('new-secret', 'New secret'), body, sub);
  },
  tasks() {
    const e = get('tasks');
    return listShell('Tasks', '', need(e, (all) => {
      const open = all.filter((t) => !t.done);
      const counts = { open: open.length, overdue: open.filter((t) => isOverdue(t)).length, week: open.filter(dueThisWeek).length, nodate: open.filter((t) => !t.due).length, done: all.length - open.length };
      return taskFilters.map(([id, n, ic]) => selItem('tasks', id, `<i data-lucide="${ic}" class="size-4 ${id === 'overdue' && counts[id] ? 'text-red' : 'text-overlay1'}"></i><span class="flex-1 text-sm text-text">${n}</span><span class="${lbl}">${counts[id]}</span>`)).join('')
        + group('By priority') + priorities.map((p) => selItem('tasks', p, `<span class="size-2 rounded-full ${p === 'critical' ? 'bg-red' : p === 'high' ? 'bg-peach' : 'bg-overlay0'}"></span><span class="flex-1 text-sm text-text capitalize">${p}</span><span class="${lbl}">${open.filter((t) => t.priority === p).length}</span>`)).join('');
    }));
  },
  ssh() {
    const e = get('targets');
    return listShell('SSH', addBtn('new-host', 'New host'), need(e, ({ targets, errors }) => {
      const hosts = targets.filter((x) => x.kind === 'host');
      const machines = targets.filter((x) => x.kind === 'machine');
      return group('Hosts') + (hosts.map((x) => selItem('ssh', x.ref, `<i data-lucide="server" class="size-4 text-overlay1"></i>${twoLine(esc(x.name), esc(`${x.user}@${x.addr}`))}`)).join('') || `<p class="px-3 py-2 ${lbl}">No hosts yet</p>`)
        + group('Machines') + (machines.map((x) => selItem('ssh', x.ref, `${dot(x.state)}${twoLine(esc(x.name), esc(`${x.profile} · ${x.addr || x.state}`))}`)).join('') || (errors.length ? '' : `<p class="px-3 py-2 ${lbl}">No machines</p>`))
        + errors.map((x) => `<p class="px-3 py-1.5 text-xs text-overlay1 truncate" title="${esc(x.error)}"><span class="font-mono">${esc(x.profile)}</span> · <span class="text-red">${esc(x.error)}</span></p>`).join('');
    }, 'Loading hosts and machines'));
  },
  aws() {
    const e = get('secrets');
    return listShell('AWS', addBtn('new-aws-secret', 'New AWS secret'), need(e, (secrets) => {
      const sso = secrets.filter((s) => s.type === 'aws-sso');
      const statics = secrets.filter((s) => s.type === 'aws-static');
      for (const s of sso) get('sso', s.name);
      return sso.map((s) => group(`SSO · ${esc(s.name)} ${ssoBadge(s.name)}`) + (s.profiles ?? []).map((p) => selItem('aws', p.ref, `<i class="fa-brands fa-aws text-yellow w-4"></i>${twoLine(esc(p.ref), esc(`${p.role_name} · ${p.region}`))}`)).join('')).join('')
        + (statics.length ? group('Static keys') + statics.map((s) => selItem('aws', s.name, `<i class="fa-brands fa-aws text-peach w-4"></i>${twoLine(esc(s.name), esc([peek('whoami', s.name)?.account, s.fields.region].filter(Boolean).join(' · ')))}`)).join('') : '')
        + group('One-off') + viewItem('aws', 'adhoc', `<i data-lucide="key-square" class="size-4 text-overlay1"></i>${twoLine('Ad-hoc keys', 'Held for one request only')}`);
    }));
  },
  machines() {
    const e = get('secrets');
    if (!('data' in e)) return listShell('Machines', '', need(e, () => ''));
    if (!profiles().length) return listShell('Machines', '', `<p class="px-3 py-6 ${lbl}">No AWS profiles in the vault yet</p>`);
    const js = jobSummary(jobs);
    const sc = get('scaffold', st.profile);
    const ms = get('machines', st.profile);
    const scLine = 'data' in sc ? (sc.data.key.state === 'missing' ? '<span class="text-peach">key not in anbu</span>' : sc.data.key.state === 'differs' ? '<span class="text-red">key differs</span>' : sc.data.ready ? 'ready' : 'not set up') : sc.error ? '<span class="text-red">unavailable</span>' : 'checking';
    const sub = `<div class="px-4 pb-3">${accountPicker()}</div>`;
    return listShell('Machines', addBtn('new-machine', 'New machine'),
      (js ? viewItem('machines', 'jobs', `<i data-lucide="loader-circle" class="size-4 text-yellow animate-spin"></i>${twoLine('Job pipeline', js)}`) : viewItem('machines', 'jobs', `<i data-lucide="list-ordered" class="size-4 text-overlay1"></i>${twoLine('Job pipeline', 'Idle')}`))
      + viewItem('machines', 'scaffold', `<i data-lucide="network" class="size-4 text-overlay1"></i>${twoLine('Scaffold', scLine)}`)
      + need(ms, (list) => group(`${list.length} machine${list.length === 1 ? '' : 's'}`) + list.map((m) => selItem('machines', m.name, `${dot(m.state)}${twoLine(esc(m.name), esc(`${m.type} · ${m.state}`))}<span class="${lbl}">${money(m.state === 'stopped' ? m.monthly_stopped : m.monthly_running)}/mo</span>`)).join(''), 'Loading machines'), sub);
  },
  tools() {
    const groups = [...new Set(tools.map((t) => t.group))];
    return listShell('Tools', '', groups.map((g) => group(g) + tools.filter((t) => t.group === g).map((t) => selItem('tools', t.id, `<i data-lucide="${t.icon}" class="size-4 text-overlay1"></i>${twoLine(esc(t.name), esc(t.desc))}`)).join('')).join(''));
  },
  settings() {
    return listShell('Settings', '', settingsSections.map(([id, n, ic]) => selItem('settings', id, `<i data-lucide="${ic}" class="size-4 text-overlay1"></i><span class="text-sm text-text">${n}</span>`)).join(''));
  },
};

function fieldRow(k, v, kind) {
  const secret = kind.includes('secret');
  const multi = kind.includes('multiline');
  if (!v) return '';
  const reveal = `<button type="button" data-reveal class="rounded-full p-1.5 text-overlay1 hover:text-text hover:bg-surface1" title="Reveal"><i data-lucide="eye" class="size-4"></i></button>`;
  const actions = (inner) => `<span class="flex shrink-0 opacity-0 group-hover:opacity-100 focus-within:opacity-100 max-md:opacity-100">${inner}</span>`;
  if (multi) return `<div class="group min-w-0 xl:col-span-2 px-4 py-3"><div class="flex items-center gap-2 mb-2"><span class="flex-1 ${lbl}">${esc(k)}</span>${actions(reveal + copyBtn(v))}</div><pre data-secret="${esc(v)}" class="${well}">${'•'.repeat(24)}</pre></div>`;
  return `<div class="group min-w-0 flex items-center gap-3 rounded-xl px-4 py-3 hover:bg-surface0"><span class="w-24 sm:w-36 shrink-0 ${lbl}">${esc(k)}</span><code ${secret ? `data-secret="${esc(v)}"` : ''} class="flex-1 min-w-0 truncate font-mono text-sm text-text">${secret ? '•'.repeat(16) : kind === 'url' ? `<a href="${esc(v)}" target="_blank" rel="noopener" class="hover:text-blue">${esc(v)}</a>` : esc(v)}</code>${actions(`${secret ? reveal : ''}${copyBtn(v)}`)}</div>`;
}

const facts = (rows, cols) => `<dl class="grid ${cols} gap-x-8 gap-y-5">${rows.map(([k, v, mono]) => `<div class="min-w-0"><dt class="${lbl} mb-1">${k}</dt><dd class="text-sm text-text truncate ${mono ? 'font-mono' : ''}">${v}</dd></div>`).join('')}</dl>`;

function totpBand(s) {
  return `<section class="flex items-center gap-5 bg-surface0 rounded-2xl px-5 py-4 mb-8">
    <svg viewBox="0 0 36 36" class="size-12 -rotate-90 shrink-0"><circle cx="18" cy="18" r="15" fill="none" stroke="var(--ctp-surface1)" stroke-width="3"/><circle data-ring cx="18" cy="18" r="15" fill="none" stroke="var(--ctp-mauve)" stroke-width="3" stroke-linecap="round" stroke-dasharray="94.25" stroke-dashoffset="94.25"/></svg>
    <div class="flex-1 min-w-0"><p class="${lbl}">One-time code · ${esc(s.totp.algorithm)} · ${s.totp.digits} digits</p><code data-totp="${esc(s.name)}" class="block whitespace-nowrap font-mono text-3xl sm:text-4xl tracking-widest text-overlay0">${groupCode('•'.repeat(s.totp.digits))}</code></div>
    <span data-remain class="${lbl} hidden sm:block"></span>
    <button type="button" data-act="copy-totp" class="${ghost}"><i data-lucide="copy" class="size-4"></i><span class="hidden sm:inline">Copy</span></button></section>`;
}

function ssoState(secret) {
  const e = get('sso', secret);
  if (!('data' in e)) return e.error ? chip('status unavailable', 'red') : '';
  const s = e.data;
  if (s.state === 'active') return chip(`● active · expires ${rel(s.expires_at)}`, 'green');
  if (s.state === 'pending') return chip('<i data-lucide="loader-circle" class="size-3 animate-spin"></i> pending', 'yellow');
  if (s.state === 'error') return chip(`● error · ${esc(s.message)}`, 'red');
  return chip('○ logged out');
}

function ssoPanel(secret) {
  const s = peek('sso', secret);
  if (s?.state !== 'pending') return '';
  return `<div class="grid sm:grid-cols-[auto_1fr] items-center gap-6 bg-base rounded-2xl p-6"><div><p class="${lbl} mb-1">Device code</p><p class="font-mono text-4xl tracking-[0.2em] text-text">${esc(s.user_code)}</p></div>
    <div class="flex flex-col gap-3"><p class="text-sm">Open the AWS verification page and confirm the code. anbu polls in the background and the session turns active on its own.</p>
    <div class="flex flex-wrap gap-2"><a href="${esc(s.verification_uri)}" target="_blank" rel="noopener" class="${primary}"><i data-lucide="external-link" class="size-4"></i>Open verification page</a><span class="inline-flex items-center gap-2 text-sm text-yellow"><i data-lucide="loader-circle" class="size-4 animate-spin"></i>pending · <span data-until="${esc(s.expires_at)}"></span> left</span></div></div></div>`;
}

function secretDetail(name) {
  return needD(get('secret', name), (s) => {
    const types = peek('types');
    get('types');
    const t = types?.[s.type];
    const refs = s.used_by ?? [];
    let extra = '';
    if (s.type === 'aws-sso') {
      const status = peek('sso', s.name);
      const first = s.profiles?.[0];
      extra = `<section class="mb-8"><div class="flex flex-wrap items-center gap-3 mb-3"><h3 class="text-sm font-medium text-subtext1">SSO session</h3>${ssoState(s.name)}<span class="flex-1"></span><button type="button" data-sso-login="${esc(s.name)}" class="${ghost}"><i data-lucide="log-in" class="size-4"></i>${status?.state === 'active' ? 'Log in again' : 'Log in'}</button></div>
        ${ssoPanel(s.name)}
        <h3 class="${h3} mt-6">Profiles</h3>${s.profiles?.length ? '' : `<p class="px-4 text-sm">No profiles yet. Log in, then use Edit and Fill from SSO accounts.</p>`}<ul class="flex flex-col gap-1">${(s.profiles ?? []).map((p) => `<li class="grid grid-cols-2 md:grid-cols-[10rem_1fr_1fr_8rem_auto] items-center gap-x-4 gap-y-1 rounded-xl px-4 py-3 hover:bg-surface0"><span class="font-mono text-sm text-text truncate">${esc(s.name)}:${esc(p.name)}</span><span class="text-sm truncate">${esc(p.role_name)}</span><span class="font-mono text-sm">${esc(p.account_id)}</span><span class="text-sm">${esc(p.region)}</span><span class="flex items-center justify-end gap-2">${copyBtn(`credential_process = anbu api secrets get ${s.name}:${p.name}`)}</span></li>`).join('')}</ul>
        ${first ? `<p class="${lbl} mt-4 mb-2">Use from a local AWS profile</p>${codeBlock(well, esc(`[profile ${s.name}-${first.name}]\ncredential_process = anbu api secrets get ${s.name}:${first.name}`))}` : ''}</section>`;
    }
    const fieldRows = (t?.fields ?? []).map((f) => fieldRow(f.key, s.fields?.[f.key], f.kind)).join('');
    const confirm = st.confirm?.kind === 'delete-secret' ? st.confirm : null;
    let confirmPanel = '';
    if (confirm?.error) {
      confirmPanel = `<section class="rounded-2xl bg-red/10 p-5 mb-8 text-sm"><p class="text-red font-medium mb-1">409 · ${esc(confirm.error.message)}</p><p>Remove these references first: ${esc((confirm.error.data?.used_by ?? refs).join(', '))}</p></section>`;
    } else if (confirm) {
      confirmPanel = `<section class="flex flex-wrap items-center gap-3 rounded-2xl bg-red/10 p-5 mb-8 text-sm"><p class="flex-1 min-w-48">Delete <span class="font-mono text-text">${esc(s.name)}</span>? This cannot be undone.</p><button type="button" data-act="cancel" class="${ghost}">Cancel</button><button type="button" data-act="confirm-delete-secret" class="inline-flex items-center gap-2 rounded-full bg-red text-crust font-medium text-sm px-4 py-2 hover:opacity-90"><i data-lucide="trash-2" class="size-4"></i>Delete</button></section>`;
    }
    const badge = `<span class="grid place-items-center size-14 shrink-0 rounded-2xl bg-${meta(s.type).color}/15 text-${meta(s.type).color}">${typeIcon(s.type, 'size-6 text-2xl')}</span>`;
    return article(`<header class="flex flex-wrap items-center gap-4 mb-8">${badge}
      <div class="flex-1 min-w-0"><h2 class="${h2} truncate">${esc(s.name)}</h2><p class="${lbl}">${esc(t?.label ?? s.type)} · updated ${rel(s.updated_at)}</p></div>
      <div class="flex gap-2 max-sm:w-full"><button type="button" data-act="edit-secret" class="${ghost}"><i data-lucide="pencil" class="size-4"></i>Edit</button><button type="button" data-act="delete-secret" class="${danger}"><i data-lucide="trash-2" class="size-4"></i>Delete</button></div></header>
      ${confirmPanel}
      ${s.totp ? totpBand(s) : ''}
      ${extra}
      ${fieldRows ? `<section class="mb-8"><h3 class="${h3} px-4">Fields</h3><dl class="grid xl:grid-cols-2 gap-x-4">${fieldRows}</dl></section>` : ''}
      ${s.custom?.length ? `<section class="mb-8"><h3 class="${h3} px-4">Custom fields</h3><dl class="grid xl:grid-cols-2 gap-x-4">${s.custom.map((c) => fieldRow(c.name, c.value, c.hidden ? 'secret' : 'text')).join('')}</dl></section>` : ''}
      ${refs.length ? `<section class="mb-8 px-4"><h3 class="${h3}">Used by</h3><div class="flex flex-wrap gap-2">${refs.map((r) => chip(esc(r), 'blue')).join('')}</div></section>` : ''}
      <p class="px-4 ${lbl} font-mono">id ${esc(s.id)}</p>`);
  });
}

const profileRow = (p = {}) => `<div data-prow data-custom="${esc(JSON.stringify(p.custom ?? []))}" class="grid grid-cols-2 lg:grid-cols-[repeat(4,minmax(0,1fr))_auto] items-end gap-3">${['name', 'account_id', 'role_name', 'region'].map((k) => `<label class="flex flex-col gap-1.5 ${lbl}">${k}<input data-p="${k}" value="${esc(p[k] ?? '')}" class="${field} font-mono"></label>`).join('')}<button type="button" data-act="del-row" title="Remove profile" class="justify-self-start grid place-items-center size-9 rounded-full text-overlay1 hover:text-red hover:bg-surface1"><i data-lucide="x" class="size-4"></i></button></div>`;
const customRow = (c = {}) => `<div data-crow class="flex gap-2"><input data-c="name" class="${field}" placeholder="name" value="${esc(c.name ?? '')}"><input data-c="value" type="${c.hidden ? 'password' : 'text'}" class="${field}" placeholder="value" value="${esc(c.value ?? '')}"><button type="button" data-act="toggle-hidden" data-hidden="${c.hidden ? '1' : '0'}" class="shrink-0 rounded-lg px-3 ${c.hidden ? 'bg-surface1 text-text' : 'bg-surface0 text-overlay1'} hover:text-text" title="Hidden"><i data-lucide="${c.hidden ? 'eye-off' : 'eye'}" class="size-4"></i></button><button type="button" data-act="del-row" title="Remove field" class="shrink-0 rounded-lg px-2 text-overlay1 hover:text-red"><i data-lucide="x" class="size-4"></i></button></div>`;

function secretForm(s) {
  return needD(get('types'), (types) => {
    const type = s ? s.type : types[st.newType] ? st.newType : Object.keys(types)[0];
    const t = types[type];
    const val = (k) => s?.fields?.[k] ?? '';
    const input = (f) => {
      const name = `${esc(f.key)}${f.required ? ' *' : ''}`;
      if (f.kind.includes('multiline')) return `<label class="md:col-span-2 flex flex-col gap-1.5 ${lbl}">${name}<textarea data-f="${esc(f.key)}" class="${field} font-mono min-h-32">${esc(val(f.key))}</textarea></label>`;
      if (f.kind === 'date') return `<div class="flex flex-col gap-1.5 ${lbl}">${name}${datePicker({ attrs: `data-f="${esc(f.key)}"`, value: val(f.key), cls: field })}</div>`;
      return `<label class="flex flex-col gap-1.5 ${lbl}">${name}<input data-f="${esc(f.key)}" type="${f.kind === 'secret' ? 'password' : 'text'}" value="${esc(val(f.key))}" class="${field} font-mono"></label>`;
    };
    const chips = s ? '' : `<div class="flex flex-wrap gap-1.5 mb-8">${Object.entries(types).map(([k, v]) => `<button type="button" data-newtype="${esc(k)}" class="inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm ${k === type ? `bg-${meta(k).color}/15 text-${meta(k).color}` : 'bg-surface0 text-overlay1 hover:text-text'}">${typeIcon(k)}${esc(v.label)}</button>`).join('')}</div>`;
    const ssoLive = s && t.profiles && get('sso', s.name).data?.state === 'active';
    return article(`<h2 class="${h2} mb-6">${s ? `Edit ${esc(s.name)}` : 'New secret'}</h2>${chips}
      <form data-secret-form data-type="${esc(type)}" autocomplete="off">
      <div class="grid md:grid-cols-2 gap-4 mb-8"><label class="md:col-span-2 flex flex-col gap-1.5 ${lbl}">name *<input name="name" ${s ? `value="${esc(s.name)}"` : 'data-keep="newName"'} class="${field} font-mono" placeholder="no / or :"></label>${t.fields.map(input).join('')}</div>
      ${type === 'ssh-key' && !s ? `<button type="button" data-act="gen-ssh-key" class="${ghost} mb-8"><i data-lucide="sparkles" class="size-4"></i>Generate an Ed25519 key instead</button>` : ''}
      ${t.profiles ? `<section class="mb-8"><div class="flex items-center gap-2 mb-3"><h3 class="flex-1 text-sm font-medium text-subtext1">Profiles</h3><button type="button" data-act="fill-profiles" class="${ghost}" ${ssoLive ? '' : `disabled title="${s ? 'Log in to this SSO session first' : 'Save the secret and log in first'}"`}><i data-lucide="list-plus" class="size-4"></i>Fill from SSO accounts</button></div>
        <div data-prows class="flex flex-col gap-4 bg-base rounded-2xl p-4">${(s?.profiles?.length ? s.profiles : [{}]).map(profileRow).join('')}</div><button type="button" data-act="add-profile" class="mt-3 text-sm text-overlay1 hover:text-text">+ Add profile</button></section>` : ''}
      <section class="grid md:grid-cols-2 gap-4 mb-8"><label class="flex flex-col gap-1.5 ${lbl} self-start">TOTP (base32 secret or otpauth:// URI)<input name="totp" value="${esc(s?.totp?.secret ?? '')}" class="${field} font-mono" placeholder="otpauth://totp/..."></label>
        <div class="flex flex-col gap-1.5 ${lbl}">Custom fields<div data-crows class="flex flex-col gap-2">${(s?.custom?.length ? s.custom : [{}]).map(customRow).join('')}</div><button type="button" data-act="add-custom" class="self-start mt-1 text-sm text-overlay1 hover:text-text">+ Add field</button></div></section>
      <div class="flex flex-wrap gap-2"><button class="${primary}"><i data-lucide="save" class="size-4"></i>Save secret</button>${s ? `<button type="button" data-act="cancel" class="${ghost}">Cancel</button>` : ''}</div></form>`);
  });
}

function collectSecret(form, orig) {
  const types = peek('types');
  const type = form.dataset.type;
  const fields = {};
  for (const el of form.querySelectorAll('[data-f]')) fields[el.dataset.f] = el.value;
  const body = { name: form.elements.name.value.trim(), type, fields };
  body.custom = [...form.querySelectorAll('[data-crow]')].map((r) => ({
    name: r.querySelector('[data-c="name"]').value.trim(),
    value: r.querySelector('[data-c="value"]').value,
    hidden: r.querySelector('[data-hidden]').dataset.hidden === '1',
  })).filter((c) => c.name || c.value);
  if (types[type].profiles) {
    body.profiles = [...form.querySelectorAll('[data-prow]')].map((r) => {
      const p = { custom: JSON.parse(r.dataset.custom || '[]') };
      for (const el of r.querySelectorAll('[data-p]')) p[el.dataset.p] = el.value.trim();
      return p;
    }).filter((p) => p.name || p.account_id || p.role_name || p.region);
  }
  const totp = form.elements.totp.value.trim();
  body.totp = orig?.totp && totp === orig.totp.secret ? orig.totp : totp;
  return body;
}

function tasksDetail() {
  const sel = st.sel.tasks;
  const titles = Object.fromEntries(taskFilters.map(([id, n]) => [id, n]));
  return needD(get('tasks'), (all) => {
    const filters = {
      open: (t) => !t.done, overdue: (t) => isOverdue(t), week: dueThisWeek, nodate: (t) => !t.done && !t.due, done: (t) => t.done,
      critical: (t) => !t.done && t.priority === 'critical', high: (t) => !t.done && t.priority === 'high', default: (t) => !t.done && t.priority === 'default',
    };
    const list = all.filter(filters[sel] ?? filters.open);
    return article(`<h2 class="${h2} mb-6 capitalize">${titles[sel] ?? sel}</h2>
      <form data-addtask class="flex flex-wrap items-center gap-2 bg-surface0 rounded-3xl md:rounded-full p-1.5 mb-6">
        <input name="text" autocomplete="off" class="flex-1 min-w-48 bg-transparent px-3 py-1.5 text-sm text-text outline-none pointer-coarse:text-[16px]" placeholder="Add a task and press Enter">
        <span class="inline-flex rounded-full bg-crust/40 p-0.5">${priorities.slice().reverse().map((p) => `<button type="button" data-prio="${p}" class="rounded-full px-3 py-1 text-xs text-overlay1">${p}</button>`).join('')}</span>
        ${datePicker({ attrs: 'name="due"', cls: 'rounded-full px-3 py-1.5 text-xs text-subtext0 hover:bg-surface1 outline-none focus:ring-2 focus:ring-mauve', placeholder: 'Due date', align: 'right' })}
        <button class="grid place-items-center size-8 rounded-full bg-mauve text-crust" title="Add task"><i data-lucide="arrow-up" class="size-4"></i></button></form>
      <ul class="flex flex-col gap-1">${list.map(taskRow).join('') || `<li class="px-4 py-10 text-center ${lbl}">Nothing here</li>`}</ul>`);
  });
}

const prioClass = (p) => `rounded-full px-3 py-1 text-xs ${st.newPrio === p ? (p === 'critical' ? 'bg-red/20 text-red' : p === 'high' ? 'bg-peach/20 text-peach' : 'bg-surface1 text-text') : 'text-overlay1'}`;

function paintPrio() {
  for (const b of document.querySelectorAll('[data-prio]')) b.className = prioClass(b.dataset.prio);
}

function taskRow(t) {
  const over = isOverdue(t);
  const pc = t.priority === 'critical' ? 'red' : t.priority === 'high' ? 'peach' : '';
  return `<li class="group flex items-center gap-3 rounded-xl px-4 py-3 ${over ? 'bg-red/10' : 'hover:bg-surface0'}">
    <button type="button" data-toggle="${esc(t.id)}" title="${t.done ? 'Reopen' : 'Complete'}" class="grid place-items-center size-5 shrink-0 rounded-full ${t.done ? 'bg-green text-crust' : 'bg-surface1 hover:bg-surface2'}">${t.done ? '<i data-lucide="check" class="size-3.5"></i>' : ''}</button>
    <span class="flex-1 min-w-0 text-sm ${t.done ? 'line-through text-overlay0' : 'text-text'}">${esc(t.text)}</span>
    ${pc && !t.done ? chip(t.priority, pc) : ''}
    ${t.due ? `<span class="w-24 text-right text-xs ${over ? 'text-red font-medium' : 'text-overlay1'}">${dueLabel(t.due)}</span>` : '<span class="w-24"></span>'}
    <button type="button" data-del-task="${esc(t.id)}" title="Delete" class="opacity-0 group-hover:opacity-100 max-md:opacity-100 rounded-full p-1 text-overlay1 hover:text-red"><i data-lucide="x" class="size-4"></i></button></li>`;
}

function hostForm(h) {
  return needD(get('secrets'), (secrets) => {
    const keys = secrets.filter((s) => s.type === 'ssh-key').map((s) => [s.id, s.name]);
    const v = (k, d = '') => esc(h?.[k] ?? d);
    const confirm = st.confirm?.kind === 'delete-host';
    return article(`<h2 class="${h2} mb-6">${h ? `Edit ${esc(h.name)}` : 'New host'}</h2><form data-host-form autocomplete="off"><div class="grid md:grid-cols-2 xl:grid-cols-4 gap-4 mb-8">${[['name', 'nas'], ['address', '192.168.1.20'], ['port', '22'], ['user', 'admin']].map(([k, p]) => `<label class="flex flex-col gap-1.5 ${lbl}">${k}<input name="${k}" ${k === 'port' ? 'inputmode="numeric"' : ''} value="${v(k, k === 'port' ? '22' : '')}" class="${field} font-mono" placeholder="${p}"></label>`).join('')}<div class="md:col-span-2 flex flex-col gap-1.5 ${lbl}">key secret${keys.length ? dropdown({ attrs: 'name="key_secret"', options: keys, value: h?.key_secret, cls: field }) : `<p class="py-2 text-sm">No SSH key secrets yet. <button type="button" data-act="new-ssh-key" class="text-mauve hover:underline">Create one</button></p>`}</div></div>
      <div class="flex flex-wrap gap-2"><button class="${primary}"><i data-lucide="save" class="size-4"></i>Save host</button>${h ? `<button type="button" data-act="cancel" class="${ghost}">Cancel</button><span class="flex-1"></span>${confirm ? `<button type="button" data-act="confirm-delete-host" class="inline-flex items-center gap-2 rounded-full bg-red text-crust font-medium text-sm px-4 py-2"><i data-lucide="trash-2" class="size-4"></i>Confirm delete</button>` : `<button type="button" data-act="delete-host" class="${danger}"><i data-lucide="trash-2" class="size-4"></i>Delete host</button>`}` : ''}</div></form>`);
  });
}

function sameMachine(targets) {
  const i = st.sel.ssh.indexOf('/');
  if (i < 0) return undefined;
  const alias = peek('machines', st.sel.ssh.slice(0, i))?.find((m) => m.name === st.sel.ssh.slice(i + 1))?.alias;
  return alias ? targets.find((x) => x.kind === 'machine' && x.alias === alias) : undefined;
}

function sshDetail() {
  if (st.view.ssh === 'new') return hostForm(null);
  get('secrets');
  return needD(get('targets'), ({ targets, errors }) => {
    const t = targets.find((x) => x.ref === st.sel.ssh) ?? sameMachine(targets);
    const failed = !t && errors.find((x) => st.sel.ssh.startsWith(`${x.profile}/`));
    if (failed) {
      const p = profiles().find((x) => x.ref === failed.profile);
      return article(failed.status === 401 && p?.kind === 'sso' ? ssoRequired(p.secret) : failure(Object.assign(new Error(failed.error), { status: failed.status }), 'targets'));
    }
    if (!t) return st.sel.ssh ? empty(`No SSH target named ${esc(st.sel.ssh)}`) : empty('No SSH targets yet', `<button type="button" data-act="new-host" class="${primary}"><i data-lucide="plus" class="size-4"></i>New host</button>`);
    if (st.editing && t.kind === 'host') return hostForm({ ...t, port: t.port });
    const tabs = seg('sshtab', [['terminal', 'Terminal'], ['info', 'Details']], st.sshTab, 'data-local');
    const live = t.state === 'running' || t.state === 'ready';
    const termPane = live ? `<div id="term" data-ref="${esc(t.ref)}" data-target="${esc(JSON.stringify(t.target))}" class="${st.sshTab === 'info' ? 'hidden ' : ''}flex-1 min-h-80 bg-base rounded-2xl overflow-hidden"></div>` : '';
    let body = termPane;
    if (st.sshTab === 'info') {
      body = `<dl class="grid xl:grid-cols-2 gap-x-4">${fieldRow('target', t.ref, 'text')}${fieldRow('address', t.addr || '-', 'text')}${fieldRow('user', t.user, 'text')}${fieldRow('key', t.key, 'text')}${fieldRow('known_hosts alias', t.alias, 'text')}</dl>
        <div class="flex flex-wrap gap-2 mt-6 px-4">${t.kind === 'host' ? `<button type="button" data-act="edit-host" class="${ghost}"><i data-lucide="pencil" class="size-4"></i>Edit host</button>` : ''}<button type="button" data-act="forget-key" data-alias="${esc(t.alias)}" class="${ghost}"><i data-lucide="shield-off" class="size-4"></i>Forget host key</button></div>${termPane}`;
    } else if (!live) {
      body = `<div class="flex-1 grid place-items-center bg-base rounded-2xl p-10 text-center"><div><p class="text-text mb-1">${esc(t.name)} is ${esc(t.state)}</p><p class="${lbl} mb-4">409 · the machine has no public IP until it starts</p><button type="button" data-href="${esc(`/machines/${segEnc(t.profile)}/${segEnc(t.name)}`)}" class="${ghost}"><i data-lucide="play" class="size-4"></i>Open in Machines</button></div></div>`;
    }
    return article(`<header class="flex flex-wrap items-center gap-4 mb-6"><div class="flex-1 min-w-0"><h2 class="${h2} truncate">${esc(t.name)}</h2><p class="${lbl} font-mono truncate">${esc(`${t.user}@${t.addr || '-'}`)} · ${esc(t.kind === 'machine' ? t.profile : 'stored host')}</p></div>${tabs}</header>${body}`, 'min-h-full flex flex-col');
  }, 'Loading hosts and machines');
}

function runner(key) {
  return `<section class="flex flex-col"><h3 class="${h3}">Command runner</h3><form data-run="${esc(key)}" class="flex items-center gap-2 bg-surface0 rounded-full pl-4 p-1.5 mb-3"><span class="font-mono text-sm text-overlay1">aws</span><input name="cmd" autocomplete="off" spellcheck="false" class="flex-1 min-w-0 bg-transparent font-mono text-sm text-text outline-none pointer-coarse:text-[16px]" value="${esc(st.awsCmd)}"><button class="${primary}"><i data-lucide="play" class="size-4"></i>Run</button></form>
    <div class="flex flex-wrap gap-1.5 mb-4">${['sts get-caller-identity', 's3 ls', 'ec2 describe-instances', 'iam list-users'].map((c) => `<button type="button" data-quickaws="${c}" class="rounded-full bg-surface0 px-3 py-1 font-mono text-xs text-overlay1 hover:text-text">${c}</button>`).join('')}</div>
    <div data-run-out></div></section>`;
}

function paintRun() {
  const el = document.querySelector('[data-run-out]');
  if (!el) return;
  const key = el.closest('section').querySelector('[data-run]').dataset.run;
  const r = st.run?.key === key ? st.run : null;
  if (!r) { el.innerHTML = ''; return; }
  if (r.loading) { el.innerHTML = loadingLine(`Running aws ${esc(r.command)}`); }
  else if (r.error) { el.innerHTML = failure(r.error); }
  else {
    const out = r.out;
    let stdout = '';
    if (out.stdout) { try { stdout = highlightJSON(JSON.stringify(JSON.parse(out.stdout), null, 2)); } catch { stdout = esc(out.stdout); } }
    el.innerHTML = `<div class="flex items-center gap-2 mb-2">${chip(`exit ${out.exit_code}`, out.exit_code ? 'red' : 'green')}${/(^|\s)--output(\s|=|$)/.test(r.command) ? '' : `<span class="${lbl}">--output json added</span>`}</div>${codeBlock(`${well} max-h-[60vh]`, `${stdout}${out.stderr ? `<span class="text-red">${esc(out.stderr)}</span>` : ''}` || `<span class="text-overlay1">no output</span>`)}`;
  }
  lucide.createIcons();
}

async function runAws(key, command, source) {
  st.awsCmd = command;
  st.run = { key, command, loading: true };
  paintRun();
  try {
    const out = await api('POST', '/api/aws/run', { source, command });
    if (st.run?.command === command) st.run = { key, command, out };
  } catch (err) {
    if (st.run?.command === command) st.run = { key, command, error: err };
  }
  paintRun();
}

function needAws(e, fn, text) {
  const p = profiles().find((x) => x.ref === st.profile);
  if (e.error?.status === 401 && p?.kind === 'sso') return article(ssoRequired(p.secret));
  return needD(e, fn, text);
}

function ssoRequired(secret) {
  return `<div class="flex flex-wrap items-center gap-4 rounded-2xl bg-yellow/10 p-5 text-sm"><p class="flex-1 min-w-48"><span class="text-yellow font-medium">401 · SSO login required.</span> Log in to the <span class="font-mono text-text">${esc(secret)}</span> session to use its profiles.</p><button type="button" data-sso-login="${esc(secret)}" class="${ghost} !bg-surface1"><i data-lucide="log-in" class="size-4"></i>SSO login</button></div>`;
}

function awsDetail() {
  if (st.view.aws === 'adhoc') {
    return article(`<h2 class="${h2} mb-2">Ad-hoc keys</h2><p class="text-sm mb-6">Keys typed here ride on one request and are never stored.</p>
      <form data-adhoc class="grid md:grid-cols-2 xl:grid-cols-4 gap-4 mb-10" autocomplete="off">${['access_key_id', 'secret_access_key', 'session_token', 'region'].map((k) => `<label class="flex flex-col gap-1.5 ${lbl}">${k}<input name="${k}" type="${k.includes('secret') || k.includes('token') ? 'password' : 'text'}" class="${field} font-mono"></label>`).join('')}</form>${runner('adhoc')}`);
  }
  return needD(get('secrets'), () => {
    const p = profiles().find((x) => x.ref === st.sel.aws);
    if (!p) return st.sel.aws ? empty(`No AWS profile named ${esc(st.sel.aws)}`) : empty('No AWS profiles yet', `<button type="button" data-act="new-aws-secret" class="${primary}"><i data-lucide="plus" class="size-4"></i>New AWS secret</button>`);
    const sso = p.kind === 'sso';
    if (sso) get('sso', p.secret);
    const w = get('whoami', p.ref);
    let identity;
    if ('data' in w) identity = facts([['Account', esc(w.data.account), 1], ['ARN', esc(w.data.arn), 1], ['Region', esc(w.data.region), 1]], 'sm:grid-cols-[auto_1fr_auto]');
    else if (w.error?.status === 401 && sso) identity = peek('sso', p.secret)?.state === 'pending' ? '' : ssoRequired(p.secret);
    else if (w.error) identity = failure(w.error, w.key);
    else identity = facts([['Account', '…'], ['ARN', '…'], ['Region', esc(p.region)]], 'sm:grid-cols-[auto_1fr_auto]');
    const color = sso ? 'yellow' : 'peach';
    return article(`<header class="flex flex-wrap items-center gap-4 mb-6"><span class="grid place-items-center size-14 rounded-2xl bg-${color}/15 text-${color} text-2xl"><i class="fa-brands fa-aws"></i></span><div class="flex-1 min-w-0"><h2 class="${h2} font-mono truncate">${esc(p.ref)}</h2><p class="${lbl}">${sso ? `${esc(p.role_name)} via SSO` : 'static keys'} · ${esc(p.region)}</p></div>
      <div class="flex flex-wrap gap-2 max-sm:w-full">${sso ? `<button type="button" data-sso-login="${esc(p.secret)}" class="${ghost}"><i data-lucide="log-in" class="size-4"></i>SSO login</button><button type="button" data-copy="${esc(`credential_process = anbu api secrets get ${p.ref}`)}" class="${ghost}"><i data-lucide="terminal-square" class="size-4"></i>credential_process</button>` : ''}</div></header>
      ${sso && ssoPanel(p.secret) ? `<div class="mb-8">${ssoPanel(p.secret)}</div>` : ''}
      ${identity ? `<div class="mb-10">${identity}</div>` : ''}
      ${runner(p.ref)}`);
  });
}

function shapeLabel(sh) {
  const [c, m] = sh.match(/\d+/g);
  return `${c} vCPU · ${m} GB`;
}

function shapeGrid(o, arch, cls, chosen, attr) {
  return `<div class="grid grid-cols-[repeat(auto-fill,minmax(11.5rem,1fr))] gap-2">${o.shapes.map((sh) => {
    const x = o.types?.[arch]?.[sh]?.[cls];
    return `<button type="button" ${attr}="${sh}" ${x ? '' : 'disabled'} class="rounded-xl px-3 py-2.5 text-left ${sh === chosen ? 'bg-mauve/20 text-mauve' : x ? 'bg-surface0 hover:bg-surface1' : 'bg-surface0/40 opacity-40 cursor-not-allowed'}"><span class="block text-sm text-text">${shapeLabel(sh)}</span><span class="block ${lbl} font-mono">${x ? `${x.type} · $${(x.hourly * 730).toFixed(0)}/mo` : 'not offered'}</span></button>`;
  }).join('')}</div>`;
}

function scaffoldBlock(sc) {
  if (!sc.ready) return 'Set up the scaffold on this account and region first.';
  if (sc.key.state === 'missing') return 'Import the key for this scaffold first.';
  if (sc.key.state === 'differs') return 'The scaffold key pair differs from the key anbu holds.';
  return '';
}

function jobsView() {
  const icon = (x) => x.status === 'running' ? '<i data-lucide="loader-circle" class="size-4 text-yellow animate-spin"></i>' : x.status === 'queued' ? '<i data-lucide="clock" class="size-4 text-overlay1"></i>' : x.status === 'failed' ? '<i data-lucide="circle-x" class="size-4 text-red"></i>' : '<i data-lucide="circle-check" class="size-4 text-green"></i>';
  const jobRow = (x) => `<li><button type="button" data-openjob="${esc(x.id)}" class="w-full grid grid-cols-[auto_1fr_auto] items-center gap-3 rounded-xl px-4 py-3 text-left hover:bg-surface0">${icon(x)}<span class="min-w-0"><span class="block text-sm text-text truncate">${esc(x.label)}</span><span class="block ${lbl} font-mono">${esc(x.kind)} · ${relTime(x.finished ?? x.started ?? x.queued)}</span></span>${x.status === 'failed' ? chip('failed', 'red') : ''}</button>
    ${st.openJob === x.id && (x.result || x.error) ? `<div class="mx-4 mb-2">${codeBlock(well, `${x.result ? highlightJSON(x.result) : ''}${x.result && x.error ? '\n' : ''}${x.error ? `<span class="text-red">${esc(x.error)}</span>` : ''}`)}</div>` : ''}</li>`;
  const none = `<li class="px-4 py-2 ${lbl}">None</li>`;
  return article(`<h2 class="${h2} mb-1">Job pipeline</h2><p class="text-sm mb-8">${jobSummary(jobs) || 'Idle'} · jobs run one at a time, in order</p>
    <h3 class="${h3} px-4">Running</h3><ul class="mb-6">${jobs.running ? jobRow(jobs.running) : none}</ul><h3 class="${h3} px-4">Upcoming</h3><ul class="mb-6">${jobs.queued.map(jobRow).join('') || none}</ul><h3 class="${h3} px-4">Finished</h3><ul>${jobs.finished.map(jobRow).join('') || none}</ul>`);
}

function scaffoldView() {
  return needAws(get('scaffold', st.profile), (sc) => {
    const res = (icon, color, name, id) => `<li class="flex items-center gap-3 py-1 min-w-0"><i data-lucide="${icon}" class="size-4 shrink-0 text-${color}"></i><span class="font-mono text-sm text-text">${esc(name)}</span><span class="${lbl} font-mono truncate">${esc(id)}</span></li>`;
    const ks = sc.key.state;
    const keyRow = ks === 'match'
      ? `<li class="flex items-center gap-3 py-1 min-w-0"><i data-lucide="circle-check" class="size-4 shrink-0 text-green"></i><span class="font-mono text-sm text-text">scaffold key</span><span class="${lbl} font-mono truncate">${esc(sc.key.fingerprint ?? '')}</span></li>`
      : ks === 'none' ? '' : `<li class="flex items-center gap-3 py-1"><i data-lucide="circle-alert" class="size-4 text-${ks === 'differs' ? 'red' : 'peach'}"></i><span class="font-mono text-sm text-text">scaffold key</span><span class="text-xs text-${ks === 'differs' ? 'red' : 'peach'}">${ks === 'differs' ? 'differs from the key pair' : 'not in anbu'}</span></li>`;
    const missing = `<section class="rounded-2xl bg-peach/10 p-5 mb-8 text-sm"><p class="text-peach font-medium mb-1">Key pair has no key in anbu</p><p class="mb-4">This account and region was set up outside anbu, most likely by sharingan. Paste the private key its machines authorize, and setup and create will accept it. For sharingan, that key is <span class="font-mono text-text">~/.config/sharingan/id_ed25519</span>.</p><textarea data-keep="adoptKey" spellcheck="false" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" class="${field} font-mono min-h-32 mb-3"></textarea><button type="button" data-act="adopt-key" class="${ghost} !bg-surface1"><i data-lucide="upload" class="size-4"></i>Import key</button></section>`;
    const differs = `<section class="rounded-2xl bg-red/10 p-5 mb-8 text-sm"><p class="text-red font-medium mb-1">Key pair differs from the scaffold key</p><p>The <span class="font-mono text-text">sharingan-key</span> key pair on this account and region does not match the key anbu holds${sc.key.fingerprint ? ` (<span class="font-mono text-text">${esc(sc.key.fingerprint)}</span>)` : ''}. Machines launched here authorize the other key, so setup and create refuse to run.</p></section>`;
    const any = sc.resources.some((r) => r.id);
    const confirm = st.confirm?.kind === 'teardown';
    return article(`<h2 class="${h2} mb-1">Scaffold</h2><p class="text-sm font-mono mb-1">${esc(sc.account)} · ${esc(sc.region)}</p><p class="${lbl} mb-8">Shared by every profile on this account and region, and by every machine it launches</p>
      ${ks === 'missing' ? missing : ks === 'differs' ? differs : ''}
      ${any ? `<ul class="grid sm:grid-cols-2 xl:grid-cols-3 gap-3 mb-8">${sc.resources.map((r) => r.id ? res('circle-check', 'green', r.resource, r.id) : res('circle-dashed', 'overlay0', r.resource, 'missing')).join('')}${keyRow}</ul>` : `<p class="bg-base rounded-2xl p-6 mb-8 text-sm">Nothing is set up on this account and region yet. Setup creates the VPC, gateway, subnet, route table, and security group, plus a new SSH key that anbu keeps for this scaffold and imports as the sharingan-key key pair.</p>`}
      <div class="flex flex-wrap gap-2">${ks !== 'missing' && ks !== 'differs' ? `<button type="button" data-act="setup-scaffold" class="${primary}"><i data-lucide="hammer" class="size-4"></i>${sc.ready ? 'Re-run setup' : 'Set up'}</button>` : ''}${any ? (confirm ? `<button type="button" data-act="cancel" class="${ghost}">Cancel</button><button type="button" data-act="confirm-teardown" class="inline-flex items-center gap-2 rounded-full bg-red text-crust font-medium text-sm px-4 py-2"><i data-lucide="trash-2" class="size-4"></i>Confirm tear down</button>` : `<button type="button" data-act="teardown" class="${danger}"><i data-lucide="trash-2" class="size-4"></i>Tear down</button>`) : ''}</div>
      ${confirm ? `<p class="mt-3 text-sm">Tear down deletes every scaffold resource and the scaffold key. It refuses while any machine exists.</p>` : ''}`);
  }, 'Checking scaffold');
}

function newMachineView() {
  return needAws(get('options', st.profile), (o) => {
    const x = o.types?.[st.mArch]?.[st.mShape]?.[st.mClass];
    const sc = peek('scaffold', st.profile);
    get('scaffold', st.profile);
    const blocked = sc ? scaffoldBlock(sc) : '';
    return article(`<h2 class="${h2} mb-6">New machine</h2>
      <div class="grid xl:grid-cols-[1fr_20rem] gap-8"><div class="flex flex-col gap-6">
      <div class="grid md:grid-cols-2 gap-4 max-w-3xl">${accountPicker()}<label class="flex flex-col gap-1.5 ${lbl}">name<input data-keep="mName" class="${field} font-mono" placeholder="devbox"></label></div>
      <div class="flex flex-wrap gap-6"><div><p class="${lbl} mb-2">Architecture</p>${seg('arch', o.arches.map((a) => [a, a]), st.mArch, 'data-local')}</div><div><p class="${lbl} mb-2">Class</p>${seg('class', o.classes.map((c) => [c, c[0].toUpperCase() + c.slice(1)]), st.mClass, 'data-local')}</div></div>
      <div><p class="${lbl} mb-2">Shape</p>${shapeGrid(o, st.mArch, st.mClass, st.mShape, 'data-shape')}</div>
      <div><p class="${lbl} mb-2">Disk (gp3, fixed after create)</p><div class="flex flex-wrap gap-1.5">${o.disks.map((d) => `<button type="button" data-disk="${d}" class="rounded-full px-4 py-1.5 text-sm ${d === st.mDisk ? 'bg-surface1 text-text' : 'bg-surface0 text-overlay1 hover:text-text'}">${d} GB</button>`).join('')}</div></div></div>
      <aside class="bg-base rounded-2xl p-5 h-fit"><p class="${lbl}">Resolved type</p><p class="font-mono text-xl text-text mb-4">${x?.type ?? 'none'}</p><p class="${lbl}">Running</p><p class="text-3xl font-display text-text">${money(x?.monthly?.[st.mDisk] ?? 0)}<span class="text-sm text-overlay1">/mo</span></p><p class="${lbl} mt-3">Stopped</p><p class="text-lg text-text mb-6">${money(o.stopped?.[st.mDisk] ?? 0)}/mo</p>${!sc ? `<p class="${lbl} mb-3">Checking the scaffold</p>` : blocked ? `<p class="text-sm text-red mb-3">${blocked}</p><button type="button" data-view="scaffold" class="${ghost} w-full justify-center"><i data-lucide="network" class="size-4"></i>Open scaffold</button>` : `<button type="button" data-act="queue-create" class="${primary} w-full justify-center" ${x ? '' : 'disabled'}><i data-lucide="rocket" class="size-4"></i>Queue create</button>`}<p class="${lbl} mt-3">Ubuntu 26.04 · bootstrap installs zsh, Docker, Homebrew, and cps</p></aside></div>`);
  }, 'Loading sizes and prices. The first load per account and region takes a while');
}

function machineView() {
  return needAws(get('machines', st.profile), (list) => {
    const m = list.find((x) => x.name === st.sel.machines);
    if (!m) return st.sel.machines ? empty(`No machine named ${esc(st.sel.machines)} on this account and region`) : empty('No machines on this account and region', `<button type="button" data-act="new-machine" class="${primary}"><i data-lucide="plus" class="size-4"></i>New machine</button>`);
    const running = m.state === 'running';
    const busy = m.state === 'pending' || m.state === 'stopping';
    let panel = '';
    if (st.modify) {
      const o = get('options', st.profile);
      panel = `<section class="bg-surface0/40 rounded-2xl p-5 mb-8"><div class="flex flex-wrap items-center gap-3 mb-4"><h3 class="flex-1 text-sm font-medium text-subtext1">Modify ${esc(m.name)}</h3>${'data' in o ? seg('modclass', o.data.classes.map((c) => [c, c[0].toUpperCase() + c.slice(1)]), st.modify.class, 'data-local') : ''}</div>
        ${need(o, (od) => `${shapeGrid(od, m.arch, st.modify.class, st.modify.shape, 'data-modshape')}<p class="${lbl} mt-3 mb-4">The machine stops if running, changes type, and starts again. Disk size stays ${m.disk} GB.</p><div class="flex gap-2"><button type="button" data-act="queue-modify" class="${primary}" ${od.types?.[m.arch]?.[st.modify.shape]?.[st.modify.class] ? '' : 'disabled'}><i data-lucide="scaling" class="size-4"></i>Queue modify</button><button type="button" data-act="cancel" class="${ghost}">Cancel</button></div>`, 'Loading sizes and prices')}</section>`;
    } else if (st.remove) {
      panel = `<section class="rounded-2xl bg-red/10 p-5 mb-8 text-sm"><p class="text-red font-medium mb-1">Remove ${esc(m.name)}</p><p class="mb-4">The instance and its disk are terminated. Type the machine name to confirm.</p><div class="flex flex-wrap gap-2"><input data-keep="removeConfirm" autocomplete="off" class="${field} font-mono max-w-xs" placeholder="${esc(m.name)}"><button type="button" data-act="queue-remove" class="inline-flex items-center gap-2 rounded-full bg-red text-crust font-medium text-sm px-4 py-2"><i data-lucide="trash-2" class="size-4"></i>Queue remove</button><button type="button" data-act="cancel" class="${ghost}">Cancel</button></div></section>`;
    }
    const cls = m.class ?? (m.type.startsWith('t') ? 'burstable' : 'dedicated');
    return article(`<header class="flex flex-wrap items-center gap-4 mb-8"><div class="flex-1 min-w-0"><div class="flex items-center gap-3"><h2 class="${h2} truncate">${esc(m.name)}</h2>${chip(`${busy ? '<i data-lucide="loader-circle" class="size-3 animate-spin"></i>' : '●'} ${esc(m.state)}`, stateColor[m.state] ?? 'overlay1')}</div><p class="${lbl} font-mono">${esc(m.instance_id)} · ${esc(m.profile)}</p></div>
      <div class="flex flex-wrap gap-2 max-sm:w-full">${running ? `<button type="button" data-href="${esc(`/ssh/${segEnc(m.profile)}/${segEnc(m.name)}`)}" class="${primary}"><i data-lucide="terminal" class="size-4"></i>Terminal</button><button type="button" data-act="machine-stop" class="${ghost}"><i data-lucide="square" class="size-4"></i>Stop</button>` : `<button type="button" data-act="machine-start" class="${primary}" ${busy ? 'disabled' : ''}><i data-lucide="play" class="size-4"></i>Start</button>`}<button type="button" data-act="modify" class="${ghost}" ${busy ? 'disabled' : ''}><i data-lucide="scaling" class="size-4"></i>Modify</button><button type="button" data-act="remove" class="${danger}"><i data-lucide="trash-2" class="size-4"></i>Remove</button></div></header>
      ${panel}
      <div class="mb-10">${facts([['Type', m.type], ['Class', cls], ['Arch', m.arch], ['vCPU', m.vcpu || '-'], ['Memory', m.memory ? `${m.memory} GB` : '-'], ['Disk', `${m.disk} GB gp3`], ['Public IP', m.public_ip || '-'], ['Created', m.created ? ymd(new Date(m.created)) : '-'], ['Hourly', m.hourly ? `$${m.hourly.toFixed(4)}` : '-'], ['Monthly running', money(m.monthly_running)], ['Monthly stopped', money(m.monthly_stopped)], ['SSH alias', m.alias ?? '-']].map(([k, v]) => [k, esc(v), /IP|Type|alias/.test(k)]), 'grid-cols-2 lg:grid-cols-4 2xl:grid-cols-6')}</div>
      <section><div class="flex items-center gap-3 mb-3"><h3 class="flex-1 text-sm font-medium text-subtext1">Bootstrap status</h3><button type="button" data-act="probe" class="${ghost}" ${running ? '' : 'disabled'}><i data-lucide="activity" class="size-4"></i>Probe</button></div><div data-probe-out data-running="${running ? 1 : 0}"></div></section>`);
  }, 'Loading machines');
}

function paintProbe() {
  const el = document.querySelector('[data-probe-out]');
  if (!el) return;
  const p = st.probe?.name === st.sel.machines ? st.probe : null;
  let inner = el.dataset.running === '1' ? '<span class="text-overlay1">Probe runs cloud-init status and checks for the bootstrap marker over SSH</span>' : '<span class="text-overlay1">Start the machine to probe it</span>';
  if (p?.loading) inner = '<span class="text-overlay1">Probing over SSH…</span>';
  else if (p?.error) inner = `<span class="text-red">${esc(p.error.message)}</span>`;
  else if (p?.report !== undefined) inner = esc(p.report).replace(/bootstrap-done/g, '<span class="text-green">bootstrap-done</span>').replace(/bootstrap-pending/g, '<span class="text-yellow">bootstrap-pending</span>');
  el.innerHTML = codeBlock(well, inner);
  lucide.createIcons();
}

function machinesDetail() {
  if (!('data' in get('secrets'))) return needD(get('secrets'), () => '');
  if (!profiles().length) return empty('Machines run on an AWS profile from the vault', `<button type="button" data-act="new-aws-secret" class="${primary}"><i data-lucide="plus" class="size-4"></i>New AWS secret</button>`);
  const v = st.view.machines;
  if (v === 'jobs') return jobsView();
  if (v === 'scaffold') return scaffoldView();
  if (v === 'new') return newMachineView();
  return machineView();
}

function toolDetail() {
  const t = tools.find((x) => x.id === st.sel.tools) ?? tools[0];
  return needD(get('settings'), (s) => article(`<header class="flex flex-wrap items-center gap-3 mb-6"><h2 class="${h2}">${esc(t.name)}</h2>${chip(t.where === 'browser' ? 'runs in browser' : 'runs on server', t.where === 'browser' ? 'teal' : 'blue')}</header><div id="tool" data-tool="${t.id}">${toolView(t.id, s)}</div>`));
}

function settingsDetail() {
  const sel = st.sel.settings;
  if (sel === 'api') return article(`<h2 class="${h2} mb-6">API and CLI</h2><div class="mb-4">${codeBlock(well, esc(`anbu api setup ${location.origin} -H "X-Proxy-Token: …"\nanbu api secrets list\nanbu api secrets get <secret>\nanbu api secrets get <secret>:<profile>\nanbu api secrets totp <secret>\nanbu api ssh targets\nanbu api machines list <profile-ref>`))}</div><p class="text-sm">Base URL ${esc(location.origin)}</p>`);
  if (sel === 'password') return article(`<h2 class="${h2} mb-2">Vault password</h2><p class="text-sm mb-6">The whole vault is re-encrypted under a new key. The password lives in the data directory, so this is about rotation, not unlocking.</p><form data-password-form autocomplete="off"><div class="grid sm:grid-cols-2 gap-5 mb-8 max-w-3xl"><label class="flex flex-col gap-1.5 ${lbl}">new password<input name="password" type="password" autocomplete="new-password" class="${field}"></label><label class="flex flex-col gap-1.5 ${lbl}">repeat<input name="repeat" type="password" autocomplete="new-password" class="${field}"></label></div><button class="${primary}"><i data-lucide="save" class="size-4"></i>Rotate password</button></form>`);
  if (sel === 'backup') {
    const r = st.importResult;
    return article(`<h2 class="${h2} mb-6">Export and import</h2><div class="grid lg:grid-cols-2 gap-4"><section class="bg-surface0 rounded-2xl p-5"><p class="text-text mb-1">Export</p><p class="text-sm mb-4">Plaintext JSON of every secret, named anbu-vault-${ymd(new Date()).replace(/-/g, '')}.json.</p><a href="/api/vault/export" download class="${ghost} !bg-surface1"><i data-lucide="download" class="size-4"></i>Download export</a></section><section class="bg-surface0 rounded-2xl p-5"><p class="text-text mb-1">Import</p><p class="text-sm mb-4">Secrets whose ID or name exists are skipped.</p><label class="${ghost} !bg-surface1 cursor-pointer"><i data-lucide="upload" class="size-4"></i>Choose file<input type="file" accept="application/json,.json" data-import class="hidden"></label>${r ? `<p class="mt-4 text-sm">${r.error ? `<span class="text-red">${esc(r.error)}</span>` : `<span class="text-green">added ${r.added.length}</span> · <span class="text-overlay1">skipped ${r.skipped.length}${r.skipped.length ? ` (${esc(r.skipped.join(', '))})` : ''}</span>`}</p>` : ''}</section></div>`);
  }
  return needD(get('settings'), (s) => {
    const num = (k, suf = '') => `<label class="flex flex-col gap-1.5 ${lbl}">${k}<span class="flex items-center gap-2">${numInput(`name="${k}" value="${esc(s[k])}"`, 'w-40', field)}${suf}</span></label>`;
    const views = {
      generators: `<h2 class="${h2} mb-6">Generators</h2><div class="grid sm:grid-cols-2 xl:grid-cols-3 gap-5 mb-8">${num('passphrase_words')}<div class="flex flex-col gap-1.5 ${lbl}">passphrase_simple${seg('passphrase_simple', [['false', 'Capital + digit'], ['true', 'Simple']], String(s.passphrase_simple))}</div>${num('random_length')}<div class="flex flex-col gap-1.5 ${lbl}">random_charset${dropdown({ attrs: 'name="random_charset"', options: charsets.map((c) => [c, c]), value: s.random_charset, cls: field })}</div><div class="flex flex-col gap-1.5 ${lbl}">uuid_version${seg('uuid_version', [['4', 'v4'], ['7', 'v7']], String(s.uuid_version))}</div></div>`,
      timeouts: `<h2 class="${h2} mb-6">Timeouts</h2><div class="grid sm:grid-cols-2 gap-5 mb-8">${num('aws_command_timeout_seconds', 'seconds')}</div>`,
      machines: `<h2 class="${h2} mb-6">Machines</h2><div class="grid sm:grid-cols-2 gap-5 mb-8"><label class="flex flex-col gap-1.5 ${lbl}">machine_timezone<input name="machine_timezone" value="${esc(s.machine_timezone)}" class="${field}" placeholder="empty uses the server's zone"></label></div>`,
    };
    return article(`<form data-settings-form autocomplete="off">${views[sel] ?? views.generators}<button class="${primary}"><i data-lucide="save" class="size-4"></i>Save</button></form>`);
  });
}

const details = {
  vault() {
    if (st.view.vault === 'new') return secretForm(null);
    if (!st.sel.vault) return needD(get('secrets'), (list) => (list.length ? '' : empty('No secrets yet', `<button type="button" data-act="new-secret" class="${primary}"><i data-lucide="plus" class="size-4"></i>New secret</button>`)));
    if (st.editing) return needD(get('secret', st.sel.vault), (s) => secretForm(s));
    return secretDetail(st.sel.vault);
  },
  tasks: tasksDetail,
  ssh: sshDetail,
  aws: awsDetail,
  machines: machinesDetail,
  tools: toolDetail,
  settings: settingsDetail,
};

function ensureDefaults() {
  const ps = peek('secrets') ? profiles() : null;
  if (ps?.length && !ps.some((p) => p.ref === st.profile)) st.profile = ps[0].ref;
  const m = st.mod;
  if (!(m in st.view) || st.view[m] || st.sel[m]) return;
  if (m === 'vault') st.sel.vault = peek('secrets')?.[0]?.name ?? '';
  if (m === 'ssh') st.sel.ssh = peek('targets')?.targets?.[0]?.ref ?? '';
  if (m === 'aws' && ps) { if (ps.length) st.sel.aws = ps[0].ref; else st.view.aws = 'adhoc'; }
  if (m === 'machines' && ps?.length) {
    const list = peek('machines', st.profile);
    if (list) { if (list.length) st.sel.machines = list[0].name; else st.view.machines = 'scaffold'; }
  }
}

const last = { rail: '', list: '', detail: '', key: '', mod: '' };
let cleanup = [];
let rendering = false;

function render() {
  if (rendering) return;
  rendering = true;
  try { draw(); } finally { rendering = false; }
}

function draw() {
  watchSSO();
  ensureDefaults();
  renderRail();
  const list = document.getElementById('list');
  const detail = document.getElementById('detail');
  const listHTML = lists[st.mod]();
  if (listHTML !== last.list) {
    const top = st.mod === last.mod ? list.querySelector('[data-scroll]')?.scrollTop ?? 0 : 0;
    list.innerHTML = listHTML;
    const sc = list.querySelector('[data-scroll]');
    if (sc) sc.scrollTop = top;
    last.list = listHTML;
  }
  const key = `${st.mod}|${st.profile}|${st.view[st.mod] ?? ''}|${st.sel[st.mod]}`;
  const detailHTML = details[st.mod]();
  if (detailHTML !== last.detail || key !== last.key) {
    for (const c of cleanup) c();
    cleanup = [];
    const top = key === last.key ? detail.scrollTop : 0;
    detail.innerHTML = detailHTML;
    detail.scrollTop = top;
    last.detail = detailHTML;
    afterDetail(detail);
  }
  last.key = key;
  last.mod = st.mod;
  const r = route();
  if (location.pathname + location.search !== r) history[st.push ? 'pushState' : 'replaceState'](null, '', r);
  st.push = false;
  list.classList.toggle('max-md:hidden', st.detail);
  detail.classList.toggle('max-md:hidden', !st.detail);
  tick();
  lucide.createIcons();
}

let liveTerm = null;
function closeTerm() {
  liveTerm?.close?.();
  liveTerm = null;
}

let totpCode = '';
function afterDetail(detail) {
  for (const el of detail.querySelectorAll('[data-keep]')) if (st.keep[el.dataset.keep] !== undefined) el.value = st.keep[el.dataset.keep];
  const term = detail.querySelector('#term');
  if (term && liveTerm?.ref === term.dataset.ref) {
    liveTerm.el.className = term.className;
    term.replaceWith(liveTerm.el);
  } else {
    closeTerm();
    if (term) {
      const t = { ref: term.dataset.ref, el: term, close: null };
      liveTerm = t;
      requestAnimationFrame(() => { if (term.isConnected && liveTerm === t) t.close = openTerminal(term, JSON.parse(term.dataset.target)); });
    }
  }
  const tool = detail.querySelector('#tool');
  if (tool) wireTool(tool, tool.dataset.tool);
  const code = detail.querySelector('[data-totp]');
  if (code) {
    totpCode = '';
    cleanup.push(watchTOTP(code.dataset.totp, ({ code: c, remaining, period }) => {
      totpCode = c;
      code.textContent = groupCode(c);
      code.classList.replace('text-overlay0', 'text-text');
      const ring = detail.querySelector('[data-ring]');
      ring.setAttribute('stroke-dashoffset', String(94.25 * (1 - remaining / period)));
      ring.setAttribute('stroke', remaining <= 5 ? 'var(--ctp-red)' : 'var(--ctp-mauve)');
      detail.querySelector('[data-remain]').textContent = `${remaining}s`;
    }));
  }
  paintRun();
  paintProbe();
  paintPrio();
  tick();
}

let ticks = 0;
function tick() {
  for (const el of document.querySelectorAll('[data-rel]')) {
    const r = relTime(el.dataset.rel);
    el.textContent = el.dataset.fmt === 'left' ? `${r.replace('in ', '')} left` : r;
  }
  for (const el of document.querySelectorAll('[data-until]')) {
    const s = Math.max(0, Math.round((new Date(el.dataset.until) - Date.now()) / 1000));
    el.textContent = `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
  }
}
setInterval(() => {
  tick();
  if (++ticks % 3) return;
  for (const [k, e] of cache) {
    if (!k.startsWith('sso:') || e.data?.state !== 'pending' || e.loading) continue;
    e.stale = true;
    load(k, () => loaders.sso(k.slice(4)));
  }
}, 1000);

function ssoChanged(secret) {
  for (const [k, e] of cache) if (/^(whoami|machines|options|scaffold):/.test(k) && k.split(':')[1] === secret) e.stale = true;
  invalidate('targets');
}

const prevSSO = new Map();
function watchSSO() {
  for (const [k, e] of cache) {
    if (!k.startsWith('sso:') || !e.data) continue;
    const was = prevSSO.get(k);
    if (was === 'pending' && e.data.state === 'active') { ssoChanged(k.slice(4)); toast(`SSO session ${k.slice(4)} is active`, 'green'); }
    prevSSO.set(k, e.data.state);
  }
}

function resetTransient() { Object.assign(st, transient()); }

function nav(fn) {
  resetTransient();
  fn();
  st.push = true;
  render();
}

function navHref(href) {
  history.pushState(null, '', href);
  resetTransient();
  applyRoute();
  render();
}

async function act(fn, ok) {
  try {
    const r = await fn();
    if (ok) toast(typeof ok === 'function' ? ok(r) : ok, 'green');
    return r;
  } catch (err) {
    toast(err.message);
    return undefined;
  }
}

async function enqueue(method, path, body) {
  const job = await act(() => api(method, path, body), (j) => `Queued: ${j.label}`);
  if (job) refreshJobs();
  return job;
}

const mpath = (name, suffix = '') => `/api/aws/machines/${enc(name)}${suffix}?${profileQ(st.profile)}`;

async function onAct(a, b) {
  switch (a) {
    case 'back': return nav(() => { st.detail = false; });
    case 'new-secret': return nav(() => { st.mod = 'vault'; st.view.vault = 'new'; st.detail = true; });
    case 'new-aws-secret': return nav(() => { st.mod = 'vault'; st.newType = 'aws-sso'; st.view.vault = 'new'; st.detail = true; });
    case 'new-ssh-key': return nav(() => { st.mod = 'vault'; st.newType = 'ssh-key'; st.view.vault = 'new'; st.detail = true; });
    case 'adopt-key': {
      const key = document.querySelector('[data-keep="adoptKey"]')?.value ?? '';
      const r = await act(() => api('PUT', `/api/aws/scaffold/key?${profileQ(st.profile)}`, { private_key: key }), 'Scaffold key imported');
      if (r) {
        st.keep.adoptKey = '';
        invalidate('scaffold');
      }
      return render();
    }
    case 'new-host': return nav(() => { st.mod = 'ssh'; st.view.ssh = 'new'; st.detail = true; });
    case 'new-machine': return nav(() => { st.view.machines = 'new'; st.detail = true; });
    case 'cancel': st.confirm = null; st.editing = false; st.modify = null; st.remove = false; return render();
    case 'edit-secret': case 'edit-host': st.editing = true; return render();
    case 'delete-secret': st.confirm = { kind: 'delete-secret' }; return render();
    case 'confirm-delete-secret': {
      const name = st.sel.vault;
      try {
        await api('DELETE', `/api/secrets/${enc(name)}`);
        invalidate('secrets');
        cache.delete(`secret:${name}`);
        toast(`Deleted ${name}`, 'green');
        return nav(() => { st.sel.vault = ''; st.detail = false; });
      } catch (err) {
        st.confirm = { kind: 'delete-secret', error: err };
        return render();
      }
    }
    case 'copy-totp': if (totpCode) copy(totpCode); return undefined;
    case 'add-profile': b.closest('section').querySelector('[data-prows]').insertAdjacentHTML('beforeend', profileRow()); return lucide.createIcons();
    case 'add-custom': b.parentElement.querySelector('[data-crows]').insertAdjacentHTML('beforeend', customRow()); return lucide.createIcons();
    case 'del-row': return b.closest('[data-prow], [data-crow]').remove();
    case 'toggle-hidden': {
      const hidden = b.dataset.hidden !== '1';
      b.dataset.hidden = hidden ? '1' : '0';
      b.className = `shrink-0 rounded-lg px-3 ${hidden ? 'bg-surface1 text-text' : 'bg-surface0 text-overlay1'} hover:text-text`;
      b.innerHTML = `<i data-lucide="${hidden ? 'eye-off' : 'eye'}" class="size-4"></i>`;
      b.parentElement.querySelector('[data-c="value"]').type = hidden ? 'password' : 'text';
      return lucide.createIcons();
    }
    case 'gen-ssh-key': {
      const name = document.querySelector('[data-secret-form] [name="name"]').value.trim();
      if (!name) return toast('Name the key first');
      const s = await act(() => api('POST', '/api/secrets/generate-ssh-key', { name }), `Generated ${name}`);
      if (!s) return undefined;
      invalidate('secrets');
      return nav(() => { st.view.vault = ''; st.sel.vault = s.name; st.detail = true; });
    }
    case 'fill-profiles': {
      const s = peek('secret', st.sel.vault);
      const accounts = await act(() => api('GET', `/api/aws/sso/${enc(s.name)}/accounts`));
      if (!accounts) return undefined;
      const rows = b.closest('section').querySelector('[data-prows]');
      const have = new Set([...rows.querySelectorAll('[data-prow]')].map((r) => `${r.querySelector('[data-p="account_id"]').value}/${r.querySelector('[data-p="role_name"]').value}`));
      for (const r of [...rows.querySelectorAll('[data-prow]')]) if (![...r.querySelectorAll('input')].some((i) => i.value)) r.remove();
      const region = document.querySelector('[data-f="sso_region"]')?.value ?? '';
      let added = 0;
      for (const acc of accounts) {
        for (const role of acc.roles) {
          if (have.has(`${acc.account_id}/${role}`)) continue;
          const name = `${acc.account_name || acc.account_id}-${role}`.toLowerCase().replace(/[^a-z0-9._-]+/g, '-');
          rows.insertAdjacentHTML('beforeend', profileRow({ name, account_id: acc.account_id, role_name: role, region }));
          added++;
        }
      }
      toast(`Added ${added} profile${added === 1 ? '' : 's'}`, 'green');
      return lucide.createIcons();
    }
    case 'forget-key': {
      const r = await act(() => api('DELETE', `/api/ssh/known-hosts/${enc(b.dataset.alias)}`), (x) => `Removed ${x.removed} host key line${x.removed === 1 ? '' : 's'}`);
      return r;
    }
    case 'delete-host': st.confirm = { kind: 'delete-host' }; return render();
    case 'confirm-delete-host': {
      const t = peek('targets')?.targets.find((x) => x.ref === st.sel.ssh);
      if (!t) return undefined;
      if (await act(() => api('DELETE', `/api/hosts/${enc(t.id)}`), `Deleted ${t.name}`) === undefined) return undefined;
      invalidate('targets', 'secret');
      return nav(() => { st.sel.ssh = ''; st.detail = false; });
    }
    case 'probe': {
      const name = st.sel.machines;
      st.probe = { name, loading: true };
      paintProbe();
      try { st.probe = { name, report: (await api('GET', mpath(name, '/status'))).report }; } catch (err) { st.probe = { name, error: err }; }
      return paintProbe();
    }
    case 'machine-start': return enqueue('POST', mpath(st.sel.machines, '/start'), {});
    case 'machine-stop': return enqueue('POST', mpath(st.sel.machines, '/stop'), {});
    case 'modify': {
      const mm = peek('machines', st.profile)?.find((x) => x.name === st.sel.machines);
      const shape = mm ? `${mm.vcpu}vcpu-${mm.memory}gb` : st.mShape;
      st.remove = false;
      st.modify = { shape, class: mm?.class ?? 'dedicated' };
      return render();
    }
    case 'queue-modify': {
      if (await enqueue('POST', mpath(st.sel.machines, '/modify'), { shape: st.modify.shape, class: st.modify.class })) { st.modify = null; render(); }
      return undefined;
    }
    case 'remove': st.modify = null; st.remove = true; return render();
    case 'queue-remove': {
      const name = st.sel.machines;
      const confirm = (st.keep.removeConfirm ?? '').trim();
      if (confirm !== name) return toast(`Type ${name} to confirm`);
      if (await enqueue('DELETE', mpath(name), { confirm })) { st.remove = false; delete st.keep.removeConfirm; render(); }
      return undefined;
    }
    case 'queue-create': {
      const name = (st.keep.mName ?? '').trim();
      if (!name) return toast('Name the machine first');
      const job = await enqueue('POST', `/api/aws/machines?${profileQ(st.profile)}`, { name, arch: st.mArch, shape: st.mShape, class: st.mClass, disk: st.mDisk });
      if (job) nav(() => { st.view.machines = 'jobs'; });
      return undefined;
    }
    case 'setup-scaffold': return enqueue('POST', `/api/aws/scaffold?${profileQ(st.profile)}`, {});
    case 'teardown': st.confirm = { kind: 'teardown' }; return render();
    case 'confirm-teardown': {
      st.confirm = null;
      await enqueue('DELETE', `/api/aws/scaffold?${profileQ(st.profile)}`, {});
      return render();
    }
    default: return undefined;
  }
}

document.addEventListener('click', async (e) => {
  const $ = (s) => e.target.closest(s);
  let b;
  if ((b = $('[data-mod]'))) return nav(() => { st.mod = b.dataset.mod; st.detail = false; });
  if ((b = $('[data-sel]'))) return nav(() => { st.sel[st.mod] = b.dataset.sel; if (st.mod in st.view) st.view[st.mod] = ''; st.detail = true; });
  if ((b = $('[data-view]'))) return nav(() => { st.view[st.mod] = b.dataset.view; st.detail = true; });
  if ((b = $('[data-href]'))) return navHref(b.dataset.href);
  if ((b = $('[data-pane]'))) return togglePane(b.dataset.pane);
  if ((b = $('[data-retry]'))) { invalidate(b.dataset.retry); return render(); }
  if ((b = $('[data-sso-login]'))) {
    const secret = b.dataset.ssoLogin;
    const p = await act(() => api('POST', `/api/aws/sso/${enc(secret)}/login`));
    if (!p) return undefined;
    const e2 = cache.get(`sso:${secret}`) ?? { key: `sso:${secret}` };
    e2.data = { state: 'pending', verification_uri: p.verification_uri, user_code: p.user_code, expires_at: new Date(Date.now() + p.expires_in * 1000).toISOString() };
    e2.error = null;
    cache.set(`sso:${secret}`, e2);
    prevSSO.set(`sso:${secret}`, 'pending');
    return render();
  }
  if ((b = $('[data-act]'))) return onAct(b.dataset.act, b);
  if ((b = $('[data-seg][data-local] button'))) {
    const name = b.parentElement.dataset.seg;
    const v = b.dataset.v;
    if (name === 'sshtab') st.sshTab = v;
    if (name === 'arch' || name === 'class') {
      if (name === 'arch') st.mArch = v; else st.mClass = v;
      const o = peek('options', st.profile);
      if (o && !o.types?.[st.mArch]?.[st.mShape]?.[st.mClass]) st.mShape = o.shapes.find((sh) => o.types?.[st.mArch]?.[sh]?.[st.mClass]) ?? st.mShape;
    }
    if (name === 'modclass') st.modify.class = v;
    return render();
  }
  if ((b = $('[data-seg]:not([data-local]) button')) && !b.closest('#tool')) return segToggle(b);
  if ((b = $('[data-shape]'))) { st.mShape = b.dataset.shape; return render(); }
  if ((b = $('[data-modshape]'))) { st.modify.shape = b.dataset.modshape; return render(); }
  if ((b = $('[data-disk]'))) { st.mDisk = +b.dataset.disk; return render(); }
  if ((b = $('[data-newtype]'))) { st.newType = b.dataset.newtype; return render(); }
  if ((b = $('[data-prio]'))) { st.newPrio = b.dataset.prio; return paintPrio(); }
  if ((b = $('[data-toggle]'))) {
    const li = b.closest('li');
    if (li.dataset.leaving) return undefined;
    const t = peek('tasks').find((x) => x.id === b.dataset.toggle);
    const done = !t.done;
    b.className = `grid place-items-center size-5 shrink-0 rounded-full ${done ? 'bg-green text-crust' : 'bg-surface1'}`;
    b.innerHTML = done ? '<i data-lucide="check" class="size-3.5"></i>' : '';
    lucide.createIcons();
    if (!calm.matches) b.animate([{ transform: 'scale(0.5)' }, { transform: 'scale(1.25)' }, { transform: 'scale(1)' }], { duration: 280, easing: 'ease-out' });
    const text = li.querySelector('span.flex-1');
    text.classList.toggle('line-through', done);
    text.classList.toggle('text-overlay0', done);
    text.classList.toggle('text-text', !done);
    const req = api('PUT', `/api/tasks/${enc(t.id)}`, { text: t.text, priority: t.priority, due: t.due ?? '', done });
    const gone = new Promise((r) => leave(li, 380, r));
    return finishTask(req, gone);
  }
  if ((b = $('[data-del-task]'))) {
    const req = api('DELETE', `/api/tasks/${enc(b.dataset.delTask)}`);
    const gone = new Promise((r) => leave(b.closest('li'), 0, r));
    return finishTask(req, gone);
  }
  if ((b = $('[data-quickaws]'))) {
    const form = b.closest('section').querySelector('[data-run]');
    form.elements.cmd.value = b.dataset.quickaws;
    return form.requestSubmit();
  }
  if ((b = $('[data-openjob]'))) { st.openJob = st.openJob === b.dataset.openjob ? null : b.dataset.openjob; return render(); }
  if ((b = $('[data-reveal]'))) {
    const target = b.closest('.group').querySelector('[data-secret]');
    const show = target.dataset.shown !== '1';
    target.textContent = show ? target.dataset.secret : '•'.repeat(target.tagName === 'PRE' ? 24 : 16);
    target.dataset.shown = show ? '1' : '0';
  }
  return undefined;
});

async function finishTask(req, gone) {
  try {
    await Promise.all([req, gone]);
  } catch (err) {
    toast(err.message);
    last.detail = '';
  }
  invalidate('tasks');
  render();
}

document.addEventListener('submit', async (e) => {
  e.preventDefault();
  const f = e.target;
  if (f.matches('[data-addtask]')) {
    const text = f.elements.text.value.trim();
    if (!text) return;
    const due = f.elements.due.value;
    if (await act(() => api('POST', '/api/tasks', { text, priority: st.newPrio, ...(due ? { due } : {}) })) === undefined) return;
    f.elements.text.value = '';
    setDate(f.querySelector('[data-dp]'), '');
    invalidate('tasks');
    render();
    return;
  }
  if (f.matches('[data-run]')) {
    const command = f.elements.cmd.value.trim().replace(/^aws\s+/, '');
    if (!command) return;
    const key = f.dataset.run;
    let source = { profile: key };
    if (key === 'adhoc') {
      const k = document.querySelector('[data-adhoc]').elements;
      const inline = { access_key_id: k.access_key_id.value.trim(), secret_access_key: k.secret_access_key.value.trim(), session_token: k.session_token.value.trim(), region: k.region.value.trim() };
      if (!inline.access_key_id || !inline.secret_access_key || !inline.region) { toast('Ad-hoc keys need access_key_id, secret_access_key, and region'); return; }
      source = { inline };
    }
    runAws(key, command, source);
    return;
  }
  if (f.matches('[data-secret-form]')) {
    const orig = st.editing ? peek('secret', st.sel.vault) : null;
    const body = collectSecret(f, orig);
    const s = await act(() => (orig ? api('PUT', `/api/secrets/${enc(orig.name)}`, body) : api('POST', '/api/secrets', body)), `Saved ${body.name}`);
    if (!s) return;
    invalidate('secrets', 'targets');
    if (orig) cache.delete(`secret:${orig.name}`);
    cache.delete(`secret:${s.name}`);
    nav(() => { st.view.vault = ''; st.sel.vault = s.name; st.detail = true; });
    return;
  }
  if (f.matches('[data-host-form]')) {
    const el = f.elements;
    const port = el.port.value.trim();
    const body = { name: el.name.value.trim(), address: el.address.value.trim(), port: port ? Number(port) : 22, user: el.user.value.trim(), key_secret: el.key_secret?.value ?? '' };
    const t = st.editing ? peek('targets')?.targets.find((x) => x.ref === st.sel.ssh) : null;
    const h = await act(() => (t ? api('PUT', `/api/hosts/${enc(t.id)}`, body) : api('POST', '/api/hosts', body)), `Saved ${body.name}`);
    if (!h) return;
    invalidate('targets', 'secret');
    nav(() => { st.view.ssh = ''; st.sel.ssh = h.name; st.detail = true; });
    return;
  }
  if (f.matches('[data-settings-form]')) {
    const cur = peek('settings');
    const next = { ...cur };
    for (const el of f.querySelectorAll('input[name]')) next[el.name] = el.type === 'number' ? Number(el.value) : el.value;
    for (const sg of f.querySelectorAll('[data-seg]')) {
      const v = segVal(f, sg.dataset.seg);
      next[sg.dataset.seg] = v === 'true' || v === 'false' ? v === 'true' : Number(v);
    }
    const saved = await act(() => api('PUT', '/api/settings', next), 'Settings saved');
    if (!saved) return;
    cache.set('settings', { key: 'settings', data: saved });
    render();
    return;
  }
  if (f.matches('[data-password-form]')) {
    const { password, repeat } = f.elements;
    if (!password.value) { toast('Enter a new password'); return; }
    if (password.value !== repeat.value) { toast('The passwords do not match'); return; }
    if (await act(() => api('PUT', '/api/settings/password', { password: password.value }), 'Vault password rotated') === undefined) return;
    f.reset();
  }
});

document.addEventListener('input', (e) => {
  const t = e.target;
  if (t.matches('[data-q]')) {
    st.q = t.value;
    const pos = t.selectionStart;
    render();
    const q = document.querySelector('[data-q]');
    q.focus();
    q.setSelectionRange(pos, pos);
  }
  if (t.dataset.keep) st.keep[t.dataset.keep] = t.value;
});

document.addEventListener('change', async (e) => {
  const t = e.target;
  if (t.matches('[data-profile]')) {
    st.profile = t.value;
    try { localStorage.setItem('anbu.profile', st.profile); } catch { }
    return nav(() => { if (!st.view.machines) st.sel.machines = ''; });
  }
  if (t.matches('[data-import]')) {
    const file = t.files[0];
    if (!file) return undefined;
    try {
      const doc = JSON.parse(await file.text());
      st.importResult = await api('POST', '/api/vault/import', doc);
      invalidate('secrets', 'secret');
    } catch (err) {
      st.importResult = { error: err.message };
    }
    t.value = '';
    return render();
  }
  return undefined;
});

document.addEventListener('scroll', (e) => {
  const el = e.target === document ? document.documentElement : e.target;
  el.classList.add('scrolling');
  clearTimeout(el.scrollIdle);
  el.scrollIdle = setTimeout(() => el.classList.remove('scrolling'), 1500);
}, true);

window.addEventListener('popstate', () => { resetTransient(); applyRoute(); render(); });
wide.addEventListener('change', render);

applyRoute();
installControls();
applyLayout();
render();
refreshJobs();
