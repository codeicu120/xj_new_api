package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	statsService "xj_comp/internal/service/stats"
)

type StatsHandler struct {
	service *statsService.Service
}

func NewStatsHandler(service *statsService.Service) *StatsHandler {
	return &StatsHandler{service: service}
}

func (h *StatsHandler) ShortcutAdd(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	if err := h.service.ShortcutAdd(c.Request.Context(), c.ClientIP()); err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("保存快捷方式统计失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: ""})
}

func (h *StatsHandler) AdAdd(c *gin.Context) {
	pos, _ := strconv.Atoi(inputValue(c, "pos"))
	click, _ := strconv.Atoi(inputValue(c, "click"))
	install, _ := strconv.Atoi(inputValue(c, "install"))
	retcode, errmsg, err := h.service.AdAdd(c.Request.Context(), authToken(c), c.ClientIP(), statsService.AdInput{
		Title:   inputValue(c, "title"),
		URL:     inputValue(c, "url"),
		Pos:     pos,
		Click:   click,
		Install: install,
	})
	h.write(c, retcode, errmsg, err)
}

func (h *StatsHandler) PlayAdd(c *gin.Context) {
	vid, _ := strconv.Atoi(inputValue(c, "vid"))
	mini, _ := strconv.Atoi(inputValue(c, "mini"))
	duration, _ := strconv.Atoi(inputValue(c, "duration"))
	played, _ := strconv.Atoi(inputValue(c, "played"))
	retcode, errmsg, err := h.service.PlayAdd(c.Request.Context(), authToken(c), c.ClientIP(), statsService.PlayInput{
		VID:      vid,
		Mini:     mini,
		Duration: duration,
		Played:   played,
	})
	h.write(c, retcode, errmsg, err)
}

func (h *StatsHandler) write(c *gin.Context, retcode int, errmsg string, err error) {
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: ""})
}
