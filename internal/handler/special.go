package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	vodService "xj_comp/internal/service/vod"
)

type SpecialHandler struct {
	service *vodService.SpecialService
}

func NewSpecialHandler(service *vodService.SpecialService) *SpecialHandler {
	return &SpecialHandler{service: service}
}

func (h *SpecialHandler) Index(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.Data(http.StatusOK, "text/html", nil)
}

func (h *SpecialHandler) Listing(c *gin.Context) {
	data, err := h.service.Listing(c.Request.Context(), vodService.SpecialListingRequest{
		PathParams:  strings.TrimPrefix(c.Param("params"), "-"),
		QueryPage:   inputValue(c, "page"),
		IsH5Request: c.GetHeader("x-cookie-auth") != "",
	})
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取专题列表失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *SpecialHandler) Detail(c *gin.Context) {
	spid, params := parseSpecialDetailPath(c.Param("spid"))
	data, err := h.service.Detail(c.Request.Context(), vodService.SpecialDetailRequest{
		SPID:        spid,
		PathParams:  params,
		IsH5Request: c.GetHeader("x-cookie-auth") != "",
	})
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, vodService.ErrSpecialNotFound) {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: -1, ErrMsg: "记录不存在或已被删除", Data: map[string]interface{}{}})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取专题详情失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *SpecialHandler) Up(c *gin.Context) {
	h.vote(c, "up")
}

func (h *SpecialHandler) Down(c *gin.Context) {
	h.vote(c, "down")
}

func (h *SpecialHandler) vote(c *gin.Context, action string) {
	spid, _ := strconv.Atoi(c.Param("spid"))
	retcode, errmsg, err := h.service.Vote(c.Request.Context(), authToken(c), spid, action)
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, vodService.ErrSpecialNotFound) {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode == 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: errmsg})
		return
	}
	if retcode == -9999 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.Error(errmsg))
}

func parseSpecialDetailPath(value string) (int, string) {
	parts := strings.SplitN(value, "-", 2)
	spid, _ := strconv.Atoi(parts[0])
	if len(parts) == 1 {
		return spid, ""
	}
	return spid, parts[1]
}
