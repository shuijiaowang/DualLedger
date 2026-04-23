package model

import "testing"

func TestMoneyArithmetic(t *testing.T) {
	cases := []struct {
		a, b    string
		addWant string
		subWant string
	}{
		{"100.00", "50.00", "150.00", "50.00"},
		{"0.10", "0.20", "0.30", "-0.10"},
		{"150.00", "150.00", "300.00", "0.00"},
		{"1.99", "0.02", "2.01", "1.97"},
	}
	for _, c := range cases {
		a := NewMoney(c.a)
		b := NewMoney(c.b)
		if got := a.Add(b).String(); got != c.addWant {
			t.Errorf("%s + %s = %s; want %s", c.a, c.b, got, c.addWant)
		}
		if got := a.Sub(b).String(); got != c.subWant {
			t.Errorf("%s - %s = %s; want %s", c.a, c.b, got, c.subWant)
		}
	}
}

func TestMoneyNegateAndCmp(t *testing.T) {
	m := NewMoney("42.50")
	if m.Negate().String() != "-42.50" {
		t.Fatalf("Negate wrong: %s", m.Negate().String())
	}
	if m.Cmp(NewMoney("42.50")) != 0 {
		t.Fatalf("equal Cmp should be 0")
	}
	if m.Cmp(NewMoney("100")) >= 0 {
		t.Fatalf("42.50 should be < 100")
	}
}

func TestMoneyNormalize(t *testing.T) {
	m := NewMoney("10")
	if m.String() != "10.00" {
		t.Fatalf("string normalize: %s", m.String())
	}
	if NewMoney(nil).String() != "0.00" {
		t.Fatal("nil should be 0.00")
	}
	if NewMoney("abc").String() != "0.00" {
		t.Fatal("invalid string should fallback to 0.00")
	}
}

func TestAmortizeRuleValidate(t *testing.T) {
	days := 30
	ok := AmortizeRule{Type: AmortizeFixedPeriod, Days: &days}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid FIXED_PERIOD rejected: %v", err)
	}
	bad := AmortizeRule{Type: AmortizeFixedPeriod}
	if err := bad.Validate(); err == nil {
		t.Fatal("FIXED_PERIOD without days should fail")
	}
	qty := 6.0
	ok = AmortizeRule{Type: AmortizeByCount, TotalQty: &qty}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid BY_COUNT rejected: %v", err)
	}
	exp := 30
	ok = AmortizeRule{Type: AmortizeDynamicByDay, ExpectedDays: &exp}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid DYNAMIC rejected: %v", err)
	}
	if err := (AmortizeRule{Type: "UNKNOWN"}).Validate(); err == nil {
		t.Fatal("unknown type should fail")
	}
}
