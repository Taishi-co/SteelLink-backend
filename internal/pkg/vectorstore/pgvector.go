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

func NewPGVectorClient(cfg config.VectorSearchConfig) *PGVectorClient {
	return &PGVectorClient{
		dsn:   cfg.ConnectionString(),
		table: cfg.ProductCollection,
	}
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
