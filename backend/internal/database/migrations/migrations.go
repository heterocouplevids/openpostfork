package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

//go:embed *.sql
var migrationFiles embed.FS

type SchemaMigration struct {
	bun.BaseModel `bun:"table:schema_migrations"`
	Version       int64 `bun:",pk"`
	AppliedAt     int64 `bun:",notnull"`
}

// RunMigrations executes all pending migrations in order.
// Migration files must be named like: 001_description.sql, 002_description.sql, etc.
func RunMigrations(db *bun.DB) error {
	ctx := context.Background()

	// Ensure migrations table exists
	if _, err := db.NewCreateTable().Model((*SchemaMigration)(nil)).IfNotExists().Exec(ctx); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Get already applied versions
	var applied []SchemaMigration
	if err := db.NewSelect().Model(&applied).Order("version ASC").Scan(ctx); err != nil {
		return fmt.Errorf("failed to list applied migrations: %w", err)
	}
	appliedSet := make(map[int64]bool)
	for _, m := range applied {
		appliedSet[m.Version] = true
	}

	// Read embedded migration files
	entries, err := migrationFiles.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version, err := parseVersion(entry.Name())
		if err != nil {
			return fmt.Errorf("invalid migration filename %q: %w", entry.Name(), err)
		}
		content, err := migrationFiles.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %q: %w", entry.Name(), err)
		}
		migrations = append(migrations, migration{
			version: version,
			name:    entry.Name(),
			sql:     normalizeMigrationSQL(db.Dialect().Name(), string(content)),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	// Run pending migrations inside transactions
	for _, m := range migrations {
		if appliedSet[m.version] {
			continue
		}

		if err := runMigration(ctx, db, m); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.name, err)
		}
	}

	return nil
}

var (
	postgresBlobTypeExpr         = regexp.MustCompile(`(?i)\bBLOB\b`)
	postgresDateTimeTypeExpr     = regexp.MustCompile(`(?i)\bDATETIME\b`)
	postgresBooleanIsActiveFalse = regexp.MustCompile(`\bis_active\s*=\s*0\b`)
	postgresBooleanIsActiveTrue  = regexp.MustCompile(`\bis_active\s*=\s*1\b`)
)

func normalizeMigrationSQL(name dialect.Name, raw string) string {
	if name != dialect.PG {
		return raw
	}

	out := postgresBlobTypeExpr.ReplaceAllString(raw, "BYTEA")
	out = postgresDateTimeTypeExpr.ReplaceAllString(out, "TIMESTAMPTZ")
	out = postgresBooleanIsActiveFalse.ReplaceAllString(out, "is_active = FALSE")
	out = postgresBooleanIsActiveTrue.ReplaceAllString(out, "is_active = TRUE")
	return out
}

type migration struct {
	version int64
	name    string
	sql     string
}

func runMigration(ctx context.Context, db *bun.DB, m migration) error {
	return db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
		// Split by ";" and execute each statement
		statements := splitStatements(m.sql)
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := tx.ExecContext(txCtx, stmt); err != nil {
				// SQLite: ignore "duplicate column name" — migration may already be applied
				// via CreateSchema on a fresh database. Postgres reports the same
				// condition as "column ... already exists".
				if isDuplicateColumnMigrationError(stmt, err) {
					continue
				}
				return fmt.Errorf("statement failed: %s: %w", stmt, err)
			}
		}

		// Record migration
		record := &SchemaMigration{
			Version:   m.version,
			AppliedAt: time.Now().Unix(),
		}
		if _, err := tx.NewInsert().Model(record).Exec(txCtx); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}
		return nil
	})
}

func isDuplicateColumnMigrationError(stmt string, err error) bool {
	if err == nil {
		return false
	}
	normalizedStmt := strings.ToUpper(strings.TrimSpace(stmt))
	if !strings.HasPrefix(normalizedStmt, "ALTER TABLE") || !strings.Contains(normalizedStmt, "ADD COLUMN") {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate column name") || strings.Contains(message, "already exists")
}

func parseVersion(filename string) (int64, error) {
	base := path.Base(filename)
	parts := strings.SplitN(base, "_", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("filename must start with a version number")
	}
	return strconv.ParseInt(parts[0], 10, 64)
}

func splitStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}
	return statements
}
