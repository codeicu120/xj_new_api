package open

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"sort"
	"strings"
	"time"

	userRepo "xj_comp/internal/repository/user"
	"xj_comp/internal/service/resourceurl"
)

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type GuestStore interface {
	GuestExists(ctx context.Context, sid string) (bool, error)
	CreateGuest(ctx context.Context, sid string, now int64) error
}

type Service struct {
	auth            AuthStore
	guest           GuestStore
	resourceBaseURL string
	now             func() time.Time
	resources       *resourceurl.Resolver
}

func (s *Service) WithResourceResolver(r *resourceurl.Resolver) *Service { s.resources = r; return s }

type appConfig struct {
	AESKey string
	MD5Key string
}

var apps = map[string]appConfig{
	"4b4131e49": {AESKey: "ZowrOhHS7FulQ2Po", MD5Key: "ndMfF4niMaPjF1JP"},
}

func NewService(auth AuthStore, guest GuestStore, resourceBaseURL string) *Service {
	return &Service{auth: auth, guest: guest, resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"), now: time.Now}
}

func (s *Service) ReqAuth(ctx context.Context, token string, ip string, appID string) (map[string]interface{}, int, string, error) {
	cfg, ok := apps[strings.TrimSpace(appID)]
	if !ok {
		return nil, -1, "请输入正确的appid", nil
	}
	user, sid, err := s.userOrGuest(ctx, token, ip)
	if err != nil {
		return nil, -1, "获取授权失败", err
	}
	now := s.now().Unix()
	resolved := resourceurl.Resolved{BaseURL: s.resourceBaseURL, Timestamp: now}
	if s.resources != nil {
		resolved = resourceurl.Resolved{Timestamp: now}
		if value, resolveErr := s.resources.ResolveContext(ctx); resolveErr == nil {
			resolved = value
		}
	}
	iv := md5Hex(cfg.MD5Key)[:16]
	authrow := map[string]interface{}{
		"headUrl":     "",
		"gender":      0,
		"nickName":    "",
		"phoneNumber": "",
	}
	openidText := ""
	if atoi(user["uid"]) == 0 {
		authrow["deviceString"] = sid
		delete(authrow, "phoneNumber")
		openidText = sid
	} else {
		authrow["phoneNumber"] = fmt.Sprint(user["mobi"])
		authrow["headUrl"] = avatarURL(resolved, fmt.Sprint(user["avatar"]))
		authrow["gender"] = atoi(user["gender"])
		nickname := fmt.Sprint(user["nickname"])
		if nickname == "" {
			nickname = fmt.Sprint(user["username"])
		}
		authrow["nickName"] = nickname
		delete(authrow, "deviceString")
		openidText = fmt.Sprint(user["uid"])
	}
	openid, err := aesCBCBase64(openidText, cfg.AESKey, iv)
	if err != nil {
		return nil, -1, "获取授权失败", err
	}
	sign := openSign(authrow, openid, now, cfg.MD5Key)
	return map[string]interface{}{
		"authrow": authrow,
		"openid":  openid,
		"sign":    sign,
		"time":    now,
	}, 0, "", nil
}

func (s *Service) userOrGuest(ctx context.Context, token string, ip string) (map[string]interface{}, string, error) {
	sid := userRepo.CleanToken(token)
	if s.auth != nil && sid != "" {
		user, err := s.auth.UserBySession(ctx, sid)
		if err != nil {
			return nil, sid, err
		}
		if user != nil && atoi(user["uid"]) > 0 {
			return user, sid, nil
		}
	}
	if sid == "" {
		sid = guestSID(ip)
	}
	if s.guest != nil {
		exists, err := s.guest.GuestExists(ctx, sid)
		if err != nil {
			return nil, sid, err
		}
		if !exists {
			if err := s.guest.CreateGuest(ctx, sid, s.now().Unix()); err != nil {
				return nil, sid, err
			}
		}
	}
	return map[string]interface{}{"uid": "0", "sid": sid}, sid, nil
}

func avatarURL(resolved resourceurl.Resolved, avatar string) string {
	if avatar == "" {
		return ""
	}
	for _, ch := range avatar {
		if ch < '0' || ch > '9' {
			return resolved.GetRes(avatar, "C1")
		}
	}
	return avatar
}

func openSign(authrow map[string]interface{}, openid string, now int64, md5Key string) string {
	values := map[string]interface{}{"openid": openid, "time": now}
	for key, value := range authrow {
		values[key] = value
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, values[key]))
	}
	return md5Hex(strings.Join(parts, "&") + "&" + md5Key)
}

func aesCBCBase64(text string, key string, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	plain := pkcs7Pad([]byte(text), block.BlockSize())
	dst := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, []byte(iv)).CryptBlocks(dst, plain)
	return base64.StdEncoding.EncodeToString(dst), nil
}

func pkcs7Pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	out := make([]byte, len(src)+padding)
	copy(out, src)
	for i := len(src); i < len(out); i++ {
		out[i] = byte(padding)
	}
	return out
}

func guestSID(ip string) string {
	crc := crc32.ChecksumIEEE([]byte(ip))
	return md5Hex(fmt.Sprintf("%x", crc))
}

func md5Hex(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
