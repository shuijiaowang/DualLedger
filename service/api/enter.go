package api

import "SService/service"

type ApiGroup struct {
	ExampleApi
	UserApi
}

var (
	exampleService = service.ExampleService{}
	userService    = service.UserService{}
)
