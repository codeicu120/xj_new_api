package ucp

import (
	"context"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"xj_comp/internal/domain"
	"xj_comp/internal/service/resourceurl"
)

type FeedbackCreateRequest struct {
	CID        int
	Content    string
	PayID      int
	PayName    string
	PayAccount string
	Device     string
	LongIDs    string
	ShortIDs   string
	IP         string
	FileCount  int
	Legacy     bool
}

func (s *Service) FeedbackIndex(ctx context.Context, token string) (domain.UCPFeedbackIndexData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPFeedbackIndexData{}, -1, "获取反馈信息失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPFeedbackIndexData{}, -9999, "您还没有登录", nil
	}

	since := s.now().Add(-30 * 24 * time.Hour).Unix()
	rows, err := s.store.PaymentsSince(ctx, uid, since, 100)
	if err != nil {
		return domain.UCPFeedbackIndexData{}, -1, "获取反馈信息失败", err
	}
	return domain.UCPFeedbackIndexData{PayRows: processPaymentRows(rows)}, 0, "", nil
}

func (s *Service) FeedbackListing(ctx context.Context, token string, page int) (domain.UCPFeedbackData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPFeedbackData{}, -1, "获取反馈列表失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPFeedbackData{}, -9999, "您还没有登录", nil
	}

	pageSize := 20
	total, err := s.store.CountFeedbacks(ctx, uid)
	if err != nil {
		return domain.UCPFeedbackData{}, -1, "获取反馈列表失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.Feedbacks(ctx, uid, page, pageSize)
	if err != nil {
		return domain.UCPFeedbackData{}, -1, "获取反馈列表失败", err
	}
	return domain.UCPFeedbackData{
		Rows:     processFeedbackRows(rows),
		PageInfo: pageInfo(total, pageSize, page, "/ucp/feedback?page=[?]"),
	}, 0, "", nil
}

func (s *Service) FeedbackNewListing(ctx context.Context, token string, feedbackType int, page int) (domain.UCPFeedbackData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPFeedbackData{}, -1, "获取反馈列表失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPFeedbackData{}, -9999, "您还没有登录", nil
	}
	feedbackType = normalizeFeedbackType(feedbackType)

	pageSize := 20
	total, err := s.store.CountFeedbacksByType(ctx, uid, feedbackType)
	if err != nil {
		return domain.UCPFeedbackData{}, -1, "获取反馈列表失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.FeedbacksByType(ctx, uid, feedbackType, page, pageSize)
	if err != nil {
		return domain.UCPFeedbackData{}, -1, "获取反馈列表失败", err
	}
	return domain.UCPFeedbackData{
		Rows:     processFeedbackRows(rows),
		PageInfo: pageInfo(total, pageSize, page, "/ucp/feedback/listing?type="+str(feedbackType)+"&page=[?]"),
	}, 0, "", nil
}

func (s *Service) FeedbackDetail(ctx context.Context, token string, id int, resourceReq ...resourceurl.Request) (domain.UCPFeedbackDetailData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPFeedbackDetailData{}, -1, "获取反馈详情失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPFeedbackDetailData{}, -9999, "您还没有登录", nil
	}

	row, err := s.store.FeedbackByID(ctx, id)
	if err != nil {
		return domain.UCPFeedbackDetailData{}, -1, "获取反馈详情失败", err
	}
	if len(row) == 0 || atoi(row["uid"]) != uid {
		return domain.UCPFeedbackDetailData{}, -1, "记录不存在或已被删除", nil
	}

	resources, err := s.resolveResources(ctx, resourceReq)
	if err != nil {
		return domain.UCPFeedbackDetailData{}, -1, "获取反馈详情失败", err
	}
	picURLs, err := s.feedbackPicURLs(ctx, str(row["aids"]), resources)
	if err != nil {
		return domain.UCPFeedbackDetailData{}, -1, "获取反馈详情失败", err
	}
	payrow := map[string]interface{}{}
	if payid := atoi(row["payid"]); payid > 0 {
		payrow, err = s.store.PaymentByID(ctx, payid)
		if err != nil {
			return domain.UCPFeedbackDetailData{}, -1, "获取反馈详情失败", err
		}
	}

	var pics interface{}
	if len(picURLs) > 0 {
		pics = picURLs
	}
	return domain.UCPFeedbackDetailData{
		Row:     processFeedbackRow(row, payrow),
		PicURLs: pics,
	}, 0, "", nil
}

