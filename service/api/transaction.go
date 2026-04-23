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

type TransactionApi struct{}

// txCreateReq 创建交易入参
type txCreateReq struct {
	Type         string        `json:"type" binding:"required"`
	Direction    string        `json:"direction"`
	OccurAt      *time.Time    `json:"occur_at"`
	Amount       model.Money   `json:"amount" binding:"required"`
	AccountID    uint64        `json:"account_id" binding:"required"`
	ToAccountID  *uint64       `json:"to_account_id"`
	CategoryCode string        `json:"category_code"`
	ResourceID   *uint64       `json:"resource_id"`
	Counterparty string        `json:"counterparty"`
	Tags         []string      `json:"tags"`
	Title        string        `json:"title"`
	Note         string        `json:"note"`
	Ext          model.JSONMap `json:"ext"`
}

// Create POST /api/transactions
func (h *TransactionApi) Create(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	var req txCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	occur := time.Now()
	if req.OccurAt != nil {
		occur = *req.OccurAt
	}
	in := service.TxInput{
		UserID:       uint64(claims.ID),
		Type:         req.Type,
		Direction:    req.Direction,
		OccurAt:      occur,
		Amount:       req.Amount,
		AccountID:    req.AccountID,
		ToAccountID:  req.ToAccountID,
		CategoryCode: req.CategoryCode,
		ResourceID:   req.ResourceID,
		Counterparty: req.Counterparty,
		Tags:         req.Tags,
		Title:        req.Title,
		Note:         req.Note,
		Ext:          req.Ext,
	}
	t, err := ledgerService.Create(in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(t, c)
}

// List GET /api/transactions
func (h *TransactionApi) List(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	q := dao.TransactionQuery{UserID: uint64(claims.ID)}
	if s := c.Query("from"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			q.From = &t
		} else if t, err := time.Parse("2006-01-02", s); err == nil {
			q.From = &t
		}
	}
	if s := c.Query("to"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			q.To = &t
		} else if t, err := time.Parse("2006-01-02", s); err == nil {
			tt := t.AddDate(0, 0, 1)
			q.To = &tt
		}
	}
	if s := c.Query("types"); s != "" {
		q.Types = splitCSV(s)
	}
	if s := c.Query("account_id"); s != "" {
		if id, err := strconv.ParseUint(s, 10, 64); err == nil {
			q.AccountID = &id
		}
	}
	q.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "50"))
	q.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	rows, total, err := dao.ListTransactions(q)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"rows": rows, "total": total}, c)
}

// Get GET /api/transactions/:id
func (h *TransactionApi) Get(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	t, err := dao.GetTransaction(uint64(claims.ID), id)
	if err != nil {
		response.FailWithMessage("未找到交易: "+err.Error(), c)
		return
	}
	response.OkWithData(t, c)
}

// Delete DELETE /api/transactions/:id
// 注意：软删除不回退账户余额。如需回退，调用 /accounts/:id/rebuild-balance
func (h *TransactionApi) Delete(c *gin.Context) {
	claims := util.GetUserInfo(c)
	if claims == nil {
		response.FailWithMessage("未登录", c)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := dao.DeleteTransaction(uint64(claims.ID), id); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已删除；如需回退余额请调用 rebuild-balance 接口", c)
}

func splitCSV(s string) []string {
	var out []string
	cur := ""
	for _, ch := range s {
		if ch == ',' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
		} else {
			cur += string(ch)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
