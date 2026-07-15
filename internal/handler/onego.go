package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	onegoService "xj_comp/internal/service/onego"
)

type OneGoHandler struct {
	service *onegoService.Service
}

func NewOneGoHandler(service *onegoService.Service) *OneGoHandler {
	return &OneGoHandler{service: service}
}

func (h *OneGoHandler) Rules(c *gin.Context) {
	data, err := h.service.Rules(c.Request.Context())
	c.Header("X-Served-By", "newbie")
	if err != nil {
		if errors.Is(err, onegoService.ErrNotOpen) {
			c.JSON(http.StatusOK, oneGoError("系统尚未开放该活动"))
			return
		}
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取一元购规则失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *OneGoHandler) Rooms(c *gin.Context) {
	data, err := h.service.Rooms(c.Request.Context())
	c.Header("X-Served-By", "newbie")
	if err != nil {
		if errors.Is(err, onegoService.ErrNotOpen) {
			c.JSON(http.StatusOK, oneGoError("系统尚未开放该活动"))
			return
		}
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取一元购房间失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *OneGoHandler) Current(c *gin.Context) {
	roomID, _ := strconv.Atoi(inputValue(c, "roomid"))
	data, err := h.service.Current(c.Request.Context(), roomID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		h.writeOneGoError(c, err, "获取一元购当前期数失败")
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *OneGoHandler) Last(c *gin.Context) {
	roomID, _ := strconv.Atoi(inputValue(c, "roomid"))
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, err := h.service.Last(c.Request.Context(), roomID, page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		h.writeOneGoError(c, err, "获取一元购上期记录失败")
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *OneGoHandler) Hash(c *gin.Context) {
	data, err := h.service.Hash(inputValue(c, "plaintext"))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		switch {
		case errors.Is(err, onegoService.ErrMissingPlaintext):
			c.JSON(http.StatusOK, oneGoError("请传入参数"))
		case errors.Is(err, onegoService.ErrHashNumberUnavailable):
			c.JSON(http.StatusOK, oneGoError("无法计算后六位数字"))
		default:
			c.JSON(http.StatusInternalServerError, legacyjson.Error("计算一元购哈希失败"))
		}
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *OneGoHandler) History(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.History(c.Request.Context(), authToken(c), page)
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

func (h *OneGoHandler) Bet(c *gin.Context) {
	quantity, _ := strconv.Atoi(inputValue(c, "quantity"))
	roomID, _ := strconv.Atoi(inputValue(c, "roomid"))
	retcode, errmsg, err := h.service.BetEdge(c.Request.Context(), authToken(c), inputValue(c, "period"), roomID, quantity)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
}

func (h *OneGoHandler) Lucky(c *gin.Context) {
	data, err := h.service.Lucky(c.Request.Context())
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取一元购幸运榜失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *OneGoHandler) BetRanks(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	roomID, _ := strconv.Atoi(inputValue(c, "roomid"))
	data, err := h.service.BetRanks(c.Request.Context(), inputValue(c, "period"), roomID, page)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		h.writeOneGoError(c, err, "获取一元购押注排行失败")
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *OneGoHandler) Marquee(c *gin.Context) {
	data, err := h.service.Marquee(c.Request.Context())
	c.Header("X-Served-By", "newbie")
	if err != nil {
		h.writeOneGoError(c, err, "获取一元购跑马灯失败")
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *OneGoHandler) writeOneGoError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, onegoService.ErrNotOpen):
		c.JSON(http.StatusOK, oneGoError("系统尚未开放该活动"))
	case errors.Is(err, onegoService.ErrSelectRoom):
		c.JSON(http.StatusOK, oneGoError("请选择场次"))
	case errors.Is(err, onegoService.ErrActivityEnded):
		c.JSON(http.StatusOK, oneGoError("活动已结束或尚未开始"))
	case errors.Is(err, onegoService.ErrNoData):
		c.JSON(http.StatusOK, oneGoError("暂无数据"))
	case errors.Is(err, onegoService.ErrInvalidRoom):
		c.JSON(http.StatusOK, oneGoError("无效场次"))
	case errors.Is(err, onegoService.ErrInvalidPeriod):
		c.JSON(http.StatusOK, oneGoError("无效的活动期号"))
	default:
		c.JSON(http.StatusInternalServerError, legacyjson.Error(fallback))
	}
}

func oneGoError(message string) legacyjson.Response {
	return legacyjson.Response{RetCode: -1, ErrMsg: message, Data: map[string]interface{}{}}
}
