package ucp

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	coinTypeSign = 1
)

func (s *Service) HighRiskActionEdge(ctx context.Context, token string, pendingMessage string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if pendingMessage == "" {
		pendingMessage = "成功分支暂未迁移"
	}
	return -1, pendingMessage, nil
}

func (s *Service) UpgradeEdge(ctx context.Context, token string, day int) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	const superVIPGID = 6
	if atoi(user["sysgid"]) == superVIPGID {
		return -1, "您已经是尊贵会员", nil
	}
	pricing := map[int]int{
		7:    100,
		30:   300,
		180:  1000,
		365:  1500,
		3650: 3000,
	}
	deductCoin, ok := pricing[day]
	if !ok {
		return -1, "请选择一个时长", nil
	}
	if day == 3650 {
		return -1, "终身尊贵VIP暂停升级", nil
	}
	quota, err := s.store.Quota(ctx, uid)
	if err != nil {
		return -1, "会员升级失败", err
	}
	if atoi(quota["goldcoin"]) < deductCoin {
		return -1, "金币不足，快做任务获取金币吧！", nil
	}
	return -1, "会员升级成功分支暂未迁移", nil
}

func (s *Service) TaskSignEdge(ctx context.Context, token string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -1, "签到失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		guest, err := s.store.GuestBySID(ctx, str(user["sid"]))
		if err != nil {
			return -1, "签到失败", err
		}
		if len(guest) == 0 {
			return -1, "请登录后操作，客户端游客请先携带信息", nil
		}
		if sameDay(atoi64(guest["signtime"]), s.now()) {
			return -1, "您今天已经签过到了", nil
		}
		return -1, "签到成功分支暂未迁移", nil
	}
	count, err := s.store.CountCoinLogsSinceByType(ctx, uid, coinTypeSign, dayStartUnix(s.now()))
	if err != nil {
		return -1, "签到失败", err
	}
	if count > 0 {
		return -1, "您今天已经签过到了", nil
	}
	return -1, "签到成功分支暂未迁移", nil
}

func (s *Service) TaskInviteCodeInputEdge(ctx context.Context, token string, inviteCode string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	count, err := s.store.CountCoinLogsSinceByType(ctx, uid, coinTypeSaveQRCode, dayStartUnix(s.now()))
	if err != nil {
		return -1, "保存二维码失败", err
	}
	if count > 0 {
		return -1, "您今天已经保存过了", nil
	}
	expected := strings.ToUpper(taskBase36(atoi(user["uniqkey"])))
	if strings.TrimSpace(inviteCode) != expected {
		return -1, "邀请码不正确", nil
	}
	return -1, "邀请码绑定成功分支暂未迁移", nil
}

func (s *Service) TaskAdviewClickEdge(ctx context.Context, token string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	count, err := s.store.CountCoinLogsSinceByType(ctx, uid, coinTypeAdViewClick, dayStartUnix(s.now()))
	if err != nil {
		return -1, "广告点击失败", err
	}
	if count > 0 {
		return -1, "您今天已经送过了", nil
	}
	return -1, "广告点击奖励成功分支暂未迁移", nil
}

func (s *Service) TaskQRCodeSaveEdge(ctx context.Context, token string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	count, err := s.store.CountCoinLogsSinceByType(ctx, uid, coinTypeSaveQRCode, dayStartUnix(s.now()))
	if err != nil {
		return -1, "保存二维码失败", err
	}
	if count > 0 {
		return -1, "您今天已经保存过了", nil
	}
	return -1, "保存二维码奖励成功分支暂未迁移", nil
}

