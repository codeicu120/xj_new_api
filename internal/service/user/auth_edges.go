package user

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	userRepo "xj_comp/internal/repository/user"
)

type AuthEdgeStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type AuthEdgeLookupStore interface {
	UserByID(ctx context.Context, uid int) (map[string]interface{}, error)
	UserByMobi(ctx context.Context, mobi string) (map[string]interface{}, error)
	UserByEmail(ctx context.Context, email string) (map[string]interface{}, error)
	UserByUsername(ctx context.Context, username string) (map[string]interface{}, error)
}

type AuthEdgePolicyStore interface {
	SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
	KeylimitCountSince(ctx context.Context, key string, since int64) (int, error)
}

type AuthEdgeDeletionStore interface {
	AccountDeletionExists(ctx context.Context, uid int) (bool, error)
}

type AuthEdgeDeletionRequestStore interface {
	RequestAccountDeletion(ctx context.Context, uid int, sid string, now int64) error
}

type AuthEdgePasswordStore interface {
	ResetPassword(ctx context.Context, uid int, passwordHash string, salt string) error
}

type AuthEdgePhoneStore interface {
	ChangePhone(ctx context.Context, uid int, mobi string) (bool, string, error)
}

type AuthEdgeLoginStore interface {
	CreateLoginSession(ctx context.Context, uid int, passwordHash string, salt string, now int64) (string, error)
	Quota(ctx context.Context, uid int) (map[string]interface{}, error)
	Goldbean(ctx context.Context, uid int) (map[string]interface{}, error)
	ClearAccountDeletion(ctx context.Context, uid int) error
}

type AuthEdgeService struct {
	store AuthEdgeStore
	now   func() time.Time
}

var usernamePattern = regexp.MustCompile(`^[\p{Han}a-z0-9_]+$`)

type AuthEdgeRequest struct {
	Token      string
	AUP        int
	Step       string
	Mobi       string
	Email      string
	Username   string
	Password   string
	SMSCode    string
	EmailCode  string
	MobiPrefix string
	RegType    int
	LoginType  int
	ClientIP   string
}

func NewAuthEdgeService(store AuthEdgeStore) *AuthEdgeService {
	return &AuthEdgeService{store: store, now: time.Now}
}

func (s *AuthEdgeService) Register(ctx context.Context, req AuthEdgeRequest, v2 bool) (int, string, error) {
	user, err := s.userByToken(ctx, req.Token)
	if err != nil {
		return -1, "注册失败", err
	}
	if atoi(user["uid"]) > 0 {
		return -1, "用户已登录", nil
	}
	if req.AUP != 1 {
		return -1, "请同意用户协议", nil
	}
	if closed, err := s.registrationClosed(ctx); err != nil {
		return -1, "注册失败", err
	} else if closed {
		return -1, "已暂时关闭了注册", nil
	}
	if limited, err := s.registrationIPLimited(ctx, req.ClientIP); err != nil {
		return -1, "注册失败", err
	} else if limited {
		return -1, "注册过于频繁，请稍后再试", nil
	}
	if !v2 {
		if !validMainlandMobile(req.MobiPrefix, req.Mobi) {
			return -1, "手机号码填写不正确", nil
		}
		if errmsg, err := s.checkMobiRegistration(ctx, normalizedMobi(req.MobiPrefix, req.Mobi)); err != nil {
			return -1, "注册失败", err
		} else if errmsg != "" {
			return -1, errmsg, nil
		}
	}
	if v2 && req.RegType == 1 {
		if errmsg, err := s.checkUsernameRegistration(ctx, req.Username); err != nil {
			return -1, "注册失败", err
		} else if errmsg != "" {
			return -1, errmsg, nil
		}
	}
	if v2 && req.RegType == 2 && !validMainlandMobile(req.MobiPrefix, req.Mobi) {
		return -1, "手机号码填写不正确", nil
	}
	if v2 && req.RegType == 2 {
		if errmsg, err := s.checkMobiRegistration(ctx, normalizedMobi(req.MobiPrefix, req.Mobi)); err != nil {
			return -1, "注册失败", err
		} else if errmsg != "" {
			return -1, errmsg, nil
		}
	}
	if v2 && req.RegType == 3 && !validEmail(req.Email) {
		return -1, "请输入正确邮箱地址", nil
	}
	if v2 && req.RegType == 3 {
		if errmsg, err := s.checkEmailRegistration(ctx, req.Email); err != nil {
			return -1, "注册失败", err
		} else if errmsg != "" {
			return -1, errmsg, nil
		}
	}
	if v2 && req.RegType == 1 && !validPassword(req.Password) {
		return -1, "密码6-16位", nil
	}
	return -1, "注册成功分支暂未迁移", nil
}

