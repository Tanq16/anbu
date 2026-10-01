package server

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/rs/zerolog/log"
	"github.com/tanq16/anbu/internal/sshx"
	"golang.org/x/crypto/ssh"
)

type startMsg struct {
	T      string         `json:"t"`
	Target sshx.TargetRef `json:"target"`
	Cols   int            `json:"cols"`
	Rows   int            `json:"rows"`
}

type inMsg struct {
	T    string `json:"t"`
	Data string `json:"data"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

func readWSJSON(ctx context.Context, c *websocket.Conn, v any) error {
	typ, data, err := c.Read(ctx)
	if err != nil {
		return err
	}
	if typ != websocket.MessageText {
		return errors.New("expected a text frame")
	}
	return json.Unmarshal(data, v)
}

func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

func (s *Server) handleTerminal(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.CloseNow()
	c.SetReadLimit(1 << 20)
	ctx := r.Context()
	sendJSON := func(v any) {
		b, _ := json.Marshal(v)
		c.Write(ctx, websocket.MessageText, b)
	}
	fail := func(err error) {
		log.Debug().Err(err).Msg("terminal failed")
		sendJSON(map[string]string{"t": "error", "msg": errMessage(err)})
	}

	var start startMsg
	startCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	err = readWSJSON(startCtx, c, &start)
	cancel()
	if err != nil || start.T != "start" {
		return
	}
	client, err := s.dialTarget(ctx, start.Target)
	if err != nil {
		fail(err)
		return
	}
	defer client.Close()
	sess, err := client.NewSession()
	if err != nil {
		fail(err)
		return
	}
	defer sess.Close()
	modes := ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}
	if err := sess.RequestPty("xterm-256color", clamp(start.Rows, 5, 200), clamp(start.Cols, 20, 500), modes); err != nil {
		fail(err)
		return
	}
	stdin, err := sess.StdinPipe()
	if err != nil {
		fail(err)
		return
	}
	stdout, err := sess.StdoutPipe()
	if err != nil {
		fail(err)
		return
	}
	if err := sess.Shell(); err != nil {
		fail(err)
		return
	}
	log.Info().Str("target", start.Target.String()).Msg("terminal opened")
	defer log.Info().Str("target", start.Target.String()).Msg("terminal closed")
	sendJSON(map[string]string{"t": "started"})

	drained := make(chan struct{})
	go func() {
		defer close(drained)
		buf := make([]byte, 32<<10)
		for {
			n, err := stdout.Read(buf)
			if n > 0 && c.Write(ctx, websocket.MessageBinary, buf[:n]) != nil {
				return
			}
			if err != nil {
				return
			}
		}
	}()
	go func() {
		code := 0
		if exitErr, ok := errors.AsType[*ssh.ExitError](sess.Wait()); ok {
			code = exitErr.ExitStatus()
		}
		<-drained
		sendJSON(map[string]any{"t": "exited", "code": code})
		c.Close(websocket.StatusNormalClosure, "")
	}()
	go func() {
		for range time.Tick(25 * time.Second) {
			pctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.Ping(pctx)
			cancel()
			if err != nil {
				c.CloseNow()
				return
			}
		}
	}()

	for {
		var msg inMsg
		if err := readWSJSON(ctx, c, &msg); err != nil {
			return
		}
		switch msg.T {
		case "input":
			io.WriteString(stdin, msg.Data)
		case "resize":
			sess.WindowChange(clamp(msg.Rows, 5, 200), clamp(msg.Cols, 20, 500))
		}
	}
}
