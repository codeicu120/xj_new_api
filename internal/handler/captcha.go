package handler

import (
	"errors"
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

func (h *CaptchaHandler) ReqV2(c *gin.Context) {
	data, err := h.service.ReqV2()
	if err != nil {
		c.Header("X-Served-By", "newbie")
		c.JSON(http.StatusInternalServerError, legacyjson.Error("验证码生成失败"))
		return
	}
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *CaptchaHandler) Pic(c *gin.Context) {
	h.image(c)
}

func (h *CaptchaHandler) PicX(c *gin.Context) {
	h.image(c)
}

func (h *CaptchaHandler) Verify(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	switch {
	case input(c, "g_ticket") != "":
		c.JSON(http.StatusOK, legacyjson.Error("人机验证失败，请重新验证"))
	case input(c, "tx_ticket") != "" || input(c, "tx_randstr") != "":
		if input(c, "tx_ticket") == "" || input(c, "tx_randstr") == "" {
			c.JSON(http.StatusOK, legacyjson.Error("未提供验证票据"))
			return
		}
		c.JSON(http.StatusOK, legacyjson.Error("验证失败，请重新验证"))
	case input(c, "self_token") != "":
		c.JSON(http.StatusOK, legacyjson.Error("验证失败，请重新验证"))
	case h.service.Verify(input(c, "captcha_key"), input(c, "captcha_code")):
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: "验证成功"})
	default:
		c.JSON(http.StatusOK, legacyjson.Error("验证失败"))
	}
}

func (h *CaptchaHandler) image(c *gin.Context) {
	body, err := h.service.PNG(c.Request.URL.RawQuery)
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, captchaService.ErrInvalidSecret) {
		c.JSON(http.StatusNotFound, legacyjson.Response{RetCode: -4, ErrMsg: "验证码无效", Data: map[string]interface{}{}})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("验证码生成失败"))
		return
	}
	c.Data(http.StatusOK, "image/png", body)
}
