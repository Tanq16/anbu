package cmd

// import (
// 	"fmt"
// 	"os"

// 	"github.com/spf13/cobra"
// 	"github.com/tanq16/anbu/internal/fileutil"
// 	"github.com/tanq16/anbu/utils"
// )

// var fileUtilFlags struct {
// 	encryptFile string
// 	decryptFile string
// 	password    string
// 	length      int
// 	count       int
// 	command     string
// }

// var fileCryptoCmd = &cobra.Command{
// 	Use:   "fileutil",
// 	Short: "File and string utility functions",
// }

// var fileCryptoEncryptCmd = &cobra.Command{
// 	Use:   "encrypt",
// 	Short: "Encrypt a file using AES-256-ECB",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		logger := utils.GetLogger("fileutil")

// 		if fileUtilFlags.encryptFile == "" {
// 			logger.Error().Msg("No input file specified")
// 			fmt.Println(utils.OutError("Error: No input file specified"))
// 			cmd.Help()
// 			os.Exit(1)
// 		}

// 		if fileUtilFlags.password == "" {
// 			logger.Error().Msg("No password specified")
// 			fmt.Println(utils.OutError("Error: No password specified"))
// 			cmd.Help()
// 			os.Exit(1)
// 		}

// 		err := fileutil.EncryptFile(fileUtilFlags.encryptFile, fileUtilFlags.password)
// 		if err != nil {
// 			logger.Error().Err(err).Msg("Failed to encrypt file")
// 			fmt.Println(utils.OutError("Error: " + err.Error()))
// 			os.Exit(1)
// 		}

// 		fmt.Println(utils.OutSuccess("File encrypted successfully: " + fileUtilFlags.encryptFile + ".enc"))
// 	},
// }

// var fileCryptoDecryptCmd = &cobra.Command{
// 	Use:   "decrypt",
// 	Short: "Decrypt a file that was encrypted using AES-256-ECB",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		logger := utils.GetLogger("fileutil")

// 		if fileUtilFlags.decryptFile == "" {
// 			logger.Error().Msg("No input file specified")
// 			fmt.Println(utils.OutError("Error: No input file specified"))
// 			cmd.Help()
// 			os.Exit(1)
// 		}

// 		if fileUtilFlags.password == "" {
// 			logger.Error().Msg("No password specified")
// 			fmt.Println(utils.OutError("Error: No password specified"))
// 			cmd.Help()
// 			os.Exit(1)
// 		}

// 		err := fileutil.DecryptFile(fileUtilFlags.decryptFile, fileUtilFlags.password)
// 		if err != nil {
// 			logger.Error().Err(err).Msg("Failed to decrypt file")
// 			fmt.Println(utils.OutError("Error: " + err.Error()))
// 			os.Exit(1)
// 		}

// 		outputFile := fileUtilFlags.decryptFile
// 		if len(outputFile) > 4 && outputFile[len(outputFile)-4:] == ".enc" {
// 			outputFile = outputFile[:len(outputFile)-4]
// 		}

// 		fmt.Println(utils.OutSuccess("File decrypted successfully: " + outputFile))
// 	},
// }

// func init() {
// 	// Add fileUtil command to root command
// 	rootCmd.AddCommand(fileUtilCmd)

// 	// Setup encrypt command
// 	encryptCmd.Flags().StringVarP(&fileUtilFlags.encryptFile, "file", "f", "", "File to encrypt")
// 	encryptCmd.Flags().StringVarP(&fileUtilFlags.password, "password", "p", "", "Encryption password")
// 	fileUtilCmd.AddCommand(encryptCmd)

// 	// Setup decrypt command
// 	decryptCmd.Flags().StringVarP(&fileUtilFlags.decryptFile, "file", "f", "", "File to decrypt")
// 	decryptCmd.Flags().StringVarP(&fileUtilFlags.password, "password", "p", "", "Decryption password")
// 	fileUtilCmd.AddCommand(decryptCmd)
// }
