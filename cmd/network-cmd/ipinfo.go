package networkCmd

import (
	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
)

var IPInfoCmd = &cobra.Command{
	Use:     "ip-info",
	Aliases: []string{"ip"},
	Short:   "Display local network interface and public IP information",
	Run: func(cmd *cobra.Command, args []string) {
		ipv6, _ := cmd.Flags().GetBool("ipv6")
		anbuNetwork.GetLocalIPInfo(ipv6)
	},
}

func init() {
	IPInfoCmd.Flags().BoolP("ipv6", "6", false, "Include IPv6 addresses in the output")
}
