package handlers

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/openpost/backend/internal/models"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

func TestApplyRandomDelayStaysWithinBounds(t *testing.T) {
	scheduledAt := time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC)
	const maxDelay = 15

	for i := 0; i < 200; i++ {
		actual := applyRandomDelay(scheduledAt, maxDelay)
		diff := actual.Sub(scheduledAt)
		if diff < -15*time.Minute || diff > 15*time.Minute {
			t.Fatalf("random delay out of bounds: got %v", diff)
		}
	}
}

func TestApplyRandomDelayWithZeroDelayReturnsScheduledTime(t *testing.T) {
	scheduledAt := time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC)

	actual := applyRandomDelay(scheduledAt, 0)
	if !actual.Equal(scheduledAt) {
		t.Fatalf("expected unchanged time, got %s want %s", actual, scheduledAt)
	}
}

func TestListPostsOrderExpressionKeepsCoalesceCall(t *testing.T) {
	sqldb, err := sql.Open("sqlite3", "file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=private")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqldb.Close()
	})

	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() {
		_ = db.Close()
	})

	query := db.NewSelect().
		Model((*models.Post)(nil))
	query = applyListPostsOrder(query).Limit(50)

	require.Contains(t, query.String(), "ORDER BY COALESCE(scheduled_at, created_at) DESC")

	_, err = db.NewCreateTable().Model((*models.Post)(nil)).IfNotExists().Exec(context.Background())
	require.NoError(t, err)

	var posts []models.Post
	query = db.NewSelect().Model(&posts)
	err = applyListPostsOrder(query).Limit(50).Scan(context.Background())
	require.NoError(t, err)
}
