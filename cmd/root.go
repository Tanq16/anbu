package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	cryptoCmd "github.com/tanq16/anbu/cmd/crypto-cmd"
	genericsCmd "github.com/tanq16/anbu/cmd/generics-cmd"
	networkCmd "github.com/tanq16/anbu/cmd/network-cmd"
)

var AnbuVersion = "dev-build"
var debug bool

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
	if debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug logging enabled")
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func init() {
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	cobra.OnInitialize(setupLogs)

	rootCmd.AddCommand(genericsCmd.LoopCmd)
	rootCmd.AddCommand(genericsCmd.StringCmd)
	rootCmd.AddCommand(genericsCmd.TimeCmd)
	rootCmd.AddCommand(genericsCmd.BulkRenameCmd)
	rootCmd.AddCommand(genericsCmd.ConvertCmd)

	rootCmd.AddCommand(cryptoCmd.FileCryptoCmd)
	rootCmd.AddCommand(cryptoCmd.KeyPairCmd)
	rootCmd.AddCommand(cryptoCmd.JwtDecodeCmd)
	rootCmd.AddCommand(cryptoCmd.SecretsScanCmd)
	rootCmd.AddCommand(cryptoCmd.SecretsCmd)

	rootCmd.AddCommand(networkCmd.TunnelCmd)
	rootCmd.AddCommand(networkCmd.HTTPServerCmd)
	rootCmd.AddCommand(networkCmd.IPInfoCmd)
}
