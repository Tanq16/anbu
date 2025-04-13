package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	anbuStringGen "github.com/tanq16/anbu/internal/stringgen"
	"github.com/tanq16/anbu/utils"
)

var stringCmd = &cobra.Command{
	Use:   "sgen",
	Short: "generate a random string, a uuid, a shorter uuid, a sequence, or a repetition",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger("string")
		// No args
		if len(args) == 0 {
			randomStr, err := anbuStringGen.GenerateRandomString(0)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate random string")
			}
			fmt.Println(randomStr)
			return
		}
		// Length arg
		if len, err := strconv.Atoi(args[0]); err == nil {
			randomStr, err := anbuStringGen.GenerateRandomString(len)
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
			sequence, err := anbuStringGen.GenerateSequence(length)
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
			repeated, err := anbuStringGen.GenerateRepetition(count, args[2])
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate repetition")
			}
			fmt.Println(repeated)
			return
		}
		if args[0] == "uuid" {
			uuid, err := anbuStringGen.GenerateUUID()
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate UUID")
			}
			fmt.Println(utils.OutSuccess(uuid))
			return
		}
		if args[0] == "ruid" {
			if len(args) < 2 {
				logger.Fatal().Msg("Missing length for RUID command")
			}
			ruid, err := anbuStringGen.GenerateRUID(args[1])
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to generate RUID")
			}
			fmt.Println(utils.OutSuccess(ruid))
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
