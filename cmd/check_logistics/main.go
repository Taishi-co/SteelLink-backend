package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"math/rand"
	"time"
)

func main() {
	dsn := "root:Steellink2026@tcp(117.72.123.14:3306)/steellink?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		return
	}
	defer db.Close()

	orders := []struct {
		orderID      uint64
		orderNo      string
		contractID   uint64
		bindingID    uint64
		quoteID      uint64
	}{
		{66, "O202605131632348180", 4, 0, 0},
		{65, "O202605131631181513", 5, 0, 0},
		{67, "O202605141318057636", 6, 0, 0},
		{68, "O202605150131321004", 7, 0, 0},
	}

	for _, o := range orders {
		fmt.Printf("\n=== 处理订单 %s ===\n", o.orderNo)

		var biddingID, quoteID uint64
		var logisticsEnterpriseID uint64

		err := db.QueryRow("SELECT bidding_id, quote_id FROM logistics_contracts WHERE id = ?", o.contractID).Scan(&biddingID, &quoteID)
		if err != nil {
			fmt.Printf("查询合同失败: %v\n", err)
			continue
		}
		fmt.Printf("bidding_id: %d, quote_id: %d\n", biddingID, quoteID)

		err = db.QueryRow("SELECT logistics_enterprise_id FROM logistics_bid_quotes WHERE id = ?", quoteID).Scan(&logisticsEnterpriseID)
		if err != nil {
			fmt.Printf("查询报价失败: %v\n", err)
			continue
		}
		fmt.Printf("logistics_enterprise_id: %d\n", logisticsEnterpriseID)

		var logisticsServiceID uint64
		err = db.QueryRow("SELECT id FROM logistics_services WHERE enterprise_id = ? LIMIT 1", logisticsEnterpriseID).Scan(&logisticsServiceID)
		if err != nil {
			fmt.Printf("查询物流服务失败: %v\n", err)
			logisticsServiceID = 1
		}
		fmt.Printf("logistics_service_id: %d\n", logisticsServiceID)

		var pickupAddress, destinationAddress string
		var vehicleInfo string
		var amount float64

		err = db.QueryRow("SELECT pickup_address, destination_address FROM logistics_biddings WHERE id = ?", biddingID).Scan(&pickupAddress, &destinationAddress)
		if err != nil {
			fmt.Printf("查询竞价地址失败: %v\n", err)
		}

		err = db.QueryRow("SELECT vehicle_info, amount FROM logistics_bid_quotes WHERE id = ?", quoteID).Scan(&vehicleInfo, &amount)
		if err != nil {
			fmt.Printf("查询报价信息失败: %v\n", err)
		}

		if pickupAddress == "" {
			pickupAddress = "北京市-北京市-大兴区"
		}
		if destinationAddress == "" {
			destinationAddress = "河北省-唐山市-路北区"
		}
		if vehicleInfo == "" {
			vehicleInfo = "标准货车"
		}
		if amount == 0 {
			amount = 1000.00
		}

		rand.Seed(time.Now().UnixNano() + int64(o.orderID))
		suffix := fmt.Sprintf("%03d", rand.Intn(1000))
		logisticsNo := fmt.Sprintf("LGAUTO%s%s", time.Now().Format("20060102150405"), suffix)

		query := `
			INSERT INTO logistics_orders (
				logistics_no, order_id, order_no, buyer_enterprise_id,
				logistics_service_id, carrier_enterprise_id, warehouse_service_id,
				origin, destination, cargo_name, weight, vehicle_type,
				estimated_price, final_price, status, current_responsible_party,
				created_by, updated_by, created_at, updated_at
			) VALUES (?, ?, ?, 5, ?, ?, 0, ?, ?, '钢材', 10.00, ?, ?, ?, 'task_created', 'logistics', 5, 5, NOW(), NOW())
		`

		result, err := db.Exec(query, logisticsNo, o.orderID, o.orderNo, logisticsServiceID, logisticsEnterpriseID, pickupAddress, destinationAddress, vehicleInfo, amount, amount)
		if err != nil {
			fmt.Printf("创建物流订单失败: %v\n", err)
			continue
		}

		id, _ := result.LastInsertId()
		fmt.Printf("成功创建物流订单: id=%d, logistics_no=%s\n", id, logisticsNo)
	}

	fmt.Println("\n=== 验证所有物流合同都有物流订单 ===")
	rows, _ := db.Query(`
		SELECT c.order_no, c.status, lo.id as logistics_order_id, lo.logistics_no
		FROM logistics_contracts c
		LEFT JOIN logistics_orders lo ON c.order_id = lo.order_id
		WHERE c.status IN ('pending', 'buyer_signed')
	`)
	for rows.Next() {
		var orderNo, contractStatus string
		var logisticsOrderID uint64
		var logisticsNo sql.NullString
		if err := rows.Scan(&orderNo, &contractStatus, &logisticsOrderID, &logisticsNo); err != nil {
			fmt.Printf("扫描失败: %v\n", err)
			continue
		}
		if logisticsOrderID > 0 {
			fmt.Printf("✓ %s (%s) -> 物流订单: %s\n", orderNo, contractStatus, logisticsNo.String)
		} else {
			fmt.Printf("✗ %s (%s) -> 没有物流订单!\n", orderNo, contractStatus)
		}
	}
	rows.Close()
}