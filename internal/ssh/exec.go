package ssh

import (
	"os"
	"os/exec"
	"strconv"
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
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
