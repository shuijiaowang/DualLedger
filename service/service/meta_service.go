package service

import (
	"SService/dao"
	"SService/model"
)

type MetaService struct{}

// Categories 返回分类静态常量（MVP 阶段用户自定义暂不开放）
func (s *MetaService) Categories() []model.Category {
	rows, err := dao.ListCategories()
	if err != nil {
		return model.PresetCategories
	}
	out := make([]model.Category, 0, len(rows))
	for _, r := range rows {
		out = append(out, model.Category{
			Code:       r.Code,
			ParentCode: r.ParentCode,
			Name:       r.Name,
			Kind:       r.Kind,
			Icon:       r.Icon,
			Sort:       r.Sort,
			Source:     r.Source,
		})
	}
	return out
}

// Tags 返回标签建议词
func (s *MetaService) Tags() []string {
	return model.PresetTags
}

// Enums 返回前后端共享的枚举
func (s *MetaService) Enums() map[string]any {
	return map[string]any{
		"transaction_type": []string{
			model.TxIncome, model.TxExpense, model.TxTransfer,
			model.TxLoan, model.TxDeposit, model.TxRefund, model.TxAdjust,
		},
		"direction":       []string{model.DirIn, model.DirOut, model.DirBoth},
		"resource_status": []string{model.ResStatusActive, model.ResStatusEnded, model.ResStatusReturned, model.ResStatusDiscarded},
		"amortize_type":   []string{model.AmortizeFixedPeriod, model.AmortizeByCount, model.AmortizeDynamicByDay},
		"accrual_source":  []string{model.AccrualManual, model.AccrualEndSettle, model.AccrualAdjust},
	}
}
