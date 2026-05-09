package entity

import (
	"testing"
)

func TestValidateStatusTransition_DirectPaymentHappyPath(t *testing.T) {
	path := []struct {
		from OrderStatus
		to   OrderStatus
	}{
		{OrderStatusDraft, OrderStatusPendingWarehouseConfirm},
		{OrderStatusPendingWarehouseConfirm, OrderStatusPendingSellerConfirm},
		{OrderStatusPendingSellerConfirm, OrderStatusPendingBuyerConfirm},
		{OrderStatusPendingBuyerConfirm, OrderStatusPendingContract},
		{OrderStatusPendingContract, OrderStatusPendingPayment},
		{OrderStatusPendingPayment, OrderStatusPaid},
		{OrderStatusPaid, OrderStatusPendingDelivery},
		{OrderStatusPendingDelivery, OrderStatusDelivering},
		{OrderStatusDelivering, OrderStatusDelivered},
		{OrderStatusDelivered, OrderStatusReceived},
		{OrderStatusReceived, OrderStatusReconciled},
		{OrderStatusReconciled, OrderStatusArchived},
	}
	for _, step := range path {
		if !ValidateStatusTransition(step.from, step.to) {
			t.Errorf("expected transition %d -> %d to be valid", step.from, step.to)
		}
	}
}

func TestValidateStatusTransition_FinancingHappyPath(t *testing.T) {
	path := []struct {
		from OrderStatus
		to   OrderStatus
	}{
		{OrderStatusDraft, OrderStatusPendingWarehouseConfirm},
		{OrderStatusPendingWarehouseConfirm, OrderStatusPendingSellerConfirm},
		{OrderStatusPendingSellerConfirm, OrderStatusPendingBuyerConfirm},
		{OrderStatusPendingBuyerConfirm, OrderStatusPendingContract},
		{OrderStatusPendingContract, OrderStatusPendingPayment},
		{OrderStatusPendingPayment, OrderStatusFinancingReview},
		{OrderStatusFinancingReview, OrderStatusPendingFinancingContract},
		{OrderStatusPendingFinancingContract, OrderStatusPaid},
		{OrderStatusPaid, OrderStatusLogisticsBidding},
		{OrderStatusLogisticsBidding, OrderStatusPendingLogisticsConfirm},
		{OrderStatusPendingLogisticsConfirm, OrderStatusShipping},
		{OrderStatusShipping, OrderStatusDelivered},
		{OrderStatusDelivered, OrderStatusReceived},
		{OrderStatusReceived, OrderStatusReconciled},
		{OrderStatusReconciled, OrderStatusArchived},
	}
	for _, step := range path {
		if !ValidateStatusTransition(step.from, step.to) {
			t.Errorf("expected transition %d -> %d to be valid", step.from, step.to)
		}
	}
}

func TestValidateStatusTransition_CrossStatusBlocked(t *testing.T) {
	blocked := []struct {
		from OrderStatus
		to   OrderStatus
	}{
		{OrderStatusDraft, OrderStatusPaid},
		{OrderStatusDraft, OrderStatusShipping},
		{OrderStatusPendingContract, OrderStatusPaid},
		{OrderStatusPendingPayment, OrderStatusShipping},
		{OrderStatusPaid, OrderStatusReceived},
		{OrderStatusPendingDelivery, OrderStatusReceived},
		{OrderStatusShipping, OrderStatusArchived},
		{OrderStatusDelivered, OrderStatusReconciled},
		{OrderStatusReceived, OrderStatusArchived},
		{OrderStatusPendingWarehouseConfirm, OrderStatusPendingBuyerConfirm},
		{OrderStatusPendingSellerConfirm, OrderStatusPendingContract},
	}
	for _, step := range blocked {
		if ValidateStatusTransition(step.from, step.to) {
			t.Errorf("expected cross-status transition %d -> %d to be blocked", step.from, step.to)
		}
	}
}

func TestValidateStatusTransition_CanCancelFromAnyStatus(t *testing.T) {
	statuses := []OrderStatus{
		OrderStatusDraft,
		OrderStatusPendingConfirm,
		OrderStatusPendingContract,
		OrderStatusPendingPayment,
		OrderStatusPaid,
		OrderStatusPendingDelivery,
		OrderStatusDelivering,
		OrderStatusDelivered,
		OrderStatusPendingWarehouseConfirm,
		OrderStatusPendingSellerConfirm,
		OrderStatusPendingBuyerConfirm,
		OrderStatusLogisticsBidding,
		OrderStatusPendingLogisticsConfirm,
		OrderStatusShipping,
		OrderStatusReceived,
		OrderStatusReconciled,
		OrderStatusFinancingReview,
		OrderStatusPendingFinancingContract,
		OrderStatusPendingWarehouseReselect,
		OrderStatusPendingLogisticsRebid,
	}
	for _, from := range statuses {
		if !ValidateStatusTransition(from, OrderStatusCancelled) {
			t.Errorf("expected any status %d can transition to Cancelled", from)
		}
	}
}

func TestValidateStatusTransition_TerminalStatesNoExit(t *testing.T) {
	terminals := []OrderStatus{OrderStatusCancelled, OrderStatusArchived}
	targets := []OrderStatus{
		OrderStatusDraft, OrderStatusPendingConfirm, OrderStatusPaid,
		OrderStatusShipping, OrderStatusDelivered, OrderStatusReceived,
	}
	for _, terminal := range terminals {
		for _, target := range targets {
			if ValidateStatusTransition(terminal, target) {
				t.Errorf("terminal status %d should not transition to %d", terminal, target)
			}
		}
	}
}

func TestValidateStatusTransition_WarehouseRejectAndReselect(t *testing.T) {
	if !ValidateStatusTransition(OrderStatusPendingWarehouseConfirm, OrderStatusPendingWarehouseReselect) {
		t.Error("warehouse reject should be allowed")
	}
	if !ValidateStatusTransition(OrderStatusPendingWarehouseReselect, OrderStatusPendingWarehouseConfirm) {
		t.Error("reselect warehouse should go back to warehouse confirm")
	}
}

func TestValidateStatusTransition_LogisticsRebid(t *testing.T) {
	if !ValidateStatusTransition(OrderStatusLogisticsBidding, OrderStatusPendingLogisticsRebid) {
		t.Error("logistics no capacity should be allowed")
	}
	if !ValidateStatusTransition(OrderStatusPendingLogisticsRebid, OrderStatusLogisticsBidding) {
		t.Error("rebid logistics should go back to bidding")
	}
	if !ValidateStatusTransition(OrderStatusPaid, OrderStatusPendingLogisticsRebid) {
		t.Error("paid status should allow logistics rebid")
	}
}

func TestValidateStatusTransition_FrozenAndException(t *testing.T) {
	if !ValidateStatusTransition(OrderStatusFrozen, OrderStatusCancelled) {
		t.Error("frozen should allow cancel")
	}
	if !ValidateStatusTransition(OrderStatusException, OrderStatusCancelled) {
		t.Error("exception should allow cancel")
	}
	statuses := []OrderStatus{OrderStatusDraft, OrderStatusPendingPayment, OrderStatusPaid}
	for _, from := range statuses {
		if !ValidateStatusTransition(from, OrderStatusFrozen) {
			t.Errorf("any active status %d should allow freeze", from)
		}
		if !ValidateStatusTransition(from, OrderStatusException) {
			t.Errorf("any active status %d should allow exception", from)
		}
	}
}