func (s *Service) FeedbackCreate(ctx context.Context, token string, req FeedbackCreateRequest) (int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return -1, "提交反馈失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	content := strings.TrimRightFunc(req.Content, unicode.IsSpace)
	if content == "" || utf8.RuneCountInString(content) > 250 {
		return -1, "内容最多250个字符", nil
	}
	if req.PayID > 0 {
		payrow, err := s.store.PaymentByID(ctx, req.PayID)
		if err != nil {
			return -1, "提交反馈失败", err
		}
		if len(payrow) == 0 || atoi(payrow["uid"]) != uid {
			return -1, "请选择订单信息", nil
		}
	}
	if !req.Legacy && req.CID == 5 && req.PayID <= 0 {
		return -1, "请选择订单信息", nil
	}
	if req.FileCount > 5 {
		return -1, "最多不允许上传超过5张图片", nil
	}
	if !req.Legacy {
		total, err := s.store.CountFeedbacksSince(ctx, uid, dayStartUnix(s.now()))
		if err != nil {
			return -1, "提交反馈失败", err
		}
		if total > 10 {
			return -1, "当日反馈内容过多", nil
		}
	}
	id, err := s.store.CreateFeedback(ctx, domain.FeedbackCreateInput{
		UID:        uid,
		CID:        req.CID,
		Content:    content,
		PayID:      req.PayID,
		PayName:    strings.TrimSpace(req.PayName),
		PayAccount: strings.TrimSpace(req.PayAccount),
		CreatedAt:  s.now().Unix(),
		IP:         req.IP,
		Device:     strings.TrimSpace(req.Device),
		LongIDs:    strings.TrimSpace(req.LongIDs),
		ShortIDs:   strings.TrimSpace(req.ShortIDs),
	})
	if err != nil {
		return -1, "提交反馈失败", err
	}
	if id == 0 {
		return -1, "提交失败，请重试", nil
	}
	return 0, "信息已反馈", nil
}

func normalizeFeedbackType(feedbackType int) int {
	if feedbackType == 1 || feedbackType == 2 {
		return feedbackType
	}
	return 0
}

func processFeedbackRows(rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, processFeedbackRow(row, nil))
	}
	return out
}

func processFeedbackRow(row map[string]interface{}, payrow map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":         str(row["id"]),
		"cid":        str(row["cid"]),
		"content":    str(row["content"]),
		"ctimestamp": formatUnixMinute(atoi64(row["ctimestamp"])),
		"replytime":  formatOptionalUnixMinute(atoi64(row["replytime"])),
		"replytext":  str(row["replytext"]),
		"payid":      str(row["payid"]),
		"payname":    str(row["payname"]),
		"payaccount": str(row["payaccount"]),
		"itemname":   feedbackPayItemName(payrow),
		"paidtime":   formatOptionalUnixMinute(atoi64(payrow["paidtime"])),
	}
}

func feedbackPayItemName(payrow map[string]interface{}) interface{} {
	if len(payrow) == 0 {
		return nil
	}
	return payrow["itemname"]
}

func (s *Service) feedbackPicURLs(ctx context.Context, aids string, resources resourceurl.Resolved) ([]string, error) {
	ids := parseIDList(aids)
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := s.store.AttachByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, len(rows))
	for _, row := range rows {
		uri := str(row["uri"])
		if uri == "" {
			continue
		}
		urls = append(urls, resources.GetRes(uri, ""))
	}
	return urls, nil
}

func parseIDList(value string) []int {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		id := atoi(strings.TrimSpace(part))
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func (s *Service) resourceURL(uri string) string {
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return uri
	}
	if s.resourceBaseURL == "" {
		return "/" + strings.TrimLeft(uri, "/")
	}
	return s.resourceBaseURL + "/" + strings.TrimLeft(uri, "/")
}
