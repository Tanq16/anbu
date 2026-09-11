package networkCmd

import (
	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
	u "github.com/tanq16/anbu/utils"
)

var ipInfoFlags struct {
	ipv6 bool
}

var IPInfoCmd = &cobra.Command{
	Use:     "ip-info",
	Aliases: []string{"ip"},
	Short:   "Display local network interface and public IP information",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuNetwork.GetLocalIPInfo(ipInfoFlags.ipv6); err != nil {
			u.PrintFatal("failed to get network interfaces", err)
		}
	},
}

func init() {
	IPInfoCmd.Flags().BoolVar(&ipInfoFlags.ipv6, "ipv6", false, "Include IPv6 addresses in the output")
}
