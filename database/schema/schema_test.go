package schema

import (
	"strings"
	"testing"
)

func TestSplitStatements(t *testing.T) {
	script := `-- table comment
CREATE TABLE IF NOT EXISTS a (id int); -- trailing comment
CREATE INDEX IF NOT EXISTS i ON a (id);

CREATE OR REPLACE FUNCTION f() RETURNS trigger AS $$
BEGIN
    NEW.x := 'a;b';  -- semicolons inside must not split
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;`
	stmts := splitStatements(script)
	if len(stmts) != 3 {
		t.Fatalf("statements = %d, want 3:\n%q", len(stmts), stmts)
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
