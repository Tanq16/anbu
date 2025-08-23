package networkCmd

import (
	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
)

var IPInfoCmd = &cobra.Command{
	Use:     "ip-info",
	Aliases: []string{"ip"},
	Short:   "Display local network interface and public IP information",
	Long: `Shows details about local network interfaces, including IPv4 addresses,
subnet masks, and MAC addresses. It also fetches and displays public IP
information, including geolocation data from ipinfo.io.

Examples:
  # Display local IPv4 and public IP information
  anbu ip-info

  # Include local IPv6 addresses in the output
  anbu ip-info -6`,
	Run: func(cmd *cobra.Command, args []string) {
		ipv6, _ := cmd.Flags().GetBool("ipv6")
		anbuNetwork.GetLocalIPInfo(ipv6)
	},
}

func init() {
	IPInfoCmd.Flags().BoolP("ipv6", "6", false, "Include IPv6 addresses in the output")
}
