package cryptoCmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
)

var fileCryptoFlags struct {
	file     string
	password string
	decrypt  bool
}

var FileCryptoCmd = &cobra.Command{
	Use:     "file-crypt <file-path>",
	Aliases: []string{"fc"},
	Short:   "Encrypt or decrypt files using AES-256-GCM with a password-based symmetric encryption",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if fileCryptoFlags.password == "" {
			log.Fatal().Msg("No password specified")
		}
		if fileCryptoFlags.decrypt {
			anbuCrypto.DecryptFileSymmetric(args[0], fileCryptoFlags.password)
		} else {
			anbuCrypto.EncryptFileSymmetric(args[0], fileCryptoFlags.password)
		}
	},
}

func init() {
	FileCryptoCmd.Flags().StringVarP(&fileCryptoFlags.password, "password", "p", "", "Encryption password")
	FileCryptoCmd.Flags().BoolVarP(&fileCryptoFlags.decrypt, "decrypt", "d", false, "Decrypt the file instead of encrypting")
}
