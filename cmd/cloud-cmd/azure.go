package cloudCmd

import (
	"errors"

	"github.com/spf13/cobra"
	anbuCloud "github.com/tanq16/anbu/internal/cloud/azure"
	u "github.com/tanq16/anbu/utils"
)

var azureSwitchFlags struct {
	subscription string
}

var AzureCmd = &cobra.Command{
	Use:     "azure",
	Aliases: []string{"az"},
	Short:   "Helper utilities for Azure",
}

var azureSwitchCmd = &cobra.Command{
	Use:     "switch-sub",
	Aliases: []string{"switch"},
	Short:   "Switch between Azure subscriptions interactively",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuCloud.SwitchSubscription(azureSwitchFlags.subscription); err != nil {
			if errors.Is(err, u.ErrNoTerminal) {
				u.PrintFatal("switch-sub needs --subscription when there is no interactive terminal", nil)
			}
			u.PrintFatal("failed to switch subscription", err)
		}
	},
}

func init() {
	azureSwitchCmd.Flags().StringVarP(&azureSwitchFlags.subscription, "subscription", "s", "", "Subscription ID or name")
	AzureCmd.AddCommand(azureSwitchCmd)
}
