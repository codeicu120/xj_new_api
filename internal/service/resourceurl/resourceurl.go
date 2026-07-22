package resourceurl

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type SettingsStore interface {
	SettingValue(ctx context.Context) (string, error)
}

type IPLocator interface {
	Find(ip string) ([]string, error)
}

type availabilityStore interface {
	Available() bool
}

type Request struct {
	HasCookieAuth bool
	ClientIP      string
}

type requestContextKey struct{}

type requestState struct {
	req      Request
	once     sync.Once
	resolved Resolved
	err      error
}

func WithRequest(ctx context.Context, req Request) context.Context {
	return context.WithValue(ctx, requestContextKey{}, &requestState{req: req})
}

func RequestFromContext(ctx context.Context) Request {
	state, _ := ctx.Value(requestContextKey{}).(*requestState)
	if state == nil {
		return Request{}
	}
	return state.req
}

type Resolver struct {
	store       SettingsStore
	locator     IPLocator
	fallbackURL string
	now         func() time.Time
}

func (r *Resolver) ResolveContext(ctx context.Context) (Resolved, error) {
	return r.Resolve(ctx, RequestFromContext(ctx))
}

type Resolved struct {
	BaseURL    string
	AuthSecret string
	Timestamp  int64
}

func NewResolver(store SettingsStore, locator IPLocator, fallbackURL string) *Resolver {
	return &Resolver{
		store:       store,
		locator:     locator,
		fallbackURL: fallbackURL,
		now:         time.Now,
	}
}

func (r *Resolver) Resolve(ctx context.Context, req Request) (Resolved, error) {
	if state, _ := ctx.Value(requestContextKey{}).(*requestState); state != nil && state.req == req {
		state.once.Do(func() { state.resolved, state.err = r.resolve(ctx, req) })
		return state.resolved, state.err
	}
	return r.resolve(ctx, req)
}

func (r *Resolver) resolve(ctx context.Context, req Request) (Resolved, error) {
	settings := map[string]string{}
	if r.store != nil {
		raw, err := r.store.SettingValue(ctx)
		if err != nil {
			return Resolved{}, fmt.Errorf("load resource URL settings: %w", err)
		}
		settings = parseSerializedStrings(raw)
	}

	baseURL := settings["resurl"]
	if req.HasCookieAuth {
		baseURL = settings["resurl_h5"]
		if r.inFreeArea(req.ClientIP, settings["resurl_h5_free_area"]) {
			baseURL = settings["resurl_h5_free"]
		}
	}
	// Legacy constructors without a settings store retain an explicit fallback for
	// isolated tests. Production injects the DB-backed store and preserves an empty
	// database value exactly as PHP runtime::resUrl does.
	storeUnavailable := r.store == nil
	if available, ok := r.store.(availabilityStore); ok && !available.Available() {
		storeUnavailable = true
	}
	if storeUnavailable && strings.TrimSpace(baseURL) == "" {
		baseURL = r.fallbackURL
	}

	now := r.now()
	hour := now.In(time.FixedZone("CST", 8*60*60)).Format("2006010215")
	return Resolved{
		BaseURL:    strings.ReplaceAll(baseURL, "{rand}", hour),
		AuthSecret: settings["resurl_auth"],
		Timestamp:  now.Unix(),
	}, nil
}

func (r *Resolver) inFreeArea(ip string, configured string) bool {
	if r.locator == nil || strings.TrimSpace(configured) == "" {
		return false
	}
	parts, err := r.locator.Find(ip)
	if err != nil {
		return false
	}
	address := strings.Join(parts, " ")
	for _, area := range strings.Split(configured, ",") {
		// PHP uses strpos($ipaddr, $loc) > 0 rather than >= 0.
		if area != "" && strings.Index(address, area) > 0 {
			return true
		}
	}
	return false
}

func (r Resolved) GetRes(uri string, prefix string) string {
	if uri == "" {
		return ""
	}
	fullURL := uri
	if !absoluteResourcePattern.MatchString(uri) {
		pathPrefix := ""
		if prefix != "" {
			pathPrefix = prefix + "/"
		}
		fullURL = r.BaseURL + "/" + pathPrefix + uri
	}
	if r.AuthSecret != "" {
		signURI := uri
		if !strings.HasPrefix(signURI, "/") {
			signURI = "/" + signURI
		}
		sum := md5.Sum([]byte(fmt.Sprintf("%s@%d@%s", signURI, r.Timestamp, r.AuthSecret)))
		fullURL += "?sign=" + hex.EncodeToString(sum[:]) + "&t=" + fmt.Sprint(r.Timestamp)
	}
	return fullURL
}

var (
	absoluteResourcePattern = regexp.MustCompile(`(?i)^[a-z]{2,5}://`)
	serializedStringPattern = regexp.MustCompile(`s:\d+:"([^"]*)";s:\d+:"([^"]*)"`)
)

func parseSerializedStrings(raw string) map[string]string {
	result := map[string]string{}
	for _, match := range serializedStringPattern.FindAllStringSubmatch(raw, -1) {
		if len(match) == 3 {
			result[match[1]] = match[2]
		}
	}
	return result
}
