package entity

import "testing"

func TestProductReturnStock(t *testing.T) {
	p := &Product{
		Quantity:     100,
		AvailableQty: 70,
		LockedQty:    30,
	}
	if !p.DeductStock(30) {
		t.Fatal("deduct stock should succeed")
	}
	if p.Quantity != 70 || p.LockedQty != 0 {
		t.Fatalf("after deduct: Quantity=%f, LockedQty=%f, want 70,0", p.Quantity, p.LockedQty)
	}
	p.ReturnStock(30)
	if p.Quantity != 100 || p.AvailableQty != 100 {
		t.Fatalf("after return: Quantity=%f, AvailableQty=%f, want 100,100", p.Quantity, p.AvailableQty)
	}
}

func TestProductDeductStock_OnlyFromLocked(t *testing.T) {
	p := &Product{
		Quantity:     100,
		AvailableQty: 70,
		LockedQty:    30,
	}
	if p.DeductStock(31) {
		t.Error("should not deduct more than locked")
	}
	if !p.DeductStock(30) {
		t.Error("should deduct exact locked amount")
	}
	if p.Quantity != 70 || p.LockedQty != 0 || p.AvailableQty != 70 {
		t.Errorf("after deduct: Quantity=%f LockedQty=%f AvailableQty=%f", p.Quantity, p.LockedQty, p.AvailableQty)
	}
}

func TestProductLockUnlockCycle(t *testing.T) {
	p := &Product{
		Quantity:     100,
		AvailableQty: 100,
		LockedQty:    0,
	}
	if !p.LockStock(50) {
		t.Fatal("lock should succeed")
	}
	if p.AvailableQty != 50 || p.LockedQty != 50 {
		t.Fatalf("after lock: AvailableQty=%f, LockedQty=%f", p.AvailableQty, p.LockedQty)
	}
	if !p.DeductStock(50) {
		t.Fatal("deduct should succeed")
	}
	if p.Quantity != 50 || p.LockedQty != 0 {
		t.Fatalf("after deduct: Quantity=%f, LockedQty=%f", p.Quantity, p.LockedQty)
	}
	p.ReturnStock(50)
	if p.Quantity != 100 || p.AvailableQty != 100 {
		t.Fatalf("after return: Quantity=%f, AvailableQty=%f", p.Quantity, p.AvailableQty)
	}
}
