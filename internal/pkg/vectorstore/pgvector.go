package vectorstore

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PGVectorClient struct {
	dsn   string
	table string
}

type ProductHit struct {
	ProductID    uint64
	ProductNo    string
	ProductName  string
	Category     string
	EnterpriseID uint64
	Score        float64
}

type ProductEmbedding struct {
	ProductID       uint64
	ProductNo       string
	ProductName     string
	Category        string
	EnterpriseID    uint64
	SourceUpdatedAt time.Time
	Embedding       []float32
}

type ProductSyncWatermark struct {
	SourceUpdatedAt time.Time
	ProductID       uint64
	Initialized     bool
}

func NewPGVectorClient(cfg config.VectorSearchConfig) *PGVectorClient {
	return &PGVectorClient{
		dsn:   cfg.ConnectionString(),
		table: cfg.ProductCollection,
	}
}

func (c *PGVectorClient) EnsureProductSchema(ctx context.Context, dimension int) error {
	if c.dsn == "" || c.table == "" {
		return fmt.Errorf("pgvector config is incomplete")
	}
	if dimension <= 0 {
		dimension = 1536
	}
	db, err := sql.Open("pgx", c.dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS vector`); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	product_id BIGINT PRIMARY KEY,
	product_no VARCHAR(64) NOT NULL,
	product_name VARCHAR(255) NOT NULL,
	category VARCHAR(100) NOT NULL DEFAULT '',
	enterprise_id BIGINT NOT NULL DEFAULT 0,
	source_updated_at TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00+00',
	synced_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	embedding vector(%d) NOT NULL
)`, quoteIdentifier(c.table), dimension))
	if err != nil {
		return err
	}
	if err := c.ensureProductMetadataColumns(ctx, db); err != nil {
		return err
	}
	if err := c.ensureProductSyncState(ctx, db); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf(
		`CREATE INDEX IF NOT EXISTS %s ON %s USING hnsw (embedding vector_cosine_ops)`,
		quoteIdentifier(c.table+"_embedding_hnsw_idx"),
		quoteIdentifier(c.table),
	)); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf(
		`CREATE INDEX IF NOT EXISTS %s ON %s (category)`,
		quoteIdentifier(c.table+"_category_idx"),
		quoteIdentifier(c.table),
	)); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf(
		`CREATE INDEX IF NOT EXISTS %s ON %s (source_updated_at, product_id)`,
		quoteIdentifier(c.table+"_source_updated_at_idx"),
		quoteIdentifier(c.table),
	)); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf(
		`CREATE INDEX IF NOT EXISTS %s ON %s (enterprise_id)`,
		quoteIdentifier(c.table+"_enterprise_id_idx"),
		quoteIdentifier(c.table),
	))
	return err
}

func (c *PGVectorClient) ensureProductMetadataColumns(ctx context.Context, db *sql.DB) error {
	statements := []string{
		fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS source_updated_at TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00+00'`, quoteIdentifier(c.table)),
		fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ NOT NULL DEFAULT now()`, quoteIdentifier(c.table)),
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	hasTextNotNull, err := c.columnIsNotNull(ctx, db, "text")
	if err != nil {
		return err
	}
	if hasTextNotNull {
		_, err = db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL`, quoteIdentifier(c.table), quoteIdentifier("text")))
		return err
	}
	return nil
}

