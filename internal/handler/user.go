package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	userService "xj_comp/internal/service/user"
)

type UserHandler struct {
	sysAvatarService *userService.SysAvatarService
}

func NewUserHandler(sysAvatarService *userService.SysAvatarService) *UserHandler {
	return &UserHandler{sysAvatarService: sysAvatarService}
}

func (h *UserHandler) SysAvatar(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.OK(h.sysAvatarService.List()))
}