func (s *AuthEdgeService) Login(ctx context.Context, req AuthEdgeRequest, v2 bool) (map[string]interface{}, int, string, error) {
	user, err := s.userByToken(ctx, req.Token)
	if err != nil {
		return nil, -1, "登录失败", err
	}
	if atoi(user["uid"]) > 0 {
		return nil, -1, "用户已登录", nil
	}
	if !v2 && req.LoginType != 1 {
		if closed, err := s.passwordLoginClosed(ctx); err != nil {
			return nil, -1, "登录失败", err
		} else if closed {
			return nil, -1, "系统已关闭密码登录", nil
		}
	}
	if v2 && strings.TrimSpace(req.Mobi) == "" && strings.TrimSpace(req.Email) == "" && strings.TrimSpace(req.Username) == "" {
		return nil, -1, "用户名未注册", nil
	}
	var row map[string]interface{}
	if v2 {
		row, err = s.lookupLoginUser(ctx, req)
		if err != nil {
			return nil, -1, "登录失败", err
		}
		if len(row) == 0 {
			switch {
			case strings.TrimSpace(req.Mobi) != "":
				return nil, -1, "手机号码未注册", nil
			case strings.TrimSpace(req.Email) != "":
				return nil, -1, "邮箱未注册", nil
			default:
				return nil, -1, "用户名未注册", nil
			}
		}
		if req.LoginType == 1 {
			key := "sms." + normalizedMobi(req.MobiPrefix, req.Mobi) + "." + strings.TrimSpace(req.SMSCode)
			errmsg := "手机验证码不正确"
			if strings.TrimSpace(req.Email) != "" && strings.TrimSpace(req.Mobi) == "" {
				key = "email." + strings.TrimSpace(req.Email) + "." + strings.TrimSpace(req.SMSCode)
				errmsg = "邮箱验证码不正确"
			}
			ok, err := s.verificationCodeValid(ctx, key, 600)
			if err != nil {
				return nil, -1, "登录失败", err
			}
			if !ok {
				return nil, -1, errmsg, nil
			}
			req.Password = str(row["password"])
		} else if strings.TrimSpace(req.Password) == "" {
			return nil, -1, "密码不能为空", nil
		}
	} else if req.LoginType == 1 {
		return nil, -1, "短信登录自动注册成功分支暂未迁移", nil
	} else {
		username := strings.TrimSpace(req.Username)
		if username == "" {
			username = normalizedMobi(req.MobiPrefix, req.Mobi)
		}
		row, err = s.lookupUsername(ctx, username)
		if err != nil {
			return nil, -1, "登录失败", err
		}
		if len(row) == 0 {
			return nil, -1, "用户名或密码不正确", nil
		}
	}
	data, errmsg, err := s.loginSuccess(ctx, row, req.Password)
	if err != nil {
		return nil, -1, "登录失败", err
	}
	if errmsg != "" {
		return nil, -1, errmsg, nil
	}
	return data, 0, "登录成功", nil
}

