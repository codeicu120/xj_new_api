package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	amazingService "xj_comp/internal/service/amazing"
)

type AmazingHandler struct {
	categoryService *amazingService.CategoryService
	listingService  *amazingService.ListingService
}

func NewAmazingHandler(categoryService *amazingService.CategoryService, listingService *amazingService.ListingService) *AmazingHandler {
	return &AmazingHandler{
		categoryService: categoryService,
		listingService:  listingService,
	}
}

func (h *AmazingHandler) Categories(c *gin.Context) {
	parentID, _ := strconv.Atoi(c.Query("parent_id"))
	data, err := h.categoryService.List(c.Request.Context(), parentID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取精彩推荐分类失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *AmazingHandler) Listing(c *gin.Context) {
	action := amazingAction(c.Request.URL.Path)
	params := strings.TrimPrefix(c.Param("params"), "-")
	data, err := h.listingService.List(c.Request.Context(), amazingService.ListingRequest{
		Action:     action,
		PathParams: params,
		QueryPage:  c.Query("page"),
	})
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取精彩推荐列表失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func amazingAction(path string) string {
	path = strings.TrimPrefix(path, "/v2/amazing/")
	if index := strings.Index(path, "-"); index >= 0 {
		path = path[:index]
	}
	return path
}
