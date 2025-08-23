package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	cryptoCmd "github.com/tanq16/anbu/cmd/crypto-cmd"
	genericsCmd "github.com/tanq16/anbu/cmd/generics-cmd"
	networkCmd "github.com/tanq16/anbu/cmd/network-cmd"
	"github.com/tanq16/anbu/utils"
)

var AnbuVersion = "dev-build"
var debugFlag bool

var rootCmd = &cobra.Command{
	Use:     "anbu",
	Short:   "anbu is a tool for performing various everyday tasks with ease",
	Version: AnbuVersion,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func setupLogs() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugFlag {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		utils.GlobalDebugFlag = true
	}
}

func init() {
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	cobra.OnInitialize(setupLogs)

	rootCmd.AddCommand(genericsCmd.StringCmd)
	rootCmd.AddCommand(genericsCmd.TimeCmd)
	rootCmd.AddCommand(genericsCmd.BulkRenameCmd)
	rootCmd.AddCommand(genericsCmd.ConvertCmd)

	rootCmd.AddCommand(cryptoCmd.FileCryptoCmd)
	rootCmd.AddCommand(cryptoCmd.KeyPairCmd)
	rootCmd.AddCommand(cryptoCmd.SecretsScanCmd)
	rootCmd.AddCommand(cryptoCmd.SecretsCmd)

	rootCmd.AddCommand(networkCmd.TunnelCmd)
	rootCmd.AddCommand(networkCmd.HTTPServerCmd)
	rootCmd.AddCommand(networkCmd.IPInfoCmd)
}
