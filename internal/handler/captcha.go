package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	captchaService "xj_comp/internal/service/captcha"
)

type CaptchaHandler struct {
	service *captchaService.Service
}

func NewCaptchaHandler(service *captchaService.Service) *CaptchaHandler {
	return &CaptchaHandler{service: service}
}

func (h *CaptchaHandler) Req(c *gin.Context) {
	data, err := h.service.Req()
	if err != nil {
		c.Header("X-Served-By", "newbie")
		c.JSON(http.StatusInternalServerError, legacyjson.Error("验证码生成失败"))
		return
	}
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.OK(data))
}
