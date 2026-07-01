package database

import (
	"fmt"
	"regexp"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

var (
	safeSQLIdentifierExpr = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)?$`)
	safeJSONKeyExpr       = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

func JSONTextExpr(db *bun.DB, column, key string) string {
	return JSONTextExprForDialect(db.Dialect().Name(), column, key)
}

func JSONTextExprForDialect(name dialect.Name, column, key string) string {
	if !safeSQLIdentifierExpr.MatchString(column) {
		panic(fmt.Sprintf("unsafe JSON SQL column expression %q", column))
	}
	if !safeJSONKeyExpr.MatchString(key) {
		panic(fmt.Sprintf("unsafe JSON key %q", key))
	}

	if name == dialect.PG {
		return fmt.Sprintf("(%s::jsonb ->> '%s')", column, key)
	}
	return fmt.Sprintf("json_extract(%s, '$.%s')", column, key)
}
