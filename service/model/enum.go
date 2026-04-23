package model

// CategoryKind 分类大类
const (
	CategoryKindIncome   = "INCOME"
	CategoryKindExpense  = "EXPENSE"
	CategoryKindTransfer = "TRANSFER"
	CategoryKindOther    = "OTHER"
)

// TransactionType 交易类型（对齐设计文档 v2 §3.6）
const (
	TxIncome   = "INCOME"
	TxExpense  = "EXPENSE"
	TxTransfer = "TRANSFER"
	TxLoan     = "LOAN"
	TxDeposit  = "DEPOSIT"
	TxRefund   = "REFUND"
	TxAdjust   = "ADJUST"
)

// TransactionDirection 交易方向
const (
	DirIn   = "IN"
	DirOut  = "OUT"
	DirBoth = "BOTH"
)

// ResourceStatus 资源状态（v2 已去掉 PAUSED）
const (
	ResStatusActive    = "ACTIVE"
	ResStatusEnded     = "ENDED"
	ResStatusReturned  = "RETURNED"
	ResStatusDiscarded = "DISCARDED"
)

// AmortizeType 摊销规则类型（v2 三种）
const (
	AmortizeFixedPeriod  = "FIXED_PERIOD"
	AmortizeByCount      = "BY_COUNT"
	AmortizeDynamicByDay = "DYNAMIC_BY_DAY"
)

// AccrualSource 权责事件来源（v2 去掉 AUTO）
const (
	AccrualManual    = "MANUAL"
	AccrualEndSettle = "END_SETTLE"
	AccrualAdjust    = "ADJUST"
)

// IsValidTxType 校验 transaction.type
func IsValidTxType(t string) bool {
	switch t {
	case TxIncome, TxExpense, TxTransfer, TxLoan, TxDeposit, TxRefund, TxAdjust:
		return true
	}
	return false
}

// IsValidDirection 校验 direction
func IsValidDirection(d string) bool {
	switch d {
	case DirIn, DirOut, DirBoth:
		return true
	}
	return false
}

// IsValidResourceStatus 校验 resource.status
func IsValidResourceStatus(s string) bool {
	switch s {
	case ResStatusActive, ResStatusEnded, ResStatusReturned, ResStatusDiscarded:
		return true
	}
	return false
}

// IsValidAccrualSource 校验 accrual_entry.source
func IsValidAccrualSource(s string) bool {
	switch s {
	case AccrualManual, AccrualEndSettle, AccrualAdjust:
		return true
	}
	return false
}

// DefaultDirectionFor 根据 type 推导默认 direction（省前端事）
func DefaultDirectionFor(txType string) string {
	switch txType {
	case TxIncome, TxRefund:
		return DirIn
	case TxExpense:
		return DirOut
	case TxTransfer:
		return DirBoth
	default:
		return ""
	}
}

// IsCashflowAccountingOnly 是否只走现金流视图（权责视图默认隐藏）
// TRANSFER / LOAN / DEPOSIT / REFUND 纯资金运输，不参与价值消耗
func IsCashflowAccountingOnly(txType string) bool {
	switch txType {
	case TxTransfer, TxLoan, TxDeposit, TxRefund:
		return true
	}
	return false
}
