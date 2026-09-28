package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/models"
)

func validateArticleExportOptions(path, format string) error {
	if format != "markdown" && format != "json" {
		return fmt.Errorf("--export-format must be markdown or json")
	}
	if format == "json" && path == "" {
		return fmt.Errorf("--export-format json requires --export <file>")
	}
	return nil
}

func writeArticleExport(articleData map[string]interface{}, tweet *models.Tweet, path, format string) error {
	if err := validateArticleExportOptions(path, format); err != nil {
		return err
	}
	if path == "" {
		return fmt.Errorf("--export <file> is required")
	}
	if format == "markdown" {
		return core.ExportArticleToFile(articleData, tweet, path)
	}

	data, err := core.ArticleToJSON(articleData, tweet)
	if err != nil {
		return fmt.Errorf("serialize article: %w", err)
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create export directory: %w", err)
		}
	}
	if err := os.WriteFile(path, []byte(data+"\n"), 0644); err != nil {
		return fmt.Errorf("write article JSON: %w", err)
	}
	return nil
}
