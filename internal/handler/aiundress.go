package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	aiundressService "xj_comp/internal/service/aiundress"
)

type AIUndressHandler struct {
	service *aiundressService.Service
}

func NewAIUndressHandler(service *aiundressService.Service) *AIUndressHandler {
	return &AIUndressHandler{service: service}
}

func (h *AIUndressHandler) Listing(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	module, _ := strconv.Atoi(inputValue(c, "module"))
	data, retcode, errmsg, err := h.service.Listing(c.Request.Context(), authToken(c), page, module)
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
