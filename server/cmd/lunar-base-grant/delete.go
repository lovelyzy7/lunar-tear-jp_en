package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"lunar-tear/server/internal/database"
)

// queryer is the subset of *sql.DB / *sql.Tx the schema introspection needs.
type queryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// runDeleteUser permanently removes one account and every row that belongs to
// it, in a single transaction.
//
// The set of child tables is discovered from sqlite_master instead of being
// hard-coded: any table carrying a `user_id` column gets
//
//	DELETE FROM <table> WHERE user_id = ?
//
// so tables added by later lunar-tear migrations are covered automatically.
// PRAGMA defer_foreign_keys makes SQLite verify foreign keys only at COMMIT,
// by which point every referencing row is gone, so deletion order does not
// matter.
func runDeleteUser(req *request) (int, error) {
	if req.UserID <= 0 {
		return 0, errors.New("user_id must be positive")
	}

	db, err := database.Open(req.DBPath)
	if err != nil {
		return 0, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	var exists bool
	if err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE user_id = ?)`, req.UserID,
	).Scan(&exists); err != nil {
		return 0, fmt.Errorf("check user: %w", err)
	}
	if !exists {
		return 0, fmt.Errorf("user %d not found", req.UserID)
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`PRAGMA defer_foreign_keys = ON`); err != nil {
		return 0, fmt.Errorf("defer foreign keys: %w", err)
	}

	tables, err := tablesWithUserID(tx)
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, table := range tables {
		res, err := tx.Exec(
			fmt.Sprintf(`DELETE FROM %s WHERE user_id = ?`, quoteIdent(table)), req.UserID,
		)
		if err != nil {
			return 0, fmt.Errorf("delete from %s: %w", table, err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("rows affected for %s: %w", table, err)
		}
		deleted += int(n)
	}

	res, err := tx.Exec(`DELETE FROM users WHERE user_id = ?`, req.UserID)
	if err != nil {
		return 0, fmt.Errorf("delete user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for users: %w", err)
	}
	deleted += int(n)

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return deleted, nil
}

// tablesWithUserID returns the name of every real table that has a `user_id`
// column, in sqlite_master order. Internal tables (sqlite_*) are skipped.
func tablesWithUserID(q queryer) ([]string, error) {
	names, err := allTableNames(q)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, name := range names {
		ok, err := hasUserIDColumn(q, name)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, name)
		}
	}
	return out, nil
}

func allTableNames(q queryer) ([]string, error) {
	rows, err := q.Query(
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan table name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	return names, nil
}

func hasUserIDColumn(q queryer, table string) (bool, error) {
	rows, err := q.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, quoteIdent(table)))
	if err != nil {
		return false, fmt.Errorf("table info for %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notNull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			return false, fmt.Errorf("scan table info for %s: %w", table, err)
		}
		if name == "user_id" {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("table info for %s: %w", table, err)
	}
	return false, nil
}

// quoteIdent safely quotes a SQLite identifier (table name).
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
