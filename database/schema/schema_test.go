package schema

import (
	"strings"
	"testing"
)

func TestSplitStatements(t *testing.T) {
	script := `-- table comment
CREATE TABLE IF NOT EXISTS a (id int); -- trailing comment
CREATE INDEX IF NOT EXISTS i ON a (id);

-- PostgreSQL has no ON UPDATE CURRENT_TIMESTAMP; emulate it so
-- update_time keeps the semantics the MySQL schema had.
CREATE OR REPLACE FUNCTION f() RETURNS trigger AS $$
BEGIN
    NEW.x := 'a;b';  -- semicolons inside must not split
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

/* block comment; with a semicolon */
SELECT 'a;b' AS x;`
	stmts := splitStatements(script)
	if len(stmts) != 4 {
		t.Fatalf("statements = %d, want 4:\n%q", len(stmts), stmts)
	}
	// The comment header and the function must form ONE statement, with
	// the ';' inside the line comment ignored (regression: it used to
	// split "emulate it so ..." into a bogus statement).
	if !strings.HasPrefix(stmts[2], "-- PostgreSQL has no ON UPDATE") {
		t.Errorf("comment header detached: %q", stmts[2])
	}
	if !strings.Contains(stmts[2], "LANGUAGE plpgsql") {
		t.Errorf("function statement truncated: %q", stmts[2])
	}
	// The dollar-quoted body must survive intact in one statement.
	for _, frag := range []string{"$$", "NEW.x :=", "'a;b'", "RETURN NEW;", "END;"} {
		if !strings.Contains(stmts[2], frag) {
			t.Errorf("function statement lost %q: %q", frag, stmts[2])
		}
	}
	if !strings.HasPrefix(stmts[3], "/* block comment") {
		t.Errorf("block comment statement wrong: %q", stmts[3])
	}
}

func TestDDLIsIdempotent(t *testing.T) {
	// Every DDL statement must be re-runnable: CREATE TABLE/INDEX with IF
	// NOT EXISTS, functions with OR REPLACE, triggers dropped before create.
	for _, stmt := range splitStatements(ddl) {
		upper := strings.ToUpper(stmt)
		switch {
		case strings.HasPrefix(upper, "CREATE TABLE ") || strings.HasPrefix(upper, "CREATE INDEX "):
			if !strings.Contains(upper, "IF NOT EXISTS") {
				t.Errorf("non-idempotent statement: %.60q", stmt)
			}
		case strings.HasPrefix(upper, "CREATE TRIGGER "):
			t.Errorf("CREATE TRIGGER without preceding DROP is not idempotent: %.60q", stmt)
		}
	}
}
