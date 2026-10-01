package awscred

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	oidctypes "github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
	"github.com/tanq16/anbu/internal/vault"
	"golang.org/x/sync/errgroup"
)

const (
	StateIdle    = "idle"
	StatePending = "pending"
	StateActive  = "active"
	StateError   = "error"

	deviceGrantType = "urn:ietf:params:oauth:grant-type:device_code"
	credentialSlack = 5 * time.Minute
)

type ssoSession struct {
	clientID, clientSecret string
	clientExpires          time.Time
	clientRegion           string
	accessToken            string
	tokenExpires           time.Time
	state                  string
	lastErr                error
	verificationURI        string
	userCode               string
	deviceExpires          time.Time
	login                  int
	cancelPoll             context.CancelFunc
	roleCreds              map[string]aws.Credentials
}

type LoginPrompt struct {
	VerificationURI string `json:"verification_uri"`
	UserCode        string `json:"user_code"`
	ExpiresIn       int32  `json:"expires_in"`
}

type SessionStatus struct {
	State           string    `json:"state"`
	ExpiresAt       time.Time `json:"expires_at,omitzero"`
	Message         string    `json:"message,omitempty"`
	VerificationURI string    `json:"verification_uri,omitempty"`
	UserCode        string    `json:"user_code,omitempty"`
}

