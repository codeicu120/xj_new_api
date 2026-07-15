package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	indexService "xj_comp/internal/service/index"
)

type IndexHandler struct {
	certService *indexService.CertService
}

func NewIndexHandler(certService *indexService.CertService) *IndexHandler {
	return &IndexHandler{certService: certService}
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
