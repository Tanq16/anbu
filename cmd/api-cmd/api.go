package apiCmd

import (
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/apiclient"
)

var ApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Call the anbu REST API and print the raw response",
}

func call(method, path string, body any) {
	cfg, err := apiclient.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load api config")
	}
	resp, err := apiclient.Do(cfg, method, path, body)
	if err != nil {
		log.Fatal().Err(err).Msg("request failed")
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read response")
	}
	if resp.StatusCode >= 400 {
		log.Fatal().Int("status", resp.StatusCode).Str("body", strings.TrimSpace(string(data))).Msg("server returned an error")
	}
	os.Stdout.Write(data)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		os.Stdout.Write([]byte("\n"))
	}
}

func segment(s string) string {
	return url.PathEscape(s)
}