func (c *PGVectorClient) ensureProductSyncState(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	target VARCHAR(64) PRIMARY KEY,
	last_source_updated_at TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00+00',
	last_product_id BIGINT NOT NULL DEFAULT 0,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`, quoteIdentifier(c.syncStateTable())))
	return err
}

func (c *PGVectorClient) UpsertProduct(ctx context.Context, item ProductEmbedding) error {
	if c.dsn == "" || c.table == "" {
		return fmt.Errorf("pgvector config is incomplete")
	}
	if item.ProductID == 0 {
		return fmt.Errorf("product_id is required")
	}
	if len(item.Embedding) == 0 {
		return fmt.Errorf("embedding is required")
	}
	db, err := sql.Open("pgx", c.dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	query := fmt.Sprintf(`
INSERT INTO %s (
	product_id, product_no, product_name, category, enterprise_id, source_updated_at, synced_at, embedding
) VALUES (
	$1, $2, $3, $4, $5, $6, now(), $7::vector
)
ON CONFLICT (product_id) DO UPDATE SET
	product_no = EXCLUDED.product_no,
	product_name = EXCLUDED.product_name,
	category = EXCLUDED.category,
	enterprise_id = EXCLUDED.enterprise_id,
	source_updated_at = EXCLUDED.source_updated_at,
	synced_at = EXCLUDED.synced_at,
	embedding = EXCLUDED.embedding`, quoteIdentifier(c.table))
	_, err = db.ExecContext(ctx, query,
		item.ProductID,
		item.ProductNo,
		item.ProductName,
		item.Category,
		item.EnterpriseID,
		item.SourceUpdatedAt,
		vectorLiteral(item.Embedding),
	)
	return err
}

func (c *PGVectorClient) GetProductSyncWatermark(ctx context.Context) (ProductSyncWatermark, error) {
	if c.dsn == "" || c.table == "" {
		return ProductSyncWatermark{}, fmt.Errorf("pgvector config is incomplete")
	}
	db, err := sql.Open("pgx", c.dsn)
	if err != nil {
		return ProductSyncWatermark{}, err
	}
	defer db.Close()

	var watermark ProductSyncWatermark
	query := fmt.Sprintf(`SELECT last_source_updated_at, last_product_id FROM %s WHERE target = $1`, quoteIdentifier(c.syncStateTable()))
	if err := db.QueryRowContext(ctx, query, "products").Scan(&watermark.SourceUpdatedAt, &watermark.ProductID); err != nil {
		if err != sql.ErrNoRows {
			return ProductSyncWatermark{}, err
		}
		rowQuery := fmt.Sprintf(`
SELECT COALESCE(MAX(source_updated_at), '1970-01-01 00:00:00+00'::timestamptz), COALESCE(MAX(product_id), 0)
FROM %s`, quoteIdentifier(c.table))
		if scanErr := db.QueryRowContext(ctx, rowQuery).Scan(&watermark.SourceUpdatedAt, &watermark.ProductID); scanErr != nil {
			return ProductSyncWatermark{}, scanErr
		}
		return watermark, nil
	}
	watermark.Initialized = true
	return watermark, nil
}

func (c *PGVectorClient) SaveProductSyncWatermark(ctx context.Context, watermark ProductSyncWatermark) error {
	if c.dsn == "" || c.table == "" {
		return fmt.Errorf("pgvector config is incomplete")
	}
	db, err := sql.Open("pgx", c.dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	query := fmt.Sprintf(`
INSERT INTO %s (target, last_source_updated_at, last_product_id, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (target) DO UPDATE SET
	last_source_updated_at = EXCLUDED.last_source_updated_at,
	last_product_id = EXCLUDED.last_product_id,
	updated_at = EXCLUDED.updated_at`, quoteIdentifier(c.syncStateTable()))
	_, err = db.ExecContext(ctx, query, "products", watermark.SourceUpdatedAt, watermark.ProductID)
	return err
}

func (c *PGVectorClient) SearchProducts(ctx context.Context, vector []float32, topK int) ([]ProductHit, error) {
	if c.dsn == "" || c.table == "" {
		return nil, fmt.Errorf("pgvector config is incomplete")
	}
	if len(vector) == 0 {
		return nil, fmt.Errorf("search vector is empty")
	}
	if topK <= 0 {
		topK = 10
	}

	db, err := sql.Open("pgx", c.dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	query := fmt.Sprintf(`
SELECT product_id, product_no, product_name, category, enterprise_id, 1 - (embedding <=> $1::vector) AS score
FROM %s
ORDER BY embedding <=> $1::vector
LIMIT $2`, quoteIdentifier(c.table))
	rows, err := db.QueryContext(ctx, query, vectorLiteral(vector), topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hits := make([]ProductHit, 0, topK)
	for rows.Next() {
		var hit ProductHit
		if err := rows.Scan(&hit.ProductID, &hit.ProductNo, &hit.ProductName, &hit.Category, &hit.EnterpriseID, &hit.Score); err != nil {
			return nil, err
		}
		if hit.ProductID == 0 {
			continue
		}
		hits = append(hits, hit)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hits, nil
}

func (c *PGVectorClient) syncStateTable() string {
	parts := strings.Split(c.table, ".")
	parts[len(parts)-1] = parts[len(parts)-1] + "_sync_state"
	return strings.Join(parts, ".")
}

func (c *PGVectorClient) columnIsNotNull(ctx context.Context, db *sql.DB, column string) (bool, error) {
	schema, table, err := c.tableSchemaAndName(ctx, db)
	if err != nil {
		return false, err
	}
	var nullable string
	err = db.QueryRowContext(ctx, `
SELECT is_nullable
FROM information_schema.columns
WHERE table_schema = $1 AND table_name = $2 AND column_name = $3`, schema, table, column).Scan(&nullable)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return strings.EqualFold(nullable, "NO"), nil
}

func (c *PGVectorClient) tableSchemaAndName(ctx context.Context, db *sql.DB) (string, string, error) {
	parts := strings.Split(c.table, ".")
	if len(parts) >= 2 {
		return strings.Trim(parts[len(parts)-2], `"`), strings.Trim(parts[len(parts)-1], `"`), nil
	}
	var schema string
	if err := db.QueryRowContext(ctx, `SELECT current_schema()`).Scan(&schema); err != nil {
		return "", "", err
	}
	return schema, strings.Trim(c.table, `"`), nil
}

func vectorLiteral(vector []float32) string {
	parts := make([]string, 0, len(vector))
	for _, value := range vector {
		parts = append(parts, strconv.FormatFloat(float64(value), 'f', -1, 32))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func quoteIdentifier(value string) string {
	parts := strings.Split(value, ".")
	for i, part := range parts {
		parts[i] = `"` + strings.ReplaceAll(part, `"`, `""`) + `"`
	}
	return strings.Join(parts, ".")
}

func ConnectionStringFromEnv(envName string) string {
	if envName == "" {
		return ""
	}
	return os.Getenv(envName)
}
