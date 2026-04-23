package middleware

import (
	"SService/config"
	"SService/util"
	"SService/util/response"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// JWTInterceptor 验证JWT令牌的中间件
func JWTInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			response.FailWithMessage("未提供token", c)
			c.Abort() //终止后续处理
			return    //返回
		}

		// 检查格式是否为Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.FailWithMessage("token格式错误", c)
			c.Abort()
			return
		}

		// 解析token
		claims, err := util.ParseToken(parts[1])
		if err != nil {
			response.FailWithMessage("无效的token", c)
			c.Abort()
			return
		}
		refreshBeforeMinutes := config.AppConfig.JWT.RefreshBeforeMinutes
		if refreshBeforeMinutes <= 0 {
			refreshBeforeMinutes = 30
		}
		// 当 token 即将过期时刷新
		if claims.ExpiresAt != nil && claims.ExpiresAt.Unix()-time.Now().Unix() < int64(refreshBeforeMinutes*60) {
			newToken, err := util.GenerateToken(claims.ID, claims.Email, claims.Nickname, claims.UUID)
			if err == nil {
				c.Header("new-token", newToken)
			}
		}
		// 将用户ID存入上下文
		//c.Set("userID", claims.UserID)
		//c.Set("email", claims.Email)
		//c.Set("userUUID", claims.UserUUID)
		c.Set("claims", claims)
		c.Next()
	}
}
