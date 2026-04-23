package api

import (
	"SService/util"
	"SService/util/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserApi struct{}

func (h *UserApi) Register(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname" binding:"required,min=2,max=20"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6,max=64"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("无效的请求格式：请输入合法昵称、邮箱和密码", c)
		return
	}

	if err := userService.Register(req.Nickname, req.Email, req.Password); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

func (h *UserApi) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("无效的请求格式", c)
		return
	}

	user, ok := userService.Login(req.Email, req.Password)
	if !ok {
		response.FailWithMessage("邮箱或密码错误", c)
		return
	}
	userUUID, err := uuid.Parse(user.UUID)
	if err != nil {
		response.FailWithMessage("UUID格式错误", c)
		return
	}
	token, err := util.GenerateToken(int(user.ID), user.Email, user.Nickname, userUUID)
	if err != nil {
		response.FailWithMessage("生成令牌失败", c)
		return
	}

	response.OkWithData(gin.H{
		"id":       user.ID,
		"nickname": user.Nickname,
		"email":    user.Email,
		"uuid":     user.UUID,
		"token":    token,
	}, c)
}
