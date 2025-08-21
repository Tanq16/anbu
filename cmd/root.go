package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	cryptoCmd "github.com/tanq16/anbu/cmd/crypto-cmd"
	genericsCmd "github.com/tanq16/anbu/cmd/generics-cmd"
	networkCmd "github.com/tanq16/anbu/cmd/network-cmd"
)

var AnbuVersion = "dev-build"

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

func init() {
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.AddCommand(genericsCmd.LoopCmd)
	rootCmd.AddCommand(genericsCmd.StringCmd)
	rootCmd.AddCommand(genericsCmd.TimeCmd)
	rootCmd.AddCommand(genericsCmd.BulkRenameCmd)
	rootCmd.AddCommand(genericsCmd.ConvertCmd)
	rootCmd.AddCommand(genericsCmd.TemplateCmd)

	rootCmd.AddCommand(cryptoCmd.FileCryptoCmd)
	rootCmd.AddCommand(cryptoCmd.KeyPairCmd)
	rootCmd.AddCommand(cryptoCmd.JwtDecodeCmd)
	rootCmd.AddCommand(cryptoCmd.SecretsScanCmd)
	rootCmd.AddCommand(cryptoCmd.SecretsCmd)

	rootCmd.AddCommand(networkCmd.TunnelCmd)
	rootCmd.AddCommand(networkCmd.HTTPServerCmd)
	rootCmd.AddCommand(networkCmd.IPInfoCmd)
}
