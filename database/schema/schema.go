// Package schema embeds the database DDL and applies it idempotently at
// startup, so deployments need nothing but the image: no SQL mounts, no
// manual bootstrap. Every statement is written with IF NOT EXISTS / OR
// REPLACE semantics and is safe to run on every boot.
package schema

import (
	"embed"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

//go:embed rm_search.sql
var files embed.FS

// ddl is the schema file content.
var ddl string

func init() {
	data, err := files.ReadFile("rm_search.sql")
	if err != nil {
		panic(fmt.Errorf("read embedded schema: %w", err))
	}
	ddl = string(data)
}

// Ensure applies the embedded DDL. Statements are split on ';' with
// awareness of single-quoted strings and dollar-quoted function bodies
// ($$ ... $$), which both may contain semicolons.
func Ensure(db *gorm.DB) error {
	for _, stmt := range splitStatements(ddl) {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("apply schema statement %.80q: %w", stmt, err)
		}
	}
	return nil
}

// splitStatements splits a PostgreSQL script into individual statements.
func splitStatements(script string) []string {
	var (
		statements []string
		sb         strings.Builder
		inSingle   bool // inside '...'
		inDollar   bool // inside $$...$$
	)
	runes := []rune(script)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		sb.WriteRune(r)
		switch {
		case inSingle:
			if r == '\'' {
				inSingle = false
			}
		case inDollar:
			if r == '$' && i+1 < len(runes) && runes[i+1] == '$' {
				inDollar = false
				sb.WriteRune(runes[i+1])
				i++
			}
		case r == '\'':
			inSingle = true
		case r == '$' && i+1 < len(runes) && runes[i+1] == '$':
			inDollar = true
			sb.WriteRune(runes[i+1])
			i++
		case r == ';':
			if s := strings.TrimSpace(sb.String()); s != ";" && strings.TrimSpace(strings.TrimSuffix(s, ";")) != "" {
				statements = append(statements, strings.TrimSpace(strings.TrimSuffix(s, ";")))
			}
			sb.Reset()
		}
	}
	if s := strings.TrimSpace(sb.String()); s != "" {
		statements = append(statements, s)
	}
	return statements
}
