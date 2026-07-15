package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	indexService "xj_comp/internal/service/index"
)

type IndexHandler struct {
	certService   *indexService.CertService
	globalService *indexService.GlobalService
}

func NewIndexHandler(certService *indexService.CertService, globalService *indexService.GlobalService) *IndexHandler {
	return &IndexHandler{certService: certService, globalService: globalService}
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
