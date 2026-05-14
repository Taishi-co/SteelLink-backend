//go:build ignore

package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"

	"github.com/go-sql-driver/mysql"
)

func main() {
	// 测试不同连接参数
	dsns := []string{
		"root:Steellink2026@tcp(117.72.123.14:3306)/steellink?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&tls=false",
		"root:Steellink2026@tcp(117.72.123.14:3306)/steellink?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&tls=skip-verify",
		"root:Steellink2026@tcp(117.72.123.14:3306)/steellink?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s",
	}

	for i, dsn := range dsns {
		fmt.Printf("\n=== 测试 DSN %d ===\n", i+1)
		fmt.Printf("DSN: %s\n", dsn)

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("Failed to open database: %v", err)
			continue
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			log.Printf("Failed to ping database: %v", err)
		} else {
			fmt.Println("Database connection successful!")
			break
		}
	}
}

func init() {
	// 注册自定义TLS配置
	mysql.RegisterTLSConfig("skip-verify", &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            x509.NewCertPool(),
	})
}
