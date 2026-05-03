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
	ProductID    uint64
	ProductNo    string
	ProductName  string
	Category     string
	EnterpriseID uint64
	Embedding    []float32
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
	embedding vector(%d) NOT NULL
)`, quoteIdentifier(c.table), dimension))
	if err != nil {
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
	_, err = db.ExecContext(ctx, fmt.Sprintf(
		`CREATE INDEX IF NOT EXISTS %s ON %s (enterprise_id)`,
		quoteIdentifier(c.table+"_enterprise_id_idx"),
		quoteIdentifier(c.table),
	))
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
	product_id, product_no, product_name, category, enterprise_id, embedding
) VALUES (
	$1, $2, $3, $4, $5, $6::vector
)
ON CONFLICT (product_id) DO UPDATE SET
	product_no = EXCLUDED.product_no,
	product_name = EXCLUDED.product_name,
	category = EXCLUDED.category,
	enterprise_id = EXCLUDED.enterprise_id,
	embedding = EXCLUDED.embedding`, quoteIdentifier(c.table))
	_, err = db.ExecContext(ctx, query,
		item.ProductID,
		item.ProductNo,
		item.ProductName,
		item.Category,
		item.EnterpriseID,
		vectorLiteral(item.Embedding),
	)
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
