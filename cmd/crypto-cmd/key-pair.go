package cryptoCmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
)

var keyPairFlags struct {
	outputPath  string
	pass        string
	keySize     int
	noSSHFormat bool
}

var KeyPairCmd = &cobra.Command{
	Use:     "key-pair",
	Aliases: []string{},
	Short:   "Generate SSH key pairs for encryption",
	Long: `Examples:
s
`,
	Run: func(cmd *cobra.Command, args []string) {
		keyName := filepath.Base(keyPairFlags.outputPath)
		keyDir := filepath.Dir(keyPairFlags.outputPath)
		if keyPairFlags.noSSHFormat {
			anbuCrypto.GenerateKeyPair(keyDir, keyName, keyPairFlags.keySize)
		} else {
			anbuCrypto.GenerateSSHKeyPair(keyDir, keyName, keyPairFlags.keySize)
		}
	},
}

func init() {
	KeyPairCmd.Flags().StringVarP(&keyPairFlags.outputPath, "output-path", "o", "./anbu-key", "Output path for key files")
	KeyPairCmd.Flags().IntVarP(&keyPairFlags.keySize, "key-size", "k", 2048, "RSA key size (2048, 3072, or 4096)")
	KeyPairCmd.Flags().BoolVarP(&keyPairFlags.noSSHFormat, "ssh", "s", false, "Generate keys in SSH format instead of PEM")
}
