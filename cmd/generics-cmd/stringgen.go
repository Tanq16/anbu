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
	Short:   "generate a random string, a sequence, a repetition, or password/passphrase",
	Long: `generate a variety of strings with the following
Examples:
	anbu string N                      # generate a random string of length N
	anbu string seq N                  # generate a sequence of length N
	anbu string rep N hello            # repeat the string hello N times
	anbu string uuid                   # generate a UUID
	anbu string ruid [N b/w 1 and 30]  # generate a reduced UUID of length N
	anbu string suid                   # generate a short UUID of length 18
	anbu string password               # generate a password of length 12
	anbu string password N             # generate a password of length N
	anbu string password N simple      # generate a password of length N w/ only letters
	anbu string passphrase             # generate a passphrase of 3 words
	anbu string passphrase N           # generate a passphrase of N words
	anbu string passphrase N "M"       # generate a passphrase of N words w/ M as separator
	anbu string passphrase N simple    # generate a passphrase of N words and - as separator`,
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
