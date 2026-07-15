package ucp

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
)

type UserStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
	Groups(ctx context.Context) ([]map[string]interface{}, error)
	CountRecommended(ctx context.Context, uid int) (int, error)
	RecommendedUsers(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
	RollTitles(ctx context.Context) ([]map[string]interface{}, error)
	CountPayments(ctx context.Context, uid int) (int, error)
	Payments(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
	SafePayLogs(ctx context.Context, uid int, since int64, limit int) ([]map[string]interface{}, error)
	PaymentsSince(ctx context.Context, uid int, since int64, limit int) ([]map[string]interface{}, error)
	Account(ctx context.Context, uid int) (map[string]interface{}, error)
	Quota(ctx context.Context, uid int) (map[string]interface{}, error)
	Goldbean(ctx context.Context, uid int) (map[string]interface{}, error)
	CountVODPlayLogsSince(ctx context.Context, uid int, since int64) (int, error)
	CountVODDownLogsSince(ctx context.Context, uid int, since int64) (int, error)
	GuestBySID(ctx context.Context, sid string) (map[string]interface{}, error)
	CountGuestVODPlayLogsSince(ctx context.Context, sid string, since int64) (int, error)
	CountGuestVODDownLogsSince(ctx context.Context, sid string, since int64) (int, error)
	CountMiniVODViewLogsSince(ctx context.Context, uid int, since int64, action int) (int, error)
	CountGuestMiniVODViewLogsSince(ctx context.Context, sid string, since int64, action int) (int, error)
	CountCoinLogsSinceByType(ctx context.Context, uid int, coinType int, since int64) (int, error)
	SumCoinLogsSinceByType(ctx context.Context, uid int, coinType int, since int64) (int, error)
	CountVODCommentsSince(ctx context.Context, uid int, since int64, unique bool) (int, error)
	CountVODFavoritesSince(ctx context.Context, uid int, since int64) (int, error)
	CountFeedbacks(ctx context.Context, uid int) (int, error)
	Feedbacks(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
	CountFeedbacksByType(ctx context.Context, uid int, feedbackType int) (int, error)
	FeedbacksByType(ctx context.Context, uid int, feedbackType int, page int, pageSize int) ([]map[string]interface{}, error)
	FeedbackByID(ctx context.Context, id int) (map[string]interface{}, error)
	CountFeedbacksSince(ctx context.Context, uid int, since int64) (int, error)
	CreateFeedback(ctx context.Context, input domain.FeedbackCreateInput) (int, error)
	PaymentByID(ctx context.Context, payid int) (map[string]interface{}, error)
	AttachByIDs(ctx context.Context, ids []int) ([]map[string]interface{}, error)
	CountMsgConversations(ctx context.Context, uid int) (int, error)
	MsgConversations(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
	MsgConversation(ctx context.Context, uid int, cid int) (map[string]interface{}, error)
	UserByID(ctx context.Context, uid int) (map[string]interface{}, error)
	BotByID(ctx context.Context, uid int) (map[string]interface{}, error)
	Bankcards(ctx context.Context, uid int) ([]map[string]interface{}, error)
	Banks(ctx context.Context) ([]map[string]interface{}, error)
	BankcardByID(ctx context.Context, uid int, cardID int) (map[string]interface{}, error)
	CreateBankcard(ctx context.Context, uid int, name string, bankname string, cardnum string, isdef int, cardType int) (int, error)
	UpdateBankcard(ctx context.Context, uid int, cardID int, name string, bankname string, cardnum string, isdef int, cardType int) (int, error)
	DeleteBankcard(ctx context.Context, uid int, cardID int) (int, error)
	SetDefaultBankcard(ctx context.Context, uid int, cardID int) error
	CountMessages(ctx context.Context, uid int, cid int) (int, error)
	Messages(ctx context.Context, uid int, cid int, page int, pageSize int) ([]map[string]interface{}, error)
	SetMsgRead(ctx context.Context, uid int, cid int) error
	CleanMsgRead(ctx context.Context, uid int) error
	DeleteMsgConversations(ctx context.Context, uid int, cids []int) error
	SendMessage(ctx context.Context, senderID int, receiverID int, content string, cid int, now int64) (int, error)
	CountBalanceLogs(ctx context.Context, uid int) (int, error)
	BalanceLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
	CountWithdraws(ctx context.Context, uid int) (int, error)
	CountWithdrawsSince(ctx context.Context, uid int, since int64) (int, error)
	Withdraws(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
	SumWithdrawAmount(ctx context.Context, uid int) (int, error)
	CoinLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
	CountCoinLogsByTypes(ctx context.Context, uid int, coinTypes []int) (int, error)
	CoinLogsByTypes(ctx context.Context, uid int, coinTypes []int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
	CoinBonusStats(ctx context.Context, uid int) (map[string]interface{}, error)
	SettingExRate(ctx context.Context) (int, error)
	SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
	CalldataByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
	PackageRows(ctx context.Context, kind string) ([]map[string]interface{}, error)
	PackageByID(ctx context.Context, kind string, pkgID int) (map[string]interface{}, error)
	PaymentChannels(ctx context.Context, gameOnly bool) ([]map[string]interface{}, error)
	CountVODOrders(ctx context.Context, uid int, status *int) (int, error)
	VODOrders(ctx context.Context, uid int, status *int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
	VODOrderByID(ctx context.Context, orderID int) (map[string]interface{}, error)
	LatestVODIssue(ctx context.Context) (map[string]interface{}, error)
	CountVODOrdersByCreateTime(ctx context.Context, start int64, end int64) (int, error)
	VODOrdersByCreateTime(ctx context.Context, start int64, end int64, page int, pageSize int) ([]map[string]interface{}, error)
	SumVODOrderCoins(ctx context.Context, uid int, status int) (int, error)
	CountVODSupports(ctx context.Context, uid int) (int, error)
	VODSupports(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
	MaxVODSupport(ctx context.Context, orderID int) (map[string]interface{}, error)
	MyVODSupportCoins(ctx context.Context, orderID int, uid int) (int, error)
	SumVODSupportCoins(ctx context.Context, uid int, onlyFrozen bool) (int, error)
	Posters(ctx context.Context) ([]map[string]interface{}, error)
	Taskboxes(ctx context.Context) ([]map[string]interface{}, error)
	TaskboxByID(ctx context.Context, taskID int) (map[string]interface{}, error)
	TaskboxLog(ctx context.Context, uid int, taskID int, dayKey int) (map[string]interface{}, error)
	TaskboxCompletedLogs(ctx context.Context, limit int) ([]map[string]interface{}, error)
	CountTaskboxLogs(ctx context.Context, uid int) (int, error)
	TaskboxLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
}

type Service struct {
	store           UserStore
	resourceBaseURL string
	now             func() time.Time
}

func NewService(store UserStore, resourceBaseURL string) *Service {
	return &Service{
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
		now:             time.Now,
	}
}

func (s *Service) MyAff(ctx context.Context, token string, page int) (domain.UCPMyAffData, int, string, error) {
	user, groups, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return domain.UCPMyAffData{}, -1, "获取用户失败", err
	}
	if atoi(user["uid"]) == 0 {
		return domain.UCPMyAffData{}, -9999, "请登录后操作", nil
	}

	pageSize := 20
	total, err := s.store.CountRecommended(ctx, atoi(user["uid"]))
	if err != nil {
		return domain.UCPMyAffData{}, -1, "获取推广列表失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.RecommendedUsers(ctx, atoi(user["uid"]), page, pageSize)
	if err != nil {
		return domain.UCPMyAffData{}, -1, "获取推广列表失败", err
	}

	return domain.UCPMyAffData{
		Rows:     s.processUsers(rows, groups),
		PageInfo: pageInfo(total, pageSize, page, "/ucp/myaff?page=[?]"),
	}, 0, "", nil
}

func (s *Service) RollTitle(ctx context.Context) (domain.UCPRollTitleData, error) {
	rows, err := s.store.RollTitles(ctx)
	if err != nil {
		return domain.UCPRollTitleData{}, fmt.Errorf("list roll titles: %w", err)
	}
	return domain.UCPRollTitleData{Messages: rows}, nil
}

func (s *Service) AffCenter(ctx context.Context, token string) (domain.UCPAffCenterData, int, string, error) {
	user, groups, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return domain.UCPAffCenterData{}, -1, "获取用户失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPAffCenterData{}, -9999, "您还没有登录", nil
	}

	quota, err := s.store.Quota(ctx, uid)
	if err != nil {
		return domain.UCPAffCenterData{}, -1, "获取推广中心失败", err
	}
	goldbean, err := s.store.Goldbean(ctx, uid)
	if err != nil {
		return domain.UCPAffCenterData{}, -1, "获取推广中心失败", err
	}
	daytime := dayStartUnix(s.now())
	playedNum, err := s.store.CountVODPlayLogsSince(ctx, uid, daytime)
	if err != nil {
		return domain.UCPAffCenterData{}, -1, "获取推广中心失败", err
	}
	downedNum, err := s.store.CountVODDownLogsSince(ctx, uid, daytime)
	if err != nil {
		return domain.UCPAffCenterData{}, -1, "获取推广中心失败", err
	}

	user["goldcoin"] = quota["goldcoin"]
	user["gold_bean"] = goldbean["gold_bean"]

	uinfo := s.affCenterInfo(user, groups, playedNum, downedNum)
	return domain.UCPAffCenterData{
		User:  singleUser(s.processUsers([]map[string]interface{}{user}, groups)),
		UInfo: uinfo,
	}, 0, "", nil
}

func (s *Service) Index(ctx context.Context, token string) (domain.UCPIndexData, int, string, error) {
	user, groups, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return domain.UCPIndexData{}, -1, "获取个人中心失败", err
	}
	uid := atoi(user["uid"])
	daytime := dayStartUnix(s.now())

	playCount, downCount, miniPlayCount, miniDownCount, err := s.indexUsageCounts(ctx, user, uid, daytime)
	if err != nil {
		return domain.UCPIndexData{}, -1, "获取个人中心失败", err
	}
	uinfo := s.indexInfo(user, groups, playCount, downCount, miniPlayCount, miniDownCount)

	if uid == 0 {
		guest, err := s.store.GuestBySID(ctx, str(user["sid"]))
		if err != nil {
			return domain.UCPIndexData{}, -1, "获取个人中心失败", err
		}
		if len(guest) == 0 {
			return domain.UCPIndexData{}, -1, "请登录后操作，客户端游客请先携带信息", nil
		}
		uinfo["goldcoin"] = guest["goldcoin"]
		uinfo["curr_group"] = nil
		uinfo["next_group"] = nil
		return domain.UCPIndexData{
			User:   singleUser(s.processUsers([]map[string]interface{}{user}, groups)),
			UInfo:  uinfo,
			Signed: signedByTimestamp(s.now(), atoi64(guest["signtime"])),
		}, 0, "", nil
	}

	quota, err := s.store.Quota(ctx, uid)
	if err != nil {
		return domain.UCPIndexData{}, -1, "获取个人中心失败", err
	}
	goldbean, err := s.store.Goldbean(ctx, uid)
	if err != nil {
		return domain.UCPIndexData{}, -1, "获取个人中心失败", err
	}
	user["goldcoin"] = quota["goldcoin"]
	user["gold_bean"] = goldbean["gold_bean"]
	uinfo["goldcoin"] = quota["goldcoin"]
	uinfo["gold_bean"] = goldbean["gold_bean"]

	signedCount, err := s.store.CountCoinLogsSinceByType(ctx, uid, 1, daytime)
	if err != nil {
		return domain.UCPIndexData{}, -1, "获取个人中心失败", err
	}
	userRow := singleUser(s.processUsers([]map[string]interface{}{user}, groups))
	clearTildeContact(userRow, "mobi")
	clearTildeContact(userRow, "email")
	return domain.UCPIndexData{
		User:   userRow,
		UInfo:  uinfo,
		Signed: boolInt(signedCount > 0),
		Groups: indexGroups(groups),
	}, 0, "", nil
}

func (s *Service) UserIndex(ctx context.Context, token string) (map[string]interface{}, int, string, error) {
	authUser, groups, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(authUser["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	user, err := s.store.UserByID(ctx, uid)
	if err != nil {
		return nil, -1, "获取用户资料失败", err
	}
	return map[string]interface{}{"user": singleUser(s.processUsers([]map[string]interface{}{user}, groups))}, 0, "", nil
}

func (s *Service) BankcardIndex(ctx context.Context, token string) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	cardRows, err := s.store.Bankcards(ctx, uid)
	if err != nil {
		return nil, -1, "获取银行卡失败", err
	}
	bankRows, err := s.store.Banks(ctx)
	if err != nil {
		return nil, -1, "获取银行卡失败", err
	}
	return map[string]interface{}{
		"cardrows":  cardRows,
		"maxallow":  3,
		"allowtype": 7,
		"banknames": []string{"工商银行", "建设银行", "中国银行", "农业银行", "交通银行", "招商银行", "中信银行", "上海浦东发展银行", "兴业银行", "民生银行"},
		"bankRows":  s.processBankRows(bankRows),
	}, 0, "", nil
}

type BankcardPostRequest struct {
	Action   string
	CardID   int
	Name     string
	BankName string
	CardNum  string
	IsDef    int
	Type     int
}

func (s *Service) BankcardPost(ctx context.Context, token string, req BankcardPostRequest) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	req.Name = strings.TrimSpace(req.Name)
	req.BankName = strings.TrimSpace(req.BankName)
	req.CardNum = strings.TrimSpace(req.CardNum)
	if req.Type == 0 || req.Type == 1 {
		req.BankName = "支付宝"
	} else if req.Type == 3 {
		req.BankName = "微信"
	}
	if req.Action == "create" {
		rows, err := s.store.Bankcards(ctx, uid)
		if err != nil {
			return -1, "操作失败", err
		}
		if len(rows) >= 5 {
			return -1, "最多可以设置3个地址", nil
		}
	} else {
		row, err := s.store.BankcardByID(ctx, uid, req.CardID)
		if err != nil {
			return -1, "操作失败", err
		}
		if len(row) == 0 {
			return -1, "修改的记录不存在", nil
		}
	}
	if req.Name == "" || len([]rune(req.Name)) > 20 {
		return -1, "姓名长度不正确", nil
	}
	if req.BankName == "" || len([]rune(req.BankName)) > 40 {
		return -1, "开户银行填写不正确", nil
	}
	if req.CardNum == "" || len([]rune(req.CardNum)) > 40 {
		return -1, "收款账户或卡号填写不正确", nil
	}
	cardID := req.CardID
	if req.Action == "create" {
		cardID, err = s.store.CreateBankcard(ctx, uid, req.Name, req.BankName, req.CardNum, req.IsDef, req.Type)
	} else {
		_, err = s.store.UpdateBankcard(ctx, uid, req.CardID, req.Name, req.BankName, req.CardNum, req.IsDef, req.Type)
	}
	if err != nil {
		return -1, "操作失败", err
	}
	if req.IsDef > 0 {
		if err := s.store.SetDefaultBankcard(ctx, uid, cardID); err != nil {
			return -1, "操作失败", err
		}
	}
	return 0, "操作成功", nil
}

func (s *Service) BankcardDelete(ctx context.Context, token string, cardID int) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	if _, err := s.store.DeleteBankcard(ctx, uid, cardID); err != nil {
		return -1, "操作失败", err
	}
	return 0, "操作成功", nil
}

func (s *Service) authenticatedUser(ctx context.Context, token string) (map[string]interface{}, []map[string]interface{}, error) {
	groups, err := s.store.Groups(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list user groups: %w", err)
	}
	sid := userRepo.CleanToken(token)
	user, err := s.store.UserBySession(ctx, sid)
	if err != nil {
		return nil, nil, fmt.Errorf("load user by session: %w", err)
	}
	if user == nil {
		user = map[string]interface{}{"uid": "0", "sid": sid}
	}
	if atoi(user["uid"]) > 0 {
		user["perms"] = initPerm(initGids(user, s.now), groups)
	} else {
		user["perms"] = initPerm([]int{0}, groups)
	}
	return user, groups, nil
}

func (s *Service) indexUsageCounts(ctx context.Context, user map[string]interface{}, uid int, daytime int64) (int, int, int, int, error) {
	if uid > 0 {
		played, err := s.store.CountVODPlayLogsSince(ctx, uid, daytime)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		downed, err := s.store.CountVODDownLogsSince(ctx, uid, daytime)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		miniPlayed, err := s.store.CountMiniVODViewLogsSince(ctx, uid, daytime, 1)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		miniDowned, err := s.store.CountMiniVODViewLogsSince(ctx, uid, daytime, 2)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		return played, downed, miniPlayed, miniDowned, nil
	}
	sid := str(user["sid"])
	played, err := s.store.CountGuestVODPlayLogsSince(ctx, sid, daytime)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	downed, err := s.store.CountGuestVODDownLogsSince(ctx, sid, daytime)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	miniPlayed, err := s.store.CountGuestMiniVODViewLogsSince(ctx, sid, daytime, 1)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	miniDowned, err := s.store.CountGuestMiniVODViewLogsSince(ctx, sid, daytime, 2)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return played, downed, miniPlayed, miniDowned, nil
}

func initGids(user map[string]interface{}, now func() time.Time) []int {
	mainGID := atoi(user["gid"])
	if atoi(user["sysgid"]) > 0 {
		mainGID = atoi(user["sysgid"])
	}
	gids := []int{mainGID}
	var extra map[string]interface{}
	switch typed := user["gids"].(type) {
	case map[string]interface{}:
		extra = typed
	case string:
		if typed != "" {
			_ = json.Unmarshal([]byte(typed), &extra)
		}
	}
	ts := now().Unix()
	for gid, exptime := range extra {
		if atoi(exptime) == 0 || atoi64(exptime) > ts {
			gids = append(gids, atoi(gid))
		}
	}
	return uniqueInts(gids)
}

func initPerm(gids []int, groups []map[string]interface{}) map[string]interface{} {
	selected := make([]map[string]interface{}, 0, len(gids))
	for _, gid := range gids {
		for _, group := range groups {
			if atoi(group["scope"]) > 0 || atoi(group["gid"]) != gid {
				continue
			}
			selected = append(selected, group)
			break
		}
	}
	sort.SliceStable(selected, func(i, j int) bool {
		return atoi(selected[i]["weight"]) > atoi(selected[j]["weight"])
	})
	multiPerms := make([]map[string]interface{}, 0, len(selected))
	for _, group := range selected {
		multiPerms = append(multiPerms, parsePermMap(group["perms"]))
	}
	return computePerm(multiPerms)
}

func parsePermMap(value interface{}) map[string]interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed
	case string:
		if typed == "" {
			return map[string]interface{}{}
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(typed), &parsed); err != nil {
			return map[string]interface{}{}
		}
		return parsed
	default:
		return map[string]interface{}{}
	}
}

func computePerm(multiPerms []map[string]interface{}) map[string]interface{} {
	keys := make([]string, 0)
	seen := map[string]struct{}{}
	for _, perms := range multiPerms {
		for key := range perms {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
	}

	result := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		switch strings.SplitN(key, ".", 2)[0] {
		case "allow", "deny":
			value := 0
			for _, perms := range multiPerms {
				if atoi(perms[key]) == 1 {
					value = 1
					break
				}
			}
			result[key] = value
		case "min":
			value := 0
			minValue := 0
			for _, perms := range multiPerms {
				if _, ok := perms[key]; ok {
					minValue = minInt(minValue, atoi(perms[key]))
					value = minValue
				}
			}
			result[key] = value
		case "max":
			value := 0
			maxValue := 0
			for _, perms := range multiPerms {
				if _, ok := perms[key]; ok {
					maxValue = maxInt(maxValue, atoi(perms[key]))
					value = maxValue
				}
			}
			result[key] = value
		case "list":
			value := ""
			for _, perms := range multiPerms {
				if str(perms[key]) == "" {
					continue
				}
				if value == "" {
					value = str(perms[key])
				} else {
					value += "," + str(perms[key])
				}
			}
			result[key] = value
		case "range":
			value := "0-0"
			minValue := 0
			maxValue := 0
			for _, perms := range multiPerms {
				if _, ok := perms[key]; !ok {
					continue
				}
				parts := strings.SplitN(str(perms[key]), "-", 2)
				if len(parts) != 2 {
					continue
				}
				minValue = minInt(minValue, atoi(parts[0]))
				maxValue = maxInt(maxValue, atoi(parts[1]))
				value = fmt.Sprintf("%d-%d", minValue, maxValue)
			}
			result[key] = value
		case "key":
			var value interface{}
			for _, perms := range multiPerms {
				if item, ok := perms[key]; ok {
					value = item
					break
				}
			}
			result[key] = value
		case "min0":
			value := 0
			minValue := 0
			for _, perms := range multiPerms {
				if _, ok := perms[key]; !ok {
					continue
				}
				if atoi(perms[key]) == 0 {
					value = atoi(perms[key])
					break
				}
				minValue = minInt(minValue, atoi(perms[key]))
				value = minValue
			}
			result[key] = value
		case "max0":
			value := 0
			maxValue := 0
			for _, perms := range multiPerms {
				if _, ok := perms[key]; !ok {
					continue
				}
				if atoi(perms[key]) == 0 {
					value = atoi(perms[key])
					break
				}
				maxValue = maxInt(maxValue, atoi(perms[key]))
				value = maxValue
			}
			result[key] = value
		case "key0":
			value := interface{}("")
			found := false
			for _, perms := range multiPerms {
				if item, ok := perms[key]; ok && (str(item) == "" || atoi(item) == 0) {
					value = item
					found = true
					break
				}
			}
			if !found {
				for _, perms := range multiPerms {
					if item, ok := perms[key]; ok {
						value = item
						break
					}
				}
			}
			result[key] = value
		}
	}
	return result
}

func (s *Service) affCenterInfo(user map[string]interface{}, groups []map[string]interface{}, playedNum int, downedNum int) map[string]interface{} {
	playDayNum := getPermInt(user["perms"], "max.vod.play.daynum")
	downDayNum := getPermInt(user["perms"], "max.vod.down.daynum")
	uinfo := map[string]interface{}{
		"goldcoin":              str(user["goldcoin"]),
		"play_daily_remainders": maxInt(playDayNum-playedNum, 0),
		"down_daily_remainders": maxInt(downDayNum-downedNum, 0),
		"curr_group":            []interface{}{},
		"next_group":            []interface{}{},
		"next_upgrade_need":     0,
		"gold_bean":             str(user["gold_bean"]),
	}

	sort.SliceStable(groups, func(i, j int) bool {
		return atoi(groups[i]["minup"]) < atoi(groups[j]["minup"])
	})
	mygid := atoi(user["gid"])
	if atoi(user["sysgid"]) > 0 {
		mygid = atoi(user["sysgid"])
	}
	for i, group := range groups {
		if atoi(group["gid"]) != mygid {
			continue
		}
		uinfo["curr_group"] = affGroup(group)
		if i+1 < len(groups) {
			next := affGroup(groups[i+1])
			uinfo["next_group"] = next
			need := atoi(next["minup"]) - atoi(user["recommend_total"])
			uinfo["next_upgrade_need"] = maxInt(need, 0)
		}
		break
	}
	return uinfo
}

func affGroup(group map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"gid":   str(group["gid"]),
		"gname": str(group["gname"]),
		"minup": str(group["minup"]),
	}
}

func (s *Service) indexInfo(user map[string]interface{}, groups []map[string]interface{}, playedNum int, downedNum int, miniPlayedNum int, miniDownedNum int) map[string]interface{} {
	playDayNum := getPermInt(user["perms"], "max.vod.play.daynum")
	downDayNum := getPermInt(user["perms"], "max.vod.down.daynum")
	miniPlayDayNum := getPermInt(user["perms"], "max.minivod.play.daynum")
	miniDownDayNum := getPermInt(user["perms"], "max.minivod.down.daynum")
	uinfo := map[string]interface{}{
		"goldcoin":                      0,
		"play_daily_remainders":         maxInt(playDayNum-playedNum, 0),
		"down_daily_remainders":         maxInt(downDayNum-downedNum, 0),
		"curr_group":                    []interface{}{},
		"next_group":                    []interface{}{},
		"next_upgrade_need":             0,
		"minivod_play_daily_remainders": maxInt(miniPlayDayNum-miniPlayedNum, 0),
		"minivod_down_daily_remainders": maxInt(miniDownDayNum-miniDownedNum, 0),
	}

	sort.SliceStable(groups, func(i, j int) bool {
		return atoi(groups[i]["minup"]) < atoi(groups[j]["minup"])
	})
	mygid := atoi(user["gid"])
	if atoi(user["sysgid"]) > 0 {
		mygid = atoi(user["sysgid"])
	}
	for i, group := range groups {
		if atoi(group["gid"]) != mygid {
			continue
		}
		uinfo["curr_group"] = indexGroup(group)
		if i+1 < len(groups) {
			next := indexGroup(groups[i+1])
			uinfo["next_group"] = next
			need := atoi(next["minup"]) - atoi(user["recommend_total"])
			uinfo["next_upgrade_need"] = maxInt(need, 0)
		}
		break
	}
	return uinfo
}

func indexGroup(group map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"gid":   str(group["gid"]),
		"gname": str(group["gname"]),
		"gicon": str(group["gicon"]),
		"minup": str(group["minup"]),
	}
}