func (s *AuthEdgeService) Forgot(ctx context.Context, req AuthEdgeRequest, v2 bool) (int, string, error) {
	if v2 && strings.TrimSpace(req.Mobi) == "" && strings.TrimSpace(req.Email) == "" {
		return -1, "请填写手机号码或者邮箱", nil
	}
	if v2 && strings.TrimSpace(req.Email) != "" && strings.TrimSpace(req.Mobi) == "" && !validEmail(req.Email) {
		return -1, "请输入正确邮箱地址", nil
	}
	if !v2 && !validMainlandMobile(req.MobiPrefix, req.Mobi) || v2 && strings.TrimSpace(req.Mobi) != "" && !validMainlandMobile(req.MobiPrefix, req.Mobi) {
		return -1, "手机号码填写不正确", nil
	}
	if strings.TrimSpace(req.Step) == "" {
		return -1, "无效的操作", nil
	}
	if req.Step == "step1" || req.Step == "step2" || req.Step == "step3" {
		row, err := s.lookupForgotUser(ctx, req, v2)
		if err != nil {
			return -1, "密码重置失败", err
		}
		if len(row) == 0 {
			if v2 && strings.TrimSpace(req.Email) != "" && strings.TrimSpace(req.Mobi) == "" {
				return -1, "输入的邮箱不存在", nil
			}
			return -1, "输入的手机号码不存在", nil
		}
		if req.Step == "step1" {
			return 0, "step1->step2", nil
		}
		key, errmsg := forgotVerificationKey(req, v2)
		ok, err := s.verificationCodeValid(ctx, key, 600)
		if err != nil {
			return -1, "密码重置失败", err
		}
		if !ok {
			return -1, errmsg, nil
		}
		if req.Step == "step3" {
			if !validPassword(req.Password) {
				return -1, "密码6-16位", nil
			}
			if atoi(row["uid"]) == 0 {
				return -1, "未获得指定用户", nil
			}
			if err := s.resetPassword(ctx, atoi(row["uid"]), req.Password); err != nil {
				return -1, "密码重置失败", err
			}
			return 0, "密码已成功设置", nil
		}
		return 0, "step2->step3", nil
	}
	return -1, "无效的操作", nil
}

