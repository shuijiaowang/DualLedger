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

type AccrualApi struct{}

// View GET /api/accrual-view?from=&to=&include_cashonly=&include_tx=
func (h *AccrualApi) View(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	from, err := parseDate(c.Query("from"))
	if err != nil {
		response.FailWithMessage("from 参数格式错误（YYYY-MM-DD 或 RFC3339）", c)
		return
	}
	to, err := parseDate(c.Query("to"))
	if err != nil {
		response.FailWithMessage("to 参数格式错误", c)
		return
	}
	includeTx := c.DefaultQuery("include_tx", "true") == "true"
	includeCashOnly := c.Query("include_cashonly") == "true"

	rows, err := accrualViewSvc.Query(service.ViewQuery{
		UserID:          uint64(claims.ID),
		From:            from,
		To:              to,
		IncludeTx:       includeTx,
		IncludeCashOnly: includeCashOnly,
	})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"rows": rows}, c)
}

// Create POST /api/accrual-entries
// 手动录入 ADJUST / MANUAL / END_SETTLE
func (h *AccrualApi) Create(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	var req struct {
		TransactionID *uint64     `json:"transaction_id"`
		ResourceID    *uint64     `json:"resource_id"`
		CategoryCode  string      `json:"category_code"`
		Amount        model.Money `json:"amount" binding:"required"`
		Qty           *float64    `json:"qty"`
		Unit          string      `json:"unit"`
		AccrueAt      *time.Time  `json:"accrue_at"`
		Source        string      `json:"source" binding:"required"`
		Tags          []string    `json:"tags"`
		Note          string      `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if !model.IsValidAccrualSource(req.Source) {
		response.FailWithMessage("非法 source", c)
		return
	}
	accrueAt := time.Now()
	if req.AccrueAt != nil {
		accrueAt = *req.AccrueAt
	}
	e := &model.AccrualEntry{
		UserID:        uint64(claims.ID),
		TransactionID: req.TransactionID,
		ResourceID:    req.ResourceID,
		CategoryCode:  req.CategoryCode,
		Amount:        req.Amount,
		Qty:           req.Qty,
		Unit:          req.Unit,
		AccrueAt:      accrueAt,
		Source:        req.Source,
		Tags:          model.JSONStrings(req.Tags),
		Note:          req.Note,
	}
	if err := dao.CreateAccrualEntry(nil, e); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(e, c)
}

// List GET /api/accrual-entries
func (h *AccrualApi) List(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	q := dao.AccrualQuery{UserID: uint64(claims.ID)}
	if f, err := parseDate(c.Query("from")); err == nil {
		q.From = &f
	}
	if t, err := parseDate(c.Query("to")); err == nil {
		q.To = &t
	}
	if s := c.Query("resource_id"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			q.ResourceID = &v
		}
	}
	q.CategoryCode = c.Query("category_code")
	q.Tag = c.Query("tag")
	q.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "200"))
	q.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	rows, err := dao.ListAccrualEntries(q)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"rows": rows}, c)
}

// Delete DELETE /api/accrual-entries/:id
func (h *AccrualApi) Delete(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := dao.DeleteAccrualEntry(uint64(claims.ID), id); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.Ok(c)
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errTimeEmpty
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
}

var errTimeEmpty = &timeEmptyErr{}

type timeEmptyErr struct{}

func (*timeEmptyErr) Error() string { return "time empty" }
