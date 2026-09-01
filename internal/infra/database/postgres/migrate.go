package postgres

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// ApplyAllMigrations 按文件名顺序执行 sql/migrations 下全部 .sql。
func ApplyAllMigrations(db *gorm.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		// *_seed.sql 由入口在引导账号后单独执行（种子依赖 demo 用户）
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") && !strings.Contains(e.Name(), "_seed") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	for _, f := range files {
		if err := ApplyMigrations(db, f); err != nil {
			return err
		}
	}
	return nil
}

// ApplyMigrations 逐个执行迁移文件中的 SQL 语句（按 ; 切分，忽略注释）。
func ApplyMigrations(db *gorm.DB, path string) error {
	demoUsername := os.Getenv("DEMO_USERNAME")
	if demoUsername == "" {
		demoUsername = "demo"
	}
	return ApplyMigrationsWithVars(db, path, map[string]string{"DEMO_USERNAME": demoUsername})
}

func ApplyMigrationsWithVars(db *gorm.DB, path string, vars map[string]string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	rawSQL, err := renderMigrationSQL(string(raw), vars)
	if err != nil {
		return fmt.Errorf("render migration %q: %w", path, err)
	}
	for _, stmt := range strings.Split(rawSQL, ";") {
		stmt = cleanSQL(stmt)
		if stmt == "" {
			continue
		}
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("exec %q: %w", stmt, err)
		}
	}
	return nil
}

var migrationVariablePattern = regexp.MustCompile(`\{\{([A-Za-z0-9_]+)\}\}`)

func renderMigrationSQL(raw string, vars map[string]string) (string, error) {
	var renderErr error
	rendered := migrationVariablePattern.ReplaceAllStringFunc(raw, func(token string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(token, "{{"), "}}")
		value, ok := vars[key]
		if !ok {
			renderErr = fmt.Errorf("missing variable %q", key)
			return token
		}
		if strings.IndexByte(value, 0) >= 0 {
			renderErr = fmt.Errorf("variable %q contains NUL", key)
			return token
		}
		return "'" + strings.ReplaceAll(value, "'", "''") + "'"
	})
	if renderErr != nil {
		return "", renderErr
	}
	return rendered, nil
}

func cleanSQL(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