type Account struct {
	AccountID   string   `json:"account_id"`
	AccountName string   `json:"account_name"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
}

func ssoSecret(sec vault.Secret) error {
	if sec.Type != vault.TypeAWSSSO {
		return fmt.Errorf("%w: %s has type %s, not %s", ErrInvalidRef, sec.Name, sec.Type, vault.TypeAWSSSO)
	}
	return nil
}

func (r *Resolver) session(secretID string) *ssoSession {
	s, ok := r.sessions[secretID]
	if !ok {
		s = &ssoSession{state: StateIdle, roleCreds: map[string]aws.Credentials{}}
		r.sessions[secretID] = s
	}
	return s
}

func (r *Resolver) StartLogin(ctx context.Context, sec vault.Secret) (LoginPrompt, error) {
	if err := ssoSecret(sec); err != nil {
		return LoginPrompt{}, err
	}
	region := sec.Fields["sso_region"]
	oidc := ssooidc.New(ssooidc.Options{Region: region})

	r.mu.Lock()
	s := r.session(sec.ID)
	clientID, clientSecret := s.clientID, s.clientSecret
	reuse := clientID != "" && s.clientRegion == region && time.Until(s.clientExpires) > 15*time.Minute
	r.mu.Unlock()

	if !reuse {
		reg, err := oidc.RegisterClient(ctx, &ssooidc.RegisterClientInput{
			ClientName: aws.String("anbu"),
			ClientType: aws.String("public"),
		})
		if err != nil {
			return LoginPrompt{}, err
		}
		clientID, clientSecret = aws.ToString(reg.ClientId), aws.ToString(reg.ClientSecret)
		r.mu.Lock()
		s.clientID, s.clientSecret = clientID, clientSecret
		s.clientExpires = time.Unix(reg.ClientSecretExpiresAt, 0)
		s.clientRegion = region
		r.mu.Unlock()
	}

	auth, err := oidc.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     aws.String(clientID),
		ClientSecret: aws.String(clientSecret),
		StartUrl:     aws.String(sec.Fields["start_url"]),
	})
	if err != nil {
		return LoginPrompt{}, err
	}

	expires := time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)
	pollCtx, cancel := context.WithDeadline(context.WithoutCancel(ctx), expires)
	r.mu.Lock()
	if s.cancelPoll != nil {
		s.cancelPoll()
	}
	s.login++
	s.cancelPoll = cancel
	s.state, s.lastErr = StatePending, nil
	s.verificationURI = aws.ToString(auth.VerificationUriComplete)
	s.userCode = aws.ToString(auth.UserCode)
	s.deviceExpires = expires
	login := s.login
	r.mu.Unlock()

	go r.pollToken(pollCtx, cancel, sec.ID, login, oidc, clientID, clientSecret, auth)
	return LoginPrompt{
		VerificationURI: aws.ToString(auth.VerificationUriComplete),
		UserCode:        aws.ToString(auth.UserCode),
		ExpiresIn:       auth.ExpiresIn,
	}, nil
}

func (r *Resolver) pollToken(ctx context.Context, cancel context.CancelFunc, secretID string, login int,
	oidc *ssooidc.Client, clientID, clientSecret string, auth *ssooidc.StartDeviceAuthorizationOutput) {
	defer cancel()
	interval := time.Duration(max(auth.Interval, 1)) * time.Second
	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				r.finishLogin(secretID, login, "", 0, errors.New("device code expired"))
			}
			return
		case <-time.After(interval):
		}
		out, err := oidc.CreateToken(ctx, &ssooidc.CreateTokenInput{
			ClientId:     aws.String(clientID),
			ClientSecret: aws.String(clientSecret),
			DeviceCode:   auth.DeviceCode,
			GrantType:    aws.String(deviceGrantType),
		})
		if _, ok := errors.AsType[*oidctypes.AuthorizationPendingException](err); ok {
			continue
		}
		if _, ok := errors.AsType[*oidctypes.SlowDownException](err); ok {
			interval += 5 * time.Second
			continue
		}
		if _, ok := errors.AsType[*oidctypes.ExpiredTokenException](err); ok {
			r.finishLogin(secretID, login, "", 0, errors.New("device code expired"))
			return
		}
		if err != nil {
			if ctx.Err() != nil {
				continue
			}
			r.finishLogin(secretID, login, "", 0, err)
			return
		}
		r.finishLogin(secretID, login, aws.ToString(out.AccessToken), out.ExpiresIn, nil)
		return
	}
}

func (r *Resolver) finishLogin(secretID string, login int, token string, expiresIn int32, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.session(secretID)
	if s.login != login {
		return
	}
	s.cancelPoll = nil
	s.verificationURI, s.userCode, s.deviceExpires = "", "", time.Time{}
	if err != nil {
		s.state, s.lastErr = StateError, err
		return
	}
	s.state, s.lastErr = StateActive, nil
	s.accessToken = token
	s.tokenExpires = time.Now().Add(time.Duration(expiresIn) * time.Second)
	clear(s.roleCreds)
}

func (r *Resolver) Status(sec vault.Secret) (SessionStatus, error) {
	if err := ssoSecret(sec); err != nil {
		return SessionStatus{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.session(sec.ID)
	switch s.state {
	case StatePending:
		return SessionStatus{State: StatePending, ExpiresAt: s.deviceExpires.UTC(), VerificationURI: s.verificationURI, UserCode: s.userCode}, nil
	case StateError:
		return SessionStatus{State: StateError, Message: s.lastErr.Error()}, nil
	case StateActive:
		if time.Now().Before(s.tokenExpires) {
			return SessionStatus{State: StateActive, ExpiresAt: s.tokenExpires.UTC()}, nil
		}
	}
	return SessionStatus{State: StateIdle}, nil
}

func (r *Resolver) accessToken(secretID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[secretID]
	if !ok || s.accessToken == "" || !time.Now().Before(s.tokenExpires) {
		return "", ErrLoginRequired
	}
	return s.accessToken, nil
}

func (r *Resolver) expireSession(secretID, token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.sessions[secretID]; ok && s.accessToken == token {
		s.accessToken, s.tokenExpires, s.state = "", time.Time{}, StateIdle
		clear(s.roleCreds)
	}
}

func (r *Resolver) Accounts(ctx context.Context, sec vault.Secret) ([]Account, error) {
	if err := ssoSecret(sec); err != nil {
		return nil, err
	}
	token, err := r.accessToken(sec.ID)
	if err != nil {
		return nil, err
	}
	client := sso.New(sso.Options{Region: sec.Fields["sso_region"]})
	var accounts []Account
	pager := sso.NewListAccountsPaginator(client, &sso.ListAccountsInput{AccessToken: aws.String(token)})
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, r.ssoErr(sec.ID, token, err)
		}
		for _, a := range page.AccountList {
			accounts = append(accounts, Account{
				AccountID:   aws.ToString(a.AccountId),
				AccountName: aws.ToString(a.AccountName),
				Email:       aws.ToString(a.EmailAddress),
				Roles:       []string{},
			})
		}
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(8)
	for i := range accounts {
		g.Go(func() error {
			roles := sso.NewListAccountRolesPaginator(client, &sso.ListAccountRolesInput{
				AccessToken: aws.String(token),
				AccountId:   aws.String(accounts[i].AccountID),
			})
			for roles.HasMorePages() {
				page, err := roles.NextPage(gctx)
				if err != nil {
					return err
				}
				for _, role := range page.RoleList {
					accounts[i].Roles = append(accounts[i].Roles, aws.ToString(role.RoleName))
				}
			}
			slices.Sort(accounts[i].Roles)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, r.ssoErr(sec.ID, token, err)
	}
	slices.SortFunc(accounts, func(a, b Account) int {
		return cmp.Or(cmp.Compare(a.AccountName, b.AccountName), cmp.Compare(a.AccountID, b.AccountID))
	})
	return accounts, nil
}

func (r *Resolver) ssoErr(secretID, token string, err error) error {
	if _, ok := errors.AsType[*ssotypes.UnauthorizedException](err); ok {
		r.expireSession(secretID, token)
		return ErrLoginRequired
	}
	return err
}

type ssoProvider struct {
	resolver  *Resolver
	secretID  string
	ssoRegion string
	accountID string
	roleName  string
}

func (p *ssoProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	r := p.resolver
	token, err := r.accessToken(p.secretID)
	if err != nil {
		return aws.Credentials{}, err
	}
	key := p.accountID + "/" + p.roleName
	r.mu.Lock()
	cached, ok := r.sessions[p.secretID].roleCreds[key]
	r.mu.Unlock()
	if ok && time.Until(cached.Expires) > credentialSlack {
		return cached, nil
	}

	client := sso.New(sso.Options{Region: p.ssoRegion})
	out, err := client.GetRoleCredentials(ctx, &sso.GetRoleCredentialsInput{
		AccessToken: aws.String(token),
		AccountId:   aws.String(p.accountID),
		RoleName:    aws.String(p.roleName),
	})
	if err != nil {
		return aws.Credentials{}, r.ssoErr(p.secretID, token, err)
	}
	rc := out.RoleCredentials
	if rc == nil {
		return aws.Credentials{}, errors.New("sso:GetRoleCredentials returned no credentials")
	}
	creds := aws.Credentials{
		AccessKeyID:     aws.ToString(rc.AccessKeyId),
		SecretAccessKey: aws.ToString(rc.SecretAccessKey),
		SessionToken:    aws.ToString(rc.SessionToken),
		Source:          "anbu-sso",
		CanExpire:       true,
		Expires:         time.Unix(0, rc.Expiration*int64(time.Millisecond)),
		AccountID:       p.accountID,
	}
	r.mu.Lock()
	if s, ok := r.sessions[p.secretID]; ok && s.accessToken == token {
		s.roleCreds[key] = creds
	}
	r.mu.Unlock()
	return creds, nil
}
