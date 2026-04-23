package model

import "testing"

func TestDefaultDirectionFor(t *testing.T) {
	cases := map[string]string{
		TxIncome:   DirIn,
		TxRefund:   DirIn,
		TxExpense:  DirOut,
		TxTransfer: DirBoth,
		TxLoan:     "",
		TxAdjust:   "",
	}
	for in, want := range cases {
		if got := DefaultDirectionFor(in); got != want {
			t.Errorf("DefaultDirectionFor(%s) = %s; want %s", in, got, want)
		}
	}
}

func TestValidators(t *testing.T) {
	if !IsValidTxType(TxIncome) || IsValidTxType("X") {
		t.Fatal("IsValidTxType wrong")
	}
	if !IsValidDirection(DirBoth) || IsValidDirection("X") {
		t.Fatal("IsValidDirection wrong")
	}
	if !IsValidResourceStatus(ResStatusActive) || IsValidResourceStatus("PAUSED") {
		t.Fatal("PAUSED should be invalid in v2")
	}
	if !IsValidAccrualSource(AccrualManual) || IsValidAccrualSource("AUTO") {
		t.Fatal("AUTO should be invalid in v2")
	}
}

func TestIsCashflowAccountingOnly(t *testing.T) {
	want := map[string]bool{
		TxTransfer: true,
		TxLoan:     true,
		TxDeposit:  true,
		TxRefund:   true,
		TxIncome:   false,
		TxExpense:  false,
		TxAdjust:   false,
	}
	for k, v := range want {
		if IsCashflowAccountingOnly(k) != v {
			t.Errorf("IsCashflowAccountingOnly(%s) wrong", k)
		}
	}
}

func TestCategoryPresetWellFormed(t *testing.T) {
	if len(PresetCategories) == 0 {
		t.Fatal("preset categories empty")
	}
	codes := map[string]bool{}
	for _, c := range PresetCategories {
		if codes[c.Code] {
			t.Fatalf("duplicate category code: %s", c.Code)
		}
		codes[c.Code] = true
		if c.Name == "" || c.Code == "" {
			t.Fatalf("category missing name/code: %+v", c)
		}
		if c.Kind != CategoryKindIncome && c.Kind != CategoryKindExpense &&
			c.Kind != CategoryKindTransfer && c.Kind != CategoryKindOther {
			t.Fatalf("category %s has invalid kind %s", c.Code, c.Kind)
		}
	}
	// parent_code 必须指向已存在的分类
	for _, c := range PresetCategories {
		if c.ParentCode != "" && !codes[c.ParentCode] {
			t.Fatalf("category %s references unknown parent %s", c.Code, c.ParentCode)
		}
	}
}
