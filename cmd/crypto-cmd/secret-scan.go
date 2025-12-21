package cryptoCmd

import (
	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
)

var printFalsePositives bool

var SecretsScanCmd = &cobra.Command{
	Use:     "secret-scan [path]",
	Aliases: []string{},
	Short:   "Scan directories for potential secrets using regex patterns and filter results by confidence level",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scanPath := "."
		if len(args) > 0 {
			scanPath = args[0]
		}
		anbuCrypto.ScanSecretsInPath(scanPath, printFalsePositives)
	},
}

func init() {
	SecretsScanCmd.Flags().BoolVarP(&printFalsePositives, "print-false", "p", false, "Include generic matches that may be false positives")
}