func (s *AuthEdgeService) Delete(ctx context.Context, req AuthEdgeRequest) (int, string, error) {
	user, err := s.userByToken(ctx, req.Token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	uid := atoi(user["uid"])
	exists, err := s.accountDeletionExists(ctx, uid)
	if err != nil {
		return -1, "账号注销失败", err
	}
	if exists {
		return -1, "该账号已申请注销，请勿重复操作", nil
	}
	row, err := s.lookupUserByID(ctx, uid)
	if err != nil {
		return -1, "账号注销失败", err
	}
	if isGuestAccount(row) {
		return -1, "游客账号无需注销", nil
	}
	if !strings.HasPrefix(str(row["mobi"]), "~") {
		ok, err := s.verificationCodeValid(ctx, "sms."+str(row["mobi"])+"."+strings.TrimSpace(req.SMSCode), 600)
		if err != nil {
			return -1, "账号注销失败", err
		}
		if !ok {
			return -1, "手机验证码不正确", nil
		}
	} else {
		ok, err := s.verificationCodeValid(ctx, "email."+str(row["email"])+"."+strings.TrimSpace(req.EmailCode), 600)
		if err != nil {
			return -1, "账号注销失败", err
		}
		if !ok {
			return -1, "邮箱验证码不正确", nil
		}
	}
	if err := s.requestAccountDeletion(ctx, uid, str(user["sid"])); err != nil {
		return -1, "账号注销失败", err
	}
	return 0, "注销后保持180天不登录，系统才会删除您的数据", nil
}

func (s *AuthEdgeService) ChangePhone(ctx context.Context, req AuthEdgeRequest) (int, string, error) {
	user, err := s.userByToken(ctx, req.Token)
	if err != nil {
		return -9999, "请登录后操作", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "请登录后操作", nil
	}
	if !validMainlandMobile(req.MobiPrefix, req.Mobi) {
		return -1, "手机号码填写不正确", nil
	}
	if req.Step != "step1" && req.Step != "step2" {
		return -1, "步骤错误", nil
	}
	mobi := normalizedMobi(req.MobiPrefix, req.Mobi)
	if strings.TrimSpace(fmt.Sprint(user["mobi"])) == mobi {
		return -1, "更换的手机号和当前手机号相同！", nil
	}
	row, err := s.lookupMobi(ctx, mobi)
	if err != nil {
		return -1, "手机号更换失败", err
	}
	if len(row) > 0 {
		return -1, "手机号已经存在", nil
	}
	if req.Step == "step1" {
		return 0, "step1->step2", nil
	}
	ok, err := s.verificationCodeValid(ctx, "sms."+mobi+"."+strings.TrimSpace(req.SMSCode), 600)
	if err != nil {
		return -1, "手机号更换失败", err
	}
	if !ok {
		return -1, "手机验证码不正确", nil
	}
	ok, errmsg, err := s.changePhone(ctx, atoi(user["uid"]), mobi)
	if err != nil {
		return -1, "手机号更换失败", err
	}
	if !ok {
		if errmsg == "" {
			errmsg = "手机号更换失败,请重试"
		}
		if errmsg == "手机号已经存在" {
			return -1, errmsg, nil
		}
		return 0, errmsg, nil
	}
	return 0, "手机号更换成功", nil
}

func (s *AuthEdgeService) lookupLoginUser(ctx context.Context, req AuthEdgeRequest) (map[string]interface{}, error) {
	switch {
	case strings.TrimSpace(req.Mobi) != "":
		return s.lookupMobi(ctx, normalizedMobi(req.MobiPrefix, req.Mobi))
	case strings.TrimSpace(req.Email) != "":
		return s.lookupEmail(ctx, strings.TrimSpace(req.Email))
	default:
		return s.lookupUsername(ctx, strings.TrimSpace(req.Username))
	}
}

func (s *AuthEdgeService) lookupUserByID(ctx context.Context, uid int) (map[string]interface{}, error) {
	lookup, ok := s.store.(AuthEdgeLookupStore)
	if !ok || lookup == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	row, err := lookup.UserByID(ctx, uid)
	if row == nil {
		row = map[string]interface{}{}
	}
	return row, err
}

func (s *AuthEdgeService) accountDeletionExists(ctx context.Context, uid int) (bool, error) {
	store, ok := s.store.(AuthEdgeDeletionStore)
	if !ok || store == nil || uid <= 0 {
		return false, nil
	}
	return store.AccountDeletionExists(ctx, uid)
}

func (s *AuthEdgeService) requestAccountDeletion(ctx context.Context, uid int, sid string) error {
	store, ok := s.store.(AuthEdgeDeletionRequestStore)
	if !ok || store == nil || uid <= 0 {
		return nil
	}
	return store.RequestAccountDeletion(ctx, uid, sid, s.now().Unix())
}

func (s *AuthEdgeService) resetPassword(ctx context.Context, uid int, password string) error {
	store, ok := s.store.(AuthEdgePasswordStore)
	if !ok || store == nil || uid <= 0 {
		return nil
	}
	salt, err := randomString(8)
	if err != nil {
		return err
	}
	return store.ResetPassword(ctx, uid, phpPassword(password+salt), salt)
}

func (s *AuthEdgeService) changePhone(ctx context.Context, uid int, mobi string) (bool, string, error) {
	store, ok := s.store.(AuthEdgePhoneStore)
	if !ok || store == nil || uid <= 0 {
		return true, "", nil
	}
	return store.ChangePhone(ctx, uid, mobi)
}

func (s *AuthEdgeService) loginSuccess(ctx context.Context, user map[string]interface{}, password string) (map[string]interface{}, string, error) {
	if len(user) == 0 {
		return nil, "用户名或密码不正确", nil
	}
	if atoi(user["locktype"]) == 1 {
		return nil, "你的账户已被锁定", nil
	}
	hash := strings.TrimSpace(password)
	if hash == "" {
		return nil, "密码不能为空", nil
	}
	if len(hash) != 40 {
		hash = phpPassword(hash + str(user["salt"]))
	}
	if hash != str(user["password"]) {
		return nil, "用户名或密码不正确", nil
	}
	loginStore, ok := s.store.(AuthEdgeLoginStore)
	if !ok || loginStore == nil {
		return nil, "登录成功分支暂未迁移", nil
	}
	uid := atoi(user["uid"])
	sid, err := loginStore.CreateLoginSession(ctx, uid, str(user["password"]), str(user["salt"]), s.now().Unix())
	if err != nil {
		return nil, "", err
	}
	if sid == "" {
		return nil, "用户名或密码不正确", nil
	}
	quota, err := loginStore.Quota(ctx, uid)
	if err != nil {
		return nil, "", err
	}
	goldbean, err := loginStore.Goldbean(ctx, uid)
	if err != nil {
		return nil, "", err
	}
	userData := procLoginUser(user, quota, goldbean, s.now())
	if err := loginStore.ClearAccountDeletion(ctx, uid); err != nil {
		return nil, "", err
	}
	return map[string]interface{}{
		"user":         userData,
		"xxx_api_auth": hex.EncodeToString([]byte(sid)),
	}, "", nil
}

func procLoginUser(user map[string]interface{}, quota map[string]interface{}, goldbean map[string]interface{}, now time.Time) map[string]interface{} {
	sysgid := atoi(user["sysgid"])
	sysgidExp := atoi64(user["sysgid_exptime"])
	gid := atoi(user["gid"])
	avatar := str(user["avatar"])
	result := map[string]interface{}{
		"uid":             str(user["uid"]),
		"uniqkey":         strings.ToUpper(base36(atoi64(user["uniqkey"]))),
		"username":        str(user["username"]),
		"nickname":        str(user["nickname"]),
		"mobi":            str(user["mobi"]),
		"email":           str(user["email"]),
		"sysgid":          str(user["sysgid"]),
		"gid":             str(user["gid"]),
		"gids":            nil,
		"gicon":           gicon(sysgid, gid),
		"isvip":           vipFlag(sysgid, sysgidExp, now),
		"regtime":         formatUnix(atoi64(user["regtime"])),
		"gender":          atoi(user["gender"]),
		"avatar":          avatar,
		"avatar_url":      avatarURL(avatar),
		"newmsg":          str(user["newmsg"]),
		"goldcoin":        atoi(quota["goldcoin"]),
		"gold_bean":       atoi(goldbean["gold_bean"]),
		"duetime":         "",
		"dueday":          "",
		"recommend_total": atoi(user["recommend_total"]),
	}
	if sysgidExp > 0 {
		result["duetime"] = formatUnix(sysgidExp)
		if sysgidExp > now.Unix() {
			result["dueday"] = "过期"
		} else {
			result["dueday"] = "已过期"
		}
	}
	return result
}

func vipFlag(sysgid int, sysgidExp int64, now time.Time) int {
	if sysgid == 6 && sysgidExp > now.Unix() {
		return 1
	}
	return 0
}

func gicon(sysgid int, gid int) string {
	if sysgid > 0 {
		return "g" + strconv.Itoa(sysgid)
	}
	if gid > 0 {
		return "g" + strconv.Itoa(gid)
	}
	return ""
}

func avatarURL(avatar string) string {
	if avatar == "" || isDigits(avatar) {
		return avatar
	}
	return avatar
}

func formatUnix(value int64) string {
	if value <= 0 {
		return "1970-01-01 00:00:00"
	}
	return time.Unix(value, 0).Format("2006-01-02 15:04:05")
}

func (s *AuthEdgeService) registrationClosed(ctx context.Context) (bool, error) {
	setting, err := s.settingMap(ctx, "user.regopt")
	if err != nil {
		return false, err
	}
	return atoi(setting["regclosed"]) == 1, nil
}

func (s *AuthEdgeService) registrationIPLimited(ctx context.Context, ip string) (bool, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return false, nil
	}
	key := "user.regiser.ip." + ip
	count30m, err := s.keylimitCount(ctx, key, 1800)
	if err != nil {
		return false, err
	}
	if count30m >= 1 {
		return true, nil
	}
	count24h, err := s.keylimitCount(ctx, key, 86400)
	if err != nil {
		return false, err
	}
	return count24h >= 2, nil
}

