package api

import (
	"SService/util"
	"SService/util/response"

	"github.com/gin-gonic/gin"
)

type DevDataApi struct{}

func (h *DevDataApi) Reset(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	if err := devDataService.ResetForUser(uint64(claims.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已清空当前用户业务数据（保留 user 表）", c)
}
