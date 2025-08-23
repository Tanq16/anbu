package cryptoCmd

import (
	"os"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
	u "github.com/tanq16/anbu/utils"
)

var fileCryptoFlags struct {
	file     string
	password string
	decrypt  bool
}

var FileCryptoCmd = &cobra.Command{
	Use:     "encrypt",
	Aliases: []string{"e"},
	Short:   "Encryption/decryption on files using AES-256-GCM symmetric encryption",
	Long: `Examples:
s
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if fileCryptoFlags.password == "" {
			u.PrintError("No password specified")
			os.Exit(1)
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
	FileCryptoCmd.Flags().BoolVarP(&fileCryptoFlags.decrypt, "decrypt", "d", false, "Decryption password")
}
