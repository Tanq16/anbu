package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	minMultiChunk = 1 * 1024 * 1024
	maxRetries    = 5
	tempDirName   = ".anbu-temp"
)

var errRangeNotSupported = errors.New("range requests are not supported")

var filenameUnsafe = regexp.MustCompile(`[^a-zA-Z0-9_\-\. ]+`)

type Config struct {
	URL         string
	OutputPath  string
	Connections int
	ProxyURL    string
	UserAgent   string
	Headers     map[string]string
}

type fileInfo struct {
	size           int64
	name           string
	rangeSupported bool
}

type Plan struct {
	URL         string
	OutputPath  string
	Size        int64
	Connections int
	useSimple   bool
}

func Prepare(ctx context.Context, cfg Config) (Plan, *Client, error) {
	parsed, err := url.Parse(cfg.URL)
	if err != nil {
		return Plan{}, nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return Plan{}, nil, fmt.Errorf("unsupported scheme: %s", parsed.Scheme)
	}
	if cfg.Connections < 1 {
		cfg.Connections = 1
	}

	client, err := NewClient(ClientConfig{
		ProxyURL:    cfg.ProxyURL,
		UserAgent:   cfg.UserAgent,
		Headers:     cfg.Headers,
		Connections: cfg.Connections,
	})
	if err != nil {
		return Plan{}, nil, err
	}

	headBlocked, finalURL, err := probeURL(ctx, client, cfg.URL)
	if err != nil {
		return Plan{}, nil, err
	}
	cfg.URL = finalURL

	info, err := statFile(ctx, client, cfg.URL, headBlocked)
	if err != nil && !errors.Is(err, errRangeNotSupported) {
		return Plan{}, nil, fmt.Errorf("error getting file info: %w", err)
	}

	if cfg.OutputPath == "" {
		cfg.OutputPath = info.name
		if cfg.OutputPath == "" {
			if pu, parseErr := url.Parse(cfg.URL); parseErr == nil {
				pathParts := strings.Split(pu.Path, "/")
				cfg.OutputPath = pathParts[len(pathParts)-1]
			}
		}
		if cfg.OutputPath == "" {
			cfg.OutputPath = "download"
		}
	}

	if existing, statErr := os.Stat(cfg.OutputPath); statErr == nil {
		if info.size > 0 && existing.Size() == info.size {
			return Plan{OutputPath: cfg.OutputPath}, nil, errAlreadyComplete{path: cfg.OutputPath}
		}
		cfg.OutputPath = renewOutputPath(cfg.OutputPath)
	}

	useSimple := !info.rangeSupported || cfg.Connections == 1 || (info.size > 0 && info.size/int64(cfg.Connections) < minMultiChunk)
	return Plan{
		URL:         cfg.URL,
		OutputPath:  cfg.OutputPath,
		Size:        info.size,
		Connections: cfg.Connections,
		useSimple:   useSimple,
	}, client, nil
}

func (p Plan) Execute(ctx context.Context, client *Client, progress io.Writer) error {
	if p.useSimple {
		return simple(ctx, p.URL, p.OutputPath, client, progress)
	}
	return multi(ctx, p.URL, p.OutputPath, p.Connections, client, p.Size, progress)
}

type errAlreadyComplete struct {
	path string
}

func (e errAlreadyComplete) Error() string {
	return e.path + " already exists"
}

func AlreadyComplete(err error) bool {
	var c errAlreadyComplete
	return errors.As(err, &c)
}

func AlreadyCompletePath(err error) string {
	var c errAlreadyComplete
	if errors.As(err, &c) {
		return c.path
	}
	return ""
}

func probeURL(ctx context.Context, client *Client, link string) (bool, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, link, nil)
	if err != nil {
		return false, "", fmt.Errorf("error creating request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("error checking URL: %w", err)
	}
	resp.Body.Close()

	if resp.Request != nil && resp.Request.URL != nil {
		link = resp.Request.URL.String()
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, "", fmt.Errorf("URL not found (404)")
	}
	if resp.StatusCode < 400 {
		return false, link, nil
	}

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return false, "", fmt.Errorf("server returned error: %d", resp.StatusCode)
	}
	getReq.Header.Set("Range", "bytes=0-0")
	getResp, err := client.Do(getReq)
	if err != nil {
		return false, "", fmt.Errorf("server returned error: %d", resp.StatusCode)
	}
	getResp.Body.Close()
	if getResp.StatusCode >= 400 {
		return false, "", fmt.Errorf("server returned error: %d (GET fallback returned: %d)", resp.StatusCode, getResp.StatusCode)
	}
	if getResp.Request != nil && getResp.Request.URL != nil {
		link = getResp.Request.URL.String()
	}
	return true, link, nil
}

func statFile(ctx context.Context, client *Client, link string, useGET bool) (fileInfo, error) {
	var resp *http.Response
	var err error
	if useGET {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
		if reqErr != nil {
			return fileInfo{}, reqErr
		}
		req.Header.Set("Range", "bytes=0-0")
		resp, err = client.Do(req)
	} else {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodHead, link, nil)
		if reqErr != nil {
			return fileInfo{}, reqErr
		}
		resp, err = client.Do(req)
	}
	if err != nil {
		return fileInfo{}, err
	}
	defer resp.Body.Close()

	info := fileInfo{name: filenameFromDisposition(resp.Header.Get("Content-Disposition"))}
	acceptRanges := resp.Header.Get("Accept-Ranges")
	info.rangeSupported = acceptRanges == "bytes" || (useGET && resp.StatusCode == http.StatusPartialContent)

	if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
		if _, after, found := strings.Cut(contentRange, "/"); found && after != "*" {
			size, parseErr := strconv.ParseInt(after, 10, 64)
			if parseErr != nil {
				return info, fmt.Errorf("invalid Content-Range total: %w", parseErr)
			}
			info.size = size
		}
	}
	if info.size <= 0 {
		contentLength := resp.Header.Get("Content-Length")
		if contentLength != "" {
			size, parseErr := strconv.ParseInt(contentLength, 10, 64)
			if parseErr != nil {
				return info, parseErr
			}
			info.size = size
		}
	}
	if !info.rangeSupported {
		return info, errRangeNotSupported
	}
	if info.size <= 0 {
		return info, errors.New("invalid file size reported by server")
	}
	return info, nil
}

func filenameFromDisposition(header string) string {
	if header == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(header)
	if err != nil {
		return ""
	}
	if fn := params["filename"]; fn != "" {
		return filenameUnsafe.ReplaceAllString(fn, "_")
	}
	fn := params["filename*"]
	if fn == "" {
		return ""
	}
	if rest, ok := strings.CutPrefix(fn, "UTF-8''"); ok {
		unescaped, _ := url.PathUnescape(rest)
		return filenameUnsafe.ReplaceAllString(unescaped, "_")
	}
	return ""
}

func renewOutputPath(outputPath string) string {
	dir := filepath.Dir(outputPath)
	base := filepath.Base(outputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	for index := 1; ; index++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s-(%d)%s", name, index, ext))
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
}

func addProgress(w io.Writer, n int64) {
	if w == nil || n == 0 {
		return
	}
	type adder interface{ Add(int64) }
	if a, ok := w.(adder); ok {
		a.Add(n)
	}
}

func removeTempDirIfEmpty(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}
	os.Remove(dir)
}

func ParseHeaders(headers []string) map[string]string {
	result := make(map[string]string)
	for _, header := range headers {
		key, value, ok := strings.Cut(header, ":")
		if !ok {
			continue
		}
		result[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return result
}
