package networkCmd

import (
	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
)

var ipInfoFlags struct {
	ipv6 bool
}

var IPInfoCmd = &cobra.Command{
	Use:     "ip-info",
	Aliases: []string{"ip"},
	Short:   "Display local network interface and public IP information",
	Run: func(cmd *cobra.Command, args []string) {
		anbuNetwork.GetLocalIPInfo(ipInfoFlags.ipv6)
	},
}

func init() {
	IPInfoCmd.Flags().BoolVarP(&ipInfoFlags.ipv6, "ipv6", "6", false, "Include IPv6 addresses in the output")
}
