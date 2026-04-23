package api

import (
	"SService/dao"
	"SService/model"
	"SService/util"
	"SService/util/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccountApi struct{}

// List GET /api/accounts
func (h *AccountApi) List(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	includeArchived := c.Query("include_archived") == "true"
	accs, err := dao.ListAccounts(uint64(claims.ID), includeArchived)
	if err != nil {
		response.FailWithMessage("查询账户失败: "+err.Error(), c)
		return
	}
	response.OkWithData(accs, c)
}

// Create POST /api/accounts
func (h *AccountApi) Create(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	var req struct {
		Name     string      `json:"name" binding:"required,max=64"`
		Balance  model.Money `json:"balance"`
		Currency string      `json:"currency"`
		Note     string      `json:"note"`
		Sort     int         `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if req.Currency == "" {
		req.Currency = "CNY"
	}
	acc := &model.Account{
		UserID:   uint64(claims.ID),
		Name:     req.Name,
		Balance:  req.Balance,
		Currency: req.Currency,
		Note:     req.Note,
		Sort:     req.Sort,
	}
	if err := dao.CreateAccount(nil, acc); err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(acc, c)
}

// Update PUT /api/accounts/:id
func (h *AccountApi) Update(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name       *string `json:"name"`
		Note       *string `json:"note"`
		Sort       *int    `json:"sort"`
		IsArchived *bool   `json:"is_archived"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Note != nil {
		updates["note"] = *req.Note
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.IsArchived != nil {
		updates["is_archived"] = *req.IsArchived
	}
	if len(updates) == 0 {
		response.FailWithMessage("无可更新字段", c)
		return
	}
	if err := dao.UpdateAccount(nil, uint64(claims.ID), id, updates); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	acc, _ := dao.GetAccount(uint64(claims.ID), id)
	response.OkWithData(acc, c)
}

// Delete DELETE /api/accounts/:id
func (h *AccountApi) Delete(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := dao.DeleteAccount(uint64(claims.ID), id); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.Ok(c)
}

// RebuildBalance POST /api/accounts/:id/rebuild-balance
func (h *AccountApi) RebuildBalance(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	net, err := ledgerService.RebuildBalance(uint64(claims.ID), id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"balance": net.String()}, c)
}
