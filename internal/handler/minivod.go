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

func (h *MiniVODHandler) ReqList(c *gin.Context) {
	debug, _ := strconv.Atoi(c.Query("debug"))
	data, err := h.service.ReqList(c.Request.Context(), authToken(c), c.GetHeader("x-cookie-auth") != "", debug)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取小视频请求列表失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *MiniVODHandler) Up(c *gin.Context) {
	h.vote(c, true)
}

func (h *MiniVODHandler) Down(c *gin.Context) {
	h.vote(c, false)
}

func (h *MiniVODHandler) ReqLong(c *gin.Context) {
	vodID, _ := strconv.Atoi(c.Param("vodid"))
	body, retcode, errmsg, err := h.service.ReqLong(c.Request.Context(), authToken(c), vodID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
		return
	}
	c.Data(http.StatusOK, "text/html", []byte(body))
}

func (h *MiniVODHandler) ParseLongM3U8(c *gin.Context) {
	vodID, _ := strconv.Atoi(c.Param("vodid"))
	_, retcode, errmsg, err := h.service.ReqLong(c.Request.Context(), authToken(c), vodID)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
		return
	}
	c.JSON(http.StatusOK, legacyjson.Error("m3u8解析成功分支暂未迁移"))
}

func (h *MiniVODHandler) ReqCoin(c *gin.Context) {
	logid, _ := strconv.Atoi(inputValue(c, "logid"))
	retcode, errmsg, err := h.service.ReqCoin(c.Request.Context(), authToken(c), logid)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}

func (h *MiniVODHandler) ThrowCoin(c *gin.Context) {
	vodID, _ := strconv.Atoi(c.Param("vodid"))
	coin, _ := strconv.Atoi(inputValue(c, "coinnum"))
	data, retcode, errmsg, err := h.service.ThrowCoinEdge(c.Request.Context(), minivodService.ThrowCoinRequest{
		Token:  authToken(c),
		VODID:  vodID,
		Method: c.Request.Method,
		Coin:   coin,
	})
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: data})
}

func (h *MiniVODHandler) ReqPlay(c *gin.Context) {
	h.reqMedia(c, true)
}

func (h *MiniVODHandler) ReqDown(c *gin.Context) {
	h.reqMedia(c, false)
}

func (h *MiniVODHandler) reqMedia(c *gin.Context, play bool) {
	vodID, _ := strconv.Atoi(c.Param("vodid"))
	playIndex, _ := strconv.Atoi(c.Query("playindex"))
	var (
		data    map[string]interface{}
		retcode int
		errmsg  string
		err     error
	)
	if play {
		data, retcode, errmsg, err = h.service.ReqPlay(c.Request.Context(), authToken(c), vodID, playIndex)
	} else {
		data, retcode, errmsg, err = h.service.ReqDown(c.Request.Context(), authToken(c), vodID, playIndex)
	}
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	if retcode != 0 {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg, Data: data})
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: errmsg, Data: data})
}

func (h *MiniVODHandler) vote(c *gin.Context, up bool) {
	vodID, _ := strconv.Atoi(c.Param("vodid"))
	retcode, errmsg, err := h.service.Vote(c.Request.Context(), authToken(c), vodID, up)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
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
