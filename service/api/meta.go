package api

import (
	"SService/util/response"

	"github.com/gin-gonic/gin"
)

type MetaApi struct{}

// Categories GET /api/meta/categories
func (h *MetaApi) Categories(c *gin.Context) {
	response.OkWithData(metaService.Categories(), c)
}

// Tags GET /api/meta/tags
func (h *MetaApi) Tags(c *gin.Context) {
	response.OkWithData(metaService.Tags(), c)
}

// Enums GET /api/meta/enums
func (h *MetaApi) Enums(c *gin.Context) {
	response.OkWithData(metaService.Enums(), c)
}
