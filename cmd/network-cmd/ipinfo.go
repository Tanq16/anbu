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
		info, err := anbuNetwork.GetLocalIPInfo()
		if err != nil {
			u.PrintFatal("failed to get network interfaces", err)
		}
		u.LineBreak()
		ipv4 := u.NewTable([]string{"Interface", "IP Address", "Subnet Mask", "MAC Address", "Status"})
		ipv4.Rows = info.IPv4
		ipv4.PrintTable()
		if ipInfoFlags.ipv6 {
			ipv6 := u.NewTable([]string{"Interface", "IPv6 Address", "MAC Address", "Status"})
			ipv6.Rows = info.IPv6
			ipv6.PrintTable()
		}
		u.LineBreak()
		if info.PublicErr != nil {
			u.PrintWarn("Could not retrieve public IP", info.PublicErr)
			return
		}
		pub := u.NewTable([]string{"Field", "Value"})
		pub.Rows = info.Public
		pub.PrintTable()
	},
}

func init() {
	IPInfoCmd.Flags().BoolVar(&ipInfoFlags.ipv6, "ipv6", false, "Include IPv6 addresses in the output")
}
