//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
)

func main() {
	if os.Getenv("STEELLINK_ALLOW_RESTORE_SCRIPT") != "1" {
		fmt.Println("Refusing to run. Set STEELLINK_ALLOW_RESTORE_SCRIPT=1 to execute this manual restore script.")
		return
	}

	database.InitDB()
	db := database.GetDB()

	// 查询最新订单ID
	var orderIDs []struct {
		ID     uint64 `db:"id"`
		Status int    `db:"status"`
	}
	db.Raw("SELECT id, status FROM orders ORDER BY id DESC LIMIT 5").Scan(&orderIDs)
	fmt.Println("订单列表：")
	for _, o := range orderIDs {
		fmt.Printf("ID: %d, Status: %d\n", o.ID, o.Status)
	}

	// 查询最新物流订单
	var logisticsOrderIDs []struct {
		ID      uint64 `db:"id"`
		OrderID uint64 `db:"order_id"`
		Status  string `db:"status"`
	}
	db.Raw("SELECT id, order_id, status FROM logistics_orders ORDER BY id DESC LIMIT 5").Scan(&logisticsOrderIDs)
	fmt.Println("\n物流订单列表：")
	for _, lo := range logisticsOrderIDs {
		fmt.Printf("ID: %d, OrderID: %d, Status: %s\n", lo.ID, lo.OrderID, lo.Status)
	}

	// 恢复订单（假设恢复 ID 最大的订单）
	targetOrderID := uint64(0)
	for _, o := range orderIDs {
		if o.Status == 19 || o.Status == 20 || o.Status == 21 || o.Status == 22 {
			targetOrderID = o.ID
			break
		}
	}

	if targetOrderID == 0 {
		fmt.Println("\n没有找到需要恢复的订单")
		return
	}

	fmt.Printf("\n恢复订单 ID: %d\n", targetOrderID)

	// 更新订单状态
	db.Exec("UPDATE orders SET status = 18, receipt_at = NULL WHERE id = ?", targetOrderID)

	// 找到对应的物流订单
	var logisticsOrderID uint64 = 0
	for _, lo := range logisticsOrderIDs {
		if lo.OrderID == targetOrderID {
			logisticsOrderID = lo.ID
			break
		}
	}

	if logisticsOrderID > 0 {
		db.Exec("UPDATE logistics_orders SET status = 'arrival_confirmed', receipt_signed_at = NULL WHERE id = ?", logisticsOrderID)
	}

	// 删除相关事件
	db.Exec("DELETE FROM order_events WHERE order_id = ? AND event_type = 'buyer_received'", targetOrderID)

	fmt.Println("恢复完成！请刷新页面。")
}
