package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	boughtService "xj_comp/internal/service/bought"
)

type BoughtHandler struct {
	service *boughtService.Service
}

func NewBoughtHandler(service *boughtService.Service) *BoughtHandler {
	return &BoughtHandler{service: service}
}

func (h *BoughtHandler) Listing(c *gin.Context) {
	page, _ := strconv.Atoi(inputValue(c, "page"))
	data, retcode, errmsg, err := h.service.Listing(c.Request.Context(), authToken(c), page, c.GetHeader("x-cookie-auth") != "")
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

func (h *BoughtHandler) Delete(c *gin.Context) {
	retcode, errmsg, err := h.service.Delete(c.Request.Context(), authToken(c), commaInts(inputValue(c, "vodids")))
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

func (h *BoughtHandler) Buy(c *gin.Context) {
	vodID, _ := strconv.Atoi(inputValue(c, "vodid"))
	if vodID == 0 {
		vodID, _ = strconv.Atoi(c.Param("vodid"))
	}
	retcode, errmsg, err := h.service.Buy(c.Request.Context(), authToken(c), vodID)
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

func commaInts(value string) []int {
	if strings.TrimSpace(value) == "" {
		return []int{}
	}
	parts := strings.Split(value, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		n, _ := strconv.Atoi(strings.TrimSpace(part))
		out = append(out, n)
	}
	return out
}
