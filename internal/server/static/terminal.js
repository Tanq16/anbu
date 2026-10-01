const theme = {
  background: '#1e1e2e', foreground: '#cdd6f4', cursor: '#f5e0dc', selectionBackground: '#585b7080',
  black: '#45475a', red: '#f38ba8', green: '#a6e3a1', yellow: '#f9e2af', blue: '#89b4fa', magenta: '#f5c2e7', cyan: '#94e2d5', white: '#bac2de',
  brightBlack: '#585b70', brightRed: '#f38ba8', brightGreen: '#a6e3a1', brightYellow: '#f9e2af', brightBlue: '#89b4fa', brightMagenta: '#f5c2e7', brightCyan: '#94e2d5', brightWhite: '#a6adc8',
};

export function openTerminal(el, target) {
  const term = new Terminal({ fontFamily: "'JetBrains Mono', monospace", fontSize: 14, cursorBlink: false, theme });
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
