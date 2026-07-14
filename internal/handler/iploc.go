package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	iplocService "xj_comp/internal/service/iploc"
)

type IPLocHandler struct {
	service *iplocService.Service
}

func NewIPLocHandler(service *iplocService.Service) *IPLocHandler {
	return &IPLocHandler{service: service}
}

func (h *IPLocHandler) Find(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.OK(h.service.Find(c.Param("ip"))))
}
