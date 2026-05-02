package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/embedding"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/vectorstore"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type syncOptions struct {
	configPath   string
	target       string
	schedule     string
	batchSize    int
	limit        int
	createSchema bool
}

func main() {
	opts := syncOptions{}
	flag.StringVar(&opts.configPath, "config", "config/config.yaml", "config file path")
	flag.StringVar(&opts.target, "target", "products", "sync target: products, suppliers, finance, all")
	flag.StringVar(&opts.schedule, "schedule", "manual", "schedule: manual, minute, hour, day")
	flag.IntVar(&opts.batchSize, "batch-size", 100, "products per batch")
	flag.IntVar(&opts.limit, "limit", 0, "max products to sync, 0 means no limit")
	flag.BoolVar(&opts.createSchema, "create-schema", true, "create pgvector extension/table/indexes when possible")
	flag.Parse()

	_ = godotenv.Load()
	if err := config.Init(opts.configPath); err != nil {
		fmt.Fprintf(os.Stderr, "init config failed: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Get()
	if err := database.InitMySQL(&cfg.Database); err != nil {
		fmt.Fprintf(os.Stderr, "connect mysql failed: %v\n", err)
		os.Exit(1)
	}
	defer database.CloseMySQL()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := runScheduled(ctx, opts); err != nil {
		fmt.Fprintf(os.Stderr, "vector sync failed: %v\n", err)
		os.Exit(1)
	}
}

func runScheduled(ctx context.Context, opts syncOptions) error {
	interval, once, err := parseSchedule(opts.schedule)
	if err != nil {
		return err
	}
	for {
		startedAt := time.Now()
		if err := runOnce(ctx, opts); err != nil {
			return err
		}
		fmt.Printf("vector sync finished schedule=%s target=%s duration=%s\n", opts.schedule, opts.target, time.Since(startedAt).Round(time.Millisecond))
		if once {
			return nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}

func parseSchedule(value string) (time.Duration, bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "manual", "once":
		return 0, true, nil
	case "minute", "minutely":
		return time.Minute, false, nil
	case "hour", "hourly":
		return time.Hour, false, nil
	case "day", "daily":
		return 24 * time.Hour, false, nil
	default:
		return 0, false, fmt.Errorf("unsupported schedule %q", value)
	}
}

func runOnce(ctx context.Context, opts syncOptions) error {
	target := strings.ToLower(strings.TrimSpace(opts.target))
	if target == "" {
		target = "products"
	}
	switch target {
	case "products":
		return syncProducts(ctx, opts)
	case "suppliers", "supplier":
		fmt.Println("supplier vector sync is reserved; no-op in this version")
		return nil
	case "finance", "finance-products", "finance_products":
		fmt.Println("finance product vector sync is reserved; no-op in this version")
		return nil
	case "all":
		if err := syncProducts(ctx, opts); err != nil {
			return err
		}
		fmt.Println("supplier vector sync is reserved; no-op in this version")
		fmt.Println("finance product vector sync is reserved; no-op in this version")
		return nil
	default:
		return fmt.Errorf("unsupported target %q", opts.target)
	}
}

func syncProducts(ctx context.Context, opts syncOptions) error {
	cfg := config.GetAIProcurement().VectorSearch
	if cfg.Provider != "pgvector" {
		return fmt.Errorf("vector provider must be pgvector, got %q", cfg.Provider)
	}
	if cfg.ConnectionString() == "" {
		return fmt.Errorf("%s is not configured", cfg.ConnectionEnv)
	}
	if opts.batchSize <= 0 {
		opts.batchSize = 100
	}

	store := vectorstore.NewPGVectorClient(cfg)
	if opts.createSchema {
		if err := store.EnsureProductSchema(ctx, cfg.Dimension); err != nil {
			return fmt.Errorf("ensure pgvector product schema failed: %w", err)
		}
	}
	embedder := embedding.NewClient(cfg)

	var synced int
	var lastID uint64
	db := database.GetDB()
	for {
		remaining := opts.limit - synced
		if opts.limit > 0 && remaining <= 0 {
			break
		}
		batchSize := opts.batchSize
		if opts.limit > 0 && remaining < batchSize {
			batchSize = remaining
		}
		products, err := listProducts(ctx, db, lastID, batchSize)
		if err != nil {
			return err
		}
		if len(products) == 0 {
			break
		}
		for _, product := range products {
			text := buildProductEmbeddingText(product)
			vector, err := embedder.Embed(ctx, text)
			if err != nil {
				return fmt.Errorf("embed product %d failed: %w", product.ID, err)
			}
			if err := store.UpsertProduct(ctx, vectorstore.ProductEmbedding{
				ProductID:    product.ID,
				ProductNo:    product.ProductNo,
				ProductName:  product.Name,
				Category:     productCategory(product),
				EnterpriseID: product.EnterpriseID,
				Text:         text,
				Embedding:    vector,
				Status:       int(product.Status),
				AvailableQty: product.AvailableQty,
				UnitPrice:    product.UnitPrice,
			}); err != nil {
				return fmt.Errorf("upsert product %d vector failed: %w", product.ID, err)
			}
			synced++
			lastID = product.ID
		}
		fmt.Printf("synced product vectors count=%d last_id=%d\n", synced, lastID)
	}
	fmt.Printf("product vector sync complete count=%d\n", synced)
	return nil
}

func listProducts(ctx context.Context, db *gorm.DB, lastID uint64, limit int) ([]*productEntity.Product, error) {
	var products []*productEntity.Product
	err := db.WithContext(ctx).
		Where("id > ?", lastID).
		Order("id ASC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

func buildProductEmbeddingText(product *productEntity.Product) string {
	parts := []string{
		"商品:" + product.Name,
		"品类:" + product.Category,
		"子品类:" + product.SubCategory,
		"钢厂:" + product.SteelMill,
		fmt.Sprintf("库存:%.2f", product.AvailableQty),
		fmt.Sprintf("价格:%.2f", product.UnitPrice),
		fmt.Sprintf("交期:%d天", product.DeliveryDays),
	}
	if product.Spec != nil {
		parts = append(parts,
			"材质:"+product.Spec.Material,
			"规格:"+product.Spec.Spec,
			"厚度:"+product.Spec.Thickness,
			"宽度:"+product.Spec.Width,
			"长度:"+product.Spec.Length,
			"表面:"+product.Spec.Surface,
			"加工:"+product.Spec.Processing,
		)
	}
	if product.WarehouseInfo != nil {
		parts = append(parts,
			"仓库:"+product.WarehouseInfo.WarehouseName,
			"省份:"+product.WarehouseInfo.Province,
			"城市:"+product.WarehouseInfo.City,
			"地址:"+product.WarehouseInfo.Address,
		)
	}
	if product.Description != "" {
		parts = append(parts, "描述:"+product.Description)
	}
	return strings.Join(compactParts(parts), "；")
}

func productCategory(product *productEntity.Product) string {
	if product.SubCategory != "" {
		return product.SubCategory
	}
	return product.Category
}

func compactParts(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && !strings.HasSuffix(part, ":") {
			out = append(out, part)
		}
	}
	return out
}
