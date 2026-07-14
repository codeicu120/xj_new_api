package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	sendfileService "xj_comp/internal/service/sendfile"
)

type SendfileHandler struct {
	service *sendfileService.Service
}

func NewSendfileHandler(service *sendfileService.Service) *SendfileHandler {
	return &SendfileHandler{service: service}
}

func (h *SendfileHandler) Play(c *gin.Context) {
	if invalidSendfileName(c.Param("file")) {
		c.Status(http.StatusNotFound)
		return
	}
	vodID, _ := strconv.Atoi(c.Query("vodid"))
	retcode, errmsg, err := h.service.Play(c.Request.Context(), authToken(c), vodID)
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, sendfileService.ErrVODNotFound) {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
}

func (h *SendfileHandler) Down(c *gin.Context) {
	if invalidSendfileName(c.Param("file")) {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
}

func invalidSendfileName(file string) bool {
	return strings.Contains(file, ".")
}
