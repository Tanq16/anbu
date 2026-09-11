package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var passphraseFlags struct {
	length int
	simple bool
}

var PassphraseCmd = &cobra.Command{
	Use:   "passphrase",
	Short: "Generate a passphrase",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		phrase, err := anbuGenerics.GeneratePassPhrase(passphraseFlags.length, passphraseFlags.simple)
		if err != nil {
			u.PrintFatal("Failed to generate passphrase", err)
		}
		u.PrintGeneric(phrase)
	},
}

func init() {
	PassphraseCmd.Flags().IntVarP(&passphraseFlags.length, "length", "l", 3, "Number of words in passphrase")
	PassphraseCmd.Flags().BoolVar(&passphraseFlags.simple, "simple", false, "Use plain words with no capital letter or digit")
}
