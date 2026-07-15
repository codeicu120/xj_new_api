package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	aiundressService "xj_comp/internal/service/aiundress"
)

type AIUndressHandler struct {
	service *aiundressService.Service
}

func NewAIUndressHandler(service *aiundressService.Service) *AIUndressHandler {
	return &AIUndressHandler{service: service}
}

func (h *AIUndressHandler) Listing(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	module, _ := strconv.Atoi(inputValue(c, "module"))
	data, retcode, errmsg, err := h.service.Listing(c.Request.Context(), authToken(c), page, module)
	respondLegacy(c, data, retcode, errmsg, err)
}

func (h *AIUndressHandler) Upload(c *gin.Context) {
	retcode, errmsg, err := h.service.RequireLoginEdge(c.Request.Context(), authToken(c), "AI 上传成功分支暂未迁移")
	respondLegacy(c, nil, retcode, errmsg, err)
}

func (h *AIUndressHandler) Undress(c *gin.Context) {
	module, _ := strconv.Atoi(inputValue(c, "module"))
	retcode, errmsg, err := h.service.UndressEdge(c.Request.Context(), authToken(c), inputValue(c, "uri"), module)
	respondLegacy(c, nil, retcode, errmsg, err)
}

func (h *AIUndressHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(inputValue(c, "id"))
	retcode, errmsg, err := h.service.DeleteEdge(c.Request.Context(), authToken(c), id)
	respondLegacy(c, nil, retcode, errmsg, err)
}

func (h *AIUndressHandler) ModuleList(c *gin.Context) {
	data, retcode, errmsg, err := h.service.ModuleList(c.Request.Context())
	respondLegacy(c, data, retcode, errmsg, err)
}

func (h *AIUndressHandler) ResourceTypeList(c *gin.Context) {
	data, retcode, errmsg, err := h.service.ResourceTypeList(c.Request.Context(), inputValue(c, "module"))
	respondLegacy(c, data, retcode, errmsg, err)
}

func (h *AIUndressHandler) ResourceList(c *gin.Context) {
	pageSize, _ := strconv.Atoi(inputValue(c, "pageSize"))
	data, retcode, errmsg, err := h.service.ResourceList(c.Request.Context(), aiundressService.ResourceListInput{
		Module:   inputValue(c, "module"),
		TypeID:   inputValue(c, "typeId"),
		PageSize: pageSize,
		Current:  inputValue(c, "page"),
	})
	respondLegacy(c, data, retcode, errmsg, err)
}

func respondLegacy(c *gin.Context, data interface{}, retcode int, errmsg string, err error) {
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
