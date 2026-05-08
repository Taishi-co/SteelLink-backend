package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, _ := sql.Open("mysql", "root:Steellink2026@tcp(117.72.123.14:3306)/steellink?parseTime=true")
	
	fmt.Println("=== 仓储企业列表 ===")
	rows, _ := db.Query(`SELECT e.id, e.name, e.enterprise_code 
		FROM enterprises e 
		JOIN enterprise_trade_capability_permissions p ON e.id = p.enterprise_id 
		WHERE p.service = "warehouse" AND p.status = 2`)
	for rows.Next() {
		var id uint64
		var name, code string
		rows.Scan(&id, &name, &code)
		fmt.Printf("ID: %d, 名称: %s, 企业编码: %s\n", id, name, code)
	}
	rows.Close()
	
	fmt.Println("\n=== 关联仓储企业的用户 ===")
	rows2, _ := db.Query(`SELECT u.id, u.phone, e.name as enterprise_name 
		FROM users u 
		JOIN enterprises e ON u.enterprise_code = e.enterprise_code 
		WHERE e.id IN (SELECT enterprise_id FROM enterprise_trade_capability_permissions WHERE service = "warehouse" AND status = 2)`)
	for rows2.Next() {
		var id uint64
		var phone, enterpriseName string
		rows2.Scan(&id, &phone, &enterpriseName)
		fmt.Printf("用户ID: %d, 手机号: %s, 所属企业: %s\n", id, phone, enterpriseName)
	}
	rows2.Close()
}