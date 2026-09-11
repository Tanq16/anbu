package genericsCmd

import (
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var uuidFlags struct {
	short bool
}

var UUIDCmd = &cobra.Command{
	Use:   "uuid",
	Short: "Generate a UUID v4",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if uuidFlags.short {
			str, err := anbuGenerics.GenerateRUIDString(18)
			if err != nil {
				u.PrintFatal("Failed to generate short UUID", err)
			}
			u.PrintGeneric(str)
			return
		}
		str, err := anbuGenerics.GenerateUUIDString()
		if err != nil {
			u.PrintFatal("Failed to generate UUID", err)
		}
		u.PrintGeneric(str)
	},
}

func init() {
	UUIDCmd.Flags().BoolVar(&uuidFlags.short, "short", false, "Generate a short UUID of length 18")
}
