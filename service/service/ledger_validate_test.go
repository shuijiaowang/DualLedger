package service

import (
	"SService/model"
	"testing"
	"time"
)

func TestValidateTxInput(t *testing.T) {
	base := func() TxInput {
		return TxInput{
			UserID:    1,
			Type:      model.TxExpense,
			Amount:    model.NewMoney("15.00"),
			AccountID: 1,
			OccurAt:   time.Now(),
		}
	}

	// 默认方向推导
	in := base()
	if err := validateTxInput(&in); err != nil {
		t.Fatalf("should pass: %v", err)
	}
	if in.Direction != model.DirOut {
		t.Fatalf("EXPENSE default direction should be OUT, got %s", in.Direction)
	}

	// amount <= 0 违反不变式 §15.1
	zero := base()
	zero.Amount = model.NewMoney("0")
	if err := validateTxInput(&zero); err == nil {
		t.Fatal("amount=0 should fail")
	}
	neg := base()
	neg.Amount = model.NewMoney("-1")
	if err := validateTxInput(&neg); err == nil {
		t.Fatal("amount<0 should fail")
	}

	// TRANSFER 必须有 to_account_id 且不同于 account_id
	tr := base()
	tr.Type = model.TxTransfer
	tr.Direction = model.DirBoth
	if err := validateTxInput(&tr); err == nil {
		t.Fatal("TRANSFER without to_account_id should fail")
	}
	same := uint64(1)
	tr.ToAccountID = &same
	if err := validateTxInput(&tr); err == nil {
		t.Fatal("TRANSFER same account should fail")
	}
	other := uint64(2)
	tr.ToAccountID = &other
	if err := validateTxInput(&tr); err != nil {
		t.Fatalf("valid TRANSFER should pass: %v", err)
	}

	// 非 TRANSFER 不允许 to_account_id
	ne := base()
	ne.ToAccountID = &other
	if err := validateTxInput(&ne); err == nil {
		t.Fatal("non-TRANSFER with to_account_id should fail")
	}

	// 非法 category_code
	bc := base()
	bc.CategoryCode = "not.exists"
	if err := validateTxInput(&bc); err == nil {
		t.Fatal("unknown category should fail")
	}

	// 非法 type
	bt := base()
	bt.Type = "BOGUS"
	if err := validateTxInput(&bt); err == nil {
		t.Fatal("bogus type should fail")
	}
}
