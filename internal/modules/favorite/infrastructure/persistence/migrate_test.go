package persistence

import (
	"strings"
	"testing"
)

func TestDuplicateFavoriteCleanupSQLPrefersActiveLegacyRows(t *testing.T) {
	sql := duplicateFavoriteCleanupSQL(true)

	if !strings.Contains(sql, "PARTITION BY user_id, product_id") {
		t.Fatalf("cleanup SQL should dedupe by owner and product: %s", sql)
	}
	if !strings.Contains(sql, "CASE WHEN deleted_at IS NULL THEN 0 ELSE 1 END, id ASC") {
		t.Fatalf("cleanup SQL should keep active legacy rows before soft-deleted rows: %s", sql)
	}
	if !strings.Contains(sql, "WHERE rn > 1") {
		t.Fatalf("cleanup SQL should delete only duplicate rows: %s", sql)
	}
}

func TestDuplicateFavoriteCleanupSQLHandlesCurrentSchema(t *testing.T) {
	sql := duplicateFavoriteCleanupSQL(false)

	if strings.Contains(sql, "deleted_at") {
		t.Fatalf("current-schema cleanup SQL should not reference deleted_at: %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY id ASC") {
		t.Fatalf("current-schema cleanup SQL should keep the earliest row: %s", sql)
	}
}
