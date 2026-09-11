package sshCmd

import (
	"strconv"

	"github.com/spf13/cobra"
	anbuSSH "github.com/tanq16/anbu/internal/ssh"
	u "github.com/tanq16/anbu/utils"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List named SSH sessions",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		sessions, err := anbuSSH.List()
		if err != nil {
			u.PrintFatal("failed to list sessions", err)
		}
		if len(sessions) == 0 {
			u.PrintInfo("there are no sessions")
			return
		}
		table := u.NewTable([]string{"Name", "User", "Host", "Port"})
		for _, s := range sessions {
			table.Rows = append(table.Rows, []string{s.Name, s.User, s.Host, strconv.Itoa(s.Port)})
		}
		table.PrintTable()
	},
}
