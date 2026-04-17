package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runInit() error {
	args := os.Args[2:]

	var projectDir = "."

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			printInitUsage()
			return nil
		default:
			projectDir = arg
		}
	}

	// Resolve to absolute path
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	fmt.Printf("\n%sInitializing Genus project in %s%s\n\n", colorCyan, absDir, colorReset)

	// Create directory structure
	dirs := []string{
		"models",
		"migrations",
	}

	for _, dir := range dirs {
		path := filepath.Join(absDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Printf("  %s[+]%s %s/\n", colorGreen, colorReset, dir)
	}

	// Create models/user.go example
	modelsPath := filepath.Join(absDir, "models", "user.go")
	if !fileExists(modelsPath) {
		if err := os.WriteFile(modelsPath, []byte(userModelTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create models/user.go: %w", err)
		}
		fmt.Printf("  %s[+]%s models/user.go\n", colorGreen, colorReset)
	} else {
		fmt.Printf("  %s[~]%s models/user.go (already exists, skipping)\n", colorYellow, colorReset)
	}

	// Create main.go example
	mainPath := filepath.Join(absDir, "main.go")
	if !fileExists(mainPath) {
		moduleName := detectModuleName(absDir)
		content := strings.ReplaceAll(mainGoTemplate, "{{MODULE}}", moduleName)
		if err := os.WriteFile(mainPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create main.go: %w", err)
		}
		fmt.Printf("  %s[+]%s main.go\n", colorGreen, colorReset)
	} else {
		fmt.Printf("  %s[~]%s main.go (already exists, skipping)\n", colorYellow, colorReset)
	}

	fmt.Printf("\n%s[OK]%s Project initialized!\n", colorGreen, colorReset)
	fmt.Printf(`
%sNext steps:%s
  1. Edit models/user.go to define your models
  2. Run %sgenus generate ./models%s to generate type-safe fields
  3. Run %sgenus generate-scanners ./models%s for optimized scanners
  4. Set %sDATABASE_URL%s and run %sgenus migrate up%s

`, colorBold, colorReset,
		colorCyan, colorReset,
		colorCyan, colorReset,
		colorCyan, colorReset,
		colorCyan, colorReset,
	)

	return nil
}

// detectModuleName tries to read the module name from go.mod.
func detectModuleName(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "myapp"
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module ")
		}
	}

	return "myapp"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func printInitUsage() {
	fmt.Println(`Initialize a new Genus project with scaffolding

Usage:
  genus init [directory]

Arguments:
  directory    Target directory (default: current directory)

Examples:
  genus init                # Initialize in current directory
  genus init ./myproject    # Initialize in myproject/

Creates:
  models/user.go       Example model with struct tags
  migrations/           Directory for migration files
  main.go              Example application with Genus setup`)
}

const userModelTemplate = `package models

import (
	"github.com/go-genus/genus/core"
)

// User is an example model with type-safe field mappings.
// Run 'genus generate ./models' to generate UserFields.
type User struct {
	core.Model
	Name  string ` + "`db:\"name\"`" + `
	Email string ` + "`db:\"email\"`" + `
	Age   int    ` + "`db:\"age\"`" + `
}

// TableName implements core.TableNamer for custom table name.
func (User) TableName() string {
	return "users"
}
`

const mainGoTemplate = `package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-genus/genus"
	_ "github.com/lib/pq" // PostgreSQL driver

	"{{MODULE}}/models"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable"
	}

	db, err := genus.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create a user
	user := &models.User{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}

	if err := db.Create(ctx, user); err != nil {
		log.Fatal(err)
	}

	// Query users with type-safe builder
	users, err := genus.Table[models.User](db).Find(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		fmt.Printf("User: %s (%s)\n", u.Name, u.Email)
	}
}
`
