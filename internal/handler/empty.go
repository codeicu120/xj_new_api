package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func EmptyHTML(c *gin.Context) {
	c.Data(http.StatusOK, "text/html", nil)
}
