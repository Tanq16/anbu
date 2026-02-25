package cryptoCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
	u "github.com/tanq16/anbu/internal/utils"
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
			u.PrintFatal("No password specified", nil)
		}
		if fileCryptoFlags.decrypt {
			outputPath, err := anbuCrypto.DecryptFileSymmetric(args[0], fileCryptoFlags.password)
			if err != nil {
				u.PrintFatal("decryption failed", err)
			}
			u.PrintGeneric(fmt.Sprintf("\nFile decrypted: %s", u.FSuccess(outputPath)))
		} else {
			outputPath, err := anbuCrypto.EncryptFileSymmetric(args[0], fileCryptoFlags.password)
			if err != nil {
				u.PrintFatal("encryption failed", err)
			}
			u.PrintGeneric(fmt.Sprintf("\nFile encrypted: %s", u.FSuccess(outputPath)))
		}
	},
}

func init() {
	FileCryptoCmd.Flags().StringVarP(&fileCryptoFlags.password, "password", "p", "", "Encryption password")
	FileCryptoCmd.Flags().BoolVarP(&fileCryptoFlags.decrypt, "decrypt", "d", false, "Decrypt the file instead of encrypting")
}
