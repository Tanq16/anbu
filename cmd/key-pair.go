package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	anbuKeypair "github.com/tanq16/anbu/internal/keypair"
	"github.com/tanq16/anbu/utils"
)

var keyPairFlags struct {
	outputPath string
	pass       string
	keySize    int
}

var keyPairCmd = &cobra.Command{
	Use:   "key-pair",
	Short: "Generate RSA key pairs for encryption",
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("key-pair")
		keyName := filepath.Base(keyPairFlags.outputPath)
		keyDir := filepath.Dir(keyPairFlags.outputPath)

		err := anbuKeypair.GenerateKeyPair(keyDir, keyName, keyPairFlags.keySize)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to generate key pair")
		}

		pubKeyPath := fmt.Sprintf("%s/%s.public.pem", keyDir, keyName)
		privKeyPath := fmt.Sprintf("%s/%s.private.pem", keyDir, keyName)
		fmt.Println(utils.OutDetail(fmt.Sprintf("RSA key pair (%d bits) generated", keyPairFlags.keySize)))
		fmt.Println(utils.OutSuccess("Public key: ") + utils.OutInfo(pubKeyPath))
		fmt.Println(utils.OutSuccess("Private key: ") + utils.OutInfo(privKeyPath))
	},
}

func init() {
	rootCmd.AddCommand(keyPairCmd)

	keyPairCmd.Flags().StringVarP(&keyPairFlags.outputPath, "output-path", "o", "./anbu-key", "Output path for key files")
	keyPairCmd.Flags().IntVarP(&keyPairFlags.keySize, "key-size", "k", 2048, "RSA key size (2048, 3072, or 4096)")
}
