package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RespondHandler struct{}

func NewRespondHandler() *RespondHandler {
	return &RespondHandler{}
}

func (h *RespondHandler) Failed(text string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Served-By", "newbie")
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(text))
	}
}
