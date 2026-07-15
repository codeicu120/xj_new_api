package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	exploreService "xj_comp/internal/service/explore"
)

type ExploreHandler struct {
	service *exploreService.Service
}

func NewExploreHandler(service *exploreService.Service) *ExploreHandler {
	return &ExploreHandler{service: service}
}

func (h *ExploreHandler) Index(c *gin.Context) {
	data, err := h.service.Index(c.Request.Context(), authToken(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取发现页失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *ExploreHandler) EmptyOK(c *gin.Context) {
	c.Header("X-Served-By", "newbie")
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: 0, ErrMsg: ""})
}

func (h *ExploreHandler) CleanNotification(c *gin.Context) {
	data, retcode, errmsg, err := h.service.CleanNotification(c.Request.Context(), authToken(c), inputValue(c, "tabkey"))
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

func (h *ExploreHandler) VodTaskShow(c *gin.Context) {
	vid, _ := strconv.Atoi(c.Param("vid"))
	data, retcode, errmsg, err := h.service.VodTaskShow(c.Request.Context(), authToken(c), vid)
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

func (h *ExploreHandler) VodTaskReqCoin(c *gin.Context) {
	logid, _ := strconv.Atoi(inputValue(c, "logid"))
	retcode, errmsg, err := h.service.VodTaskReqCoin(c.Request.Context(), authToken(c), logid)
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error(errmsg))
		return
	}
	c.JSON(http.StatusOK, legacyjson.Response{RetCode: retcode, ErrMsg: errmsg})
}
