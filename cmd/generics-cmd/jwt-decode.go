package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	"github.com/tanq16/anbu/utils"
)

var JwtDecodeCmd = &cobra.Command{
	Use:   "jwt-decode",
	Short: "Decode a JWT token",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("jwt-decode")
		token := args[0]
		err := anbuGenerics.JwtParse(token)
		if err != nil {
			logger.Fatal().Err(err).Msg("Bulk rename operation failed")
		}
	},
}
