package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	vodService "xj_comp/internal/service/vod"
)

type VODHandler struct {
	listingService *vodService.ListingService
}

func NewVODHandler(listingService *vodService.ListingService) *VODHandler {
	return &VODHandler{listingService: listingService}
}

func (h *VODHandler) Listing(c *gin.Context) {
	action := vodAction(c.Request.URL.Path)
	params := strings.TrimPrefix(c.Param("params"), "-")
	data, err := h.listingService.List(c.Request.Context(), vodService.ListingRequest{
		Action:      action,
		PathParams:  params,
		QueryPage:   c.Query("page"),
		IsH5Request: c.GetHeader("x-cookie-auth") != "",
	})
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取视频列表失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *VODHandler) LikeRows(c *gin.Context) {
	data, err := h.listingService.LikeRows(c.Request.Context(), c.GetHeader("x-cookie-auth") != "")
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取猜你喜欢失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *VODHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	data, err := h.listingService.Search(
		c.Request.Context(),
		c.Query("wd"),
		c.Query("free") == "1",
		page,
		c.GetHeader("x-cookie-auth") != "",
	)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("搜索失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *VODHandler) MiniSearch(c *gin.Context) {
	page, _ := strconv.Atoi(miniSearchInput(c, "page"))
	data, err := h.listingService.MiniSearch(
		c.Request.Context(),
		miniSearchInput(c, "wd"),
		page,
		c.GetHeader("x-cookie-auth") != "",
	)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("搜索失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func miniSearchInput(c *gin.Context, key string) string {
	if value := c.PostForm(key); value != "" {
		return value
	}
	return c.Query(key)
}

func (h *VODHandler) Show(c *gin.Context) {
	vodID, _ := strconv.Atoi(c.Param("vodid"))
	data, err := h.listingService.Show(c.Request.Context(), vodID, c.GetHeader("x-cookie-auth") != "")
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, vodService.ErrVODNotFound) {
		c.JSON(http.StatusOK, legacyjson.Error("记录不存在或已删除"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取视频详情失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *VODHandler) Preview(c *gin.Context) {
	vodID, _ := strconv.Atoi(c.Param("vodid"))
	body, err := h.listingService.Preview(c.Request.Context(), vodID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.Data(http.StatusOK, "vnd.apple.mpegurl", nil)
		return
	}
	c.Data(http.StatusOK, "vnd.apple.mpegurl", []byte(body))
}

func (h *VODHandler) Up(c *gin.Context) {
	h.vote(c, true)
}

func (h *VODHandler) Down(c *gin.Context) {
	h.vote(c, false)
}

func (h *VODHandler) Breaking(c *gin.Context) {
	data, retcode, errmsg, err := h.listingService.Breaking(c.Request.Context())
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

func (h *VODHandler) ErrorReport(c *gin.Context) {
	vodID, _ := strconv.Atoi(inputValue(c, "vodid"))
	retcode, errmsg, err := h.listingService.ErrorReport(c.Request.Context(), vodService.ErrorReportRequest{
		Token:      authToken(c),
		VODID:      vodID,
		PlayURL:    inputValue(c, "play_url"),
		AppVersion: inputValue(c, "app_ver"),
		SysVersion: inputValue(c, "sys_ver"),
		Model:      inputValue(c, "model"),
		Channel:    inputValue(c, "channel"),
		Network:    inputValue(c, "network"),
		Details:    inputValue(c, "details"),
		ClientIP:   c.ClientIP(),
	})
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: ""})
}

func (h *VODHandler) vote(c *gin.Context, up bool) {
	vodID, _ := strconv.Atoi(c.Param("vodid"))
	retcode, errmsg, err := h.listingService.Vote(c.Request.Context(), authToken(c), vodID, up)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}

func vodAction(path string) string {
	path = strings.TrimPrefix(path, "/v2/vod/")
	path = strings.TrimPrefix(path, "/vod/")
	if index := strings.Index(path, "-"); index >= 0 {
		path = path[:index]
	}
	return path
}
