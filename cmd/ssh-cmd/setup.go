package sshCmd

import (
	"cmp"
	"os"
	"os/user"

	"github.com/spf13/cobra"
	anbuSSH "github.com/tanq16/anbu/internal/ssh"
	u "github.com/tanq16/anbu/utils"
)

var setupFlags struct {
	host string
	user string
	port int
}

func currentUser() string {
	if usr, err := user.Current(); err == nil {
		return cmp.Or(usr.Username, os.Getenv("USER"))
	}
	return os.Getenv("USER")
}

var setupCmd = &cobra.Command{
	Use:   "setup <name>",
	Short: "Create a named SSH session and key pair",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pub, err := anbuSSH.Setup(anbuSSH.SetupConfig{
			Name: args[0],
			Host: setupFlags.host,
			User: setupFlags.user,
			Port: setupFlags.port,
		})
		if err != nil {
			u.PrintFatal("failed to create session", err)
		}
		u.PrintSuccess("created session " + args[0])
		u.PrintGeneric(pub)
	},
}

func init() {
	setupCmd.Flags().StringVarP(&setupFlags.host, "host", "H", "", "Host name or address")
	setupCmd.Flags().StringVarP(&setupFlags.user, "user", "u", currentUser(), "SSH user")
	setupCmd.Flags().IntVarP(&setupFlags.port, "port", "p", 22, "SSH port")
	_ = setupCmd.MarkFlagRequired("host")
}
