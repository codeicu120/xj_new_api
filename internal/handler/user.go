package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	userService "xj_comp/internal/service/user"
)

type UserHandler struct {
	sysAvatarService *userService.SysAvatarService
	logoutService    *userService.LogoutService
	authEdgeService  *userService.AuthEdgeService
}

func NewUserHandler(sysAvatarService *userService.SysAvatarService, logoutService *userService.LogoutService, authEdgeService *userService.AuthEdgeService) *UserHandler {
	return &UserHandler{sysAvatarService: sysAvatarService, logoutService: logoutService, authEdgeService: authEdgeService}
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

func (h *UserHandler) Register(c *gin.Context) {
	retcode, errmsg, err := h.authEdgeService.Register(c.Request.Context(), authEdgeRequest(c), strings.HasPrefix(c.FullPath(), "/v2/"))
	respondUserEdge(c, retcode, errmsg, err)
}

func (h *UserHandler) Login(c *gin.Context) {
	retcode, errmsg, err := h.authEdgeService.Login(c.Request.Context(), authEdgeRequest(c), false)
	respondUserEdge(c, retcode, errmsg, err)
}

func (h *UserHandler) LoginV2(c *gin.Context) {
	retcode, errmsg, err := h.authEdgeService.Login(c.Request.Context(), authEdgeRequest(c), true)
	respondUserEdge(c, retcode, errmsg, err)
}

func (h *UserHandler) Forgot(c *gin.Context) {
	retcode, errmsg, err := h.authEdgeService.Forgot(c.Request.Context(), authEdgeRequest(c), false)
	respondUserEdge(c, retcode, errmsg, err)
}

func (h *UserHandler) ForgotV2(c *gin.Context) {
	retcode, errmsg, err := h.authEdgeService.Forgot(c.Request.Context(), authEdgeRequest(c), true)
	respondUserEdge(c, retcode, errmsg, err)
}

func (h *UserHandler) Delete(c *gin.Context) {
	retcode, errmsg, err := h.authEdgeService.Delete(c.Request.Context(), authToken(c))
	respondUserEdge(c, retcode, errmsg, err)
}

func (h *UserHandler) ChangePhone(c *gin.Context) {
	retcode, errmsg, err := h.authEdgeService.ChangePhone(c.Request.Context(), authEdgeRequest(c))
	respondUserEdge(c, retcode, errmsg, err)
}

func authEdgeRequest(c *gin.Context) userService.AuthEdgeRequest {
	aup, _ := strconv.Atoi(inputValue(c, "aup"))
	regType, _ := strconv.Atoi(inputValue(c, "regtype"))
	loginType, _ := strconv.Atoi(inputValue(c, "logintype"))
	return userService.AuthEdgeRequest{
		Token:      authToken(c),
		AUP:        aup,
		Step:       inputValue(c, "step"),
		Mobi:       inputValue(c, "mobi"),
		Email:      inputValue(c, "email"),
		Username:   inputValue(c, "username"),
		Password:   inputValue(c, "password"),
		MobiPrefix: inputValue(c, "mobiprefix"),
		RegType:    regType,
		LoginType:  loginType,
		ClientIP:   c.ClientIP(),
	}
}

func respondUserEdge(c *gin.Context, retcode int, errmsg string, err error) {
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}
