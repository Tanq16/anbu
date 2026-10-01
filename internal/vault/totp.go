package vault

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"encoding/json/v2"
	"errors"
	"fmt"
	"hash"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TOTP struct {
	Secret    string `json:"secret"`
	Algorithm string `json:"algorithm"`
	Digits    int    `json:"digits"`
	Period    int    `json:"period"`
}

var algorithms = map[string]func() hash.Hash{
	"SHA1":   sha1.New,
	"SHA256": sha256.New,
	"SHA512": sha512.New,
}

var b32 = base32.StdEncoding.WithPadding(base32.NoPadding)

func (t *TOTP) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err == nil {
		if strings.TrimSpace(raw) == "" {
			*t = TOTP{}
			return nil
		}
		parsed, err := Parse(raw)
		if err != nil {
			return err
		}
		*t = parsed
		return nil
	}
	type plain TOTP
	var p plain
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	*t = TOTP(p)
	return nil
}

func Parse(input string) (TOTP, error) {
	input = strings.TrimSpace(input)
	t := TOTP{Algorithm: "SHA1", Digits: 6, Period: 30}
	if !strings.HasPrefix(strings.ToLower(input), "otpauth://") {
		t.Secret = input
		return t, t.normalize()
	}

	u, err := url.Parse(input)
	if err != nil {
		return TOTP{}, err
	}
	if !strings.EqualFold(u.Host, "totp") {
		return TOTP{}, fmt.Errorf("unsupported otpauth type %q, only totp is supported", u.Host)
	}
	q := u.Query()
	t.Secret = q.Get("secret")
	if a := q.Get("algorithm"); a != "" {
		t.Algorithm = a
	}
	for key, dst := range map[string]*int{"digits": &t.Digits, "period": &t.Period} {
		if v := q.Get(key); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil {
				return TOTP{}, fmt.Errorf("invalid %s %q", key, v)
			}
			*dst = n
		}
	}
	return t, t.normalize()
}

func (t *TOTP) normalize() error {
	t.Secret = strings.TrimRight(strings.ToUpper(strings.NewReplacer(" ", "", "-", "").Replace(t.Secret)), "=")
	if t.Secret == "" {
		return errors.New("secret is empty")
	}
	switch len(t.Secret) % 8 {
	case 1, 3, 6:
		return fmt.Errorf("secret has invalid base32 length %d", len(t.Secret))
	}
	if _, err := b32.DecodeString(t.Secret); err != nil {
		return fmt.Errorf("secret is not valid base32: %w", err)
	}
	t.Algorithm = strings.ToUpper(t.Algorithm)
	if _, ok := algorithms[t.Algorithm]; !ok {
		return fmt.Errorf("unsupported algorithm %q", t.Algorithm)
	}
	if t.Digits < 6 || t.Digits > 8 {
		return fmt.Errorf("digits must be between 6 and 8, got %d", t.Digits)
	}
	if t.Period <= 0 {
		return fmt.Errorf("period must be positive, got %d", t.Period)
	}
	return nil
}

func (t TOTP) Code(at time.Time) (string, error) {
	key, err := b32.DecodeString(t.Secret)
	if err != nil {
		return "", err
	}
	newHash, ok := algorithms[t.Algorithm]
	if !ok {
		return "", fmt.Errorf("unsupported algorithm %q", t.Algorithm)
	}
	counter := make([]byte, 8)
	binary.BigEndian.PutUint64(counter, uint64(at.Unix())/uint64(t.Period))
	mac := hmac.New(newHash, key)
	mac.Write(counter)
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	bin := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	mod := uint32(1)
	for range t.Digits {
		mod *= 10
	}
	return fmt.Sprintf("%0*d", t.Digits, bin%mod), nil
}
