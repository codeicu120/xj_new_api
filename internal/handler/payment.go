package handler

import (
	"net/http"
	"strconv"

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

func (h *PaymentHandler) Query(c *gin.Context) {
	payID, _ := strconv.Atoi(inputValue(c, "payid"))
	data, retcode, errmsg, err := h.service.Query(c.Request.Context(), authToken(c), payID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *PaymentHandler) Payways(c *gin.Context) {
	payID, _ := strconv.Atoi(inputValue(c, "payid"))
	data, retcode, errmsg, err := h.service.Payways(c.Request.Context(), authToken(c), payID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
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