func (s *AuthEdgeService) passwordLoginClosed(ctx context.Context) (bool, error) {
	setting, err := s.settingMap(ctx, "setting")
	if err != nil {
		return false, err
	}
	value, ok := setting["pswdLoginStatus"]
	if !ok {
		return false, nil
	}
	return atoi(value) != 1, nil
}

func (s *AuthEdgeService) settingMap(ctx context.Context, uuid string) (map[string]interface{}, error) {
	policy, ok := s.store.(AuthEdgePolicyStore)
	if !ok || policy == nil {
		return map[string]interface{}{}, nil
	}
	row, err := policy.SettingByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return parsePHPSerializedMap(str(row["value"])), nil
}

func (s *AuthEdgeService) keylimitCount(ctx context.Context, key string, seconds int64) (int, error) {
	policy, ok := s.store.(AuthEdgePolicyStore)
	if !ok || policy == nil {
		return 0, nil
	}
	since := s.now().Unix() - seconds
	return policy.KeylimitCountSince(ctx, key, since)
}

func (s *AuthEdgeService) verificationCodeValid(ctx context.Context, key string, seconds int64) (bool, error) {
	count, err := s.keylimitCount(ctx, key, seconds)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func parsePHPSerializedMap(value string) map[string]interface{} {
	out := map[string]interface{}{}
	re := regexp.MustCompile(`s:\d+:"([^"]+)";(?:s:\d+:"([^"]*)"|i:(-?\d+)|d:([0-9.]+)|N;)`)
	for _, match := range re.FindAllStringSubmatch(value, -1) {
		key := match[1]
		switch {
		case match[2] != "":
			out[key] = match[2]
		case match[3] != "":
			out[key] = atoi(match[3])
		case match[4] != "":
			out[key] = match[4]
		default:
			out[key] = nil
		}
	}
	return out
}

func (s *AuthEdgeService) checkMobiRegistration(ctx context.Context, mobi string) (string, error) {
	if strings.TrimSpace(mobi) == "" {
		return "手机号码不能为空", nil
	}
	row, err := s.lookupMobi(ctx, mobi)
	if err != nil {
		return "", err
	}
	if len(row) > 0 {
		return "手机号码已被注册", nil
	}
	return "", nil
}

func (s *AuthEdgeService) checkEmailRegistration(ctx context.Context, email string) (string, error) {
	if !validEmail(email) {
		return "请输入正确邮箱地址", nil
	}
	row, err := s.lookupEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		return "", err
	}
	if len(row) > 0 {
		return "该邮箱已经被注册，您可以通过邮箱找回密码", nil
	}
	return "", nil
}

