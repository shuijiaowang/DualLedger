package service

import (
	"SService/dao"
	"SService/model"
	"fmt"
	"time"
)

type AccrualViewService struct{}

// ViewEntry 合并后的权责视图条目（动态虚拟 + 真实事件 + 即买即耗 tx）
type ViewEntry struct {
	Source        string    `json:"source"` // DYNAMIC_VIRTUAL / ACCRUAL_REAL / TX_IMMEDIATE
	AccrueAt      time.Time `json:"accrue_at"`
	Amount        string    `json:"amount"` // 有符号；消耗为正，冲减为负
	Direction     string    `json:"direction,omitempty"`
	CategoryCode  string    `json:"category_code,omitempty"`
	ResourceID    *uint64   `json:"resource_id,omitempty"`
	TransactionID *uint64   `json:"transaction_id,omitempty"`
	RealEntryID   *uint64   `json:"real_entry_id,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
	Note          string    `json:"note,omitempty"`
	Title         string    `json:"title,omitempty"`
}

// ViewQuery 查询参数
type ViewQuery struct {
	UserID          uint64
	From            time.Time // 含
	To              time.Time // 不含
	IncludeTx       bool      // 是否包含即买即耗的 EXPENSE/INCOME（默认 true）
	IncludeCashOnly bool      // 是否包含 TRANSFER/LOAN/DEPOSIT/REFUND（"全部视图"开关；默认 false）
}

// Query 生成权责视图条目（动态 + 真实 + 即买即耗，按 accrue_at 升序）
func (s *AccrualViewService) Query(q ViewQuery) ([]ViewEntry, error) {
	if q.From.After(q.To) || q.From.Equal(q.To) {
		return nil, fmt.Errorf("from 必须早于 to")
	}

	var out []ViewEntry

	// --- 1) 动态虚拟行：扫描 ACTIVE + ENDED 的 FIXED_PERIOD / DYNAMIC_BY_DAY 资源 ---
	resources, err := dao.ListResources(q.UserID, []string{model.ResStatusActive, model.ResStatusEnded})
	if err != nil {
		return nil, err
	}
	resourceDirectionMap, err := dao.ResourceTxDirectionMap(q.UserID)
	if err != nil {
		return nil, err
	}
	for i := range resources {
		r := &resources[i]
		rows := generateDynamicRows(r, q.From, q.To, resourceDirectionMap[uint64(r.ID)])
		out = append(out, rows...)
	}

	// --- 2) 真实事件行 ---
	entries, err := dao.ListAccrualEntries(dao.AccrualQuery{
		UserID: q.UserID,
		From:   &q.From,
		To:     &q.To,
		Limit:  10000,
	})
	if err != nil {
		return nil, err
	}
	for i := range entries {
		e := &entries[i]
		id := uint64(e.ID)
		var rid *uint64
		if e.ResourceID != nil {
			v := uint64(*e.ResourceID)
			rid = &v
		}
		out = append(out, ViewEntry{
			Source:       "ACCRUAL_REAL",
			AccrueAt:     e.AccrueAt,
			Amount:       model.Money(e.Amount).String(),
			Direction:    resourceDirectionMap[*e.ResourceID],
			CategoryCode: e.CategoryCode,
			ResourceID:   rid,
			RealEntryID:  &id,
			Tags:         []string(e.Tags),
			Note:         e.Note,
		})
	}

	// --- 3) 即买即耗的 transaction（视图层展示） ---
	if q.IncludeTx {
		types := []string{model.TxExpense, model.TxIncome, model.TxAdjust}
		if q.IncludeCashOnly {
			types = append(types, model.TxTransfer, model.TxLoan, model.TxDeposit, model.TxRefund)
		}
		txs, _, err := dao.ListTransactions(dao.TransactionQuery{
			UserID: q.UserID,
			From:   &q.From,
			To:     &q.To,
			Types:  types,
			Limit:  10000,
		})
		if err != nil {
			return nil, err
		}
		for i := range txs {
			t := &txs[i]
			// 关联 resource 的 transaction：权责已由动态行/真实事件承担，这里不重复展示
			if t.ResourceID != nil {
				continue
			}
			// TRANSFER/LOAN/DEPOSIT/REFUND：默认不进权责，仅在 IncludeCashOnly=true 时作为参考展示
			if model.IsCashflowAccountingOnly(t.Type) && !q.IncludeCashOnly {
				continue
			}
			signedAmount := model.Money(t.Amount).String()
			if t.Direction == model.DirOut {
				signedAmount = "-" + signedAmount
			}
			id := uint64(t.ID)
			out = append(out, ViewEntry{
				Source:        "TX_IMMEDIATE",
				AccrueAt:      t.OccurAt,
				Amount:        signedAmount,
				Direction:     t.Direction,
				CategoryCode:  t.CategoryCode,
				TransactionID: &id,
				Tags:          tagsFromExt(t.Ext),
				Title:         t.Title,
				Note:          t.Note,
			})
		}
	}

	// 按 accrue_at 升序（前端再自行分组）
	sortByAccrueAt(out)
	return out, nil
}

// generateDynamicRows 按规则生成动态虚拟行
// 约定：消耗/产出 amount 记为正数；INCOME 资源（工资包）也记正数，由前端按分类区分收支。
func generateDynamicRows(r *model.Resource, from, to time.Time, direction string) []ViewEntry {
	if r.AmortizeRule.Type != model.AmortizeFixedPeriod &&
		r.AmortizeRule.Type != model.AmortizeDynamicByDay {
		return nil
	}

	// 起点：优先 StartUseAt，再退回 PurchaseAt
	start := r.PurchaseAt
	if r.StartUseAt != nil {
		start = *r.StartUseAt
	}
	if r.AmortizeRule.Type == model.AmortizeFixedPeriod && r.AmortizeRule.Start != nil {
		if t, err := time.Parse("2006-01-02", *r.AmortizeRule.Start); err == nil {
			start = t
		}
	}
	start = dayStart(start)

	// 终点：资源规则的自然结束日 + 资源自身 EndAt 做 min
	var naturalEnd time.Time
	switch r.AmortizeRule.Type {
	case model.AmortizeFixedPeriod:
		if r.AmortizeRule.Days == nil {
			return nil
		}
		naturalEnd = start.AddDate(0, 0, *r.AmortizeRule.Days)
	case model.AmortizeDynamicByDay:
		if r.AmortizeRule.ExpectedDays != nil && *r.AmortizeRule.ExpectedDays > 0 {
			naturalEnd = start.AddDate(0, 0, *r.AmortizeRule.ExpectedDays)
		} else {
			// expected_days 为空时，默认摊到今天（含今天）
			naturalEnd = dayStart(time.Now()).AddDate(0, 0, 1)
		}
	}

	effectiveEnd := naturalEnd
	if r.EndAt != nil && r.EndAt.Before(effectiveEnd) {
		effectiveEnd = dayStart(*r.EndAt).AddDate(0, 0, 1) // 含结束当天
	}

	// 与查询区间求交
	loopFrom := start
	if loopFrom.Before(dayStart(from)) {
		loopFrom = dayStart(from)
	}
	loopTo := effectiveEnd
	if loopTo.After(dayStart(to)) {
		loopTo = dayStart(to)
	}
	if !loopFrom.Before(loopTo) {
		return nil
	}

	// 单日金额（MVP 简化：DYNAMIC_BY_DAY 与 FIXED_PERIOD 都按 total_cost/天数 均摊）
	var days int
	switch r.AmortizeRule.Type {
	case model.AmortizeFixedPeriod:
		days = *r.AmortizeRule.Days
	case model.AmortizeDynamicByDay:
		if r.AmortizeRule.ExpectedDays != nil && *r.AmortizeRule.ExpectedDays > 0 {
			days = *r.AmortizeRule.ExpectedDays
		} else {
			days = int(dayStart(time.Now()).Sub(start).Hours()/24) + 1
		}
	}
	if days <= 0 {
		days = 1
	}
	totalFloat := parseFloat(string(r.TotalCost))
	perDay := totalFloat / float64(days)
	perDayStr := fmt.Sprintf("%.2f", perDay)

	var rows []ViewEntry
	for d := loopFrom; d.Before(loopTo); d = d.AddDate(0, 0, 1) {
		rid := uint64(r.ID)
		rows = append(rows, ViewEntry{
			Source:       "DYNAMIC_VIRTUAL",
			AccrueAt:     d,
			Amount:       perDayStr,
			Direction:    direction,
			CategoryCode: r.CategoryCode,
			ResourceID:   &rid,
			Tags:         tagsFromExt(r.Ext),
			Note:         fmt.Sprintf("规则动态[%s] %s", r.AmortizeRule.Type, r.Name),
			Title:        r.Name,
		})
	}
	return rows
}

func dayStart(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func sortByAccrueAt(rows []ViewEntry) {
	// 简单插入排序（条目量不会太大，避免引入 sort.Slice 闭包开销；真要大改）
	for i := 1; i < len(rows); i++ {
		j := i
		for j > 0 && rows[j-1].AccrueAt.After(rows[j].AccrueAt) {
			rows[j-1], rows[j] = rows[j], rows[j-1]
			j--
		}
	}
}
