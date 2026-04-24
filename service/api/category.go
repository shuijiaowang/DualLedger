package api

import (
	"SService/dao"
	"SService/model"
	"SService/util/response"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CategoryApi struct{}

func (h *CategoryApi) List(c *gin.Context) {
	rows, err := categoryService.List()
	if err != nil {
		response.FailWithMessage("查询分类失败: "+err.Error(), c)
		return
	}
	response.OkWithData(rows, c)
}

func (h *CategoryApi) Create(c *gin.Context) {
	var req struct {
		Code       string `json:"code"`
		Name       string `json:"name"`
		Kind       string `json:"kind"`
		ParentCode string `json:"parent_code"`
		ParentName string `json:"parent_name"`
		Icon       string `json:"icon"`
		Sort       int    `json:"sort"`
		Source     string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	parentCode := strings.TrimSpace(req.ParentCode)
	if parentCode == "" && strings.TrimSpace(req.ParentName) != "" {
		parent, err := dao.GetCategoryByName(strings.TrimSpace(req.ParentName))
		if err != nil {
			response.FailWithMessage("父分类名称不存在: "+req.ParentName, c)
			return
		}
		parentCode = parent.Code
	}
	entity := model.CategoryEntity{
		Code:       strings.TrimSpace(req.Code),
		Name:       strings.TrimSpace(req.Name),
		Kind:       strings.TrimSpace(req.Kind),
		ParentCode: parentCode,
		Icon:       strings.TrimSpace(req.Icon),
		Sort:       req.Sort,
		Source:     strings.TrimSpace(req.Source),
	}
	row, err := categoryService.Create(entity)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(row, c)
}

func (h *CategoryApi) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if parentNameRaw, ok := req["parent_name"]; ok {
		parentName, _ := parentNameRaw.(string)
		parentName = strings.TrimSpace(parentName)
		if parentName == "" {
			req["parent_code"] = ""
		} else {
			parent, err := dao.GetCategoryByName(parentName)
			if err != nil {
				response.FailWithMessage("父分类名称不存在: "+parentName, c)
				return
			}
			req["parent_code"] = parent.Code
		}
		delete(req, "parent_name")
	}
	if err := categoryService.Update(id, req); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.Ok(c)
}

func (h *CategoryApi) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := categoryService.Delete(id); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.Ok(c)
}
