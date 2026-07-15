package handler

import (
	"encoding/base64"
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

func (h *PaymentHandler) ChPayway(c *gin.Context) {
	payID, _ := strconv.Atoi(inputValue(c, "payid"))
	retcode, errmsg, err := h.service.ChangePayway(c.Request.Context(), authToken(c), payID, inputValue(c, "paycode"))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{
		RetCode: 0,
		ErrMsg:  errmsg,
	})
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

func (h *PaymentHandler) ReqPay(c *gin.Context) {
	payID, _ := strconv.Atoi(inputValue(c, "payid"))
	data, retcode, errmsg, err := h.service.ReqPay(c.Request.Context(), authToken(c), payID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: data})
}

func (h *PaymentHandler) Pay12Req(c *gin.Context) {
	payID, _ := strconv.Atoi(inputValue(c, "payid"))
	_, _, errmsg, err := h.service.ReqPay(c.Request.Context(), authToken(c), payID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(""))
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(h.service.PayErrorHTML(c.Request.Context(), errmsg)))
}

func (h *PaymentHandler) SuccessHTML(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(h.service.SuccessHTML(c.Request.Context())))
}

func (h *PaymentHandler) Pay7Submit(c *gin.Context) {
	rawParams, _ := base64.StdEncoding.DecodeString(inputValue(c, "p"))
	c.Header("X-Served-By", "newbie")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(h.service.SubmitHTML(c.Request.Context(), inputValue(c, "gateway"), string(rawParams))))
}

func (h *PaymentHandler) Pay11(c *gin.Context) {
	qrlink := inputValue(c, "qrlink")
	c.Header("X-Served-By", "newbie")
	if qrlink == "" {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(h.service.SuccessHTML(c.Request.Context())))
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(h.service.QRCodeHTML(c.Request.Context(), qrlink)))
}

func (h *PaymentHandler) WapPay1(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: h.service.SuccessMessage(c.Request.Context())})
}

func (h *PaymentHandler) WapPay2(c *gin.Context) {
	payID, _ := strconv.Atoi(inputValue(c, "payid"))
	if payID > 0 {
		html, err := h.service.PaymentHTML(c.Request.Context(), payID)
		c.Header("X-Served-By", "newbie")
		if err != nil {
			c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(""))
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		return
	}
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: h.service.SuccessMessage(c.Request.Context())})
}
