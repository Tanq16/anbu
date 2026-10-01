package apiCmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/apiclient"
)

var setupFlags struct {
	headers []string
}

var setupCmd = &cobra.Command{
	Use:   "setup <url>",
	Short: "Save the server URL and headers that every api call uses",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := apiclient.Config{URL: args[0], Headers: map[string]string{}}
		for _, h := range setupFlags.headers {
			key, value, err := apiclient.ParseHeader(h)
			if err != nil {
				log.Fatal().Err(err).Msg("invalid header")
			}
			cfg.Headers[key] = value
		}
		path, err := apiclient.Save(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to save api config")
		}
		log.Info().Str("path", path).Str("url", cfg.URL).Int("headers", len(cfg.Headers)).Msg("api config saved")
	},
}

func init() {
	ApiCmd.AddCommand(setupCmd)
	setupCmd.Flags().StringArrayVarP(&setupFlags.headers, "header", "H", nil, `Header sent on every call, as "Key: Value" (repeatable)`)
}
