package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	minivodService "xj_comp/internal/service/minivod"
)

type MiniVODHandler struct {
	service *minivodService.Service
}

func NewMiniVODHandler(service *minivodService.Service) *MiniVODHandler {
	return &MiniVODHandler{service: service}
}

func (h *MiniVODHandler) Listing(c *gin.Context) {
	action := minivodAction(c.Request.URL.Path)
	params := strings.TrimPrefix(c.Param("params"), "-")
	data, err := h.service.Listing(c.Request.Context(), minivodService.ListingRequest{
		Action:      action,
		PathParams:  params,
		QueryPage:   c.Query("page"),
		IsH5Request: c.GetHeader("x-cookie-auth") != "",
	})
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取小视频列表失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *MiniVODHandler) Show(c *gin.Context) {
	vodID, _ := strconv.Atoi(c.Param("vodid"))
	data, err := h.service.Show(c.Request.Context(), vodID, c.GetHeader("x-cookie-auth") != "")
	c.Header("X-Served-By", "newbie")
	switch {
	case errors.Is(err, minivodService.ErrVODNotFound):
		c.JSON(http.StatusOK, legacyjson.Error("记录不存在或已删除"))
	case errors.Is(err, minivodService.ErrAuthorNotFound):
		c.JSON(http.StatusOK, legacyjson.Error("作者不存在或已被删除"))
	case err != nil:
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取小视频详情失败"))
	default:
		c.JSON(http.StatusOK, legacyjson.OK(data))
	}
}

func (h *MiniVODHandler) Author(c *gin.Context) {
	authorID, _ := strconv.Atoi(c.Param("authorid"))
	action := strings.TrimPrefix(c.Param("action"), "/")
	if action != "" && action != "index" && action != "listing" {
		c.Header("X-Served-By", "newbie")
		c.JSON(http.StatusOK, legacyjson.Error("用户不存在或已被删除"))
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	data, err := h.service.AuthorListing(c.Request.Context(), authorID, page, c.GetHeader("x-cookie-auth") != "")
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, minivodService.ErrAuthorNotFound) {
		c.JSON(http.StatusOK, legacyjson.Error("用户不存在或已被删除"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取作者主页失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func minivodAction(path string) string {
	path = strings.TrimPrefix(path, "/minivod/")
	if index := strings.Index(path, "-"); index >= 0 {
		path = path[:index]
	}
	return path
}
