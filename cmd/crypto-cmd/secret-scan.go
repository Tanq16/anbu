package cryptoCmd

import (
	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
	"github.com/tanq16/anbu/utils"
)

var SecretsScanCmd = &cobra.Command{
	Use:   "secrets [path]",
	Short: "Scan files in a directory for potential secrets and sensitive information",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("secrets")
		scanPath := args[0]
		err := anbuCrypto.ScanSecretsInPath(scanPath)
		if err != nil {
			logger.Fatal().Err(err).Msg("Secret scanning failed")
		}
	},
}
