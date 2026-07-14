package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	gameService "xj_comp/internal/service/game"
)

type GameHandler struct {
	platformService  *gameService.PlatformService
	categoryService  *gameService.CategoryService
	listingService   *gameService.ListingService
	broadcastService *gameService.BroadcastService
}

func NewGameHandler(platformService *gameService.PlatformService, categoryService *gameService.CategoryService, listingService *gameService.ListingService, broadcastService *gameService.BroadcastService) *GameHandler {
	return &GameHandler{
		platformService:  platformService,
		categoryService:  categoryService,
		listingService:   listingService,
		broadcastService: broadcastService,
	}
}

func (h *GameHandler) Platforms(c *gin.Context) {
	data, err := h.platformService.List(c.Request.Context())
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取游戏平台失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *GameHandler) Games(c *gin.Context) {
	platformID, _ := strconv.Atoi(c.Query("platform_id"))
	categoryID, _ := strconv.Atoi(c.Query("category_id"))
	data, err := h.listingService.List(c.Request.Context(), platformID, categoryID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取游戏列表失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *GameHandler) WaliGames(c *gin.Context) {
	categoryID, _ := strconv.Atoi(c.Query("category_id"))
	c.Header("X-Served-By", "newbie")
	if categoryID == 5 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: -9999, ErrMsg: "您还没有登录", Data: map[string]interface{}{}})
		return
	}

	data, err := h.listingService.List(c.Request.Context(), 1, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取游戏列表失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *GameHandler) Broadcasts(c *gin.Context) {
	data, err := h.broadcastService.List(c.Request.Context())
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取游戏广播失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *GameHandler) Categories(c *gin.Context) {
	parentID, _ := strconv.Atoi(c.Query("parent_id"))
	data, err := h.categoryService.List(c.Request.Context(), parentID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取游戏分类失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}
