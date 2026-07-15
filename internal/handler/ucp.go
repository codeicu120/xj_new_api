package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	ucpService "xj_comp/internal/service/ucp"
)

type UCPHandler struct {
	service *ucpService.Service
}

func NewUCPHandler(service *ucpService.Service) *UCPHandler {
	return &UCPHandler{service: service}
}

func (h *UCPHandler) MyAff(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	data, retcode, errmsg, err := h.service.MyAff(c.Request.Context(), authToken(c), page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode == -1 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) RollTitle(c *gin.Context) {
	data, err := h.service.RollTitle(c.Request.Context())
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取滚动消息失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) TaskSharePic(c *gin.Context) {
	data, err := h.service.TaskSharePic(c.Request.Context())
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取分享图片失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) TaskboxIndex(c *gin.Context) {
	data, err := h.service.TaskboxIndex(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取任务宝箱失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) TaskboxLog(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.TaskboxLogListing(c.Request.Context(), authToken(c), page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) AffCenter(c *gin.Context) {
	data, retcode, errmsg, err := h.service.AffCenter(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode == -1 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) Index(c *gin.Context) {
	data, retcode, errmsg, err := h.service.Index(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode == -1 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) UserIndex(c *gin.Context) {
	data, retcode, errmsg, err := h.service.UserIndex(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) BankcardIndex(c *gin.Context) {
	data, retcode, errmsg, err := h.service.BankcardIndex(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) FeedbackListing(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		h.FeedbackCreateLegacy(c)
		return
	}
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.FeedbackListing(c.Request.Context(), authToken(c), page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) FeedbackIndex(c *gin.Context) {
	data, retcode, errmsg, err := h.service.FeedbackIndex(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) FeedbackNewListing(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	feedbackType, _ := strconv.Atoi(inputValue(c, "type"))
	data, retcode, errmsg, err := h.service.FeedbackNewListing(c.Request.Context(), authToken(c), feedbackType, page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) FeedbackDetail(c *gin.Context) {
	id, _ := strconv.Atoi(inputValue(c, "id"))
	data, retcode, errmsg, err := h.service.FeedbackDetail(c.Request.Context(), authToken(c), id)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode == -1 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) FeedbackCreate(c *gin.Context) {
	h.feedbackCreate(c, false)
}

func (h *UCPHandler) FeedbackCreateLegacy(c *gin.Context) {
	h.feedbackCreate(c, true)
}

func (h *UCPHandler) feedbackCreate(c *gin.Context, legacy bool) {
	cid, _ := strconv.Atoi(inputValue(c, "cid"))
	payid, _ := strconv.Atoi(inputValue(c, "payid"))
	fileCount := multipartFileCount(c, "upfiles")
	retcode, errmsg, err := h.service.FeedbackCreate(c.Request.Context(), authToken(c), ucpService.FeedbackCreateRequest{
		CID:        cid,
		Content:    inputValue(c, "content"),
		PayID:      payid,
		PayName:    inputValue(c, "payname"),
		PayAccount: inputValue(c, "payaccount"),
		Device:     inputValue(c, "device"),
		LongIDs:    inputValue(c, "longids"),
		ShortIDs:   inputValue(c, "shortids"),
		IP:         c.ClientIP(),
		FileCount:  fileCount,
		Legacy:     legacy,
	})
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}

func (h *UCPHandler) MsgListing(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.MsgListing(c.Request.Context(), authToken(c), page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) MsgDetail(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	cid, _ := strconv.Atoi(inputValue(c, "cid"))
	data, retcode, errmsg, err := h.service.MsgDetail(c.Request.Context(), authToken(c), cid, page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode == -1 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) MsgSetRead(c *gin.Context) {
	retcode, errmsg, err := h.service.MsgSetRead(c.Request.Context(), authToken(c), intArrayValue(c, "cids"))
	h.msgActionResponse(c, retcode, errmsg, err)
}

func (h *UCPHandler) MsgCleanRead(c *gin.Context) {
	retcode, errmsg, err := h.service.MsgCleanRead(c.Request.Context(), authToken(c))
	h.msgActionResponse(c, retcode, errmsg, err)
}

func (h *UCPHandler) MsgDelete(c *gin.Context) {
	retcode, errmsg, err := h.service.MsgDelete(c.Request.Context(), authToken(c), intArrayValue(c, "cids"))
	h.msgActionResponse(c, retcode, errmsg, err)
}

func (h *UCPHandler) msgActionResponse(c *gin.Context, retcode int, errmsg string, err error) {
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: errmsg})
}

func (h *UCPHandler) PaymentListing(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.PaymentListing(c.Request.Context(), authToken(c), page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) SafePayLog(c *gin.Context) {
	data, retcode, errmsg, err := h.service.SafePayLog(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) AccountIndex(c *gin.Context) {
	data, retcode, errmsg, err := h.service.AccountIndex(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) BalanceLog(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.BalanceLog(c.Request.Context(), authToken(c), page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) CoinLogIndex(c *gin.Context) {
	data, retcode, errmsg, err := h.service.CoinLogIndex(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) CoinLogBonusLog(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.CoinLogBonusLog(c.Request.Context(), authToken(c), page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *UCPHandler) CoinLogInviteLog(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.CoinLogInviteLog(c.Request.Context(), authToken(c), page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func authToken(c *gin.Context) string {
	if c.Request.Method != http.MethodOptions {
		if token, err := c.Cookie("xxx_api_auth"); err == nil && token != "" {
			return token
		}
	}
	if token := c.GetHeader("x-cookie-auth"); token != "" {
		return token
	}
	if token, err := c.Cookie("xxx_api_auth"); err == nil {
		return token
	}
	return ""
}

func inputValue(c *gin.Context, key string) string {
	if value := c.Query(key); value != "" {
		return value
	}
	return c.PostForm(key)
}

func multipartFileCount(c *gin.Context, key string) int {
	form, err := c.MultipartForm()
	if err != nil || form == nil || form.File == nil {
		return 0
	}
	return len(form.File[key])
}

func intArrayValue(c *gin.Context, key string) []int {
	values := append([]string{}, c.QueryArray(key)...)
	values = append(values, c.QueryArray(key+"[]")...)
	values = append(values, c.PostFormArray(key)...)
	values = append(values, c.PostFormArray(key+"[]")...)
	if len(values) == 0 {
		values = []string{inputValue(c, key)}
	}
	out := []int{}
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			n, _ := strconv.Atoi(part)
			out = append(out, n)
		}
	}
	return out
}
