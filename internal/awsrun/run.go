package awsrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var (
	ErrNotInstalled = errors.New("aws CLI not installed on the server")
	ErrInvalid      = errors.New("invalid command")
)

type Result struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

func Installed() error {
	if _, err := exec.LookPath("aws"); err != nil {
		return ErrNotInstalled
	}
	return nil
}

func Run(ctx context.Context, dataDir string, creds aws.Credentials, region, command string) (Result, error) {
	args, err := Split(strings.TrimPrefix(strings.TrimSpace(command), "aws "))
	if err != nil {
		return Result{}, err
	}
	if len(args) == 0 {
		return Result{}, fmt.Errorf("%w: command is empty", ErrInvalid)
	}
	if err := Installed(); err != nil {
		return Result{}, err
	}
	if !slices.Contains(args, "--output") {
		args = append(args, "--output", "json")
	}
	awsDir := filepath.Join(dataDir, "aws")
	cmd := exec.CommandContext(ctx, "aws", args...)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + filepath.Join(awsDir, "home"),
		"AWS_CONFIG_FILE=" + filepath.Join(awsDir, "config"),
		"AWS_SHARED_CREDENTIALS_FILE=" + filepath.Join(awsDir, "credentials"),
		"AWS_ACCESS_KEY_ID=" + creds.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + creds.SecretAccessKey,
		"AWS_SESSION_TOKEN=" + creds.SessionToken,
		"AWS_REGION=" + region,
		"AWS_DEFAULT_REGION=" + region,
		"AWS_PAGER=",
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	res := Result{Stdout: stdout.String(), Stderr: stderr.String()}
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		res.ExitCode = exitErr.ExitCode()
		return res, nil
	}
	return res, err
}

func Split(s string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inArg := false
	var quote rune
	escaped := false
	for _, r := range s {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case quote == '\'':
			if r == '\'' {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\\':
			escaped, inArg = true, true
		case quote == '"':
			if r == '"' {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote, inArg = r, true
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			if inArg {
				args = append(args, cur.String())
				cur.Reset()
				inArg = false
			}
		default:
			cur.WriteRune(r)
			inArg = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("%w: unterminated quote", ErrInvalid)
	}
	if escaped {
		return nil, fmt.Errorf("%w: trailing backslash", ErrInvalid)
	}
	if inArg {
		args = append(args, cur.String())
	}
	return args, nil
}
