package starlive

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
)

type InfoStore interface {
	Info(ctx context.Context) (domain.StarLiveInfo, error)
}

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type GuestStore interface {
	GuestBySID(ctx context.Context, sid string) (map[string]interface{}, error)
}

type QuotaStore interface {
	Quota(ctx context.Context, uid int) (map[string]interface{}, error)
}

type Service struct {
	infoStore  InfoStore
	authStore  AuthStore
	guestStore GuestStore
	quotaStore QuotaStore
}

func NewService(infoStore InfoStore, authStore AuthStore, guestStore GuestStore, quotaStore QuotaStore) *Service {
	return &Service{
		infoStore:  infoStore,
		authStore:  authStore,
		guestStore: guestStore,
		quotaStore: quotaStore,
	}
}

func (s *Service) Index(ctx context.Context, token string) (domain.StarLiveIndexData, int, string, error) {
	userID, retcode, errmsg, err := s.memberID(ctx, token)
	if err != nil || retcode != 0 {
		return domain.StarLiveIndexData{}, retcode, errmsg, err
	}
	if s.infoStore == nil {
		return domain.StarLiveIndexData{}, -1, "暂未开放", nil
	}
	info, err := s.infoStore.Info(ctx)
	if err != nil {
		return domain.StarLiveIndexData{}, -1, "暂未开放", err
	}
	if strings.TrimSpace(info.AppID) == "" || strings.TrimSpace(info.SecKey) == "" {
		return domain.StarLiveIndexData{}, -1, "暂未开放", nil
	}
	encryptUID, err := encryptAES128CBC(fmt.Sprint(userID), info.SecKey, "16-Bytes--String")
	if err != nil {
		return domain.StarLiveIndexData{}, -1, "暂未开放", err
	}
	tokenHash := md5.Sum([]byte(info.AppID + "_" + fmt.Sprint(userID) + "_" + info.SecKey))
	return domain.StarLiveIndexData{Data: map[string]interface{}{
		"appId":      info.AppID,
		"encryptUid": encryptUID,
		"token":      hex.EncodeToString(tokenHash[:]),
		"apiHost":    info.APIHost,
		"env":        info.Env,
		"src":        info.Src,
		"liveHost":   info.LiveHost,
	}}, 0, "", nil
}

func (s *Service) QueryCoinBalance(ctx context.Context, memberID string) (domain.StarLiveBalanceResponse, error) {
	memberID = strings.TrimSpace(memberID)
	if len(memberID) > 12 {
		return domain.StarLiveBalanceResponse{Code: 0, Data: map[string]interface{}{"balance": 0}}, nil
	}
	uid, _ := strconv.Atoi(memberID)
	if uid <= 0 || s.quotaStore == nil {
		return domain.StarLiveBalanceResponse{Code: -1, Data: map[string]interface{}{"msg": "未知用户"}}, nil
	}
	quota, err := s.quotaStore.Quota(ctx, uid)
	if err != nil {
		return domain.StarLiveBalanceResponse{}, err
	}
	if len(quota) == 0 {
		return domain.StarLiveBalanceResponse{Code: -1, Data: map[string]interface{}{"msg": "未知用户"}}, nil
	}
	return domain.StarLiveBalanceResponse{Code: 0, Data: map[string]interface{}{"balance": atoi(quota["goldcoin"]) * 10}}, nil
}

func (s *Service) GameBetEdge(params map[string]interface{}) domain.StarLiveBalanceResponse {
	return starLiveMemberEdge(params)
}

func (s *Service) GameWinEdge(params map[string]interface{}) domain.StarLiveBalanceResponse {
	return starLiveMemberEdge(params)
}

func (s *Service) TranslateEdge(params map[string]interface{}) domain.StarLiveBalanceResponse {
	return starLiveMemberEdge(params)
}

func (s *Service) TryAgainEdge(params map[string]interface{}) domain.StarLiveBalanceResponse {
	busiType := atoi(params["busiType"])
	if busiType != 0 && busiType != 1 && busiType != 2 {
		return starLiveFail("未知业务类型")
	}
	return starLiveFail("重试成功分支暂未迁移")
}

func starLiveMemberEdge(params map[string]interface{}) domain.StarLiveBalanceResponse {
	memberID := strings.TrimSpace(fmt.Sprint(params["memberId"]))
	if len(memberID) > 12 {
		return starLiveFail("游客用户请先登录")
	}
	if atoi(memberID) <= 0 {
		return starLiveFail("未知用户")
	}
	return starLiveFail("直播资产成功分支暂未迁移")
}

func starLiveFail(message string) domain.StarLiveBalanceResponse {
	return domain.StarLiveBalanceResponse{Code: -1, Data: map[string]interface{}{"msg": message}}
}

func (s *Service) memberID(ctx context.Context, token string) (interface{}, int, string, error) {
	sid := userRepo.CleanToken(token)
	if sid == "" {
		return nil, -9999, "客户端游客请先携带信息", nil
	}
	if s.authStore != nil {
		user, err := s.authStore.UserBySession(ctx, sid)
		if err != nil {
			return nil, -1, "暂未开放", err
		}
		if uid := atoi(user["uid"]); uid > 0 {
			return uid, 0, "", nil
		}
	}
	if s.guestStore != nil {
		guest, err := s.guestStore.GuestBySID(ctx, sid)
		if err != nil {
			return nil, -1, "暂未开放", err
		}
		if len(guest) > 0 {
			return sid, 0, "", nil
		}
	}
	return nil, -9999, "客户端游客请先携带信息", nil
}

func encryptAES128CBC(data string, secret string, iv string) (string, error) {
	key := make([]byte, aes.BlockSize)
	copy(key, []byte(secret))
	ivBytes := make([]byte, aes.BlockSize)
	copy(ivBytes, []byte(iv+"0000000000000000"))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	plain := pkcs7Pad([]byte(data), aes.BlockSize)
	ciphertext := make([]byte, len(plain))
	mode := cipher.NewCBCEncrypter(block, ivBytes)
	mode.CryptBlocks(ciphertext, plain)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
