const field = 'w-full bg-surface0 rounded-lg px-3 py-2 text-text text-sm font-mono placeholder:text-overlay0 outline-none focus:ring-2 focus:ring-mauve pointer-coarse:text-[16px]';
const pill = 'rounded-full px-3 py-1.5 text-sm';
const lbl = 'text-xs text-overlay1';

export const esc = (s) => String(s ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

const pad = (n) => String(n).padStart(2, '0');
export const ymd = (dt) => `${dt.getFullYear()}-${pad(dt.getMonth() + 1)}-${pad(dt.getDate())}`;

export async function copy(text) {
  try { await navigator.clipboard.writeText(text); } catch { }
}

export const seg = (name, opts, active, attrs = '') => `<span class="inline-flex w-fit bg-surface0 rounded-full p-1" data-seg="${name}" ${attrs}>${opts.map(([v, l]) => `<button type="button" data-v="${esc(v)}" class="${pill} ${String(v) === String(active) ? 'bg-surface1 text-text' : 'text-overlay1 hover:text-text'}">${l}</button>`).join('')}</span>`;

export function segToggle(b) {
  for (const s of b.parentElement.children) s.className = `${pill} ${s === b ? 'bg-surface1 text-text' : 'text-overlay1 hover:text-text'}`;
}

export const segVal = (root, name) => root.querySelector(`[data-seg="${name}"] .bg-surface1`)?.dataset.v;

export const copyBtn = (v) => `<button type="button" data-copy="${esc(v)}" class="shrink-0 rounded-full p-1.5 text-overlay1 hover:text-text hover:bg-surface1" title="Copy"><i data-lucide="copy" class="size-4"></i></button>`;

export const codeBlock = (preClass, inner = '', attrs = '') => `<div class="group/code relative min-w-0"><pre ${attrs} class="${preClass}">${inner}</pre><button type="button" data-copy-pre title="Copy" class="absolute top-2 right-3 rounded-full bg-surface0 p-1.5 text-overlay1 hover:text-text hover:bg-surface1 opacity-0 group-hover/code:opacity-100 focus:opacity-100 max-md:opacity-100"><i data-lucide="copy" class="size-4"></i></button></div>`;

export const numInput = (attrs, wrap = '', input = field) => `<span class="relative inline-flex ${wrap}"><input type="number" ${attrs} class="${input} w-full pr-9"><span class="absolute inset-y-1 right-1 flex flex-col"><button type="button" tabindex="-1" data-step="1" class="flex-1 grid place-items-center w-6 rounded-md text-overlay1 hover:text-text hover:bg-surface1"><i data-lucide="chevron-up" class="size-3.5"></i></button><button type="button" tabindex="-1" data-step="-1" class="flex-1 grid place-items-center w-6 rounded-md text-overlay1 hover:text-text hover:bg-surface1"><i data-lucide="chevron-down" class="size-3.5"></i></button></span></span>`;

export function dropdown({ attrs = '', options, value, cls = field, lead = '', align = 'left', placeholder = 'None' }) {
  const cur = options.find(([v]) => v === value) ?? options[0] ?? ['', placeholder];
  const opt = ([v, l, sub]) => `<li><button type="button" data-dd-opt="${esc(v)}" data-label="${esc(l)}" class="w-full flex items-center gap-3 rounded-lg px-3 py-2 text-left text-sm whitespace-nowrap outline-none hover:bg-surface1 focus:bg-surface1 ${v === cur[0] ? 'text-mauve' : 'text-text'}"><span class="flex-1">${esc(l)}</span>${sub ? `<span class="${lbl} font-mono">${esc(sub)}</span>` : ''}<i data-lucide="check" class="size-4 ${v === cur[0] ? '' : 'invisible'}"></i></button></li>`;
  return `<div class="relative" data-dd><button type="button" data-dd-btn class="${cls.replaceAll('focus:', 'focus-visible:')} flex items-center gap-2 text-left cursor-pointer hover:bg-surface1">${lead}<span data-dd-label class="flex-1 truncate">${esc(cur[1])}</span><i data-lucide="chevron-down" class="size-4 shrink-0 text-overlay1"></i></button><input type="hidden" ${attrs} value="${esc(cur[0])}"><ul data-dd-list class="hidden absolute z-30 ${align === 'right' ? 'right-0' : 'left-0'} mt-1.5 min-w-full max-h-72 overflow-y-auto rounded-xl bg-surface0 p-1 shadow-xl shadow-crust/60">${options.map(opt).join('')}</ul></div>`;
}

const fmtDate = (v) => new Date(`${v}T00:00`).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });

export function datePicker({ attrs = '', value = '', cls = field, placeholder = 'No date', align = 'left' }) {
  return `<div class="relative" data-dp><button type="button" data-dp-btn class="${cls.replaceAll('focus:', 'focus-visible:')} flex items-center gap-2 text-left cursor-pointer"><i data-lucide="calendar" class="size-4 shrink-0 text-overlay1"></i><span data-dp-label data-placeholder="${esc(placeholder)}" class="whitespace-nowrap ${value ? '' : 'text-overlay0'}">${value ? fmtDate(value) : esc(placeholder)}</span></button><input type="hidden" ${attrs} value="${esc(value)}"><div data-dp-pop class="hidden absolute z-30 ${align === 'right' ? 'right-0' : 'left-0'} mt-1.5 w-72 rounded-2xl bg-surface0 p-3 shadow-xl shadow-crust/60"></div></div>`;
}

export function setDate(dp, v) {
  const label = dp.querySelector('[data-dp-label]');
  label.textContent = v ? fmtDate(v) : label.dataset.placeholder;
  label.classList.toggle('text-overlay0', !v);
  dp.querySelector('input').value = v;
}

