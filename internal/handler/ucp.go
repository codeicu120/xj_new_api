package handler

import (
	"net/http"
	"strconv"

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

func (h *UCPHandler) AffCenter(c *gin.Context) {
	data, retcode, errmsg, err := h.service.AffCenter(c.Request.Context(), authToken(c))
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

func (h *UCPHandler) FeedbackListing(c *gin.Context) {
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
