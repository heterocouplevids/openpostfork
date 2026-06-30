package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitDBPreservesSQLiteDefault(t *testing.T) {
	db, err := InitDB("file::memory:?cache=shared")
	require.NoError(t, err)
	defer db.Close()

	var one int
	require.NoError(t, db.QueryRow("SELECT 1").Scan(&one))
	require.Equal(t, 1, one)
}

func TestInitDBWithDriverInitializesSQLite(t *testing.T) {
	db, err := InitDBWithDriver("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)
	defer db.Close()

	var one int
	require.NoError(t, db.QueryRow("SELECT 1").Scan(&one))
	require.Equal(t, 1, one)
}

func TestInitDBWithDriverRejectsUnsupportedDriver(t *testing.T) {
	db, err := InitDBWithDriver("mysql", "mysql://example")
	require.Nil(t, db)
	require.ErrorContains(t, err, "unsupported database driver")
}

func TestInitDBWithDriverBuildsPostgresHandle(t *testing.T) {
	db, err := InitDBWithDriver("postgres", "postgres://openpost:secret@localhost:5432/openpost?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	require.NotNil(t, db)
}

func TestCreateSchemaRunsPublicationMigration(t *testing.T) {
	db, err := InitDBWithDriver("sqlite", "file:"+t.Name()+"?mode=memory&cache=private")
	require.NoError(t, err)
	defer db.Close()
	ctx := context.Background()

	require.NoError(t, CreateSchema(db))

	for _, table := range []string{"publications", "publication_assets"} {
		var count int
		require.NoError(t, db.NewSelect().
			ColumnExpr("COUNT(*)").
			TableExpr("sqlite_master").
			Where("type = 'table' AND name = ?", table).
			Scan(ctx, &count))
		require.Equal(t, 1, count, table)
	}

	row := db.QueryRow("SELECT sql FROM sqlite_master WHERE name = 'posts'")
	var postSchema string
	require.NoError(t, row.Scan(&postSchema))
	require.Contains(t, postSchema, "publication_id")
}
