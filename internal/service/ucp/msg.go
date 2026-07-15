package ucp

import (
	"context"
	"regexp"
	"strings"

	"xj_comp/internal/domain"
)

var htmlHrefPattern = regexp.MustCompile(`(?i)<a.+?href\s*=\s*["']?([^'"\s<>]+)["']?.*?>`)

func (s *Service) MsgListing(ctx context.Context, token string, page int) (domain.UCPMsgListingData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPMsgListingData{}, -1, "获取消息列表失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPMsgListingData{}, -9999, "您还没有登录", nil
	}

	pageSize := 20
	total, err := s.store.CountMsgConversations(ctx, uid)
	if err != nil {
		return domain.UCPMsgListingData{}, -1, "获取消息列表失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.MsgConversations(ctx, uid, page, pageSize)
	if err != nil {
		return domain.UCPMsgListingData{}, -1, "获取消息列表失败", err
	}
	return domain.UCPMsgListingData{
		Rows:     processMsgRows(rows),
		PageInfo: pageInfo(total, pageSize, page, "/ucp/msg?page=[?]"),
	}, 0, "", nil
}

func (s *Service) MsgDetail(ctx context.Context, token string, cid int, page int) (domain.UCPMsgDetailData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPMsgDetailData{}, -1, "获取消息详情失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPMsgDetailData{}, -9999, "您还没有登录", nil
	}

	crow, err := s.store.MsgConversation(ctx, uid, cid)
	if err != nil {
		return domain.UCPMsgDetailData{}, -1, "获取消息详情失败", err
	}
	if len(crow) == 0 {
		return domain.UCPMsgDetailData{}, -1, "您的会话不存在", nil
	}
	cuser, err := s.store.UserByID(ctx, atoi(crow["ruid"]))
	if err != nil {
		return domain.UCPMsgDetailData{}, -1, "获取消息详情失败", err
	}
	var cuserValue interface{} = cuser
	if len(cuser) == 0 {
		cuserValue = []interface{}{}
	}
	pageSize := 100
	total, err := s.store.CountMessages(ctx, uid, cid)
	if err != nil {
		return domain.UCPMsgDetailData{}, -1, "获取消息详情失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.Messages(ctx, uid, cid, page, pageSize)
	if err != nil {
		return domain.UCPMsgDetailData{}, -1, "获取消息详情失败", err
	}
	if err := s.store.SetMsgRead(ctx, uid, cid); err != nil {
		return domain.UCPMsgDetailData{}, -1, "获取消息详情失败", err
	}
	return domain.UCPMsgDetailData{
		Crow:     crow,
		CUser:    cuserValue,
		Rows:     processMsgRows(rows),
		PageInfo: pageInfo(total, pageSize, page, "/ucp/msg/show?cid="+str(cid)+"&page=[?]"),
	}, 0, "", nil
}

func (s *Service) MsgSetRead(ctx context.Context, token string, cids []int) (int, string, error) {
	uid, retcode, errmsg, err := s.msgActionUser(ctx, token)
	if retcode != 0 || err != nil {
		return retcode, errmsg, err
	}
	for _, cid := range cids {
		if cid <= 0 {
			continue
		}
		if err := s.store.SetMsgRead(ctx, uid, cid); err != nil {
			return -1, "操作失败", err
		}
	}
	return 0, "操作成功", nil
}

func (s *Service) MsgCleanRead(ctx context.Context, token string) (int, string, error) {
	uid, retcode, errmsg, err := s.msgActionUser(ctx, token)
	if retcode != 0 || err != nil {
		return retcode, errmsg, err
	}
	if err := s.store.CleanMsgRead(ctx, uid); err != nil {
		return -1, "操作失败", err
	}
	return 0, "操作成功", nil
}

func (s *Service) MsgDelete(ctx context.Context, token string, cids []int) (int, string, error) {
	uid, retcode, errmsg, err := s.msgActionUser(ctx, token)
	if retcode != 0 || err != nil {
		return retcode, errmsg, err
	}
	if err := s.store.DeleteMsgConversations(ctx, uid, cids); err != nil {
		return -1, "操作失败", err
	}
	return 0, "操作成功", nil
}

func (s *Service) msgActionUser(ctx context.Context, token string) (int, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return 0, -1, "操作失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return 0, -9999, "您还没有登录", nil
	}
	return uid, 0, "", nil
}

func processMsgRows(rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := cloneMap(row)
		item["content"] = normalizeMsgContent(str(item["content"]))
		item["__url__"] = "/ucp/msg/show?cid=" + str(item["cid"])
		out = append(out, item)
	}
	return out
}

func cloneMap(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row)+1)
	for key, value := range row {
		out[key] = value
	}
	return out
}

func normalizeMsgContent(content string) string {
	return htmlHrefPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := htmlHrefPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		url := parts[1]
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "/") {
			return match
		}
		return strings.Replace(match, url, "/"+strings.TrimLeft(url, "/"), 1)
	})
}
