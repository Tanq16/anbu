package cryptoCmd

import (
	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
)

var JwtDecodeCmd = &cobra.Command{
	Use:   "jwt-decode",
	Short: "Decode a JWT token",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token := args[0]
		anbuCrypto.JwtParse(token)
	},
}
