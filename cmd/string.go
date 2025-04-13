package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	anbuString "github.com/tanq16/anbu/internal/stringops"
	"github.com/tanq16/anbu/utils"
)

var stringCmd = &cobra.Command{
	Use:   "string",
	Short: "generate a random string, a sequence, or a repetitions",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("string")
		// No args
		if len(args) == 0 {
			randomStr, err := anbuString.GenerateRandomString(0)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate random string")
			}
			fmt.Println(randomStr)
			return
		}
		// Length arg
		if len, err := strconv.Atoi(args[0]); err == nil {
			randomStr, err := anbuString.GenerateRandomString(len)
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
			sequence, err := anbuString.GenerateSequence(length)
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
			repeated, err := anbuString.GenerateRepetition(count, args[2])
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate repetition")
			}
			fmt.Println(repeated)
			return
		}
		// Invalid command
		logger.Fatal().Msg("Unknown command")
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(stringCmd)
}
