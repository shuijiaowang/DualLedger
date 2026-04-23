package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// Money 金额字符串（对应 DB DECIMAL(14,2)）
// - 存储层始终保留两位小数字符串，避免 float 精度损失
// - 运算走 math/big.Float，完全标准库依赖
type Money string

const moneyPrec = 64 // big.Float 精度（远超 14 位十进制）

// NewMoney 从字符串/数字构造 Money，失败返回 zero
func NewMoney(v any) Money {
	switch x := v.(type) {
	case string:
		return normalizeMoney(x)
	case nil:
		return "0.00"
	case float64:
		return normalizeMoney(fmt.Sprintf("%.2f", x))
	case int:
		return normalizeMoney(fmt.Sprintf("%d", x))
	case int64:
		return normalizeMoney(fmt.Sprintf("%d", x))
	default:
		return "0.00"
	}
}

// Big 转成 big.Float
func (m Money) Big() *big.Float {
	if m == "" {
		return new(big.Float).SetPrec(moneyPrec).SetFloat64(0)
	}
	f, _, err := big.ParseFloat(string(m), 10, moneyPrec, big.ToNearestEven)
	if err != nil {
		return new(big.Float).SetPrec(moneyPrec).SetFloat64(0)
	}
	return f
}

// String 统一两位小数
func (m Money) String() string {
	if m == "" {
		return "0.00"
	}
	return string(normalizeMoney(string(m)))
}

// IsZero 是否为 0
func (m Money) IsZero() bool {
	return m.Big().Cmp(big.NewFloat(0)) == 0
}

// Cmp 小于返回 -1，等于 0，大于 1
func (m Money) Cmp(other Money) int {
	return m.Big().Cmp(other.Big())
}

// Add 返回 m + other
func (m Money) Add(other Money) Money {
	r := new(big.Float).SetPrec(moneyPrec).Add(m.Big(), other.Big())
	return moneyFromBig(r)
}

// Sub 返回 m - other
func (m Money) Sub(other Money) Money {
	r := new(big.Float).SetPrec(moneyPrec).Sub(m.Big(), other.Big())
	return moneyFromBig(r)
}

// MulInt 返回 m * n
func (m Money) MulInt(n int64) Money {
	r := new(big.Float).SetPrec(moneyPrec).Mul(m.Big(), new(big.Float).SetInt64(n))
	return moneyFromBig(r)
}

// Negate 返回 -m
func (m Money) Negate() Money {
	r := new(big.Float).SetPrec(moneyPrec).Neg(m.Big())
	return moneyFromBig(r)
}

// Scan 实现 sql.Scanner（从 DB 读出）
func (m *Money) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*m = "0.00"
	case []byte:
		*m = normalizeMoney(string(v))
	case string:
		*m = normalizeMoney(v)
	case float64:
		*m = normalizeMoney(fmt.Sprintf("%.2f", v))
	default:
		return fmt.Errorf("Money.Scan: unsupported type %T", src)
	}
	return nil
}

// Value 实现 driver.Valuer
func (m Money) Value() (driver.Value, error) {
	return m.String(), nil
}

// MarshalJSON 输出为字符串，保持精度
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// UnmarshalJSON 支持字符串或数字
func (m *Money) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "" || s == "null" {
		*m = "0.00"
		return nil
	}
	if s[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		*m = normalizeMoney(raw)
		return nil
	}
	*m = normalizeMoney(s)
	return nil
}

func normalizeMoney(s string) Money {
	s = strings.TrimSpace(s)
	if s == "" {
		return "0.00"
	}
	f, _, err := big.ParseFloat(s, 10, moneyPrec, big.ToNearestEven)
	if err != nil {
		return "0.00"
	}
	return moneyFromBig(f)
}

func moneyFromBig(f *big.Float) Money {
	return Money(f.Text('f', 2))
}

// ---------------- JSON 辅助类型 ----------------

// JSONStrings 字符串数组，对应 MySQL JSON 列
type JSONStrings []string

// Scan 实现 sql.Scanner
func (s *JSONStrings) Scan(src any) error {
	if src == nil {
		*s = nil
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("JSONStrings.Scan: unsupported type %T", src)
	}
	if len(b) == 0 {
		*s = nil
		return nil
	}
	return json.Unmarshal(b, s)
}

// Value 实现 driver.Valuer
func (s JSONStrings) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// JSONMap 通用 JSON 对象
type JSONMap map[string]any

func (m *JSONMap) Scan(src any) error {
	if src == nil {
		*m = nil
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("JSONMap.Scan: unsupported type %T", src)
	}
	if len(b) == 0 {
		*m = nil
		return nil
	}
	return json.Unmarshal(b, m)
}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// ---------------- AmortizeRule ----------------

// AmortizeRule resource.amortize_rule JSON 字段
// type 必填；其他字段根据 type 选填。
type AmortizeRule struct {
	Type            string   `json:"type"` // FIXED_PERIOD / BY_COUNT / DYNAMIC_BY_DAY
	Days            *int     `json:"days,omitempty"`
	Start           *string  `json:"start,omitempty"`             // YYYY-MM-DD，缺省取 resource.start_use_at
	TotalQty        *float64 `json:"total_qty,omitempty"`         // BY_COUNT
	ExpectedDays    *int     `json:"expected_days,omitempty"`     // DYNAMIC_BY_DAY，可空（空=默认到今天）
	IncludeStartGap *bool    `json:"include_start_gap,omitempty"` // FIXED_PERIOD
}

func (r *AmortizeRule) Scan(src any) error {
	if src == nil {
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("AmortizeRule.Scan: unsupported type %T", src)
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, r)
}

func (r AmortizeRule) Value() (driver.Value, error) {
	if r.Type == "" {
		return "{}", nil
	}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Validate 校验规则自洽性
func (r AmortizeRule) Validate() error {
	switch r.Type {
	case AmortizeFixedPeriod:
		if r.Days == nil || *r.Days <= 0 {
			return errors.New("FIXED_PERIOD 规则缺少 days（>0）")
		}
	case AmortizeByCount:
		if r.TotalQty == nil || *r.TotalQty <= 0 {
			return errors.New("BY_COUNT 规则缺少 total_qty（>0）")
		}
	case AmortizeDynamicByDay:
		if r.ExpectedDays != nil && *r.ExpectedDays <= 0 {
			return errors.New("DYNAMIC_BY_DAY 规则中 expected_days 必须 > 0")
		}
	case "":
		return errors.New("amortize_rule.type 不能为空")
	default:
		return fmt.Errorf("未知 amortize_rule.type: %s", r.Type)
	}
	return nil
}
