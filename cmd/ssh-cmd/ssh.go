package sshCmd

import (
	"github.com/spf13/cobra"
)

var SSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Manage named SSH sessions",
}

func init() {
	SSHCmd.AddCommand(setupCmd)
	SSHCmd.AddCommand(listCmd)
	SSHCmd.AddCommand(deleteCmd)
	SSHCmd.AddCommand(execCmd)
}
