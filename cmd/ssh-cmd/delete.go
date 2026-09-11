package sshCmd

import (
	"github.com/spf13/cobra"
	anbuSSH "github.com/tanq16/anbu/internal/ssh"
	u "github.com/tanq16/anbu/utils"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a named SSH session and its keys",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuSSH.Delete(args[0]); err != nil {
			u.PrintFatal("failed to delete session", err)
		}
		u.PrintSuccess("deleted session " + args[0])
	},
}
