package genericsCmd

import (
	"strconv"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/utils"
)

var (
	passwordLength int
	passwordSimple bool

	passphraseLength     int
	passphraseSeparator  string
	passphraseCapitalize bool

	seqLength  int
	ruidLength int
)

var StringCmd = &cobra.Command{
	Use:     "string [length]",
	Aliases: []string{"s"},
	Short:   "Generate random strings, sequences, passwords, and passphrases",
	Long: `Generate random strings, sequences, passwords, and passphrases.

Examples:
  anbu string 23                       # generate 23 random alphanumeric chars (default 100)
  anbu string seq 29                   # prints "abcdefghijklmnopqrstuvwxyz" until desired length
  anbu string rep 23 str2rep           # prints "str2rep" repeated 23 times
  anbu string uuid                     # generates a UUID v4
  anbu string ruid 16                  # generates a short UUID of length 1-32
  anbu string suid                     # generates a short UUID of length 18
  anbu string password                 # generate a 12-character complex password
  anbu string password 16              # generate a 16-character complex password
  anbu string password 8 -s            # generate an 8-letter simple password
  anbu string passphrase               # generate a 3-word passphrase with hyphens
  anbu string passphrase -l 5          # generate a 5-word passphrase
  anbu string passphrase -l 4 -s '@'   # generate a 4-word passphrase with a custom separator
  anbu string passphrase -c           # generate a passphrase with capitalization and digits`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		length := 100
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil {
				length = l
			}
		}
		str, err := anbuGenerics.GenerateRandomString(length)
		if err != nil {
			u.PrintError("Failed to generate random string", err)
			return
		}
		u.PrintGeneric(str)
	},
}

var seqCmd = &cobra.Command{
	Use:   "seq [length]",
	Short: "Generate sequence string",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		length := seqLength
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
			u.PrintError("Invalid repetition count", err)
			return
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
			u.PrintError("Failed to generate UUID", err)
			return
		}
		u.PrintGeneric(str)
	},
}

var ruidCmd = &cobra.Command{
	Use:   "ruid [length]",
	Short: "Generate a short UUID",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		length := ruidLength
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil {
				length = l
			}
		}
		str, err := anbuGenerics.GenerateRUIDString(length)
		if err != nil {
			u.PrintError("Failed to generate RUID", err)
			return
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
			u.PrintError("Failed to generate SUID", err)
			return
		}
		u.PrintGeneric(str)
	},
}

var passwordCmd = &cobra.Command{
	Use:   "password [length]",
	Short: "Generate a random password",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		length := passwordLength
		simple := passwordSimple
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil {
				length = l
			}
		}
		if len(args) > 1 && args[1] == "simple" {
			simple = true
		}
		pwd, err := anbuGenerics.GeneratePassword(length, simple)
		if err != nil {
			u.PrintError("Failed to generate password", err)
			return
		}
		u.PrintGeneric(pwd)
	},
}

var passphraseCmd = &cobra.Command{
	Use:   "passphrase [length]",
	Short: "Generate a passphrase",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		length := passphraseLength
		sep := passphraseSeparator
		cap := passphraseCapitalize

		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil {
				length = l
			}
		}
		if len(args) > 1 {
			if args[1] == "simple" {
				cap = false
			} else {
				sep = args[1]
			}
		}
		if len(args) > 2 {
			if args[2] == "simple" {
				cap = false
			} else {
				sep = args[2]
			}
		}

		phrase, err := anbuGenerics.GeneratePassPhrase(length, sep, cap)
		if err != nil {
			u.PrintError("Failed to generate passphrase", err)
			return
		}
		u.PrintGeneric(phrase)
	},
}

func init() {
	seqCmd.Flags().IntVarP(&seqLength, "length", "l", 100, "Length of sequence string")
	ruidCmd.Flags().IntVarP(&ruidLength, "length", "l", 18, "Length of RUID (1-30)")

	passwordCmd.Flags().IntVarP(&passwordLength, "length", "l", 12, "Length of password")
	passwordCmd.Flags().BoolVarP(&passwordSimple, "simple", "s", false, "Use simple lowercase password")

	passphraseCmd.Flags().IntVarP(&passphraseLength, "length", "l", 3, "Number of words in passphrase")
	passphraseCmd.Flags().StringVarP(&passphraseSeparator, "separator", "s", "-", "Word separator")
	passphraseCmd.Flags().BoolVarP(&passphraseCapitalize, "capitalize", "c", false, "Capitalize words and add digits")

	StringCmd.AddCommand(seqCmd)
	StringCmd.AddCommand(repCmd)
	StringCmd.AddCommand(uuidCmd)
	StringCmd.AddCommand(ruidCmd)
	StringCmd.AddCommand(suidCmd)
	StringCmd.AddCommand(passwordCmd)
	StringCmd.AddCommand(passphraseCmd)
}
