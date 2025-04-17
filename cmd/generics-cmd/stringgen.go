package genericsCmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	"github.com/tanq16/anbu/utils"
)

var StringCmd = &cobra.Command{
	Use:   "string",
	Short: "generate a random string, a sequence, a repetition, or password/passphrase",
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
		logger := utils.GetLogger("string")
		// No args
		if len(args) == 0 {
			randomStr, err := anbuGenerics.GenerateRandomString(0)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate random string")
			}
			fmt.Println(randomStr)
			return
		}
		// Length arg
		if len, err := strconv.Atoi(args[0]); err == nil {
			randomStr, err := anbuGenerics.GenerateRandomString(len)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate random string")
			}
			fmt.Println(randomStr)
			return
		}
		// Sequence command arg
		if args[0] == "seq" {
			if len(args) < 2 {
				logger.Fatal().Msg("Missing length for sequence command")
			}
			length, err := strconv.Atoi(args[1])
			if err != nil {
				logger.Fatal().Err(err).Msg("Not a valid length")
			}
			sequence, err := anbuGenerics.GenerateSequenceString(length)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate sequence")
			}
			fmt.Println(sequence)
			return
		}
		// Repeat string command
		if args[0] == "rep" {
			if len(args) < 3 {
				logger.Fatal().Msg("Missing count or string for repetition")
			}
			count, err := strconv.Atoi(args[1])
			if err != nil {
				logger.Fatal().Err(err).Msg("Not a valid count")
			}
			repeated, err := anbuGenerics.GenerateRepetitionString(count, args[2])
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate repetition")
			}
			fmt.Println(repeated)
			return
		}
		if args[0] == "uuid" {
			uuid, err := anbuGenerics.GenerateUUIDString()
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate UUID")
			}
			fmt.Println(uuid)
			return
		}
		if args[0] == "ruid" {
			if len(args) < 2 {
				logger.Fatal().Msg("Missing length for RUID command")
			}
			ruid, err := anbuGenerics.GenerateRUIDString(args[1])
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate RUID")
			}
			fmt.Println(ruid)
			return
		}
		if args[0] == "suid" {
			suid, err := anbuGenerics.GenerateRUIDString("18")
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate SUID")
			}
			fmt.Println(suid)
			return
		}
		if args[0] == "password" {
			var password string
			var err error
			if len(args) < 2 {
				password, err = anbuGenerics.GeneratePassword("12", false)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to generate password")
				}
			} else if len(args) == 2 {
				password, err = anbuGenerics.GeneratePassword(args[1], false)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to generate password")
				}
			} else if len(args) == 3 {
				password, err = anbuGenerics.GeneratePassword(args[1], true)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to generate password")
				}
			}
			fmt.Println(password)
			return
		}
		if args[0] == "passphrase" {
			var passphrase string
			var err error
			if len(args) < 2 {
				passphrase, err = anbuGenerics.GeneratePassPhrase("3", "-", false)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to generate passphrase")
				}
			} else if len(args) == 2 {
				passphrase, err = anbuGenerics.GeneratePassPhrase(args[1], "-", false)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to generate passphrase")
				}
			} else if len(args) == 3 {
				passphrase, err = anbuGenerics.GeneratePassPhrase(args[1], args[2], false)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to generate passphrase")
				}
			} else if len(args) == 4 {
				passphrase, err = anbuGenerics.GeneratePassPhrase(args[1], args[2], true)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to generate passphrase")
				}
			}
			fmt.Println(passphrase)
			return
		}
		// Invalid command
		logger.Fatal().Msg("Unknown command")
		cmd.Help()
	},
}
