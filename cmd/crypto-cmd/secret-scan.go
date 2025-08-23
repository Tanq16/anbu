package cryptoCmd

import (
	"github.com/spf13/cobra"
	anbuCrypto "github.com/tanq16/anbu/internal/crypto"
)

var printFalsePositives bool

var SecretsScanCmd = &cobra.Command{
	Use:     "secret-scan [path]",
	Aliases: []string{},
	Short:   "Scan a directory for potential secrets using regex patterns",
	Long: `Scans files in a given directory for secrets like API keys and tokens.
If no path is provided, it scans the current directory.
The scan ignores binary files, dotfiles, and common development directories.

Examples:
  # Scan the current directory for high-confidence secrets
  anbu secret-scan

  # Scan a specific project path
  anbu secret-scan /path/to/my-project

  # Scan and include generic, lower-confidence matches
  anbu secret-scan /path/to/my-project -p`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scanPath := "."
		if len(args) > 0 {
			scanPath = args[0]
		}
		anbuCrypto.ScanSecretsInPath(scanPath, printFalsePositives)
	},
}

func init() {
	SecretsScanCmd.Flags().BoolVarP(&printFalsePositives, "print-false", "p", false, "Include generic matches that may be false positives")
}
