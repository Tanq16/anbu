package networkCmd

import (
	"github.com/spf13/cobra"
	anbuNetwork "github.com/tanq16/anbu/internal/network"
	"github.com/tanq16/anbu/utils"
)

var IPInfoCmd = &cobra.Command{
	Use:   "ipinfo",
	Short: "Display network interface information including IPs, gateway, DNS, and more",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("ipinfo")
		var err error
		if len(args) > 0 {
			includeIPv6 := args[0] == "ipv6"
			err = anbuNetwork.GetLocalIPInfo(includeIPv6)
		} else {
			err = anbuNetwork.GetLocalIPInfo(false)
		}
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to retrieve network information")
		}
	},
}
