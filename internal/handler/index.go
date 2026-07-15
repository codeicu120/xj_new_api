package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	indexService "xj_comp/internal/service/index"
)

type IndexHandler struct {
	certService   *indexService.CertService
	globalService *indexService.GlobalService
	initService   *indexService.InitService
	homeService   *indexService.HomeService
	coverService  *indexService.CoverService
}

func NewIndexHandler(certService *indexService.CertService, globalService *indexService.GlobalService, initService *indexService.InitService, homeService *indexService.HomeService, coverService *indexService.CoverService) *IndexHandler {
	handler := &IndexHandler{certService: certService, globalService: globalService}
	handler.initService = initService
	handler.homeService = homeService
	handler.coverService = coverService
	return handler
}

func (h *IndexHandler) Index(c *gin.Context) {
	data, err := h.homeService.Index(c.Request.Context(), isH5Request(c))
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取首页失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *IndexHandler) GetCertUUID(c *gin.Context) {
	data, err := h.certService.GetCertUUID(c.Request.Context(), c.Query("uuid"))
	c.Header("X-Served-By", "newbie")
	if indexService.IsCertNotFound(err) {
		c.JSON(http.StatusOK, legacyjson.Response{RetCode: -1, ErrMsg: "记录不存在或已被删除", Data: map[string]interface{}{}})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取证书失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(map[string]interface{}{"data": data}))
}

func (h *IndexHandler) GetGlobalData(c *gin.Context) {
	data, err := h.globalService.GetGlobalData(c.Request.Context(), indexService.GlobalRequest{
		Pkg:       firstNonEmpty(c.Query("pkg"), c.GetHeader("x-channel")),
		Version:   c.Query("ver"),
		XVersion:  c.GetHeader("x-version"),
		UserAgent: c.GetHeader("user-agent"),
	})
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取全局配置失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *IndexHandler) Init(c *gin.Context) {
	data, err := h.initService.Init(c.Request.Context(), indexService.InitRequest{
		Token:     authToken(c),
		Pkg:       firstNonEmpty(c.Query("pkg"), c.GetHeader("x-channel")),
		Version:   c.Query("ver"),
		XVersion:  c.GetHeader("x-version"),
		UserAgent: c.GetHeader("user-agent"),
		ClientIP:  c.ClientIP(),
	})
	c.Header("X-Served-By", "newbie")
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("初始化失败"))
		return
	}
	c.JSON(http.StatusOK, legacyjson.OK(data))
}

func (h *IndexHandler) GetCover(c *gin.Context) {
	data, err := h.coverService.GetCover(c.Request.Context(), c.Query("pic"))
	c.Header("X-Served-By", "newbie")
	if errors.Is(err, indexService.ErrCoverNotFound) {
		c.JSON(http.StatusOK, legacyjson.Error("记录不存在或已被删除"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, legacyjson.Error("获取封面失败"))
		return
	}
	c.Header("Cache-Control", "max-age=86400")
	c.JSON(http.StatusOK, legacyjson.OK(map[string]interface{}{"data": data}))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func isH5Request(c *gin.Context) bool {
	return c.GetHeader("x-cookie-auth") != ""
}
