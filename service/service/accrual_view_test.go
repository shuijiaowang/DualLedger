package service

import (
	"SService/model"
	"testing"
	"time"
)

// 场景 2：电动牙刷 150 元 · DYNAMIC_BY_DAY 30 天
// 4-10 开始使用，查询 4-10 ~ 4-24（含 14 天），每天 5 元 → 共 14 行
func TestGenerateDynamicRows_DynamicByDay(t *testing.T) {
	start := time.Date(2026, 4, 10, 0, 0, 0, 0, time.Local)
	exp := 30
	r := &model.Resource{
		Name:         "电动牙刷",
		TotalCost:    model.NewMoney("150.00"),
		AmortizeRule: model.AmortizeRule{Type: model.AmortizeDynamicByDay, ExpectedDays: &exp},
		StartUseAt:   &start,
		PurchaseAt:   start.AddDate(0, 0, -2),
		Status:       model.ResStatusActive,
	}
	r.ID = 500

	from := time.Date(2026, 4, 10, 0, 0, 0, 0, time.Local)
	to := time.Date(2026, 4, 24, 0, 0, 0, 0, time.Local)

	rows := generateDynamicRows(r, from, to)
	if len(rows) != 14 {
		t.Fatalf("want 14 rows, got %d", len(rows))
	}
	for _, row := range rows {
		if row.Amount != "5.00" {
			t.Fatalf("per-day want 5.00 got %s", row.Amount)
		}
		if row.Source != "DYNAMIC_VIRTUAL" {
			t.Fatalf("source wrong")
		}
	}
}

// 场景 7：工资 4500 · FIXED_PERIOD 30 天 · 查询区间内天数
func TestGenerateDynamicRows_FixedPeriod(t *testing.T) {
	start := time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local)
	days := 30
	r := &model.Resource{
		Name:         "2026-04 工资",
		TotalCost:    model.NewMoney("4500.00"),
		AmortizeRule: model.AmortizeRule{Type: model.AmortizeFixedPeriod, Days: &days},
		StartUseAt:   &start,
		PurchaseAt:   start,
		Status:       model.ResStatusActive,
	}
	r.ID = 550

	// 整月查询 2026-04 共 30 天生成
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local)
	to := time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)

	rows := generateDynamicRows(r, from, to)
	if len(rows) != 30 {
		t.Fatalf("want 30 rows, got %d", len(rows))
	}
	// 累计闭合：30 × 150 = 4500
	total := model.NewMoney("0")
	for _, row := range rows {
		total = total.Add(model.NewMoney(row.Amount))
	}
	if total.String() != "4500.00" {
		t.Fatalf("closure invariant broken; sum=%s", total.String())
	}
}

// BY_COUNT 不产生动态行（完全靠打卡）
func TestGenerateDynamicRows_ByCountNoRows(t *testing.T) {
	qty := 6.0
	r := &model.Resource{
		Name:         "一箱苹果",
		TotalCost:    model.NewMoney("18.00"),
		AmortizeRule: model.AmortizeRule{Type: model.AmortizeByCount, TotalQty: &qty},
		PurchaseAt:   time.Date(2026, 4, 20, 0, 0, 0, 0, time.Local),
		Status:       model.ResStatusActive,
	}
	from := time.Date(2026, 4, 20, 0, 0, 0, 0, time.Local)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.Local)
	rows := generateDynamicRows(r, from, to)
	if len(rows) != 0 {
		t.Fatalf("BY_COUNT should not generate dynamic rows, got %d", len(rows))
	}
}

// 结束资源后动态行到 end_at 截止（场景 12：退货）
func TestGenerateDynamicRows_RespectEndAt(t *testing.T) {
	start := time.Date(2026, 4, 15, 0, 0, 0, 0, time.Local)
	exp := 365
	end := time.Date(2026, 4, 16, 0, 0, 0, 0, time.Local) // 4-17 退货，end_at 写到 4-16
	r := &model.Resource{
		Name:         "机械键盘",
		TotalCost:    model.NewMoney("400.00"),
		AmortizeRule: model.AmortizeRule{Type: model.AmortizeDynamicByDay, ExpectedDays: &exp},
		StartUseAt:   &start,
		PurchaseAt:   start,
		EndAt:        &end,
		Status:       model.ResStatusReturned,
	}
	from := time.Date(2026, 4, 15, 0, 0, 0, 0, time.Local)
	to := time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)
	rows := generateDynamicRows(r, from, to)
	// end_at 含当天 → 4-15, 4-16 两天
	if len(rows) != 2 {
		t.Fatalf("should stop at end_at, got %d rows", len(rows))
	}
}

// 验证排序稳定性
func TestSortByAccrueAt(t *testing.T) {
	rows := []ViewEntry{
		{AccrueAt: time.Date(2026, 4, 10, 0, 0, 0, 0, time.Local), Note: "c"},
		{AccrueAt: time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local), Note: "a"},
		{AccrueAt: time.Date(2026, 4, 5, 0, 0, 0, 0, time.Local), Note: "b"},
	}
	sortByAccrueAt(rows)
	if rows[0].Note != "a" || rows[1].Note != "b" || rows[2].Note != "c" {
		t.Fatalf("sort wrong: %+v", rows)
	}
}
