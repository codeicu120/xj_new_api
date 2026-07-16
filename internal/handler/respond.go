package handler

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	respondService "xj_comp/internal/service/respond"
)

type RespondHandler struct {
	service *respondService.Service
}

func NewRespondHandler(service *respondService.Service) *RespondHandler {
	return &RespondHandler{service: service}
}

func (h *RespondHandler) Failed(text string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Served-By", "newbie")
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(text))
	}
}

func (h *RespondHandler) Provider(action string, echoErr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Served-By", "newbie")
		req := respondService.CallbackRequest{
			Action: action,
			Form:   callbackForm(c),
		}
		if c.Request.Body != nil {
			body, _ := io.ReadAll(c.Request.Body)
			req.Raw = body
		}
		result := respondService.VerificationResult{Echo: echoErr}
		if h.service != nil {
			result = h.service.VerifyProvider(c.Request.Context(), req, echoErr)
		}
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(result.Echo))
	}
}

func (h *RespondHandler) Chan1(c *gin.Context) {
	const secret = "lpeg0Jt2YbxeoeiK25sWIXtX5oIWzDnC"
	mobi := inputValue(c, "mobi")
	token := inputValue(c, "token")
	sum := md5.Sum([]byte(mobi + "|" + secret))
	c.Header("X-Served-By", "newbie")
	if token != hex.EncodeToString(sum[:]) {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: 1, ErrMsg: "校验失败"})
		return
	}
	if h.service == nil {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: -1, ErrMsg: "chan1 成功分支暂未迁移"})
		return
	}
	retcode, errmsg, err := h.service.Chan1(c.Request.Context(), mobi)
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}

func callbackForm(c *gin.Context) url.Values {
	_ = c.Request.ParseForm()
	form := make(url.Values, len(c.Request.Form))
	for key, values := range c.Request.Form {
		form[key] = append([]string(nil), values...)
	}
	return form
}
