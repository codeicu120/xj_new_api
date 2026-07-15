package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	captchaService "xj_comp/internal/service/captcha"
)

type TestHandler struct {
	service *captchaService.TestImageService
}

func NewTestHandler(service *captchaService.TestImageService) *TestHandler {
	return &TestHandler{service: service}
}

func (h *TestHandler) Test(c *gin.Context) {
	data, err := h.service.PNG()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("X-Served-By", "newbie")
	c.Data(http.StatusOK, "image/png", data)
}