func (s *AuthEdgeService) checkUsernameRegistration(ctx context.Context, username string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(username))
	switch {
	case value == "":
		return "请填写用户名", nil
	case isDigits(value):
		return "用户名不能是纯数字", nil
	case len(value) < 6 || len(value) > 24 || utf8.RuneCountInString(value) > 16:
		return "用户名2-8个汉字，英文6-16个字符", nil
	case !usernamePattern.MatchString(value):
		return "用户名只允许中英文、数字及下划线组成", nil
	}
	row, err := s.lookupUsername(ctx, value)
	if err != nil {
		return "", err
	}
	if len(row) > 0 {
		return "用户名已存在", nil
	}
	return "", nil
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func isGuestAccount(row map[string]interface{}) bool {
	mobi := str(row["mobi"])
	email := str(row["email"])
	return strings.HasPrefix(mobi, "~") && strings.HasPrefix(email, "~")
}

func (s *AuthEdgeService) lookupForgotUser(ctx context.Context, req AuthEdgeRequest, v2 bool) (map[string]interface{}, error) {
	if v2 && strings.TrimSpace(req.Email) != "" && strings.TrimSpace(req.Mobi) == "" {
		return s.lookupEmail(ctx, strings.TrimSpace(req.Email))
	}
	return s.lookupMobi(ctx, normalizedMobi(req.MobiPrefix, req.Mobi))
}

func forgotVerificationKey(req AuthEdgeRequest, v2 bool) (string, string) {
	if v2 && strings.TrimSpace(req.Email) != "" && strings.TrimSpace(req.Mobi) == "" {
		code := strings.TrimSpace(req.EmailCode)
		if code == "" {
			code = strings.TrimSpace(req.SMSCode)
		}
		return "email." + strings.TrimSpace(req.Email) + "." + code, "邮箱验证码不正确"
	}
	return "sms." + normalizedMobi(req.MobiPrefix, req.Mobi) + "." + strings.TrimSpace(req.SMSCode), "手机验证码不正确"
}

func (s *AuthEdgeService) lookupMobi(ctx context.Context, mobi string) (map[string]interface{}, error) {
	lookup, ok := s.store.(AuthEdgeLookupStore)
	if !ok || lookup == nil {
		return map[string]interface{}{}, nil
	}
	row, err := lookup.UserByMobi(ctx, mobi)
	if row == nil {
		row = map[string]interface{}{}
	}
	return row, err
}

