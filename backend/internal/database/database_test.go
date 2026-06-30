package database

import (
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
