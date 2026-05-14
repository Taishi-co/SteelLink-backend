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

	fmt.Println("=== 查看订单事件表结构 ===")
	rows, err := db.Query("DESCRIBE order_events")
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	for rows.Next() {
		var field, typ, null, key string
		var extra interface{}
		if err := rows.Scan(&field, &typ, &null, &key, &extra); err != nil {
			continue
		}
		fmt.Printf("%s: %s\n", field, typ)
	}
	rows.Close()

	fmt.Println("\n=== 查看具体订单 69 的事件流程 ===")
	rows2, err := db.Query("SELECT id, event_type, from_status, to_status, created_at FROM order_events WHERE order_id = 69 ORDER BY id")
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	for rows2.Next() {
		var id uint64
		var eventType string
		var fromStatus, toStatus sql.NullInt64
		var createdAt string
		if err := rows2.Scan(&id, &eventType, &fromStatus, &toStatus, &createdAt); err != nil {
			fmt.Printf("扫描失败: %v\n", err)
			continue
		}
		fmt.Printf("id=%d: event=%s, from=%v, to=%v, time=%s\n", id, eventType, fromStatus, toStatus, createdAt)
	}
	rows2.Close()
}
