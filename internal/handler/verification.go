package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	verificationService "xj_comp/internal/service/verification"
)

type VerificationHandler struct {
	service *verificationService.Service
}

func NewVerificationHandler(service *verificationService.Service) *VerificationHandler {
	return &VerificationHandler{service: service}
}

func (h *VerificationHandler) SMSSendV(c *gin.Context) {
	msg, err := h.service.SendV(c.Request.Context(), verificationService.SendSMSRequest{
		MobiPrefix:  input(c, "mobiprefix"),
		Mobi:        input(c, "mobi"),
		CaptchaKey:  strings.TrimSpace(input(c, "captcha_key")),
		CaptchaCode: strings.TrimSpace(input(c, "captcha_code")),
		GTicket:     input(c, "g_ticket"),
		TXTicket:    input(c, "tx_ticket"),
		TXRandstr:   input(c, "tx_randstr"),
		SelfToken:   input(c, "self_token"),
		SendCount:   inputInt(c, "sendcount"),
		UserAgent:   c.GetHeader("user-agent"),
		ClientIP:    c.ClientIP(),
	})
	c.Header("X-Served-By", "newbie")
	writeVerification(c, msg, err)
}

func (h *VerificationHandler) SMSSendU(c *gin.Context) {
	msg, err := h.service.SendU(c.Request.Context(), verificationService.SendSMSRequest{
		Token:      authToken(c),
		MobiPrefix: input(c, "mobiprefix"),
		Mobi:       input(c, "mobi"),
		SendCount:  inputInt(c, "sendcount"),
		ClientIP:   c.ClientIP(),
	})
	c.Header("X-Served-By", "newbie")
	writeVerification(c, msg, err)
}

func (h *VerificationHandler) EmailSend(c *gin.Context) {
	msg, err := h.service.SendEmail(c.Request.Context(), verificationService.SendEmailRequest{
		Email:       input(c, "email"),
		CaptchaKey:  strings.TrimSpace(input(c, "captcha_key")),
		CaptchaCode: strings.TrimSpace(input(c, "captcha_code")),
		GTicket:     input(c, "g_ticket"),
		TXTicket:    input(c, "tx_ticket"),
		TXRandstr:   input(c, "tx_randstr"),
		SelfToken:   input(c, "self_token"),
		SendCount:   inputInt(c, "sendcount"),
		ClientIP:    c.ClientIP(),
	})
	c.Header("X-Served-By", "newbie")
	writeVerification(c, msg, err)
}

func writeVerification(c *gin.Context, msg string, err error) {
	if errors.Is(err, verificationService.ErrLoginRequired) {
		c.JSON(http.StatusOK, legacyjson.Error("请先登录"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("发送失败"))
		return
	}
	if msg == "短信已成功发送" || msg == "验证码已发送至您的邮箱，请10分钟内验证并确认" {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: msg})
		return
	}
	c.JSON(http.StatusOK, legacyjson.Error(msg))
}

func input(c *gin.Context, key string) string {
	if value := c.PostForm(key); value != "" {
		return value
	}
	return c.Query(key)
}

func inputInt(c *gin.Context, key string) int {
	value, _ := strconv.Atoi(input(c, key))
	return value
}
