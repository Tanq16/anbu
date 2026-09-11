package genericsCmd

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/archive"
	u "github.com/tanq16/anbu/utils"
)

var archiveFlags struct {
	output  string
	include []string
	exclude []string
	bare    bool
	encrypt string
}

var archiveExtractFlags struct {
	bare     bool
	password string
}

var ArchiveCmd = &cobra.Command{
	Use:   "archive <path> [path...]",
	Short: "Create a zip archive from files and directories",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		include := compileRegexes(archiveFlags.include, "include")
		exclude := compileRegexes(archiveFlags.exclude, "exclude")
		password := archiveFlags.encrypt
		encrypt := cmd.Flags().Changed("encrypt")
		if encrypt && password == "" {
			entered, err := u.PromptPassword("Password:")
			if errors.Is(err, u.ErrNoTerminal) {
				u.PrintFatal("archive needs --encrypt, or --encrypt -", nil)
			}
			if err != nil {
				u.PrintFatal("TUI error", err)
			}
			password = entered
		}
		cfg := archive.CreateConfig{
			Paths:    args,
			Output:   archiveFlags.output,
			Include:  include,
			Exclude:  exclude,
			Bare:     archiveFlags.bare,
			Encrypt:  encrypt,
			Password: password,
		}
		if err := archive.Create(cfg); err != nil {
			u.PrintFatal("failed to create archive", err)
		}
		u.PrintSuccess(fmt.Sprintf("created %s", archiveFlags.output))
	},
}

var archiveExtractCmd = &cobra.Command{
	Use:   "extract <file>",
	Short: "Extract a zip archive",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		encrypted, err := archive.IsEncrypted(args[0])
		if err != nil {
			u.PrintFatal("failed to read archive", err)
		}
		password := archiveExtractFlags.password
		if encrypted && password == "" {
			entered, err := u.PromptPassword("Password:")
			if errors.Is(err, u.ErrNoTerminal) {
				u.PrintFatal("archive extract needs --password, or --password -", nil)
			}
			if err != nil {
				u.PrintFatal("TUI error", err)
			}
			password = entered
		}
		cfg := archive.ExtractConfig{
			Archive:  args[0],
			Dest:     ".",
			Bare:     archiveExtractFlags.bare,
			Password: password,
		}
		if err := archive.Extract(cfg); err != nil {
			u.PrintFatal("failed to extract archive", err)
		}
		u.PrintSuccess("extracted")
	},
}

func compileRegexes(pats []string, flagName string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, 0, len(pats))
	for _, p := range pats {
		re, err := regexp.Compile(p)
		if err != nil {
			u.PrintFatal("invalid --"+flagName+" regex", err)
		}
		out = append(out, re)
	}
	return out
}

func init() {
	ArchiveCmd.Flags().StringVarP(&archiveFlags.output, "output", "o", "archive.zip", "Output zip path")
	ArchiveCmd.Flags().StringSliceVar(&archiveFlags.include, "include", nil, "Include only zip paths matching regex (repeatable)")
	ArchiveCmd.Flags().StringSliceVar(&archiveFlags.exclude, "exclude", nil, "Exclude zip paths matching regex (repeatable)")
	ArchiveCmd.Flags().BoolVar(&archiveFlags.bare, "bare", false, "Store given paths at zip root with no wrapper directory")
	ArchiveCmd.Flags().StringVar(&archiveFlags.encrypt, "encrypt", "", "Password to encrypt the zip, or - to read it from stdin")
	_ = u.MarkStdinLine(ArchiveCmd, "encrypt")

	archiveExtractCmd.Flags().BoolVar(&archiveExtractFlags.bare, "bare", false, "Strip the first path component when extracting")
	archiveExtractCmd.Flags().StringVar(&archiveExtractFlags.password, "password", "", "Password for an encrypted archive, or - to read it from stdin")
	_ = u.MarkStdinLine(archiveExtractCmd, "password")

	ArchiveCmd.AddCommand(archiveExtractCmd)
}
