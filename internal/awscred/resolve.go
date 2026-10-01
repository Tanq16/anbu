package awscred

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/tanq16/anbu/internal/datadir"
	"github.com/tanq16/anbu/internal/vault"
)

var (
	ErrLoginRequired = errors.New("sso login required")
	ErrInvalidRef    = errors.New("invalid profile reference")
	ErrNotFound      = errors.New("aws profile not found")
)

const AdHocLabel = "ad-hoc"

type Source struct {
	Profile string  `json:"profile,omitempty"`
	Inline  *Inline `json:"inline,omitempty"`
}

type Inline struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token,omitempty"`
	Region          string `json:"region"`
}

func (s Source) Label() string {
	if s.Inline != nil {
		return AdHocLabel
	}
	return s.Profile
}

type Profile struct {
	Ref    string
	Secret vault.Secret
	SSO    *vault.AWSProfile
	Region string
}

type Resolver struct {
	vault    *vault.Store
	dataDir  string
	mu       sync.Mutex
	sessions map[string]*ssoSession
}

func NewResolver(v *vault.Store, dataDir string) *Resolver {
	return &Resolver{vault: v, dataDir: dataDir, sessions: map[string]*ssoSession{}}
}

func (r *Resolver) Lookup(ref string) (Profile, error) {
	secretRef, profileName, isSSO := strings.Cut(ref, ":")
	if secretRef == "" || (isSSO && profileName == "") {
		return Profile{}, fmt.Errorf("%w %q", ErrInvalidRef, ref)
	}
	sec, err := r.vault.Get(secretRef)
	if errors.Is(err, vault.ErrNotFound) {
		return Profile{}, fmt.Errorf("%w: no secret named %s", ErrNotFound, secretRef)
	}
	if err != nil {
		return Profile{}, err
	}
	if !isSSO {
		if sec.Type != vault.TypeAWSStatic {
			return Profile{}, fmt.Errorf("%w: %s has type %s, not %s", ErrInvalidRef, sec.Name, sec.Type, vault.TypeAWSStatic)
		}
		return Profile{Ref: sec.Name, Secret: sec, Region: sec.Fields["region"]}, nil
	}
	if sec.Type != vault.TypeAWSSSO {
		return Profile{}, fmt.Errorf("%w: %s has type %s, not %s", ErrInvalidRef, sec.Name, sec.Type, vault.TypeAWSSSO)
	}
	p, ok := sec.Profile(profileName)
	if !ok {
		return Profile{}, fmt.Errorf("%w: %s has no profile %s", ErrNotFound, sec.Name, profileName)
	}
	return Profile{Ref: sec.Name + ":" + p.Name, Secret: sec, SSO: &p, Region: p.Region}, nil
}

func (r *Resolver) Profiles() []Profile {
	var out []Profile
	for _, sec := range r.vault.List() {
		switch sec.Type {
		case vault.TypeAWSStatic:
			out = append(out, Profile{Ref: sec.Name, Secret: sec, Region: sec.Fields["region"]})
		case vault.TypeAWSSSO:
			for _, p := range sec.Profiles {
				out = append(out, Profile{Ref: sec.Name + ":" + p.Name, Secret: sec, SSO: &p, Region: p.Region})
			}
		}
	}
	return out
}

func (r *Resolver) provider(src Source) (aws.CredentialsProvider, string, error) {
	if src.Inline != nil {
		in := src.Inline
		if in.AccessKeyID == "" || in.SecretAccessKey == "" || in.Region == "" {
			return nil, "", fmt.Errorf("%w: ad-hoc keys need access_key_id, secret_access_key, and region", ErrInvalidRef)
		}
		return credentials.NewStaticCredentialsProvider(in.AccessKeyID, in.SecretAccessKey, in.SessionToken), in.Region, nil
	}
	if src.Profile == "" {
		return nil, "", fmt.Errorf("%w: a profile or ad-hoc keys are required", ErrInvalidRef)
	}
	p, err := r.Lookup(src.Profile)
	if err != nil {
		return nil, "", err
	}
	if p.SSO == nil {
		f := p.Secret.Fields
		return credentials.NewStaticCredentialsProvider(f["access_key_id"], f["secret_access_key"], f["session_token"]), p.Region, nil
	}
	return &ssoProvider{
		resolver:  r,
		secretID:  p.Secret.ID,
		ssoRegion: p.Secret.Fields["sso_region"],
		accountID: p.SSO.AccountID,
		roleName:  p.SSO.RoleName,
	}, p.Region, nil
}

func (r *Resolver) Validate(src Source) error {
	_, _, err := r.provider(src)
	return err
}

func (r *Resolver) Config(ctx context.Context, src Source) (aws.Config, error) {
	provider, region, err := r.provider(src)
	if err != nil {
		return aws.Config{}, err
	}
	return config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.NewCredentialsCache(provider)),
		config.WithSharedConfigFiles([]string{filepath.Join(r.dataDir, datadir.AWSConfig)}),
		config.WithSharedCredentialsFiles([]string{filepath.Join(r.dataDir, datadir.AWSCredentials)}),
	)
}

func (r *Resolver) Credentials(ctx context.Context, src Source) (aws.Credentials, string, error) {
	provider, region, err := r.provider(src)
	if err != nil {
		return aws.Credentials{}, "", err
	}
	creds, err := provider.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, "", err
	}
	return creds, region, nil
}
