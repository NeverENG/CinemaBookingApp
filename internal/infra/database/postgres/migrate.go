package postgres

import (
	"fmt"
	"os"
	"path/filepath"
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
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
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
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, stmt := range strings.Split(string(raw), ";") {
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
