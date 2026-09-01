package postgres

import (
	"strings"
	"testing"
)

func TestRenderMigrationSQLQuotesVariables(t *testing.T) {
	rendered, err := renderMigrationSQL("SELECT {{DEMO_USERNAME}}", map[string]string{"DEMO_USERNAME": "o'conn"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if rendered != "SELECT 'o''conn'" {
		t.Fatalf("unexpected SQL: %s", rendered)
	}
}

func TestRenderMigrationSQLRejectsMissingVariables(t *testing.T) {
	_, err := renderMigrationSQL("SELECT {{DEMO_USERNAME}}", map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "DEMO_USERNAME") {
		t.Fatalf("expected missing variable error, got %v", err)
	}
}
