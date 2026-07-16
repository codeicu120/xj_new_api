package respond

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"xj_comp/internal/config"
)

const (
	verifierMD5FormConcatSortedValues = "md5_form_concat_sorted_values"
	verifierMD5FormQuerySecretSuffix  = "md5_form_query_secret_suffix"
)

var (
	ErrMissingConfig      = errors.New("respond provider config missing")
	ErrInvalidSignature   = errors.New("respond provider signature invalid")
	ErrAccountingDisabled = errors.New("respond accounting disabled")
)

type CallbackRequest struct {
	Action string
	Form   url.Values
	Raw    []byte
}

type VerifiedPayment struct {
	PayID       int
	OutTradeID  string
	PayAmount   int
	ProviderRaw url.Values
}

type VerificationResult struct {
	Echo   string
	Reason error
}

type Verifier interface {
	Verify(ctx context.Context, req CallbackRequest) (VerifiedPayment, error)
}

type Provider struct {
	Action            string
	EchoOK            string
	EchoErr           string
	AccountingEnabled bool
	Verifier          Verifier
}

type Registry struct {
	providers map[string]Provider
}

func NewRegistry(providers ...Provider) *Registry {
	registry := &Registry{providers: make(map[string]Provider, len(providers))}
	for _, provider := range providers {
		provider.Action = strings.TrimSpace(provider.Action)
		if provider.Action == "" {
			continue
		}
		registry.providers[provider.Action] = provider
	}
	return registry
}

func NewRegistryFromConfig(configs []config.RespondProviderConfig) *Registry {
	providers := make([]Provider, 0, len(configs))
	for _, cfg := range configs {
		provider := providerFromConfig(cfg)
		if provider.Action != "" {
			providers = append(providers, provider)
		}
	}
	return NewRegistry(providers...)
}

func (r *Registry) Provider(action string) (Provider, bool) {
	if r == nil {
		return Provider{}, false
	}
	provider, ok := r.providers[action]
	return provider, ok
}

func providerFromConfig(cfg config.RespondProviderConfig) Provider {
	provider := Provider{
		Action:            strings.TrimSpace(cfg.Action),
		EchoOK:            cfg.EchoOK,
		EchoErr:           cfg.EchoErr,
		AccountingEnabled: cfg.AccountingEnabled,
	}
	if provider.EchoErr == "" {
		provider.EchoErr = "failed"
	}
	if provider.EchoOK == "" {
		provider.EchoOK = "success"
	}
	switch cfg.Verifier {
	case verifierMD5FormConcatSortedValues:
		provider.Verifier = MD5FormVerifier{
			Secret:          cfg.Secret,
			SignField:       cfg.SignField,
			StatusField:     cfg.StatusField,
			SuccessStatus:   cfg.SuccessStatus,
			OrderIDField:    cfg.OrderIDField,
			OutTradeIDField: cfg.OutTradeIDField,
			AmountField:     cfg.AmountField,
		}
	case verifierMD5FormQuerySecretSuffix:
		provider.Verifier = MD5FormQuerySecretSuffixVerifier{
			Secret:          cfg.Secret,
			SignField:       cfg.SignField,
			StatusField:     cfg.StatusField,
			SuccessStatus:   cfg.SuccessStatus,
			OrderIDField:    cfg.OrderIDField,
			OutTradeIDField: cfg.OutTradeIDField,
			AmountField:     cfg.AmountField,
		}
	}
	return provider
}

type MD5FormVerifier struct {
	Secret          string
	SignField       string
	StatusField     string
	SuccessStatus   string
	OrderIDField    string
	OutTradeIDField string
	AmountField     string
}

func (v MD5FormVerifier) Verify(_ context.Context, req CallbackRequest) (VerifiedPayment, error) {
	if strings.TrimSpace(v.Secret) == "" || strings.TrimSpace(v.SignField) == "" {
		return VerifiedPayment{}, ErrMissingConfig
	}
	form := req.Form
	sign := form.Get(v.SignField)
	if sign == "" || !strings.EqualFold(v.sign(form), sign) {
		return VerifiedPayment{}, ErrInvalidSignature
	}
	if v.StatusField != "" && form.Get(v.StatusField) != v.SuccessStatus {
		return VerifiedPayment{}, ErrInvalidSignature
	}
	return VerifiedPayment{
		PayID:       atoiString(form.Get(v.OrderIDField)),
		OutTradeID:  form.Get(v.OutTradeIDField),
		PayAmount:   amountToCents(form.Get(v.AmountField)),
		ProviderRaw: form,
	}, nil
}

func (v MD5FormVerifier) sign(form url.Values) string {
	keys := make([]string, 0, len(form)+1)
	values := make(map[string]string, len(form)+1)
	for key, vals := range form {
		if key == v.SignField {
			continue
		}
		keys = append(keys, key)
		if len(vals) > 0 {
			values[key] = vals[0]
		}
	}
	keys = append(keys, "app_secret")
	values["app_secret"] = v.Secret
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(values[key])
	}
	sum := md5.Sum([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}

type MD5FormQuerySecretSuffixVerifier struct {
	Secret          string
	SignField       string
	StatusField     string
	SuccessStatus   string
	OrderIDField    string
	OutTradeIDField string
	AmountField     string
}

func (v MD5FormQuerySecretSuffixVerifier) Verify(_ context.Context, req CallbackRequest) (VerifiedPayment, error) {
	if strings.TrimSpace(v.Secret) == "" || strings.TrimSpace(v.SignField) == "" {
		return VerifiedPayment{}, ErrMissingConfig
	}
	form := req.Form
	sign := form.Get(v.SignField)
	if sign == "" || !strings.EqualFold(v.sign(form), sign) {
		return VerifiedPayment{}, ErrInvalidSignature
	}
	if v.StatusField != "" && form.Get(v.StatusField) != v.SuccessStatus {
		return VerifiedPayment{}, ErrInvalidSignature
	}
	return VerifiedPayment{
		PayID:       atoiString(form.Get(v.OrderIDField)),
		OutTradeID:  form.Get(v.OutTradeIDField),
		PayAmount:   amountToCents(form.Get(v.AmountField)),
		ProviderRaw: form,
	}, nil
}

func (v MD5FormQuerySecretSuffixVerifier) sign(form url.Values) string {
	keys := make([]string, 0, len(form))
	for key, vals := range form {
		if key == v.SignField || key == "sign_type" || len(vals) != 1 || vals[0] == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for i, key := range keys {
		if i > 0 {
			builder.WriteByte('&')
		}
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(form.Get(key))
	}
	builder.WriteString(v.Secret)
	sum := md5.Sum([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}

func atoiString(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func amountToCents(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if parsed, err := strconv.Atoi(value); err == nil {
		return parsed
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return int(math.Round(parsed * 100))
}