func (s *AuthEdgeService) lookupEmail(ctx context.Context, email string) (map[string]interface{}, error) {
	lookup, ok := s.store.(AuthEdgeLookupStore)
	if !ok || lookup == nil {
		return map[string]interface{}{}, nil
	}
	row, err := lookup.UserByEmail(ctx, email)
	if row == nil {
		row = map[string]interface{}{}
	}
	return row, err
}

func (s *AuthEdgeService) lookupUsername(ctx context.Context, username string) (map[string]interface{}, error) {
	lookup, ok := s.store.(AuthEdgeLookupStore)
	if !ok || lookup == nil {
		return map[string]interface{}{}, nil
	}
	row, err := lookup.UserByUsername(ctx, username)
	if row == nil {
		row = map[string]interface{}{}
	}
	return row, err
}

func (s *AuthEdgeService) userByToken(ctx context.Context, token string) (map[string]interface{}, error) {
	if s.store == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	sid := userRepo.CleanToken(token)
	if sid == "" {
		return map[string]interface{}{"uid": "0"}, nil
	}
	user, err := s.store.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	return user, nil
}

func validMainlandMobile(prefix string, mobi string) bool {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "+86"
	}
	raw := strings.TrimSpace(mobi)
	if prefix != "+86" && prefix != "86" {
		return raw != ""
	}
	if len(raw) != 11 || raw[0] != '1' {
		return false
	}
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func validEmail(email string) bool {
	email = strings.TrimSpace(email)
	at := strings.Index(email, "@")
	dot := strings.LastIndex(email, ".")
	return at > 0 && dot > at+1 && dot < len(email)-1
}

func validPassword(password string) bool {
	n := len(password)
	return n >= 6 && n <= 16
}

func phpPassword(password string) string {
	if len(password) != 32 {
		sum := md5.Sum([]byte(password))
		password = hex.EncodeToString(sum[:])
	}
	type saltItem struct {
		key string
		pos int
		val int
	}
	items := []saltItem{}
	add := func(key string, raw int, posMask int, posXor int) {
		pos := raw & posMask
		if posXor > 0 {
			pos ^= posXor
		}
		items = append(items, saltItem{key: key, pos: pos, val: (raw >> 4) ^ 0xf})
	}
	hexAt := func(start int) int {
		if start < 0 {
			start = len(password) + start
		}
		if start < 0 || start >= len(password) {
			return 0
		}
		end := start + 2
		if end > len(password) {
			end = len(password)
		}
		value, _ := strconv.ParseInt(password[start:end], 16, 64)
		return int(value)
	}
	add("x0", hexAt(0), 0x1f, 0)
	add("x1", hexAt(items[0].pos), 0x1f, 0)
	add("x2", hexAt(items[0].val), 0x0f, 0x0f)
	add("x3", hexAt(-1), 0x1f, 0)
	add("x4", hexAt(items[3].pos), 0x1f, 0)
	add("x5", hexAt(items[3].val), 0x1f, 0)
	add("x6", hexAt(14), 0x1f, 0)
	add("x7", hexAt(16), 0x1f, 0)
	sort.SliceStable(items, func(i, j int) bool {
		left := fmt.Sprintf("%d%s", items[i].pos, strings.TrimPrefix(items[i].key, "x"))
		right := fmt.Sprintf("%d%s", items[j].pos, strings.TrimPrefix(items[j].key, "x"))
		return atoi(left) < atoi(right)
	})
	for i, item := range items {
		pos := item.pos + i
		if pos < 0 {
			pos = 0
		}
		if pos > len(password) {
			pos = len(password)
		}
		password = password[:pos] + fmt.Sprintf("%x", item.val) + password[pos:]
	}
	return password
}

func randomString(length int) (string, error) {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		buf[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(buf), nil
}

func normalizedMobi(prefix string, mobi string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "+86"
	}
	return strings.Trim(strings.TrimSpace(prefix)+"."+strings.TrimSpace(mobi), "+")
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}

func atoi64(value interface{}) int64 {
	var n int64
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}

func base36(value int64) string {
	if value <= 0 {
		return "0"
	}
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz"
	result := ""
	for value > 0 {
		result = string(chars[value%36]) + result
		value /= 36
	}
	return result
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}
