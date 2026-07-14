package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	soService "xj_comp/internal/service/so"
)

type SOHandler struct {
	configService *soService.ConfigService
}

func NewSOHandler(configService *soService.ConfigService) *SOHandler {
	return &SOHandler{configService: configService}
}

func (h *SOHandler) List(c *gin.Context) {
	version, _ := strconv.Atoi(c.Query("version"))
	data, err := h.configService.List(c.Request.Context(), version, c.Query("arm"), c.Query("channel"))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取插件配置失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}
