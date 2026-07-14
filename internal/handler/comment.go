package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	commentService "xj_comp/internal/service/comment"
)

type CommentHandler struct {
	service *commentService.Service
}

func NewCommentHandler(service *commentService.Service) *CommentHandler {
	return &CommentHandler{service: service}
}

func (h *CommentHandler) Listing(c *gin.Context) {
	params := strings.TrimPrefix(c.Param("params"), "-")
	data, err := h.service.Listing(c.Request.Context(), commentService.ListingRequest{
		PathParams: params,
		QueryPage:  c.Query("page"),
	})
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, commentService.ErrVODNotFound) {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: -1, ErrMsg: "记录不存在或已被删除", Data: map[string]interface{}{}})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取评论列表失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}