function calendar(pop, value, month) {
  const first = new Date(month.getFullYear(), month.getMonth(), 1);
  const today = ymd(new Date());
  const days = Array.from({ length: 42 }, (_, i) => new Date(first.getFullYear(), first.getMonth(), 1 - first.getDay() + i));
  const nav = (dir, icon) => `<button type="button" data-dp-nav="${dir}" class="grid place-items-center size-8 rounded-full text-overlay1 hover:text-text hover:bg-surface1"><i data-lucide="${icon}" class="size-4"></i></button>`;
  const day = (d) => {
    const v = ymd(d);
    const tone = v === value ? 'bg-mauve text-crust font-medium' : v === today ? 'text-mauve hover:bg-surface1' : d.getMonth() === first.getMonth() ? 'text-text hover:bg-surface1' : 'text-overlay0 hover:bg-surface1';
    return `<button type="button" data-dp-set="${v}" class="h-9 rounded-full text-sm ${tone}">${d.getDate()}</button>`;
  };
  pop.dataset.month = ymd(first);
  pop.innerHTML = `<div class="flex items-center gap-1 mb-2"><span class="flex-1 px-2 text-sm font-medium text-text">${first.toLocaleDateString(undefined, { month: 'long', year: 'numeric' })}</span>${nav(-1, 'chevron-left')}${nav(1, 'chevron-right')}</div>
    <div class="grid grid-cols-7 text-center text-[11px] text-overlay0 mb-1">${['S', 'M', 'T', 'W', 'T', 'F', 'S'].map((d) => `<span class="py-1">${d}</span>`).join('')}</div>
    <div class="grid grid-cols-7 gap-0.5">${days.map(day).join('')}</div>
    <div class="flex justify-between mt-2"><button type="button" data-dp-set="" class="rounded-full px-3 py-1 text-sm text-overlay1 hover:text-text hover:bg-surface1">Clear</button><button type="button" data-dp-set="${today}" class="rounded-full px-3 py-1 text-sm text-mauve hover:bg-surface1">Today</button></div>`;
  lucide.createIcons();
}

function closePopups(except) {
  for (const p of document.querySelectorAll('[data-dd-list]:not(.hidden), [data-dp-pop]:not(.hidden)')) if (p !== except) p.classList.add('hidden');
}

function commit(input, value) {
  input.value = value;
  input.dispatchEvent(new Event('input', { bubbles: true }));
  input.dispatchEvent(new Event('change', { bubbles: true }));
}

export function installControls() {
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closePopups();
    const opt = e.target.closest?.('[data-dd-opt]');
    if (opt && (e.key === 'ArrowDown' || e.key === 'ArrowUp')) {
      e.preventDefault();
      opt.parentElement[e.key === 'ArrowDown' ? 'nextElementSibling' : 'previousElementSibling']?.querySelector('button').focus();
    }
  });
  document.addEventListener('click', (e) => {
    const ddBtn = e.target.closest('[data-dd-btn]');
    if (ddBtn) {
      const list = ddBtn.parentElement.querySelector('[data-dd-list]');
      closePopups(list);
      list.classList.toggle('hidden');
      list.querySelector('.text-mauve')?.focus();
      return;
    }
    const opt = e.target.closest('[data-dd-opt]');
    if (opt) {
      const dd = opt.closest('[data-dd]');
      for (const o of dd.querySelectorAll('[data-dd-opt]')) {
        o.classList.toggle('text-mauve', o === opt);
        o.classList.toggle('text-text', o !== opt);
        o.querySelector('svg:last-child')?.classList.toggle('invisible', o !== opt);
      }
      dd.querySelector('[data-dd-label]').textContent = opt.dataset.label;
      closePopups();
      dd.querySelector('[data-dd-btn]').focus();
      commit(dd.querySelector('input'), opt.dataset.ddOpt);
      return;
    }
    const dpBtn = e.target.closest('[data-dp-btn]');
    if (dpBtn) {
      const pop = dpBtn.parentElement.querySelector('[data-dp-pop]');
      const v = dpBtn.parentElement.querySelector('input').value;
      closePopups(pop);
      if (pop.classList.toggle('hidden')) return;
      calendar(pop, v, v ? new Date(`${v}T00:00`) : new Date());
      return;
    }
    const nav = e.target.closest('[data-dp-nav]');
    if (nav) {
      const pop = nav.closest('[data-dp-pop]');
      const m = new Date(`${pop.dataset.month}T00:00`);
      m.setMonth(m.getMonth() + +nav.dataset.dpNav);
      calendar(pop, pop.parentElement.querySelector('input').value, m);
      return;
    }
    const set = e.target.closest('[data-dp-set]');
    if (set) {
      const dp = set.closest('[data-dp]');
      setDate(dp, set.dataset.dpSet);
      closePopups();
      commit(dp.querySelector('input'), set.dataset.dpSet);
      return;
    }
    if (!e.target.closest('[data-dd], [data-dp]')) closePopups();
    const step = e.target.closest('[data-step]');
    if (step) {
      const input = step.closest('.relative').querySelector('input');
      if (step.dataset.step === '1') input.stepUp(); else input.stepDown();
      input.dispatchEvent(new Event('input', { bubbles: true }));
      return;
    }
    const b = e.target.closest('[data-copy], [data-copy-pre]');
    if (!b) return;
    copy(b.dataset.copy ?? b.previousElementSibling.textContent);
    const old = b.innerHTML;
    b.innerHTML = '<i data-lucide="check" class="size-4 text-green"></i>';
    lucide.createIcons();
    setTimeout(() => { b.innerHTML = old; }, 900);
  });
}
