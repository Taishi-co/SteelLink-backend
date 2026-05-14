package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:Steellink2026@tcp(117.72.123.14:3306)/steellink?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("=== 检查订单状态分布 ===")
	rows, err := db.Query("SELECT status, COUNT(*) as count FROM orders GROUP BY status ORDER BY status")
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	for rows.Next() {
		var status, count int
		if err := rows.Scan(&status, &count); err != nil {
			fmt.Printf("扫描失败: %v\n", err)
			continue
		}
		fmt.Printf("状态 %d: %d 个订单\n", status, count)
	}
	rows.Close()

	fmt.Println("\n=== 查看状态19,20,21的订单详情 ===")
	rows2, err := db.Query("SELECT id, order_no, status, created_at, updated_at FROM orders WHERE status IN (19, 20, 21) ORDER BY id DESC LIMIT 20")
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	for rows2.Next() {
		var id uint64
		var orderNo string
		var status int
		var createdAt, updatedAt string
		if err := rows2.Scan(&id, &orderNo, &status, &createdAt, &updatedAt); err != nil {
			fmt.Printf("扫描失败: %v\n", err)
			continue
		}
		fmt.Printf("订单: id=%d, order_no=%s, status=%d, created=%s, updated=%s\n", id, orderNo, status, createdAt, updatedAt)
	}
	rows2.Close()

	fmt.Println("\n=== 查看一些订单的事件 ===")
	rows3, err := db.Query("SELECT order_id, event_type, actor_type, created_at FROM order_events WHERE order_id IN (SELECT id FROM orders WHERE status IN (19, 20, 21)) ORDER BY id DESC LIMIT 30")
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	for rows3.Next() {
		var orderId uint64
		var eventType, actorType, createdAt string
		if err := rows3.Scan(&orderId, &eventType, &actorType, &createdAt); err != nil {
			fmt.Printf("扫描失败: %v\n", err)
			continue
		}
		fmt.Printf("订单%d: event=%s, actor=%s, time=%s\n", orderId, eventType, actorType, createdAt)
	}
	rows3.Close()
}
