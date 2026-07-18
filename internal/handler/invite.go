package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	inviteService "xj_comp/internal/service/invite"
)

type InviteHandler struct {
	service *inviteService.Service
}

func NewInviteHandler(service *inviteService.Service) *InviteHandler {
	return &InviteHandler{service: service}
}

func (h *InviteHandler) Info(c *gin.Context) {
	data, retcode, errmsg, err := h.service.Info(c.Request.Context(), authToken(c))
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

func (h *InviteHandler) Bind(c *gin.Context) {
	data, retcode, errmsg, err := h.service.Bind(c.Request.Context(), authToken(c), inputValue(c, "invitecode"))
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
