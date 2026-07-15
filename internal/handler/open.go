package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	openService "xj_comp/internal/service/open"
)

type OpenHandler struct {
	service *openService.Service
}

func NewOpenHandler(service *openService.Service) *OpenHandler {
	return &OpenHandler{service: service}
}

func (h *OpenHandler) Index(c *gin.Context) {
	c.Data(http.StatusOK, "text/html", nil)
}

func (h *OpenHandler) ReqAuth(c *gin.Context) {
	data, retcode, errmsg, err := h.service.ReqAuth(c.Request.Context(), authToken(c), c.ClientIP(), inputValue(c, "appid"))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}
