package sshCmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	anbuSSH "github.com/tanq16/anbu/internal/ssh"
	u "github.com/tanq16/anbu/utils"
)

var execFlags struct {
	command string
}

var execCmd = &cobra.Command{
	Use:   "exec <name>",
	Short: "Connect to a named SSH session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if execFlags.command == "" && !u.StdinIsTerminal {
			u.PrintFatal("exec needs a terminal or --command", nil)
		}
		err := anbuSSH.Exec(anbuSSH.ExecConfig{
			Name:    args[0],
			Command: execFlags.command,
		})
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				u.PrintFatal("ssh not found in PATH", err)
			}
			if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
				u.PrintError("remote command failed", err)
				os.Exit(exitErr.ExitCode())
			}
			u.PrintFatal("failed to exec session", err)
		}
	},
}

func init() {
	execCmd.Flags().StringVarP(&execFlags.command, "command", "c", "", "Remote command to run instead of an interactive session")
}