func (s *Service) TaskboxOpenEdge(ctx context.Context, token string, taskID int) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	taskrow, err := s.store.TaskboxByID(ctx, taskID)
	if err != nil {
		return -1, "任务宝箱开启失败", err
	}
	if len(taskrow) == 0 || atoi(taskrow["showtype"]) != 0 {
		return -1, "任务不存在或已停用", nil
	}
	if atoi(taskrow["mincoin"]) == 0 && atoi(taskrow["maxcoin"]) == 0 {
		return -1, "宝箱赠送金币为0", nil
	}
	now := s.now()
	nowUnix := now.Unix()
	dayKeyDaily, dayKeyWeekly, weekday, startTime := taskboxTimes(now)
	taskID = atoi(taskrow["taskid"])
	switch taskID {
	case 1022:
		if nowUnix < startTime {
			return -1, "每日神秘宝箱领取时间未开始", nil
		}
		if nowUnix >= startTime+300 {
			return -1, "每日神秘宝箱领取时间已结束", nil
		}
		checkrow, err := s.store.TaskboxLog(ctx, uid, taskID, dayKeyDaily)
		if err != nil {
			return -1, "任务宝箱开启失败", err
		}
		if len(checkrow) > 0 {
			return -1, "每日神秘宝箱已领过了", nil
		}
	case 1622:
		if weekday != 6 {
			return -1, "每周神秘宝箱周六晚开始", nil
		}
		if nowUnix < startTime {
			return -1, "每周神秘宝箱领取时间未开始", nil
		}
		if nowUnix >= startTime+300 {
			return -1, "每周神秘宝箱领取时间已结束", nil
		}
		checkrow, err := s.store.TaskboxLog(ctx, uid, taskID, dayKeyWeekly)
		if err != nil {
			return -1, "任务宝箱开启失败", err
		}
		if len(checkrow) > 0 {
			return -1, "每周神秘宝箱已领过了", nil
		}
	default:
		checkrow, err := s.store.TaskboxLog(ctx, uid, taskID, 0)
		if err != nil {
			return -1, "任务宝箱开启失败", err
		}
		if len(checkrow) > 0 {
			return -1, "推广任务宝箱已领过了", nil
		}
		currentUser, err := s.store.UserByID(ctx, uid)
		if err != nil {
			return -1, "任务宝箱开启失败", err
		}
		recommendTotal := 0
		if atoi(currentUser["uid"]) != 0 {
			recommendTotal = atoi(currentUser["recommend_total"])
		}
		if recommendTotal < taskID {
			return -1, "推广人数未达标，继续加油哦", nil
		}
	}
	return -1, "任务宝箱开启成功分支暂未迁移", nil
}

func sameDay(ts int64, now time.Time) bool {
	if ts <= 0 {
		return false
	}
	return ts >= dayStartUnix(now)
}

func (s *Service) UserCheckEmailEdge(ctx context.Context, token string, email string) (int, string, error) {
	email = strings.TrimSpace(email)
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if !validEmail(email) {
		return -1, "请输入正确的邮箱地址", nil
	}
	if msg, err := s.emailRateLimitMessage(ctx, "checkemail."+email+"."+s.now().Format("20060102"), 10*time.Second, 50); err != nil {
		return -1, "邮箱检查失败", err
	} else if msg != "" {
		return -1, msg, nil
	}
	existing, err := s.store.UserByEmail(ctx, email)
	if err != nil {
		return -1, "邮箱检查失败", err
	}
	if len(existing) > 0 {
		return -1, "邮箱已经被使用了", nil
	}
	return 0, "邮箱可用", nil
}

func (s *Service) UserSendEmailEdge(ctx context.Context, token string, email string) (int, string, error) {
	email = strings.TrimSpace(email)
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if !validEmail(email) {
		return -1, "请输入正确的邮箱地址", nil
	}
	if msg, err := s.emailRateLimitMessage(ctx, "bindemail."+email+"."+s.now().Format("20060102"), time.Minute, 10); err != nil {
		return -1, "邮箱验证码发送失败", err
	} else if msg != "" {
		return -1, msg, nil
	}
	existing, err := s.store.UserByEmail(ctx, email)
	if err != nil {
		return -1, "邮箱验证码发送失败", err
	}
	if len(existing) > 0 {
		return -1, "邮箱已经被使用了", nil
	}
	settingRow, err := s.store.SettingByUUID(ctx, "setting")
	if err != nil {
		return -1, "邮箱验证码发送失败", err
	}
	setting := parseTaskPHPSerializedMap(str(settingRow["value"]))
	conf := map[string]interface{}{}
	if err := json.Unmarshal([]byte(str(setting["mailconf"])), &conf); err != nil || len(conf) == 0 {
		return -1, "邮箱功能暂未开启，请稍后重试", nil
	}
	return -1, "邮箱验证码发送成功分支暂未迁移", nil
}

func (s *Service) UserVerifyEmailEdge(ctx context.Context, token string, email string, emailCode string) (int, string, error) {
	email = strings.TrimSpace(email)
	emailCode = strings.TrimSpace(emailCode)
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	keydata, err := s.store.KeylimitDataSince(ctx, "email."+email+"."+emailCode, s.now().Add(-time.Hour).Unix())
	if err != nil {
		return -1, "邮箱验证失败", err
	}
	if keydata == "" {
		return -1, "验证码不存在或已失效", nil
	}
	existing, err := s.store.UserByEmail(ctx, email)
	if err != nil {
		return -1, "邮箱验证失败", err
	}
	if len(existing) > 0 {
		return -1, "邮箱已经被使用", nil
	}
	return -1, "邮箱验证绑定成功分支暂未迁移", nil
}

