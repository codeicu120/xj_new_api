package handler

import (
	"errors"
	"net/http"
	"strconv"
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

func (h *CommentHandler) Post(c *gin.Context) {
	vodID, _ := strconv.Atoi(inputValue(c, "vodid"))
	parentID, _ := strconv.Atoi(inputValue(c, "parentid"))
	data, retcode, errmsg, err := h.service.Post(c.Request.Context(), authToken(c), vodID, parentID, inputValue(c, "content"), c.ClientIP())
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: errmsg, Data: data})
}

func (h *CommentHandler) Up(c *gin.Context) {
	h.vote(c, true)
}

func (h *CommentHandler) Down(c *gin.Context) {
	h.vote(c, false)
}

func (h *CommentHandler) vote(c *gin.Context, up bool) {
	id, _ := strconv.Atoi(inputValue(c, "id"))
	retcode, errmsg, err := h.service.Vote(c.Request.Context(), authToken(c), id, up)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}
