package routes

import (
	"SService/api"
	"SService/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Cors())
	r.Use(middleware.ErrorHandler())

	userApi := api.UserApi{}
	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/login", userApi.Login)
		userGroup.POST("/register", userApi.Register)
	}

	// 元数据：公开读取（前端登录前也可预加载）
	metaApi := api.MetaApi{}
	metaGroup := r.Group("/api/meta")
	{
		metaGroup.GET("/categories", metaApi.Categories)
		metaGroup.GET("/tags", metaApi.Tags)
		metaGroup.GET("/enums", metaApi.Enums)
	}

	// 以下接口统一需要 JWT
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.JWTInterceptor())
	{
		exampleApi := api.ExampleApi{}
		apiGroup.POST("/example/test", exampleApi.Test)

		accountApi := api.AccountApi{}
		apiGroup.GET("/accounts", accountApi.List)
		apiGroup.POST("/accounts", accountApi.Create)
		apiGroup.PUT("/accounts/:id", accountApi.Update)
		apiGroup.DELETE("/accounts/:id", accountApi.Delete)
		apiGroup.POST("/accounts/:id/rebuild-balance", accountApi.RebuildBalance)

		categoryApi := api.CategoryApi{}
		apiGroup.GET("/categories", categoryApi.List)
		apiGroup.POST("/categories", categoryApi.Create)
		apiGroup.PUT("/categories/:id", categoryApi.Update)
		apiGroup.DELETE("/categories/:id", categoryApi.Delete)

		txApi := api.TransactionApi{}
		apiGroup.GET("/transactions", txApi.List)
		apiGroup.POST("/transactions", txApi.Create)
		apiGroup.GET("/transactions/:id", txApi.Get)
		apiGroup.DELETE("/transactions/:id", txApi.Delete)

		resApi := api.ResourceApi{}
		apiGroup.GET("/resources", resApi.List)
		apiGroup.POST("/resources", resApi.Create)
		apiGroup.GET("/resources/:id", resApi.Get)
		apiGroup.POST("/resources/:id/punch", resApi.Punch)
		apiGroup.POST("/resources/:id/end", resApi.End)

		accrualApi := api.AccrualApi{}
		apiGroup.GET("/accrual-view", accrualApi.View)
		apiGroup.GET("/accrual-entries", accrualApi.List)
		apiGroup.POST("/accrual-entries", accrualApi.Create)
		apiGroup.DELETE("/accrual-entries/:id", accrualApi.Delete)
	}

	return r
}
