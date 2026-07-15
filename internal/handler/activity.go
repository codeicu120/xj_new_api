package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	activityService "xj_comp/internal/service/activity"
)

type ActivityHandler struct {
	service *activityService.Service
}

func NewActivityHandler(service *activityService.Service) *ActivityHandler {
	return &ActivityHandler{service: service}
}

func (h *ActivityHandler) LuckyPrizes(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.OK(h.service.LuckyPrizes()))
}

func (h *ActivityHandler) Index(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.Index(c.Request.Context(), page)
	h.jsonResponse(c, data, retcode, errmsg, err)
}

func (h *ActivityHandler) Details(c *gin.Context) {
	aid, _ := strconv.Atoi(inputValue(c, "aid"))
	data, retcode, errmsg, err := h.service.Details(c.Request.Context(), aid)
	h.jsonResponse(c, data, retcode, errmsg, err)
}

func (h *ActivityHandler) LuckyDrawHistory(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.LuckyDrawHistory(c.Request.Context(), authToken(c), page)
	h.jsonResponse(c, data, retcode, errmsg, err)
}

func (h *ActivityHandler) Ranking(c *gin.Context) {
	aid, _ := strconv.Atoi(inputValue(c, "aid"))
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.Ranking(c.Request.Context(), authToken(c), aid, page)
	h.jsonResponse(c, data, retcode, errmsg, err)
}

func (h *ActivityHandler) Receive(c *gin.Context) {
	aid, _ := strconv.Atoi(inputValue(c, "aid"))
	data, retcode, errmsg, err := h.service.Receive(c.Request.Context(), authToken(c), aid)
	h.jsonResponse(c, data, retcode, errmsg, err)
}

func (h *ActivityHandler) Recommends(c *gin.Context) {
	aid, _ := strconv.Atoi(inputValue(c, "aid"))
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.Recommends(c.Request.Context(), authToken(c), aid, page)
	h.jsonResponse(c, data, retcode, errmsg, err)
}

func (h *ActivityHandler) NewYear2020(c *gin.Context) {
	retcode, errmsg := h.service.NewYear2020()
	h.expiredActivity(c, retcode, errmsg)
}

func (h *ActivityHandler) LuckyDraw(c *gin.Context) {
	retcode, errmsg := h.service.LuckyDraw()
	h.expiredActivity(c, retcode, errmsg)
}

func (h *ActivityHandler) expiredActivity(c *gin.Context, retcode int, errmsg string) {
	c.Header("X-Served-By", "newbie")
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: ""})
}

func (h *ActivityHandler) jsonResponse(c *gin.Context, data map[string]interface{}, retcode int, errmsg string, err error) {
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
