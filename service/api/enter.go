package api

import "SService/service"

type ApiGroup struct {
	ExampleApi
	UserApi
	AccountApi
	TransactionApi
	ResourceApi
	AccrualApi
	MetaApi
	CategoryApi
}

var (
	exampleService  = service.ExampleService{}
	userService     = service.UserService{}
	ledgerService   = service.LedgerService{}
	resourceService = service.ResourceService{}
	accrualViewSvc  = service.AccrualViewService{}
	metaService     = service.MetaService{}
	categoryService = service.CategoryService{}
)
