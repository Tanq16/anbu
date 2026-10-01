function theme() {
  const css = getComputedStyle(document.documentElement);
  const c = (name) => css.getPropertyValue(`--ctp-${name}`).trim();
  return {
    background: c('base'), foreground: c('text'), cursor: c('rosewater'), selectionBackground: `${c('surface2')}80`,
    black: c('surface1'), red: c('red'), green: c('green'), yellow: c('yellow'), blue: c('blue'), magenta: c('pink'), cyan: c('teal'), white: c('subtext1'),
    brightBlack: c('surface2'), brightRed: c('red'), brightGreen: c('green'), brightYellow: c('yellow'), brightBlue: c('blue'), brightMagenta: c('pink'), brightCyan: c('teal'), brightWhite: c('subtext0'),
  };
}

export function openTerminal(el, target) {
  const term = new Terminal({ fontFamily: "'JetBrains Mono', monospace", fontSize: 14, cursorBlink: false, theme: theme() });
  const fit = new FitAddon.FitAddon();
  term.loadAddon(fit);
  term.open(el);
  try { term.loadAddon(new WebglAddon.WebglAddon()); } catch { }
  fit.fit();
  document.fonts.ready.then(() => requestAnimationFrame(() => fit.fit()));

  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  const ws = new WebSocket(`${proto}://${location.host}/ws/terminal`);
  ws.binaryType = 'arraybuffer';
  const send = (m) => ws.readyState === WebSocket.OPEN && ws.send(JSON.stringify(m));

  ws.onopen = () => send({ t: 'start', target, cols: term.cols, rows: term.rows });
  ws.onmessage = (ev) => {
    if (ev.data instanceof ArrayBuffer) return term.write(new Uint8Array(ev.data));
    const msg = JSON.parse(ev.data);
    if (msg.t === 'exited') term.write(`\r\n[process exited with code ${msg.code}]\r\n`);
    if (msg.t === 'error') term.write(`\r\n[error: ${msg.msg}]\r\n`);
  };
  ws.onclose = () => term.write('\r\n[connection closed]\r\n');

  term.onData((data) => send({ t: 'input', data }));
  term.onResize(({ cols, rows }) => send({ t: 'resize', cols, rows }));
  const ro = new ResizeObserver(() => fit.fit());
  ro.observe(el);
  term.focus();

  return () => { ro.disconnect(); ws.onclose = null; ws.close(); term.dispose(); };
}
