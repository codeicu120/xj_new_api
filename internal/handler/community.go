package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	communityService "xj_comp/internal/service/community"
)

type CommunityHandler struct {
	service *communityService.Service
}

func NewCommunityHandler(service *communityService.Service) *CommunityHandler {
	return &CommunityHandler{service: service}
}

func (h *CommunityHandler) Listing(c *gin.Context) {
	action := communityAction(c.Request.URL.Path)
	params := strings.TrimPrefix(c.Param("params"), "-")
	data, err := h.service.Listing(c.Request.Context(), communityService.ListingRequest{
		Action:      action,
		PathParams:  params,
		QueryPage:   inputValue(c, "page"),
		IsH5Request: c.GetHeader("x-cookie-auth") != "",
		Token:       authToken(c),
	})
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, communityService.ErrLoginRequired) {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: -9999, ErrMsg: "请登录后操作"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取社区列表失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *CommunityHandler) CommentListing(c *gin.Context) {
	params := strings.TrimPrefix(c.Param("params"), "-")
	tid, _ := strconv.Atoi(inputValue(c, "tid"))
	data, err := h.service.CommentListing(c.Request.Context(), communityService.CommentListingRequest{
		PathParams: params,
		QueryPage:  inputValue(c, "page"),
		QueryOrder: inputValue(c, "orderby"),
		TID:        tid,
		Token:      authToken(c),
	})
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, communityService.ErrTopicNotFound) {
		c.JSON(http.StatusOK, legacyjson.Error("记录不存在或已被删除"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取社区评论失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *CommunityHandler) Show(c *gin.Context) {
	tid, _ := strconv.Atoi(inputValue(c, "tid"))
	data, err := h.service.Show(c.Request.Context(), communityService.ShowRequest{
		TID:        tid,
		QueryOrder: inputValue(c, "orderby"),
		Token:      authToken(c),
	})
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, communityService.ErrTopicNotFound) {
		c.JSON(http.StatusOK, legacyjson.Error("记录不存在或已删除"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取社区详情失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *CommunityHandler) Up(c *gin.Context) {
	tid, _ := strconv.Atoi(inputValue(c, "tid"))
	retcode, errmsg, err := h.service.UpTopic(c.Request.Context(), authToken(c), tid)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}

func (h *CommunityHandler) Attention(c *gin.Context) {
	tid, _ := strconv.Atoi(inputValue(c, "tid"))
	retcode, errmsg, err := h.service.Attention(c.Request.Context(), authToken(c), tid, intArrayValue(c, "tids"))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}

func (h *CommunityHandler) UpComment(c *gin.Context) {
	cid, _ := strconv.Atoi(inputValue(c, "cid"))
	retcode, errmsg, err := h.service.UpComment(c.Request.Context(), authToken(c), cid)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}

func communityAction(path string) string {
	path = strings.TrimPrefix(path, "/community/")
	if index := strings.Index(path, "-"); index >= 0 {
		path = path[:index]
	}
	return path
}
