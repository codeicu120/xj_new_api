package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	favoriteRepo "xj_comp/internal/repository/favorite"
	favoriteService "xj_comp/internal/service/favorite"
)

type FavoriteHandler struct {
	service *favoriteService.Service
}

func NewFavoriteHandler(service *favoriteService.Service) *FavoriteHandler {
	return &FavoriteHandler{service: service}
}

func (h *FavoriteHandler) Listing(c *gin.Context) {
	h.listing(c, favoriteRepo.KindVOD)
}

func (h *FavoriteHandler) MiniListing(c *gin.Context) {
	h.listing(c, favoriteRepo.KindMini)
}

func (h *FavoriteHandler) MiniV2Listing(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.MiniV2Listing(c.Request.Context(), authToken(c), page, inputValue(c, "wd"), c.GetHeader("x-cookie-auth") != "")
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *FavoriteHandler) Remove(c *gin.Context) {
	h.remove(c, favoriteRepo.KindVOD)
}

func (h *FavoriteHandler) MiniRemove(c *gin.Context) {
	h.remove(c, favoriteRepo.KindMini)
}

func (h *FavoriteHandler) Add(c *gin.Context) {
	h.add(c, favoriteRepo.KindVOD)
}

func (h *FavoriteHandler) MiniAdd(c *gin.Context) {
	h.add(c, favoriteRepo.KindMini)
}

func (h *FavoriteHandler) listing(c *gin.Context, kind favoriteRepo.Kind) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.Listing(c.Request.Context(), authToken(c), kind, page, inputValue(c, "wd"), c.GetHeader("x-cookie-auth") != "")
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *FavoriteHandler) remove(c *gin.Context, kind favoriteRepo.Kind) {
	vodid, _ := strconv.Atoi(inputValue(c, "vodid"))
	vodids := commaInts(inputValue(c, "vodids"))
	if vodid > 0 {
		vodids = []int{vodid}
	}
	retcode, errmsg, err := h.service.Remove(c.Request.Context(), authToken(c), kind, vodids)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}

func (h *FavoriteHandler) add(c *gin.Context, kind favoriteRepo.Kind) {
	vodid, _ := strconv.Atoi(inputValue(c, "vodid"))
	data, retcode, errmsg, err := h.service.Add(c.Request.Context(), authToken(c), kind, vodid)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: errmsg, Data: data})
}
