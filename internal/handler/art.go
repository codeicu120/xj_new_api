package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	artService "xj_comp/internal/service/art"
)

type ArtHandler struct {
	service *artService.Service
}

func NewArtHandler(service *artService.Service) *ArtHandler {
	return &ArtHandler{service: service}
}

func (h *ArtHandler) Index(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.Data(http.StatusOK, "text/html", nil)
}

func (h *ArtHandler) Announce(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	data, err := h.service.Announce(c.Request.Context(), page)
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, artService.ErrCategoryNotFound) {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: -1, ErrMsg: "请求分类错误", Data: map[string]interface{}{}})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取公告失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *ArtHandler) Show(c *gin.Context) {
	artID, _ := strconv.Atoi(c.Query("artid"))
	data, err := h.service.Show(c.Request.Context(), artID)
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, artService.ErrArtNotFound) {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: -1, ErrMsg: "记录不存在或已被删除", Data: map[string]interface{}{}})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取文章失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}
