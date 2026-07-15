package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	starliveService "xj_comp/internal/service/starlive"
)

type StarLiveHandler struct {
	service *starliveService.Service
}

func NewStarLiveHandler(service *starliveService.Service) *StarLiveHandler {
	return &StarLiveHandler{service: service}
}

func (h *StarLiveHandler) Index(c *gin.Context) {
	data, retcode, errmsg, err := h.service.Index(c.Request.Context(), authToken(c))
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

func (h *StarLiveHandler) QueryCoinBalance(c *gin.Context) {
	var body struct {
		MemberID string `json:"memberId"`
	}
	_ = json.NewDecoder(c.Request.Body).Decode(&body)
	data, err := h.service.QueryCoinBalance(c.Request.Context(), body.MemberID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "data": gin.H{"msg": "未知用户"}})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *StarLiveHandler) GameBet(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, h.service.GameBetEdge(starLiveBody(c)))
}

func (h *StarLiveHandler) GameWin(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, h.service.GameWinEdge(starLiveBody(c)))
}

func (h *StarLiveHandler) Translate(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, h.service.TranslateEdge(starLiveBody(c)))
}

func (h *StarLiveHandler) TryAgain(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, h.service.TryAgainEdge(starLiveBody(c)))
}

func starLiveBody(c *gin.Context) map[string]interface{} {
	var body map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil || body == nil {
		return map[string]interface{}{}
	}
	return body
}
