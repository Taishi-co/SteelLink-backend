//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := os.Getenv("STEELLINK_RESTORE_DSN")
	if dsn == "" {
		log.Fatal("STEELLINK_RESTORE_DSN is required")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	fmt.Println("=== 查询最新订单 ===")
	rows, err := db.Query("SELECT id, order_no, status FROM orders ORDER BY id DESC LIMIT 5")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	orders := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var orderNo string
		var status int
		if err := rows.Scan(&id, &orderNo, &status); err != nil {
			log.Printf("Scan failed: %v", err)
			continue
		}
		fmt.Printf("ID: %d, 订单号: %s, 状态: %d\n", id, orderNo, status)
		orders = append(orders, map[string]interface{}{"id": id, "status": status})
	}

	fmt.Println("\n=== 查询最新物流订单 ===")
	loRows, err := db.Query("SELECT id, order_id, status FROM logistics_orders ORDER BY id DESC LIMIT 5")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer loRows.Close()

	logisticsOrders := make([]map[string]interface{}, 0)
	for loRows.Next() {
		var id int64
		var orderID int64
		var status string
		if err := loRows.Scan(&id, &orderID, &status); err != nil {
			log.Printf("Scan failed: %v", err)
			continue
		}
		fmt.Printf("ID: %d, 订单ID: %d, 状态: %s\n", id, orderID, status)
		logisticsOrders = append(logisticsOrders, map[string]interface{}{"id": id, "order_id": orderID, "status": status})
	}

	// 找到需要恢复的订单
	targetOrderID := int64(0)
	for _, o := range orders {
		st := o["status"].(int)
		if st == 19 || st == 20 || st == 21 || st == 22 {
			targetOrderID = o["id"].(int64)
			break
		}
	}

	if targetOrderID == 0 {
		fmt.Println("\n没有找到需要恢复的订单")
		return
	}

	fmt.Printf("\n准备恢复订单 ID: %d\n", targetOrderID)

	// 更新订单状态
	_, err = db.Exec("UPDATE orders SET status = 18, receipt_at = NULL WHERE id = ?", targetOrderID)
	if err != nil {
		log.Fatalf("Update order failed: %v", err)
	}

	// 更新物流订单
	targetLogisticsOrderID := int64(0)
	for _, lo := range logisticsOrders {
		if lo["order_id"].(int64) == targetOrderID {
			targetLogisticsOrderID = lo["id"].(int64)
			break
		}
	}

	if targetLogisticsOrderID > 0 {
		_, err = db.Exec("UPDATE logistics_orders SET status = 'arrival_confirmed', receipt_signed_at = NULL WHERE id = ?", targetLogisticsOrderID)
		if err != nil {
			log.Printf("Update logistics order failed: %v", err)
		}
	}

	// 删除相关事件
	_, err = db.Exec("DELETE FROM order_events WHERE order_id = ? AND event_type = 'buyer_received'", targetOrderID)
	if err != nil {
		log.Printf("Delete events failed: %v", err)
	}

	fmt.Println("\n✅ 恢复完成！请刷新前端页面重新测试！")
}
