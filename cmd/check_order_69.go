//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := os.Getenv("STEELLINK_CHECK_DSN")
	if dsn == "" {
		fmt.Println("STEELLINK_CHECK_DSN is required")
		return
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("=== 订单 69 的详细信息 ===")
	var id uint64
	var orderNo string
	var status int
	var createdAt, updatedAt string
	err = db.QueryRow("SELECT id, order_no, status, created_at, updated_at FROM orders WHERE id = 69").Scan(&id, &orderNo, &status, &createdAt, &updatedAt)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	fmt.Printf("订单: id=%d, order_no=%s, status=%d, created=%s, updated=%s\n", id, orderNo, status, createdAt, updatedAt)

	fmt.Println("\n=== 订单 69 的完整事件流 ===")
	rows, err := db.Query("SELECT event_type, actor_service, from_status, to_status, created_at FROM order_events WHERE order_id = 69 ORDER BY id ASC")
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	for rows.Next() {
		var eventType, actorService string
		var fromStatus, toStatus int
		var createdAt string
		if err := rows.Scan(&eventType, &actorService, &fromStatus, &toStatus, &createdAt); err != nil {
			fmt.Printf("扫描失败: %v\n", err)
			continue
		}
		fmt.Printf("%s: %s (actor=%s, %d -> %d)\n", createdAt, eventType, actorService, fromStatus, toStatus)
	}
	rows.Close()
}
