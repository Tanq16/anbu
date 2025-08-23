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
	Short:   "Generate RSA key pairs in PEM or SSH format",
	Long: `Generates RSA public and private key pairs.
Supports standard PEM format for general use and OpenSSH format for SSH authentication.

Examples:
  # Generate a 4096-bit RSA key pair in PEM format
  anbu key-pair -o my-key -k 4096

  # Generate a default 2048-bit RSA key pair in SSH format
  anbu key-pair --ssh -o my-ssh-key`,
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
	KeyPairCmd.Flags().StringVarP(&keyPairFlags.outputPath, "output-path", "o", "./anbu-key", "Output path and name for the key files")
	KeyPairCmd.Flags().IntVarP(&keyPairFlags.keySize, "key-size", "k", 2048, "RSA key size (e.g., 2048, 3072, 4096)")
	KeyPairCmd.Flags().BoolVarP(&keyPairFlags.noSSHFormat, "ssh", "s", false, "Generate keys in SSH format instead of PEM")
}
