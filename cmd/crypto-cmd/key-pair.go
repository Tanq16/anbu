package cryptoCmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
)

var keyPairFlags struct {
	outputPath string
	pass       string
	keySize    int
	sshFormat  bool
}

var KeyPairCmd = &cobra.Command{
	Use:   "key-pair",
	Short: "Generate RSA key pairs for encryption",
	Run: func(cmd *cobra.Command, args []string) {
		keyName := filepath.Base(keyPairFlags.outputPath)
		keyDir := filepath.Dir(keyPairFlags.outputPath)
		if keyPairFlags.sshFormat {
			anbuCrypto.GenerateSSHKeyPair(keyDir, keyName, keyPairFlags.keySize)
		} else {
			anbuCrypto.GenerateKeyPair(keyDir, keyName, keyPairFlags.keySize)
		}
	},
}

func init() {
	KeyPairCmd.Flags().StringVarP(&keyPairFlags.outputPath, "output-path", "o", "./anbu-key", "Output path for key files")
	KeyPairCmd.Flags().IntVarP(&keyPairFlags.keySize, "key-size", "k", 2048, "RSA key size (2048, 3072, or 4096)")
	KeyPairCmd.Flags().BoolVarP(&keyPairFlags.sshFormat, "ssh", "s", false, "Generate keys in SSH format instead of PEM")
}
