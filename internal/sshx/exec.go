package sshx

import (
	"bytes"
	"context"
	"errors"

	"golang.org/x/crypto/ssh"
)

type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

func Exec(ctx context.Context, client *ssh.Client, command string) (ExecResult, error) {
	sess, err := client.NewSession()
	if err != nil {
		return ExecResult{}, err
	}
	defer sess.Close()
	var stdout, stderr bytes.Buffer
	sess.Stdout, sess.Stderr = &stdout, &stderr
	stop := context.AfterFunc(ctx, func() { sess.Signal(ssh.SIGKILL); sess.Close() })
	defer stop()
	err = sess.Run(command)
	res := ExecResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if exitErr, ok := errors.AsType[*ssh.ExitError](err); ok {
		res.ExitCode = exitErr.ExitStatus()
		return res, nil
	}
	if err == nil {
		return res, nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return res, ctxErr
	}
	return res, err
}
