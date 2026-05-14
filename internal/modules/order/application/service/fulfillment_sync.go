package service

import (
	"fmt"
	"strings"
	"time"

	orderEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	fulfillmentEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/servicefulfillment/domain/entity"
	offerEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	"gorm.io/gorm"
)

func buildOrderCargoName(order *orderEntity.Order) string {
	if order == nil || len(order.Items) == 0 {
		return orderNumberFallback(order)
	}
	names := make([]string, 0, len(order.Items))
	for _, item := range order.Items {
		name := strings.TrimSpace(item.ProductName)
		if name == "" {
			name = strings.TrimSpace(item.Category)
		}
		if name == "" {
			name = strings.TrimSpace(item.ProductNo)
		}
		if name != "" {
			names = append(names, name)
		}
		if len(names) >= 3 {
			break
		}
	}
	if len(names) == 0 {
		return orderNumberFallback(order)
	}
	return strings.Join(names, " / ")
}

func syncSuffix() string {
	return fmt.Sprintf("%04d", time.Now().UnixNano()%10000)
}

func orderNumberFallback(order *orderEntity.Order) string {
	if order == nil {
		return "钢材订单"
	}
	if strings.TrimSpace(order.OrderNo) == "" {
		return "钢材订单"
	}
	return "订单" + order.OrderNo
}

func buildPickupOrigin(order *orderEntity.Order, bidding *orderEntity.LogisticsBidding) string {
	if bidding != nil && strings.TrimSpace(bidding.PickupAddress) != "" {
		return strings.TrimSpace(bidding.PickupAddress)
	}
	if order != nil && order.DeliveryInfo != nil {
		if strings.TrimSpace(order.DeliveryInfo.WarehouseAddress) != "" {
			return strings.TrimSpace(order.DeliveryInfo.WarehouseAddress)
		}
		if strings.TrimSpace(order.DeliveryInfo.WarehouseName) != "" {
			return strings.TrimSpace(order.DeliveryInfo.WarehouseName)
		}
	}
	return "待确认提货点"
}

func buildDestination(order *orderEntity.Order, bidding *orderEntity.LogisticsBidding) string {
	if bidding != nil && strings.TrimSpace(bidding.DestinationAddress) != "" {
		return strings.TrimSpace(bidding.DestinationAddress)
	}
	if order != nil && order.DeliveryInfo != nil {
		if strings.TrimSpace(order.DeliveryInfo.DestWarehouseAddress) != "" {
			return strings.TrimSpace(order.DeliveryInfo.DestWarehouseAddress)
		}
		if strings.TrimSpace(order.DeliveryInfo.Address) != "" {
			return strings.TrimSpace(order.DeliveryInfo.Address)
		}
	}
	return "待确认目的地"
}

func resolveLogisticsServiceID(tx *gorm.DB, logisticsEnterpriseID uint64) uint64 {
	if logisticsEnterpriseID == 0 {
		return 0
	}
	var service offerEntity.LogisticsService
	err := tx.Where("enterprise_id = ? AND status = ?", logisticsEnterpriseID, offerEntity.OfferStatusApproved).
		Order("id asc").
		First(&service).Error
	if err == nil {
		return service.ID
	}
	err = tx.Where("enterprise_id = ?", logisticsEnterpriseID).Order("id asc").First(&service).Error
	if err == nil {
		return service.ID
	}
	return 0
}