func (s *Service) emailRateLimitMessage(ctx context.Context, key string, recentWindow time.Duration, dayLimit int) (string, error) {
	recentCount, err := s.store.KeylimitCountSince(ctx, key, s.now().Add(-recentWindow).Unix())
	if err != nil {
		return "", err
	}
	if recentCount > 0 {
		return "发送太频率请稍后重试", nil
	}
	dayCount, err := s.store.KeylimitCountSince(ctx, key, 0)
	if err != nil {
		return "", err
	}
	if dayCount >= dayLimit {
		return "系统维护稍后重试", nil
	}
	return "", nil
}

func (s *Service) UserBindMobiEdge(ctx context.Context, token string, mobiPrefix string, mobi string, smsCode string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if mobi := str(user["mobi"]); mobi != "" && !strings.HasPrefix(mobi, "~") {
		return -1, "您已绑定手机", nil
	}
	mobiPrefix = strings.TrimSpace(mobiPrefix)
	mobi = strings.TrimSpace(mobi)
	smsCode = strings.TrimSpace(smsCode)
	if mobiPrefix == "" {
		mobiPrefix = "+86"
	}
	fullMobi := strings.Trim(mobiPrefix+"."+mobi, "+")
	count, err := s.store.KeylimitCountSince(ctx, "sms."+fullMobi+"."+smsCode, s.now().Add(-10*time.Minute).Unix())
	if err != nil {
		return -1, "手机验证失败", err
	}
	if count == 0 {
		return -1, "手机验证码不正确", nil
	}
	if _, err := s.store.UserByMobi(ctx, fullMobi); err != nil {
		return -1, "手机验证失败", err
	}
	return -1, "手机验证绑定成功分支暂未迁移", nil
}

func (s *Service) UserProfileEdge(ctx context.Context, token string, gender int, nickname string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	user, err = s.store.UserByID(ctx, uid)
	if err != nil {
		return -1, "资料设置失败", err
	}
	if gender != 1 && gender != 2 {
		gender = 1
	}
	nickname = strings.TrimSpace(nickname)
	if nickname != "" && nickname != str(user["nickname"]) {
		if ok, message := validProfileNickname(nickname); !ok {
			return -1, message, nil
		}
		nickrows, err := s.store.Nicknames(ctx)
		if err != nil {
			return -1, "资料设置失败", err
		}
		found := false
		for _, row := range nickrows {
			if atoi(row["gender"]) == gender && str(row["name"]) == nickname {
				found = true
				break
			}
		}
		if !found {
			return -1, "如需修改昵称，请联系客服修改", nil
		}
		if err := s.store.UpdateUserProfile(ctx, uid, gender, &nickname); err != nil {
			return -1, "资料设置失败", err
		}
		return 0, "资料设置成功", nil
	}
	if err := s.store.UpdateUserProfile(ctx, uid, gender, nil); err != nil {
		return -1, "资料设置失败", err
	}
	return 0, "资料设置成功", nil
}

func (s *Service) UserPasswdEdge(ctx context.Context, token string, passwordOld string, password string, passwordConfirm string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	user, err = s.store.UserByID(ctx, uid)
	if err != nil {
		return -1, "密码修改失败", err
	}
	if str(user["password"]) != "" && phpPassword(passwordOld+str(user["salt"])) != str(user["password"]) {
		return -1, "原密码不正确", nil
	}
	if len(password) < 6 || len(password) > 16 {
		return -1, "密码6-16位", nil
	}
	if password != passwordConfirm {
		return -1, "两次输入密码不一致", nil
	}
	return -1, "密码修改成功分支暂未迁移", nil
}

func validProfileNickname(nickname string) (bool, string) {
	lowered := strings.ToLower(nickname)
	if len(lowered) < 6 || len(lowered) > 24 || utf8.RuneCountInString(lowered) > 16 {
		return false, "昵称2-8个汉字，英文6-16个字符"
	}
	if !profileNicknamePattern.MatchString(lowered) {
		return false, "昵称只允许中英文、数字及下划线组成"
	}
	return true, ""
}

var profileNicknamePattern = regexp.MustCompile(`^[\p{Han}a-z0-9_]+$`)

func phpPassword(password string) string {
	if len(password) != 32 {
		sum := md5.Sum([]byte(password))
		password = hex.EncodeToString(sum[:])
	}
	pos := map[string]int{}
	val := map[string]int{}
	keys := map[int]string{}
	set := func(name string, source string, mask int, xor int) {
		n := phpHexByte(source)
		pos[name] = n & mask
		val[name] = (n >> 4) ^ xor
		suffix := int(name[1] - '0')
		keys[pos[name]*10+suffix] = name
	}
	set("x0", password[0:2], 0x1f, 0xf)
	set("x1", phpSubstr(password, pos["x0"], 2), 0x1f, 0xf)
	set("x2", phpSubstr(password, val["x0"], 2), 0x0f, 0xf)
	set("x3", phpSubstr(password, -1, 2), 0x1f, 0xf)
	set("x4", phpSubstr(password, pos["x3"], 2), 0x1f, 0xf)
	set("x5", phpSubstr(password, val["x3"], 2), 0x1f, 0xf)
	set("x6", phpSubstr(password, 14, 2), 0x1f, 0xf)
	set("x7", phpSubstr(password, 16, 2), 0x1f, 0xf)

	sorted := make([]int, 0, len(keys))
	for key := range keys {
		sorted = append(sorted, key)
	}
	sort.Ints(sorted)
	for i, key := range sorted {
		name := keys[key]
		insertAt := pos[name] + i
		password = password[:insertAt] + fmt.Sprintf("%x", val[name]) + password[insertAt:]
	}
	return password
}

