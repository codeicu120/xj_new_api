package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	gameService "xj_comp/internal/service/game"
	"xj_comp/internal/service/resourceurl"
)

type GameHandler struct {
	platformService  *gameService.PlatformService
	categoryService  *gameService.CategoryService
	listingService   *gameService.ListingService
	broadcastService *gameService.BroadcastService
	waliService      *gameService.WaliService
	lotteryService   *gameService.LotteryService
}

func NewGameHandler(platformService *gameService.PlatformService, categoryService *gameService.CategoryService, listingService *gameService.ListingService, broadcastService *gameService.BroadcastService, waliService *gameService.WaliService, lotteryService ...*gameService.LotteryService) *GameHandler {
	var lottery *gameService.LotteryService
	if len(lotteryService) > 0 {
		lottery = lotteryService[0]
	}
	return &GameHandler{
		platformService:  platformService,
		categoryService:  categoryService,
		listingService:   listingService,
		broadcastService: broadcastService,
		waliService:      waliService,
		lotteryService:   lottery,
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
	data, err := h.listingService.ListForRequest(c.Request.Context(), platformID, categoryID, resourceurl.Request{HasCookieAuth: isH5Request(c), ClientIP: c.ClientIP()})
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

	data, err := h.listingService.ListForRequest(c.Request.Context(), 1, categoryID, resourceurl.Request{HasCookieAuth: isH5Request(c), ClientIP: c.ClientIP()})
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取游戏列表失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *GameHandler) LotteryGames(c *gin.Context) {
	categoryID, _ := strconv.Atoi(c.Query("category_id"))
	c.Header("X-Served-By", "newbie")
	if categoryID == 5 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: -9999, ErrMsg: "您还没有登录", Data: map[string]interface{}{}})
		return
	}

	data, err := h.listingService.ListForRequest(c.Request.Context(), 0, categoryID, resourceurl.Request{HasCookieAuth: isH5Request(c), ClientIP: c.ClientIP()})
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

func (h *GameHandler) WaliTest(c *gin.Context) {
	data, err := h.waliService.Ping(c.Request.Context())
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusOK, legacyjson.Error("测试失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *GameHandler) WaliBalance(c *gin.Context) {
	data, retcode, errmsg, err := h.waliService.Balance(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *GameHandler) WaliEnter(c *gin.Context) {
	gameURL, retcode, errmsg, err := h.waliService.Enter(c.Request.Context(), authToken(c), inputValue(c, "game"))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(gameURL))
}

func (h *GameHandler) LotteryEnter(c *gin.Context) {
	gameURL, retcode, errmsg, err := h.lotteryService.Enter(c.Request.Context(), authToken(c), inputValue(c, "lotid"))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(gameURL))
}

func (h *GameHandler) LotteryBalance(c *gin.Context) {
	data, retcode, errmsg, err := h.lotteryService.Balance(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: map[string]interface{}{}})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(map[string]interface{}{"data": data}))
}

func (h *GameHandler) HighRiskAction(message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		retcode, errmsg, err := h.waliService.ActionEdge(c.Request.Context(), authToken(c), message)
		c.Header("X-Served-By", "newbie")
		if err != nil {
			c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
			return
		}
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
	}
}

func (h *GameHandler) TransferTopup(message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		retcode, errmsg, err := h.waliService.TopupEdge(c.Request.Context(), authToken(c), inputValue(c, "amount"), message)
		c.Header("X-Served-By", "newbie")
		if err != nil {
			c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
			return
		}
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
	}
}

func (h *GameHandler) TransferWithdraw(message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		retcode, errmsg, err := h.waliService.WithdrawEdge(c.Request.Context(), authToken(c), inputValue(c, "amount"), message)
		c.Header("X-Served-By", "newbie")
		if err != nil {
			c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
			return
		}
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
	}
}

func (h *GameHandler) Categories(c *gin.Context) {
	parentID, _ := strconv.Atoi(c.Query("parent_id"))
	data, err := h.categoryService.ListForRequest(c.Request.Context(), parentID, resourceurl.Request{HasCookieAuth: isH5Request(c), ClientIP: c.ClientIP()})
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取游戏分类失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}
