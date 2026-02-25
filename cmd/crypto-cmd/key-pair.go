package cryptoCmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
	u "github.com/tanq16/anbu/internal/utils"
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
	Run: func(cmd *cobra.Command, args []string) {
		keyName := filepath.Base(keyPairFlags.outputPath)
		keyDir := filepath.Dir(keyPairFlags.outputPath)
		var result *anbuCrypto.KeyPairResult
		var err error
		if keyPairFlags.noSSHFormat {
			result, err = anbuCrypto.GenerateKeyPair(keyDir, keyName, keyPairFlags.keySize)
		} else {
			result, err = anbuCrypto.GenerateSSHKeyPair(keyDir, keyName, keyPairFlags.keySize)
		}
		if err != nil {
			u.PrintFatal("key pair generation failed", err)
		}
		u.LineBreak()
		if keyPairFlags.noSSHFormat {
			u.PrintGeneric(fmt.Sprintf("RSA key pair (%d bits) generated", result.KeySize))
		} else {
			u.PrintGeneric(fmt.Sprintf("SSH key pair (%d bits) generated", result.KeySize))
		}
		u.PrintGeneric(fmt.Sprintf("Public key: %s", u.FInfo(result.PublicKeyPath)))
		u.PrintGeneric(fmt.Sprintf("Private key: %s", u.FInfo(result.PrivateKeyPath)))
	},
}

func init() {
	KeyPairCmd.Flags().StringVarP(&keyPairFlags.outputPath, "output-path", "o", "./anbu-key", "Output path and name for the key files")
	KeyPairCmd.Flags().IntVarP(&keyPairFlags.keySize, "key-size", "k", 2048, "RSA key size (e.g., 2048, 3072, 4096)")
	KeyPairCmd.Flags().BoolVarP(&keyPairFlags.noSSHFormat, "ssh", "s", false, "Generate keys in SSH format instead of PEM")
}
