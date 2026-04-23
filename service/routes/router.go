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

	exampleApi := api.ExampleApi{}
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.JWTInterceptor())
	{
		apiGroup.POST("/example/test", exampleApi.Test)
	}

	return r
}
