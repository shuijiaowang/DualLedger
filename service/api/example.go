package api

import (
	"SService/util"
	"SService/util/response"

	"github.com/gin-gonic/gin"
)

type ExampleApi struct{}

// Test 需 JWT：演示鉴权后的最小接口，可替换为真实业务。
func (h *ExampleApi) Test(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	msg := exampleService.AddExample(uint(claims.ID))
	response.OkWithData(gin.H{"message": msg}, c)
}
