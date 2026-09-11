package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var randomFlags struct {
	length int
}

func runRandom(_ *cobra.Command, _ []string) {
	str, err := anbuGenerics.GenerateRandomString(randomFlags.length)
	if err != nil {
		u.PrintFatal("Failed to generate random string", err)
	}
	u.PrintGeneric(str)
}

var RandomCmd = &cobra.Command{
	Use:   "random",
	Short: "Generate a random alphanumeric string",
	Args:  cobra.NoArgs,
	Run:   runRandom,
}

var randomStringCmd = &cobra.Command{
	Use:   "string",
	Short: "Generate a random alphanumeric string",
	Args:  cobra.NoArgs,
	Run:   runRandom,
}

func init() {
	RandomCmd.PersistentFlags().IntVarP(&randomFlags.length, "length", "l", 100, "Length of random string")
	RandomCmd.AddCommand(randomStringCmd)
}
