package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	historyRepo "xj_comp/internal/repository/history"
	historyService "xj_comp/internal/service/history"
)

type HistoryHandler struct {
	service *historyService.Service
}

func NewHistoryHandler(service *historyService.Service) *HistoryHandler {
	return &HistoryHandler{service: service}
}

func (h *HistoryHandler) PlayListing(c *gin.Context) {
	h.listing(c, historyRepo.KindPlay, "获取播放记录失败")
}

func (h *HistoryHandler) DownListing(c *gin.Context) {
	h.listing(c, historyRepo.KindDown, "获取下载记录失败")
}

func (h *HistoryHandler) MiniPlayListing(c *gin.Context) {
	h.listing(c, historyRepo.KindMiniPlay, "获取小视频播放记录失败")
}

func (h *HistoryHandler) PlayRemove(c *gin.Context) {
	h.remove(c, historyRepo.KindPlay, "删除播放记录失败")
}

func (h *HistoryHandler) DownRemove(c *gin.Context) {
	h.remove(c, historyRepo.KindDown, "删除下载记录失败")
}

func (h *HistoryHandler) MiniPlayRemove(c *gin.Context) {
	h.remove(c, historyRepo.KindMiniPlay, "删除小视频播放记录失败")
}

func (h *HistoryHandler) listing(c *gin.Context, kind historyRepo.Kind, errmsg string) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	timeline, _ := strconv.Atoi(inputValue(c, "timeline"))
	data, err := h.service.Listing(c.Request.Context(), authToken(c), kind, page, timeline, c.GetHeader("x-cookie-auth") != "")
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *HistoryHandler) remove(c *gin.Context, kind historyRepo.Kind, errmsg string) {
	vodid, _ := strconv.Atoi(inputValue(c, "vodid"))
	vodids := commaInts(inputValue(c, "vodids"))
	if vodid > 0 {
		vodids = []int{vodid}
	}
	msg, err := h.service.Remove(c.Request.Context(), authToken(c), kind, vodids)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: msg})
}
