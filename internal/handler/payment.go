package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	paymentService "xj_comp/internal/service/payment"
)

type PaymentHandler struct {
	service *paymentService.Service
}

func NewPaymentHandler(service *paymentService.Service) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) Unpaid(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.OK(h.service.Unpaid(c.Request.Context())))
}

func (h *PaymentHandler) Success(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.Response{
		RetCode: 0,
		ErrMsg:  h.service.SuccessMessage(c.Request.Context()),
	})
}

func (h *PaymentHandler) Failed(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.Error(h.service.FailedMessage(c.Request.Context())))
}
