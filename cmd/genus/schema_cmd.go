package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-genus/genus/core"
	"github.com/go-genus/genus/migrate"
)

func runSchema() error {
	if len(os.Args) < 3 {
		printSchemaUsage()
		return nil
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "diff":
		return runSchemaDiff()
	case "help", "--help", "-h":
		printSchemaUsage()
		return nil
	default:
		return fmt.Errorf("unknown schema subcommand: %s", subcommand)
	}
}

func runSchemaDiff() error {
	args := os.Args[3:]

	var (
		modelsDir = "./models"
		generate  = false
	)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--models", "-m":
			if i+1 < len(args) {
				modelsDir = args[i+1]
				i++
			}
		case "--generate", "-g":
			generate = true
		case "-h", "--help":
			printSchemaDiffUsage()
			return nil
		}
	}

	// Connect to database
	db, dialect, err := connectDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	differ := migrate.NewSchemaDiffer(db, dialect)

	// Get current database schema
	fmt.Printf("%sFetching database schema...%s\n", colorCyan, colorReset)
	currentSchema, err := differ.GetCurrentSchema(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current schema: %w", err)
	}

	// Get target schema from model files
	fmt.Printf("%sAnalyzing models in %s...%s\n", colorCyan, modelsDir, colorReset)
	targetSchema, err := getSchemaFromModelDir(modelsDir, dialect)
	if err != nil {
		return fmt.Errorf("failed to analyze models: %w", err)
	}

	// Compute diff
	changes := differ.Diff(currentSchema, targetSchema)

	if len(changes) == 0 {
		fmt.Printf("\n%s[OK]%s Schema is up to date. No changes detected.\n", colorGreen, colorReset)
		return nil
	}

	// Display diff
	fmt.Printf("\n%sSchema Changes:%s\n", colorBold, colorReset)
	fmt.Println(strings.Repeat("-", 70))

	for _, change := range changes {
		printSchemaChange(change)
	}

	fmt.Printf("\n%d change(s) detected.\n", len(changes))

	// Generate migration if requested
	if generate {
		return generateMigrationFromChanges(changes, differ)
	}

	fmt.Printf("\nRun with %s--generate%s to create a migration file.\n", colorCyan, colorReset)
	return nil
}

func printSchemaChange(change migrate.SchemaChange) {
	var prefix, color string

	switch change.Type {
	case migrate.ChangeAddTable, migrate.ChangeAddColumn, migrate.ChangeAddIndex, migrate.ChangeAddForeignKey:
		prefix = "+"
		color = "\033[32m" // green
	case migrate.ChangeDropTable, migrate.ChangeDropColumn, migrate.ChangeDropIndex, migrate.ChangeDropForeignKey:
		prefix = "-"
		color = "\033[31m" // red
	case migrate.ChangeModifyColumn:
		prefix = "~"
		color = "\033[33m" // yellow
	default:
		prefix = " "
		color = ""
	}

	fmt.Printf("  %s%s %s%s\n", color, prefix, change.Description, colorReset)

	if change.SQL != "" {
		// Indent SQL
		lines := strings.Split(change.SQL, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				fmt.Printf("    %s%s%s\n", colorDim, line, colorReset)
			}
		}
	}
}

func generateMigrationFromChanges(changes []migrate.SchemaChange, differ *migrate.SchemaDiffer) error {
	migrationsDir := "./migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	timestamp := time.Now().Unix()
	name := "auto_schema_diff"
	filename := fmt.Sprintf("%d_%s.sql", timestamp, name)
	filePath := filepath.Join(migrationsDir, filename)

	// Build migration SQL
	var upSQL, downSQL strings.Builder

	upSQL.WriteString("-- Auto-generated migration from schema diff\n")
	upSQL.WriteString(fmt.Sprintf("-- Generated at: %s\n\n", time.Now().Format(time.RFC3339)))

	downSQL.WriteString("-- Rollback migration\n\n")

	for _, change := range changes {
		if change.SQL != "" {
			upSQL.WriteString(change.SQL)
			upSQL.WriteString(";\n\n")
		}
		if change.ReverseSQL != "" {
			downSQL.WriteString(change.ReverseSQL)
			downSQL.WriteString(";\n\n")
		}
	}

	content := fmt.Sprintf("-- Migration: %s\n-- Version: %d\n\n-- +migrate Up\n%s\n-- +migrate Down\n%s",
		name, timestamp, upSQL.String(), downSQL.String())

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write migration file: %w", err)
	}

	fmt.Printf("\n%s[OK]%s Created migration: %s\n", colorGreen, colorReset, filePath)
	return nil
}