func indexGroups(groups []map[string]interface{}) []map[string]interface{} {
	sorted := append([]map[string]interface{}(nil), groups...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return atoi(sorted[i]["minup"]) < atoi(sorted[j]["minup"])
	})
	out := make([]map[string]interface{}, 0, len(sorted))
	for _, group := range sorted {
		if str(group["gicon"]) == "" {
			continue
		}
		perms := parsePermMap(group["perms"])
		out = append(out, map[string]interface{}{
			"gname":               str(group["gname"]),
			"gicon":               str(group["gicon"]),
			"minup":               str(group["minup"]),
			"play_daynum":         atoi(perms["max.vod.play.daynum"]),
			"down_daynum":         atoi(perms["max.vod.down.daynum"]),
			"comment_daynum":      atoi(perms["max.comment.post.daynum"]),
			"minivod_play_daynum": atoi(perms["max.minivod.play.daynum"]),
			"minivod_down_daynum": atoi(perms["max.minivod.down.daynum"]),
		})
	}
	return out
}

func signedByTimestamp(now time.Time, timestamp int64) int {
	if timestamp <= 0 {
		return 0
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	nowLocal := now.In(loc)
	signedLocal := time.Unix(timestamp, 0).In(loc)
	if nowLocal.Year() == signedLocal.Year() && nowLocal.YearDay() == signedLocal.YearDay() {
		return 1
	}
	return 0
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func clearTildeContact(row map[string]interface{}, key string) {
	value := str(row[key])
	if strings.HasPrefix(value, "~") {
		row[key] = ""
	}
}

func getPermInt(perms interface{}, key string) int {
	var values map[string]interface{}
	switch typed := perms.(type) {
	case map[string]interface{}:
		values = typed
	case string:
		if typed == "" {
			return 0
		}
		if err := json.Unmarshal([]byte(typed), &values); err != nil {
			return 0
		}
	default:
		return 0
	}
	return atoi(values[key])
}

func dayStartUnix(now time.Time) int64 {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	local := now.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc).Unix()
}

func singleUser(users []map[string]interface{}) map[string]interface{} {
	if len(users) == 0 {
		return map[string]interface{}{}
	}
	return users[0]
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func uniqueInts(values []int) []int {
	out := make([]int, 0, len(values))
	seen := map[int]struct{}{}
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func (s *Service) processUsers(rows []map[string]interface{}, groups []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	now := s.now().Unix()
	for _, row := range rows {
		sysgidExptime := atoi64(row["sysgid_exptime"])
		duetime := ""
		dueday := ""
		if sysgidExptime > 0 {
			duetime = formatUnix(sysgidExptime)
			if remaining := sysgidExptime - now; remaining > 0 {
				dueday = formatRemain(remaining) + "过期"
			} else {
				dueday = "已过期"
			}
		}
		out = append(out, map[string]interface{}{
			"uid":             str(row["uid"]),
			"uniqkey":         strings.ToUpper(strconv.FormatInt(int64(atoi(row["uniqkey"])), 36)),
			"username":        str(row["username"]),
			"nickname":        str(row["nickname"]),
			"mobi":            str(row["mobi"]),
			"email":           str(row["email"]),
			"sysgid":          str(row["sysgid"]),
			"gid":             str(row["gid"]),
			"gids":            nil,
			"gicon":           gicon(row, groups),
			"isvip":           vip(row, now),
			"regtime":         formatUnix(atoi64(row["regtime"])),
			"gender":          atoi(row["gender"]),
			"avatar":          str(row["avatar"]),
			"avatar_url":      s.avatarURL(str(row["avatar"])),
			"newmsg":          str(row["newmsg"]),
			"goldcoin":        atoi(row["goldcoin"]),
			"gold_bean":       atoi(row["gold_bean"]),
			"duetime":         duetime,
			"dueday":          dueday,
			"recommend_total": atoi(row["recommend_total"]),
		})
	}
	return out
}

func (s *Service) processBankRows(rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		coverpic := str(row["coverpic"])
		if coverpic != "" && !strings.HasPrefix(coverpic, "http://") && !strings.HasPrefix(coverpic, "https://") {
			coverpic = s.resourceBaseURL + "/" + strings.TrimLeft(coverpic, "/")
		}
		out = append(out, map[string]interface{}{
			"bankid":   atoi(row["bankid"]),
			"bankname": str(row["bankname"]),
			"coverpic": coverpic,
		})
	}
	return out
}

func (s *Service) avatarURL(avatar string) string {
	if avatar == "" {
		return s.resourceBaseURL + "/sysavatar/noavatar.png"
	}
	if strings.HasPrefix(avatar, "http://") || strings.HasPrefix(avatar, "https://") {
		return avatar
	}
	if strings.HasPrefix(avatar, "sysavatar/") {
		return s.resourceBaseURL + "/" + strings.TrimLeft(avatar, "/")
	}
	return s.resourceBaseURL + "/C1/avatar/" + strings.TrimLeft(avatar, "/")
}

func pageInfo(total int, pageSize int, page int, url string) map[string]interface{} {
	if total < 0 {
		total = 0
	}
	if pageSize < 1 {
		pageSize = 1
	}
	totalPage := totalPages(total, pageSize)
	page = normalizePage(total, pageSize, page)
	start := 0
	if total > 0 {
		start = (page-1)*pageSize + 1
	}
	end := start + pageSize - 1
	if end > total {
		end = total
	}
	currURL := strings.ReplaceAll(url, "[?]", strconv.Itoa(page))
	firstURL := strings.ReplaceAll(url, "[?]", "1")
	prevPage := 0
	if page > 1 {
		prevPage = page - 1
	}
	nextPage := 0
	if page < totalPage {
		nextPage = page + 1
	}
	prevURLPage := 1
	if prevPage > 0 {
		prevURLPage = prevPage
	}
	nextURLPage := totalPage
	if nextPage > 0 {
		nextURLPage = nextPage
	}
	return map[string]interface{}{
		"plist":     plist(page, totalPage, url),
		"pagesize":  pageSize,
		"total":     total,
		"totalpage": totalPage,
		"page":      page,
		"start":     start,
		"end":       end,
		"prev":      prevPage,
		"next":      nextPage,
		"curr_url":  currURL,
		"first_url": firstURL,
		"prev_url":  strings.ReplaceAll(url, "[?]", strconv.Itoa(prevURLPage)),
		"next_url":  strings.ReplaceAll(url, "[?]", strconv.Itoa(nextURLPage)),
		"last_url":  strings.ReplaceAll(url, "[?]", strconv.Itoa(totalPage)),
		"page_url":  url,
		"pages":     pageSelector(page, totalPage),
	}
}

func plist(page int, totalPage int, pageURL string) []map[string]interface{} {
	len0 := 5
	len1 := 4
	pages := []int{}
	outnum0 := 0
	page0 := 0
	p := page
	for i := 0; i < len0; i++ {
		p--
		if p > 0 {
			pages = append(pages, p)
			page0 = p
		} else {
			outnum0++
		}
	}
	for i, j := 0, len(pages)-1; i < j; i, j = i+1, j-1 {
		pages[i], pages[j] = pages[j], pages[i]
	}
	pages = append(pages, page)

	outnum1 := 0
	page1 := 0
	p = page
	for i := 0; i < len1; i++ {
		p++
		if p > totalPage {
			outnum1++
		} else {
			pages = append(pages, p)
			page1 = p
		}
	}
	if outnum0 > 0 && outnum1 == 0 {
		p = page1
		for i := 0; i < outnum0; i++ {
			p++
			if p > totalPage {
				break
			}
			pages = append(pages, p)
		}
	} else if outnum0 == 0 && outnum1 > 0 {
		p = page0
		for i := 0; i < outnum1; i++ {
			p--
			if p < 1 {
				break
			}
			pages = append([]int{p}, pages...)
		}
	}

	result := []map[string]interface{}{}
	if page-len0 > 1 {
		result = append(result, pageLink("first", 1, "FirstPage", pageURL))
		if page-len0 > 2 {
			result = append(result, pageLink("more", 0, "...", ""))
		}
	}
	if page0 > 1 {
		result = append(result, pageLink("prev", page-1, "PrevPage", pageURL))
	}
	for _, p := range pages {
		pos := ""
		if p == page {
			pos = "curr"
		}
		result = append(result, pageLink(pos, p, p, pageURL))
	}
	if page1 > 0 && page1 < totalPage {
		result = append(result, pageLink("next", page+1, "NextPage", pageURL))
	}
	if totalPage-page > len1 {
		if totalPage-page > len1+1 {
			result = append(result, pageLink("more", 0, "...", ""))
		}
		result = append(result, pageLink("last", totalPage, "LastPage", pageURL))
	}
	return result
}

func pageLink(pos string, page int, text interface{}, pageURL string) map[string]interface{} {
	urlValue := ""
	if pageURL != "" {
		urlValue = strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(page))
	}
	return map[string]interface{}{"pos": pos, "page": page, "text": text, "url": urlValue}
}

func pageSelector(pageNow int, totalPage int) []int {
	showAll := 50
	sliceStart := 5
	sliceEnd := 5
	percent := 20
	rangeSize := 10
	if totalPage < showAll {
		pages := make([]int, 0, totalPage)
		for i := 1; i <= totalPage; i++ {
			pages = append(pages, i)
		}
		return pages
	}
	pages := []int{}
	for i := 1; i <= sliceStart; i++ {
		pages = append(pages, i)
	}
	for i := totalPage - sliceEnd; i <= totalPage; i++ {
		pages = append(pages, i)
	}

	increment := int(math.Floor(float64(totalPage) / float64(percent)))
	if increment < 1 {
		increment = 1
	}
	pageNowMinusRange := pageNow - rangeSize
	pageNowPlusRange := pageNow + rangeSize
	i := sliceStart
	x := totalPage - sliceEnd
	metBoundary := false
	for i <= x {
		if i >= pageNowMinusRange && i <= pageNowPlusRange {
			i++
			metBoundary = true
		} else {
			i += increment
			if i > pageNowMinusRange && !metBoundary {
				i = pageNowMinusRange
			}
		}
		if i > 0 && i <= x {
			pages = append(pages, i)
		}
	}

	i = pageNow
	dist := 1
	for i < x {
		dist *= 2
		i = pageNow + dist
		if i > 0 && i <= x {
			pages = append(pages, i)
		}
	}

	i = pageNow
	dist = 1
	for i > 0 {
		dist *= 2
		i = pageNow - dist
		if i > 0 && i <= x {
			pages = append(pages, i)
		}
	}

	sort.Ints(pages)
	unique := pages[:0]
	var last int
	for idx, page := range pages {
		if idx == 0 || page != last {
			unique = append(unique, page)
			last = page
		}
	}
	return unique
}

func normalizePage(total int, pageSize int, page int) int {
	totalPage := totalPages(total, pageSize)
	if page < 1 {
		page = 1
	}
	if page > totalPage {
		page = totalPage
	}
	return page
}

func totalPages(total int, pageSize int) int {
	if pageSize < 1 {
		pageSize = 1
	}
	totalPage := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPage < 1 {
		totalPage = 1
	}
	return totalPage
}

func gicon(row map[string]interface{}, groups []map[string]interface{}) string {
	gid := atoi(row["gid"])
	if atoi(row["sysgid"]) > 0 {
		gid = atoi(row["sysgid"])
	}
	for _, group := range groups {
		if atoi(group["gid"]) == gid {
			return str(group["gicon"])
		}
	}
	return ""
}

func vip(row map[string]interface{}, now int64) int {
	if atoi(row["sysgid"]) == 6 && atoi64(row["sysgid_exptime"]) > now {
		return 1
	}
	return 0
}

func formatUnix(ts int64) string {
	if ts <= 0 {
		return "1970-01-01 08:00:00"
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Unix(ts, 0).In(loc).Format("2006-01-02 15:04:05")
}

func formatRemain(seconds int64) string {
	days := seconds / 86400
	seconds %= 86400
	hours := seconds / 3600
	seconds %= 3600
	minutes := seconds / 60
	seconds %= 60
	if days > 0 {
		return fmt.Sprintf("%d天后%d小时后%d分钟后%d秒后", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时后%d分钟后%d秒后", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d分钟后%d秒后", minutes, seconds)
	}
	return fmt.Sprintf("%d秒后", seconds)
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func atoi(value interface{}) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(str(value)))
	return parsed
}

func atoi64(value interface{}) int64 {
	parsed, _ := strconv.ParseInt(strings.TrimSpace(str(value)), 10, 64)
	return parsed
}
