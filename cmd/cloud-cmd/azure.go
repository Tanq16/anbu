package cloudCmd

import (
	"github.com/spf13/cobra"
	anbuCloud "github.com/tanq16/anbu/internal/cloud/azure"
	u "github.com/tanq16/anbu/internal/utils"
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
			u.PrintFatal("failed to switch subscription", err)
		}
	},
}

func init() {
	AzureCmd.AddCommand(azureSwitchCmd)
}
