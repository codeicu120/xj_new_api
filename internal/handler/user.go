package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	userService "xj_comp/internal/service/user"
)

type UserHandler struct {
	sysAvatarService *userService.SysAvatarService
	logoutService    *userService.LogoutService
}

func NewUserHandler(sysAvatarService *userService.SysAvatarService, logoutService *userService.LogoutService) *UserHandler {
	return &UserHandler{sysAvatarService: sysAvatarService, logoutService: logoutService}
}

func (h *UserHandler) SysAvatar(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.OK(h.sysAvatarService.List()))
}

func (h *UserHandler) Logout(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	if err := h.logoutService.Logout(c.Request.Context(), authToken(c)); err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("退出失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: "已退出"})
}