// getSchemaFromModelDir builds a target schema by parsing Go struct files in the given directory.
// It uses the SchemaDiffer's model parsing capabilities.
func getSchemaFromModelDir(dir string, dialect core.Dialect) (map[string]*migrate.TableSchema, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("models directory not found: %s", dir)
	}

	// We need a temporary DB connection for model parsing via SchemaDiffer
	// Use a dummy in-memory sqlite for parsing only
	tmpDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		// If sqlite3 not available, use the provided dialect directly
		return parseModelsManually(dir, dialect)
	}
	defer tmpDB.Close()

	differ := migrate.NewSchemaDiffer(tmpDB, dialect)
	_ = differ

	return parseModelsManually(dir, dialect)
}

// parseModelsManually parses Go files for struct definitions with db tags
// and builds a target schema map.
func parseModelsManually(dir string, dialect core.Dialect) (map[string]*migrate.TableSchema, error) {
	schemas := make(map[string]*migrate.TableSchema)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_test.go") || strings.HasSuffix(entry.Name(), ".gen.go") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		fileSchemas := parseStructsFromSource(string(content), dialect)
		for name, schema := range fileSchemas {
			schemas[name] = schema
		}
	}

	return schemas, nil
}

// parseStructsFromSource does a simplified parse of Go source to extract struct schemas.
// This is intentionally simple — for full accuracy, use 'genus generate' + AutoMigrate.
func parseStructsFromSource(source string, dialect core.Dialect) map[string]*migrate.TableSchema {
	schemas := make(map[string]*migrate.TableSchema)

	lines := strings.Split(source, "\n")
	var currentStruct string
	var columns []migrate.ColumnSchema
	inStruct := false
	hasModel := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect struct declaration
		if strings.Contains(trimmed, "type ") && strings.Contains(trimmed, " struct {") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				currentStruct = parts[1]
				inStruct = true
				hasModel = false
				columns = nil
			}
			continue
		}

		// End of struct
		if inStruct && trimmed == "}" {
			if hasModel && currentStruct != "" {
				tableName := structToTableName(currentStruct)
				// Add model fields
				modelColumns := []migrate.ColumnSchema{
					{Name: "id", Type: dialect.GetType("int64"), PrimaryKey: true, AutoIncrement: true},
					{Name: "created_at", Type: dialect.GetType("time.Time"), Nullable: true},
					{Name: "updated_at", Type: dialect.GetType("time.Time"), Nullable: true},
				}
				allColumns := append(modelColumns, columns...)
				schemas[tableName] = &migrate.TableSchema{
					Name:    tableName,
					Columns: allColumns,
				}
			}
			inStruct = false
			currentStruct = ""
			continue
		}

		if !inStruct {
			continue
		}

		// Detect embedded core.Model
		if strings.Contains(trimmed, "core.Model") {
			hasModel = true
			continue
		}

		// Parse db tag
		if !strings.Contains(trimmed, "`") || !strings.Contains(trimmed, "db:\"") {
			continue
		}

		// Extract db tag value
		dbTag := extractDBTag(trimmed)
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Extract Go type
		goType := extractGoType(trimmed)
		if goType == "" {
			continue
		}

		sqlType := dialect.GetType(goType)
		nullable := strings.Contains(goType, "*") || strings.Contains(goType, "Optional")

		columns = append(columns, migrate.ColumnSchema{
			Name:     dbTag,
			Type:     sqlType,
			Nullable: nullable,
		})
	}

	return schemas
}

func extractDBTag(line string) string {
	idx := strings.Index(line, "db:\"")
	if idx < 0 {
		return ""
	}
	rest := line[idx+4:]
	endIdx := strings.Index(rest, "\"")
	if endIdx < 0 {
		return ""
	}
	tag := rest[:endIdx]
	// Handle options like db:"name,omitempty"
	if comma := strings.Index(tag, ","); comma >= 0 {
		tag = tag[:comma]
	}
	return tag
}

func extractGoType(line string) string {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return ""
	}
	// Type is usually the second field (after field name)
	goType := fields[1]
	// Skip if it looks like a tag
	if strings.HasPrefix(goType, "`") {
		return ""
	}
	return goType
}

func structToTableName(name string) string {
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func printSchemaUsage() {
	fmt.Println(`Schema tools for Genus ORM

Usage:
  genus schema <subcommand> [arguments]

Subcommands:
  diff      Compare Go models with database schema
  help      Show this help message

Run 'genus schema <subcommand> --help' for more information.`)
}

func printSchemaDiffUsage() {
	fmt.Println(`Compare Go models with the current database schema

Usage:
  genus schema diff [flags]

Flags:
  -m, --models <dir>     Models directory (default: "./models")
  -g, --generate         Generate migration file from diff
  -h, --help             Show this help message

Environment Variables:
  DATABASE_URL    Database connection string

Examples:
  genus schema diff                          # Show diff for ./models
  genus schema diff -m ./internal/models     # Custom models directory
  genus schema diff --generate               # Generate migration from diff

The diff will show:
  + Added tables/columns (green)
  - Removed tables/columns (red)
  ~ Modified columns (yellow)`)
}
