package ucp

import (
	"context"
	"fmt"
	"strings"
	"time"
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

func (s *Service) TaskboxOpenEdge(ctx context.Context, token string, taskID int) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
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
	return -1, "任务宝箱开启成功分支暂未迁移", nil
}

func sameDay(ts int64, now time.Time) bool {
	if ts <= 0 {
		return false
	}
	return ts >= dayStartUnix(now)
}

func (s *Service) UserEmailEdge(ctx context.Context, token string, email string, pendingMessage string) (int, string, error) {
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
	if pendingMessage == "" {
		pendingMessage = "邮箱成功分支暂未迁移"
	}
	return -1, pendingMessage, nil
}

func (s *Service) UserVerifyEmailEdge(ctx context.Context, token string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	return -1, "验证码不存在或已失效", nil
}

func (s *Service) UserBindMobiEdge(ctx context.Context, token string) (int, string, error) {
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
	return -1, "手机验证码不正确", nil
}

func (s *Service) UserProfileEdge(ctx context.Context, token string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	return -1, "资料设置成功分支暂未迁移", nil
}

func (s *Service) UserPasswdEdge(ctx context.Context, token string, password string, passwordConfirm string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if len(password) < 6 || len(password) > 16 {
		return -1, "密码6-16位", nil
	}
	if password != passwordConfirm {
		return -1, "两次输入密码不一致", nil
	}
	return -1, "密码修改成功分支暂未迁移", nil
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
