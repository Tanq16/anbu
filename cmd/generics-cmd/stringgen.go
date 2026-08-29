package genericsCmd

import (
	"strconv"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var seqFlags struct {
	length int
}

var ruidFlags struct {
	length int
}

var passwordFlags struct {
	length int
	simple bool
}

var passphraseFlags struct {
	length     int
	separator  string
	capitalize bool
}

var StringCmd = &cobra.Command{
	Use:     "string",
	Aliases: []string{"s"},
	Short:   "Generate random strings, sequences, passwords, and passphrases",
}

var randomCmd = &cobra.Command{
	Use:   "random [length]",
	Short: "Generate a random alphanumeric string",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		length := 100
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil {
				length = l
			}
		}
		str, err := anbuGenerics.GenerateRandomString(length)
		if err != nil {
			u.PrintFatal("Failed to generate random string", err)
		}
		u.PrintGeneric(str)
	},
}

var seqCmd = &cobra.Command{
	Use:   "seq [length]",
	Short: "Generate sequence string",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		length := seqFlags.length
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil {
				length = l
			}
		}
		u.PrintGeneric(anbuGenerics.GenerateSequenceString(length))
	},
}

var repCmd = &cobra.Command{
	Use:   "rep <count> <string>",
	Short: "Generate repeated string",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		count, err := strconv.Atoi(args[0])
		if err != nil {
			u.PrintFatal("Invalid repetition count", err)
		}
		u.PrintGeneric(anbuGenerics.GenerateRepetitionString(count, args[1]))
	},
}

var uuidCmd = &cobra.Command{
	Use:   "uuid",
	Short: "Generate a UUID v4",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		str, err := anbuGenerics.GenerateUUIDString()
		if err != nil {
			u.PrintFatal("Failed to generate UUID", err)
		}
		u.PrintGeneric(str)
	},
}

var ruidCmd = &cobra.Command{
	Use:   "ruid [length]",
	Short: "Generate a short UUID",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		length := ruidFlags.length
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil {
				length = l
			}
		}
		str, err := anbuGenerics.GenerateRUIDString(length)
		if err != nil {
			u.PrintFatal("Failed to generate RUID", err)
		}
		u.PrintGeneric(str)
	},
}

var suidCmd = &cobra.Command{
	Use:   "suid",
	Short: "Generate a short UUID of length 18",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		str, err := anbuGenerics.GenerateRUIDString(18)
		if err != nil {
			u.PrintFatal("Failed to generate SUID", err)
		}
		u.PrintGeneric(str)
	},
}

var passwordCmd = &cobra.Command{
	Use:   "password [length]",
	Short: "Generate a random password",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		length := passwordFlags.length
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil {
				length = l
			}
		}
		pwd, err := anbuGenerics.GeneratePassword(length, passwordFlags.simple)
		if err != nil {
			u.PrintFatal("Failed to generate password", err)
		}
		u.PrintGeneric(pwd)
	},
}

var passphraseCmd = &cobra.Command{
	Use:   "passphrase [length]",
	Short: "Generate a passphrase",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		length := passphraseFlags.length
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil {
				length = l
			}
		}
		phrase, err := anbuGenerics.GeneratePassPhrase(length, passphraseFlags.separator, passphraseFlags.capitalize)
		if err != nil {
			u.PrintFatal("Failed to generate passphrase", err)
		}
		u.PrintGeneric(phrase)
	},
}

func init() {
	seqCmd.Flags().IntVarP(&seqFlags.length, "length", "l", 100, "Length of sequence string")
	ruidCmd.Flags().IntVarP(&ruidFlags.length, "length", "l", 18, "Length of RUID (1-30)")

	passwordCmd.Flags().IntVarP(&passwordFlags.length, "length", "l", 12, "Length of password")
	passwordCmd.Flags().BoolVar(&passwordFlags.simple, "simple", false, "Use simple lowercase password")

	passphraseCmd.Flags().IntVarP(&passphraseFlags.length, "length", "l", 3, "Number of words in passphrase")
	passphraseCmd.Flags().StringVarP(&passphraseFlags.separator, "separator", "s", "-", "Word separator")
	passphraseCmd.Flags().BoolVar(&passphraseFlags.capitalize, "capitalize", false, "Capitalize words and add digits")

	StringCmd.AddCommand(randomCmd)
	StringCmd.AddCommand(seqCmd)
	StringCmd.AddCommand(repCmd)
	StringCmd.AddCommand(uuidCmd)
	StringCmd.AddCommand(ruidCmd)
	StringCmd.AddCommand(suidCmd)
	StringCmd.AddCommand(passwordCmd)
	StringCmd.AddCommand(passphraseCmd)
}
