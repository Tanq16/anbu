package cryptoCmd

import (
	"github.com/spf13/cobra"
)

var printFalsePositives bool

var SecretsScanCmd = &cobra.Command{
	Use:   "secrets [path]",
	Short: "Scan files in a directory for potential secrets and sensitive information",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// scanPath := "."
		// if len(args) > 0 {
		// 	scanPath := args[0]
		// 	if _, err := os.Stat(scanPath); os.IsNotExist(err) {
		// 		logger.Fatal().Str("path", scanPath).Msg("Path does not exist")
		// 	}
		// }
		// err := anbuCrypto.ScanSecretsInPath(scanPath, printFalsePositives)
		// if err != nil {
		// 	logger.Fatal().Err(err).Msg("Secret scanning failed")
		// }
	},
}

func init() {
	SecretsScanCmd.Flags().BoolVarP(&printFalsePositives, "print-false", "p", false, "Print false positives")
}
