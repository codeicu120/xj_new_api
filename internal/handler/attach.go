package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	attachService "xj_comp/internal/service/attach"
)

type AttachHandler struct {
	service *attachService.Service
}

func NewAttachHandler(service *attachService.Service) *AttachHandler {
	return &AttachHandler{service: service}
}

func (h *AttachHandler) Index(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
}

func (h *AttachHandler) UpAvatar(c *gin.Context) {
	retcode, errmsg, err := h.service.UpAvatar(c.Request.Context(), authToken(c), inputValue(c, "avatarid"))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode == -1 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(map[string]interface{}{}))
}
