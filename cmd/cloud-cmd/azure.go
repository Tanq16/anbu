package cloudCmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	anbuCloud "github.com/tanq16/anbu/internal/cloud/azure"
)

var AzureCmd = &cobra.Command{
	Use:     "azure",
	Aliases: []string{"az"},
	Short:   "Helper utilities for Azure",
}

var azureSwitchCmd = &cobra.Command{
	Use:     "switch-sub",
	Aliases: []string{"switch"},
	Short:   "Switch between Azure subscriptions interactively",
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCloud.SwitchSubscription(); err != nil {
			log.Fatal().Err(err).Msg("failed to switch subscription")
		}
	},
}

func init() {
	AzureCmd.AddCommand(azureSwitchCmd)
}
