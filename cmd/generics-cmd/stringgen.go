package genericsCmd

import (
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
)

var StringCmd = &cobra.Command{
	Use:     "string",
	Aliases: []string{"s"},
	Short:   "Generate random strings, sequences, passwords, and passphrases",
	Long: `Generate random strings, sequences, passwords, and passphrases.
Examples:
  anbu string 23                       # generate 23 (100 if not specified) random alphanumeric chars
  anbu string seq 29                   # prints "abcdefghijklmnopqrstuvxyz" until desired length
  anbu string rep 23 str2rep           # prints "str2repstr2rep...23 times"
  anbu string uuid                     # generates a uuid
  anbu string ruid 16                  # generates a short uuid of length b/w 1-32
  anbu string suid                     # generates a short uuid of length 18
  anbu string password                 # generate a 12-character complex password
  anbu string password 16              # generate a 16-character complex password
  anbu string password 8 simple        # generate an 8-letter lowercase password
  anbu string passphrase               # generate a 3-word passphrase with hyphens
  anbu string passphrase 5             # generate a 5-word passphrase with hyphens
  anbu string passphrase 4 '@'         # generate a 4-word passphrase with a custom separator`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// No args
		if len(args) == 0 {
			anbuGenerics.GenerateRandomString(0)
			return
		}
		// Length arg
		if len, err := strconv.Atoi(args[0]); err == nil {
			anbuGenerics.GenerateRandomString(len)
			return
		}
		// Sequence command arg
		if args[0] == "seq" {
			if len(args) < 2 {
				log.Fatal().Msg("Missing length for sequence command")
			}
			length, err := strconv.Atoi(args[1])
			if err != nil {
				log.Fatal().Msg("Not a valid length")
			}
			anbuGenerics.GenerateSequenceString(length)
			return
		}
		// Repeat string command
		if args[0] == "rep" {
			if len(args) < 3 {
				log.Fatal().Msg("Missing count or string for repetition")
			}
			count, err := strconv.Atoi(args[1])
			if err != nil {
				log.Fatal().Msg("Not a valid count")
			}
			anbuGenerics.GenerateRepetitionString(count, args[2])
			return
		}
		if args[0] == "uuid" {
			anbuGenerics.GenerateUUIDString()
			return
		}
		if args[0] == "ruid" {
			if len(args) < 2 {
				log.Fatal().Msg("Missing length for RUID command")
			}
			anbuGenerics.GenerateRUIDString(args[1])
			return
		}
		if args[0] == "suid" {
			anbuGenerics.GenerateRUIDString("18")
			return
		}
		if args[0] == "password" {
			if len(args) < 2 {
				anbuGenerics.GeneratePassword("12", false)
			} else if len(args) == 2 {
				anbuGenerics.GeneratePassword(args[1], false)
			} else if len(args) == 3 {
				anbuGenerics.GeneratePassword(args[1], true)
			}
			return
		}
		if args[0] == "passphrase" {
			if len(args) < 2 {
				anbuGenerics.GeneratePassPhrase("3", "-", false)
			} else if len(args) == 2 {
				anbuGenerics.GeneratePassPhrase(args[1], "-", false)
			} else if len(args) == 3 {
				if args[2] == "simple" {
					anbuGenerics.GeneratePassPhrase(args[1], "-", true)
				} else {
					anbuGenerics.GeneratePassPhrase(args[1], args[2], false)
				}
			}
			return
		}
		// Invalid command
		cmd.Help()
	},
}
