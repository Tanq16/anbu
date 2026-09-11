package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type ExecConfig struct {
	Name    string
	Command string
}

func Exec(cfg ExecConfig) error {
	bin, err := exec.LookPath("ssh")
	if err != nil {
		return err
	}
	sess, err := get(cfg.Name)
	if err != nil {
		return err
	}
	priv := identityFile(sess.Identity)
	args := []string{
		"-F", "/dev/null",
		"-i", priv,
		"-o", "IdentitiesOnly=yes",
		"-o", "UserKnownHostsFile=" + knownHostsPath(),
		"-o", "StrictHostKeyChecking=accept-new",
		"-p", strconv.Itoa(sess.Port),
		sess.User + "@" + sess.Host,
	}
	if cfg.Command != "" {
		args = append(args, cfg.Command)
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if cfg.Command == "" {
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if detail := strings.TrimSpace(stderr.String()); detail != "" {
			return fmt.Errorf("%s: %w", detail, err)
		}
		return err
	}
	return nil
}
