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

func authToken(c *gin.Context) string {
	if token := c.GetHeader("x-cookie-auth"); token != "" {
		return token
	}
	if token, err := c.Cookie("xxx_api_auth"); err == nil {
		return token
	}
	return ""
}
