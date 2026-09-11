package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var randomFlags struct {
	length int
	hex    bool
	digits bool
	alpha  bool
	all    bool
}

var RandomCmd = &cobra.Command{
	Use:     "random-string",
	Aliases: []string{"random"},
	Short:   "Generate a random alphanumeric string",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		charset := anbuGenerics.CharsetAlphaNum
		switch {
		case randomFlags.hex:
			charset = anbuGenerics.CharsetHex
		case randomFlags.digits:
			charset = anbuGenerics.CharsetDigits
		case randomFlags.alpha:
			charset = anbuGenerics.CharsetAlpha
		case randomFlags.all:
			charset = anbuGenerics.CharsetAll
		}
		str, err := anbuGenerics.GenerateRandomStringCharset(randomFlags.length, charset)
		if err != nil {
			u.PrintFatal("Failed to generate random string", err)
		}
		u.PrintGeneric(str)
	},
}

func init() {
	RandomCmd.Flags().IntVarP(&randomFlags.length, "length", "l", 48, "Length of random string")
	RandomCmd.Flags().BoolVar(&randomFlags.hex, "hex", false, "Use hexadecimal characters only")
	RandomCmd.Flags().BoolVar(&randomFlags.digits, "digits", false, "Use digits only")
	RandomCmd.Flags().BoolVar(&randomFlags.alpha, "alpha", false, "Use letters only")
	RandomCmd.Flags().BoolVar(&randomFlags.all, "all", false, "Use letters, digits, and special characters")
	RandomCmd.MarkFlagsMutuallyExclusive("hex", "digits", "alpha", "all")
}
