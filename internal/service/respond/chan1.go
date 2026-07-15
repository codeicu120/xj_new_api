package respond

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

const payTypeBuyPkg = 8

var digitsOnly = regexp.MustCompile(`^\d+$`)

type Chan1Store interface {
	UserByMobi(ctx context.Context, mobi string) (map[string]interface{}, error)
	CountPaymentsByUIDPayTypePayway(ctx context.Context, uid int, payType int, payway string) (int, error)
	VIPPackageByID(ctx context.Context, pkgID int) (map[string]interface{}, error)
}

type Service struct {
	store Chan1Store
}

func NewService(store Chan1Store) *Service {
	return &Service{store: store}
}

func (s *Service) Chan1(ctx context.Context, mobi string) (int, string, error) {
	if s.store == nil {
		return 2, "用户不存在", nil
	}
	mobi = normalizeMobi(mobi)
	user, err := s.store.UserByMobi(ctx, mobi)
	if err != nil {
		return -1, "chan1 请求失败", fmt.Errorf("query chan1 user: %w", err)
	}
	if len(user) == 0 {
		return 2, "用户不存在", nil
	}
	uid := atoi(user["uid"])
	total, err := s.store.CountPaymentsByUIDPayTypePayway(ctx, uid, payTypeBuyPkg, "chan1")
	if err != nil {
		return -1, "chan1 请求失败", fmt.Errorf("count chan1 payment: %w", err)
	}
	if total > 0 {
		return 3, "该用户已经送过会员了", nil
	}
	pkg, err := s.store.VIPPackageByID(ctx, 1)
	if err != nil {
		return -1, "chan1 请求失败", fmt.Errorf("query chan1 vip package: %w", err)
	}
	if len(pkg) == 0 || atoi(pkg["showtype"]) != 0 {
		return -1, "套餐不存在或未启用", nil
	}
	return -1, "chan1 成功分支暂未迁移", nil
}

func normalizeMobi(mobi string) string {
	mobi = strings.TrimSpace(mobi)
	if digitsOnly.MatchString(mobi) {
		return "86." + mobi
	}
	return mobi
}

func atoi(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		var out int
		_, _ = fmt.Sscanf(n, "%d", &out)
		return out
	default:
		return 0
	}
}
