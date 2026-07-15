package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	picService "xj_comp/internal/service/pic"
)

type PicHandler struct {
	service *picService.Service
}

func NewPicHandler(service *picService.Service) *PicHandler {
	return &PicHandler{service: service}
}

func (h *PicHandler) Index(c *gin.Context) {
	uri := strings.TrimPrefix(c.Param("uri"), "/")
	data, contentType, err := h.service.Image(c.Request.Context(), c.Param("size"), uri)
	if errors.Is(err, picService.ErrInvalidRequest) || errors.Is(err, picService.ErrNotFound) {
		c.Data(http.StatusNotFound, "text/html", nil)
		return
	}
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/html", nil)
		return
	}
	c.Data(http.StatusOK, contentType, data)
}
