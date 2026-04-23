package api

import (
	"SService/dao"
	"SService/model"
	"SService/service"
	"SService/util"
	"SService/util/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ResourceApi struct{}

type resourceCreateReq struct {
	Name         string             `json:"name" binding:"required"`
	CategoryCode string             `json:"category_code"`
	Unit         string             `json:"unit"`
	TotalQty     *float64           `json:"total_qty"`
	TotalCost    model.Money        `json:"total_cost" binding:"required"`
	AmortizeRule model.AmortizeRule `json:"amortize_rule" binding:"required"`
	PurchaseAt   *time.Time         `json:"purchase_at"`
	StartUseAt   *time.Time         `json:"start_use_at"`
	Note         string             `json:"note"`
	Ext          model.JSONMap      `json:"ext"`

	AccountID uint64 `json:"account_id"`
	TxType    string `json:"tx_type"` // EXPENSE / INCOME / ""（不创建交易）
	TxTitle   string `json:"tx_title"`
}

// Create POST /api/resources
func (h *ResourceApi) Create(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	var req resourceCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	purchase := time.Now()
	if req.PurchaseAt != nil {
		purchase = *req.PurchaseAt
	}
	in := service.CreateResourceInput{
		UserID:       uint64(claims.ID),
		Name:         req.Name,
		CategoryCode: req.CategoryCode,
		Unit:         req.Unit,
		TotalQty:     req.TotalQty,
		TotalCost:    req.TotalCost,
		AmortizeRule: req.AmortizeRule,
		PurchaseAt:   purchase,
		StartUseAt:   req.StartUseAt,
		Note:         req.Note,
		Ext:          req.Ext,
		AccountID:    req.AccountID,
		TxType:       req.TxType,
		TxTitle:      req.TxTitle,
	}
	out, err := resourceService.Create(in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(out, c)
}

// List GET /api/resources
func (h *ResourceApi) List(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	statuses := []string{}
	if s := c.Query("statuses"); s != "" {
		statuses = splitCSV(s)
	}
	rs, err := dao.ListResources(uint64(claims.ID), statuses)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(rs, c)
}

// Get GET /api/resources/:id
func (h *ResourceApi) Get(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	r, err := dao.GetResource(uint64(claims.ID), id)
	if err != nil {
		response.FailWithMessage("未找到: "+err.Error(), c)
		return
	}
	response.OkWithData(r, c)
}

// Punch POST /api/resources/:id/punch
func (h *ResourceApi) Punch(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Qty       float64    `json:"qty" binding:"required"`
		AccrueAt  *time.Time `json:"accrue_at"`
		Tags      []string   `json:"tags"`
		Note      string     `json:"note"`
		MarkEnded bool       `json:"mark_ended"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	accrueAt := time.Now()
	if req.AccrueAt != nil {
		accrueAt = *req.AccrueAt
	}
	e, err := resourceService.Punch(service.PunchInput{
		UserID:     uint64(claims.ID),
		ResourceID: id,
		Qty:        req.Qty,
		AccrueAt:   accrueAt,
		Tags:       req.Tags,
		Note:       req.Note,
		MarkEnded:  req.MarkEnded,
	})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(e, c)
}

// End POST /api/resources/:id/end
func (h *ResourceApi) End(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Status      string     `json:"status" binding:"required"`
		AccrueAt    *time.Time `json:"accrue_at"`
		Note        string     `json:"note"`
		WriteSettle bool       `json:"write_settle"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	accrueAt := time.Now()
	if req.AccrueAt != nil {
		accrueAt = *req.AccrueAt
	}
	r, e, err := resourceService.End(service.EndResourceInput{
		UserID:      uint64(claims.ID),
		ResourceID:  id,
		Status:      req.Status,
		AccrueAt:    accrueAt,
		Note:        req.Note,
		WriteSettle: req.WriteSettle,
	})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"resource": r, "settle_entry": e}, c)
}
