package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var uuidFlags struct {
	short bool
	v4    bool
}

var UUIDCmd = &cobra.Command{
	Use:   "uuid",
	Short: "Generate a UUID v7, or v4 with --v4",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var str string
		var err error
		if uuidFlags.short {
			str, err = anbuGenerics.GenerateShortUUIDString(uuidFlags.v4)
		} else {
			str, err = anbuGenerics.GenerateUUIDString(uuidFlags.v4)
		}
		if err != nil {
			u.PrintFatal("Failed to generate UUID", err)
		}
		u.PrintGeneric(str)
	},
}

func init() {
	UUIDCmd.Flags().BoolVar(&uuidFlags.short, "short", false, "Generate a short UUID of length 18, stripping non-random bits")
	UUIDCmd.Flags().BoolVar(&uuidFlags.v4, "v4", false, "Generate a UUID v4 instead of v7")
}
