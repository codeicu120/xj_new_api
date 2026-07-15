package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	hgameService "xj_comp/internal/service/hgame"
)

type HGameHandler struct {
	service *hgameService.Service
}

func NewHGameHandler(service *hgameService.Service) *HGameHandler {
	return &HGameHandler{service: service}
}

func (h *HGameHandler) Index(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.Index(c.Request.Context(), page)
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