func phpHexByte(value string) int {
	var n int
	_, _ = fmt.Sscanf(value, "%x", &n)
	return n
}

func phpSubstr(value string, start int, length int) string {
	if start < 0 {
		start = len(value) + start
	}
	if start < 0 {
		start = 0
	}
	if start > len(value) {
		return ""
	}
	end := start + length
	if end > len(value) {
		end = len(value)
	}
	return value[start:end]
}

func (s *Service) CoinLogExchangeEdge(ctx context.Context, token string, extype int, exnum int) (int, string, error) {
	exrate, err := s.store.SettingExRate(ctx)
	if err != nil {
		return -1, "金币兑换失败", err
	}
	if exrate == 0 {
		return -1, "系统已关闭兑换功能", nil
	}
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if extype != 1 && extype != 2 {
		return -1, "请指定兑换类型", nil
	}
	if exnum == 0 {
		return -1, "请指定兑换数量", nil
	}
	if exnum > 1000000 {
		return -1, "兑换数量100万以上请分次兑换", nil
	}
	if extype == 1 {
		if exnum < exrate {
			return -1, fmt.Sprintf("提交金币最小数量为:%d", exrate), nil
		}
		if exnum/exrate == 0 {
			return -1, "兑换计算所得人民币为0", nil
		}
	}
	return -1, "金币兑换成功分支暂未迁移", nil
}

func (s *Service) VODOrderCreateEdge(ctx context.Context, token string, vodserial string, vodname string, coins int) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if strings.TrimSpace(vodserial) == "" && strings.TrimSpace(vodname) == "" {
		return -1, "请填写视频番号或者视频名称", nil
	}
	if coins < 100 {
		return -1, "求片金币不能低于100", nil
	}
	quota, err := s.store.Quota(ctx, atoi(user["uid"]))
	if err != nil {
		return -1, "求片创建失败", err
	}
	goldcoin := ""
	if quota["goldcoin"] != nil {
		goldcoin = str(quota["goldcoin"])
	}
	if len(quota) == 0 || atoi(goldcoin) < coins {
		return -1, "金币不足:" + goldcoin, nil
	}
	return -1, "求片创建成功分支暂未迁移", nil
}

func (s *Service) VODOrderSupportEdge(ctx context.Context, token string, orderID int, coins int) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if orderID <= 0 {
		return -1, "您助力的求片记录不存在", nil
	}
	order, err := s.store.VODOrderByID(ctx, orderID)
	if err != nil {
		return -1, "求片助力失败", err
	}
	if len(order) == 0 || atoi(order["id"]) == 0 {
		return -1, "您助力的求片记录不存在", nil
	}
	now := s.now().Unix()
	if atoi(order["uid"]) == atoi(user["uid"]) {
		if now > int64(atoi(order["stop_time"])) {
			return -1, "该求片已停止助力", nil
		}
	} else if now < int64(atoi(order["start_time"])) || now > int64(atoi(order["stop_time"])) {
		return -1, "该求片助力时间为" + formatUnixTime(atoi(order["start_time"])) + "~" + formatUnixTime(atoi(order["stop_time"])), nil
	}
	if coins < 1 {
		return -1, "助力求片金币不能低于1", nil
	}
	quota, err := s.store.Quota(ctx, atoi(user["uid"]))
	if err != nil {
		return -1, "求片助力失败", err
	}
	goldcoin := ""
	if quota["goldcoin"] != nil {
		goldcoin = str(quota["goldcoin"])
	}
	if len(quota) == 0 || atoi(goldcoin) < coins {
		return -1, "金币不足:" + goldcoin, nil
	}
	return -1, "求片助力成功分支暂未迁移", nil
}

func formatUnixTime(ts int) string {
	return time.Unix(int64(ts), 0).Format("2006-01-02 15:04:05")
}

func validEmail(email string) bool {
	email = strings.TrimSpace(email)
	at := strings.Index(email, "@")
	dot := strings.LastIndex(email, ".")
	return at > 0 && dot > at+1 && dot < len(email)-1
}
