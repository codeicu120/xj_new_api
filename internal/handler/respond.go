package handler

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
)

type RespondHandler struct{}

func NewRespondHandler() *RespondHandler {
	return &RespondHandler{}
}

func (h *RespondHandler) Failed(text string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Served-By", "newbie")
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(text))
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
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: -1, ErrMsg: "chan1 成功分支暂未迁移"})
}