func shipmentTaskToLogisticsStatus(task *orderEntity.ShipmentTask, order *orderEntity.Order) fulfillmentEntity.LogisticsOrderStatus {
	if task != nil {
		switch task.Status {
		case orderEntity.ShipmentStatusPendingPickup:
			return fulfillmentEntity.LogisticsOrderTaskCreated
		case orderEntity.ShipmentStatusPickedUp:
			return fulfillmentEntity.LogisticsOrderInTransit
		case orderEntity.ShipmentStatusDelivered:
			return fulfillmentEntity.LogisticsOrderArrivalConfirmed
		case orderEntity.ShipmentStatusSigned:
			return fulfillmentEntity.LogisticsOrderReceiptSigned
		}
	}
	if order != nil {
		switch order.Status {
		case orderEntity.OrderStatusReceived, orderEntity.OrderStatusReconciled, orderEntity.OrderStatusArchived:
			return fulfillmentEntity.LogisticsOrderReceiptSigned
		case orderEntity.OrderStatusDelivered:
			return fulfillmentEntity.LogisticsOrderArrivalConfirmed
		case orderEntity.OrderStatusShipping, orderEntity.OrderStatusDelivering:
			return fulfillmentEntity.LogisticsOrderInTransit
		}
	}
	return fulfillmentEntity.LogisticsOrderTaskCreated
}

func ensureLogisticsOrderSyncTx(tx *gorm.DB, order *orderEntity.Order, bidding *orderEntity.LogisticsBidding, quote *orderEntity.LogisticsBidQuote, task *orderEntity.ShipmentTask, operatorID uint64) error {
	if order == nil {
		return nil
	}
	logisticsEnterpriseID := uint64(0)
	if quote != nil && quote.LogisticsEnterpriseID != 0 {
		logisticsEnterpriseID = quote.LogisticsEnterpriseID
	}
	if logisticsEnterpriseID == 0 && task != nil && task.LogisticsEnterpriseID != 0 {
		logisticsEnterpriseID = task.LogisticsEnterpriseID
	}
	if logisticsEnterpriseID == 0 {
		return nil
	}

	var item fulfillmentEntity.LogisticsOrder
	err := tx.Where("order_id = ?", order.ID).First(&item).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	isNew := err == gorm.ErrRecordNotFound
	if isNew {
		item = fulfillmentEntity.LogisticsOrder{
			LogisticsNo: "LGAUTO" + time.Now().Format("20060102150405") + syncSuffix(),
			OrderID:     order.ID,
			OrderNo:     order.OrderNo,
			CreatedBy:   operatorID,
		}
	}

	item.BuyerEnterpriseID = order.BuyerID
	item.LogisticsServiceID = resolveLogisticsServiceID(tx, logisticsEnterpriseID)
	item.CarrierEnterpriseID = logisticsEnterpriseID
	if order.DeliveryInfo != nil {
		item.WarehouseServiceID = order.DeliveryInfo.WarehouseServiceID
	}
	item.Origin = buildPickupOrigin(order, bidding)
	item.Destination = buildDestination(order, bidding)
	item.CargoName = buildOrderCargoName(order)
	item.Weight = order.TotalQuantity
	if quote != nil {
		item.VehicleType = strings.TrimSpace(quote.VehicleInfo)
		item.EstimatedPrice = quote.Amount
		item.FinalPrice = quote.Amount
		item.DriverName = strings.TrimSpace(quote.DriverInfo)
	}
	if order.DeliveryInfo != nil {
		if strings.TrimSpace(order.DeliveryInfo.PlateNumber) != "" {
			item.PlateNumber = strings.TrimSpace(order.DeliveryInfo.PlateNumber)
		}
		if strings.TrimSpace(order.DeliveryInfo.DriverName) != "" {
			item.DriverName = strings.TrimSpace(order.DeliveryInfo.DriverName)
		}
		if strings.TrimSpace(order.DeliveryInfo.DriverPhone) != "" {
			item.DriverPhone = strings.TrimSpace(order.DeliveryInfo.DriverPhone)
		}
	}
	item.Status = shipmentTaskToLogisticsStatus(task, order)
	item.CurrentResponsibleParty = fulfillmentEntity.ResponsibilityForLogisticsStatus(item.Status)
	item.UpdatedBy = operatorID

	now := time.Now()
	switch item.Status {
	case fulfillmentEntity.LogisticsOrderTaskCreated:
		if quote != nil && strings.TrimSpace(quote.EstimatedPickupAt) != "" && item.PickupAppointmentAt == nil {
			item.PickupAppointmentAt = &now
		}
	case fulfillmentEntity.LogisticsOrderInTransit:
		if item.LoadingConfirmedAt == nil {
			item.LoadingConfirmedAt = &now
		}
	case fulfillmentEntity.LogisticsOrderArrivalConfirmed:
		if item.ArrivalConfirmedAt == nil {
			item.ArrivalConfirmedAt = &now
		}
	case fulfillmentEntity.LogisticsOrderReceiptSigned:
		if item.ReceiptSignedAt == nil {
			if task != nil && task.SignedAt != nil {
				item.ReceiptSignedAt = task.SignedAt
			} else if order.ReceiptAt != nil {
				item.ReceiptSignedAt = order.ReceiptAt
			} else {
				item.ReceiptSignedAt = &now
			}
		}
	}

	if isNew {
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		return ensureLogisticsTrackingEventTx(tx, &item, operatorID, "task_created", item.Origin, "从订单主流程同步生成物流履约单")
	}

	if err := tx.Save(&item).Error; err != nil {
		return err
	}
	return nil
}

