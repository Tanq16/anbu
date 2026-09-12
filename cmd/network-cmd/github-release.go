package networkCmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/download"
	"github.com/tanq16/anbu/internal/ghrelease"
	u "github.com/tanq16/anbu/utils"
)

var ghReleaseFlags struct {
	output    string
	proxy     string
	userAgent string
	headers   []string
	token     string
	asset     string
	manual    bool
}

var GitHubReleaseCmd = &cobra.Command{
	Use:     "github-release <owner/repo or url>",
	Aliases: []string{"ghr", "ghrelease"},
	Short:   "Download a GitHub release asset",
	Args:    cobra.ExactArgs(1),
	Run:     runGitHubRelease,
}

func runGitHubRelease(cmd *cobra.Command, args []string) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	owner, repo, err := ghrelease.ParseRepo(args[0])
	if err != nil {
		u.PrintFatal("invalid GitHub repository", err)
	}

	headers := download.ParseHeaders(ghReleaseFlags.headers)
	if ghReleaseFlags.token != "" {
		if headers == nil {
			headers = map[string]string{}
		}
		headers["Authorization"] = "Bearer " + ghReleaseFlags.token
	}
	client, err := download.NewClient(download.ClientConfig{
		ProxyURL:  ghReleaseFlags.proxy,
		UserAgent: ghReleaseFlags.userAgent,
		Headers:   headers,
	})
	if err != nil {
		u.PrintFatal("failed to create HTTP client", err)
	}

	u.PrintRunning("checking latest release")
	release, err := ghrelease.Latest(ctx, owner, repo, client)
	u.ClearLines(1)
	if err != nil {
		u.PrintFatal("failed to fetch release info", err)
	}

	asset, ok, err := pickAsset(release)
	if errors.Is(err, u.ErrNoTerminal) {
		u.PrintFatal("github-release --manual needs a terminal, or pass --asset", nil)
	}
	if err != nil {
		u.PrintFatal("failed to select asset", err)
	}
	if !ok {
		return
	}

	output := ghReleaseFlags.output
	if output == "" {
		output = asset.Name
	}

	cfg := download.Config{
		URL:         asset.URL,
		OutputPath:  output,
		Connections: 1,
		ProxyURL:    ghReleaseFlags.proxy,
		UserAgent:   ghReleaseFlags.userAgent,
		Headers:     headers,
	}
	plan, dlClient, err := download.Prepare(ctx, cfg)
	if download.AlreadyComplete(err) {
		u.PrintSuccess(download.AlreadyCompletePath(err) + " already exists")
		return
	}
	if err != nil {
		u.PrintFatal("failed to start download", err)
	}

	m := u.NewMeter("Downloading", filepath.Base(plan.OutputPath), plan.Size, u.UnitBytes)
	if err := plan.Execute(ctx, dlClient, m); err != nil {
		m.Fail(err)
		u.PrintFatal("download failed", err)
	}
	m.Done()
}

func pickAsset(release ghrelease.Release) (ghrelease.Asset, bool, error) {
	if ghReleaseFlags.asset != "" {
		asset, ok := ghrelease.FindAsset(release.Assets, ghReleaseFlags.asset)
		if !ok {
			return ghrelease.Asset{}, false, fmt.Errorf("no asset matching %q", ghReleaseFlags.asset)
		}
		return asset, true, nil
	}
	if !ghReleaseFlags.manual {
		asset, ok := ghrelease.SelectForPlatform(release.Assets)
		if !ok {
			return ghrelease.Asset{}, false, fmt.Errorf("could not automatically select asset for platform %s/%s, use --manual or --asset", runtime.GOOS, runtime.GOARCH)
		}
		return asset, true, nil
	}

	options := make([]string, len(release.Assets))
	for i, asset := range release.Assets {
		options[i] = fmt.Sprintf("%s (%s)", asset.Name, formatAssetSize(asset.Size))
	}
	idx, err := u.PromptSelect(fmt.Sprintf("Release %s", release.Tag), options)
	if err != nil {
		return ghrelease.Asset{}, false, err
	}
	if idx < 0 {
		return ghrelease.Asset{}, false, nil
	}
	return release.Assets[idx], true, nil
}

func formatAssetSize(n int64) string {
	const k = 1024.0
	v := float64(n)
	switch {
	case v < k:
		return fmt.Sprintf("%d B", n)
	case v < k*k:
		return fmt.Sprintf("%.1f KB", v/k)
	case v < k*k*k:
		return fmt.Sprintf("%.1f MB", v/(k*k))
	default:
		return fmt.Sprintf("%.1f GB", v/(k*k*k))
	}
}

func init() {
	GitHubReleaseCmd.Flags().StringVarP(&ghReleaseFlags.output, "output", "o", "", "Output file path")
	GitHubReleaseCmd.Flags().StringVarP(&ghReleaseFlags.proxy, "proxy", "p", "", "HTTP/HTTPS proxy URL")
	GitHubReleaseCmd.Flags().StringVarP(&ghReleaseFlags.userAgent, "user-agent", "a", "Anbu-CLI", "User agent")
	GitHubReleaseCmd.Flags().StringArrayVarP(&ghReleaseFlags.headers, "header", "H", nil, "Custom header as Key: Value (repeatable)")
	GitHubReleaseCmd.Flags().StringVar(&ghReleaseFlags.token, "token", os.Getenv("GITHUB_TOKEN"), "GitHub token (or GITHUB_TOKEN env)")
	GitHubReleaseCmd.Flags().StringVar(&ghReleaseFlags.asset, "asset", "", "Release asset name to download")
	GitHubReleaseCmd.Flags().BoolVar(&ghReleaseFlags.manual, "manual", false, "Select the release asset interactively")
	GitHubReleaseCmd.MarkFlagsMutuallyExclusive("manual", "asset")
}
