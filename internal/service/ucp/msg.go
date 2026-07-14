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