func ensureLogisticsTrackingEventTx(tx *gorm.DB, item *fulfillmentEntity.LogisticsOrder, operatorID uint64, eventType, location, description string) error {
	if item == nil || item.ID == 0 || strings.TrimSpace(eventType) == "" {
		return nil
	}
	var count int64
	if err := tx.Model(&fulfillmentEntity.LogisticsTrackingEvent{}).
		Where("logistics_order_id = ? AND event_type = ?", item.ID, eventType).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	event := fulfillmentEntity.LogisticsTrackingEvent{
		LogisticsOrderID: item.ID,
		EventType:        eventType,
		Location:         strings.TrimSpace(location),
		Description:      strings.TrimSpace(description),
		OccurredAt:       time.Now(),
		CreatedBy:        operatorID,
	}
	return tx.Create(&event).Error
}

func ensureWarehouseMovementTx(tx *gorm.DB, order *orderEntity.Order, movementType fulfillmentEntity.WarehouseMovementType, warehouseServiceID, warehouseEnterpriseID uint64, operatorID uint64, scheduledAt, remark string) error {
	if order == nil || warehouseServiceID == 0 || warehouseEnterpriseID == 0 {
		return nil
	}
	var item fulfillmentEntity.WarehouseInventoryMovement
	err := tx.Where("order_id = ? AND warehouse_service_id = ? AND movement_type = ?", order.ID, warehouseServiceID, movementType).
		First(&item).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	isNew := err == gorm.ErrRecordNotFound
	if isNew {
		item = fulfillmentEntity.WarehouseInventoryMovement{
			MovementNo:            fmt.Sprintf("WM%s%s", time.Now().Format("20060102150405"), syncSuffix()),
			WarehouseServiceID:    warehouseServiceID,
			WarehouseEnterpriseID: warehouseEnterpriseID,
			OrderID:               order.ID,
			OrderNo:               order.OrderNo,
			MovementType:          movementType,
			CreatedBy:             operatorID,
		}
	}
	item.RequestEnterpriseID = order.BuyerID
	if len(order.Items) > 0 {
		item.ProductID = order.Items[0].ProductID
		item.ProductName = buildOrderCargoName(order)
		item.BundleNo = strings.TrimSpace(order.Items[0].BundleNo)
	}
	item.Quantity = order.TotalQuantity
	item.Status = fulfillmentEntity.WarehouseMovementPending
	item.ScheduledAt = strings.TrimSpace(scheduledAt)
	if strings.TrimSpace(remark) != "" {
		item.Remark = strings.TrimSpace(remark)
	}
	item.UpdatedBy = operatorID
	if isNew {
		return tx.Create(&item).Error
	}
	return tx.Save(&item).Error
}
