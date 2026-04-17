package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-genus/genus/codegen"
)

const version = "7.0.0"

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// commands maps command names to their descriptions for help and suggestions.
var commands = map[string]string{
	"generate":          "Generate type-safe field definitions from Go structs",
	"generate-scanners": "Generate optimized, reflection-free scanners",
	"init":              "Initialize a new Genus project with scaffolding",
	"migrate":           "Manage database migrations (up, down, status, create, visualize)",
	"schema":            "Schema tools (diff, validate)",
	"repl":              "Interactive query builder REPL",
	"playground":        "Start web-based query playground",
	"version":           "Print version information",
	"help":              "Show this help message",
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		if err := runGenerate(); err != nil {
			fmt.Fprintf(os.Stderr, "%sError:%s %v\n", colorBold, colorReset, err)
			os.Exit(1)
		}
	case "generate-scanners":
		if err := runGenerateScanners(); err != nil {
			fmt.Fprintf(os.Stderr, "%sError:%s %v\n", colorBold, colorReset, err)
			os.Exit(1)
		}
	case "init":
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "%sError:%s %v\n", colorBold, colorReset, err)
			os.Exit(1)
		}
	case "migrate":
		if err := runMigrate(); err != nil {
			fmt.Fprintf(os.Stderr, "%sError:%s %v\n", colorBold, colorReset, err)
			os.Exit(1)
		}
	case "schema":
		if err := runSchema(); err != nil {
			fmt.Fprintf(os.Stderr, "%sError:%s %v\n", colorBold, colorReset, err)
			os.Exit(1)
		}
	case "repl":
		if err := runREPL(); err != nil {
			fmt.Fprintf(os.Stderr, "%sError:%s %v\n", colorBold, colorReset, err)
			os.Exit(1)
		}
	case "playground":
		if err := runPlayground(); err != nil {
			fmt.Fprintf(os.Stderr, "%sError:%s %v\n", colorBold, colorReset, err)
			os.Exit(1)
		}
	case "version", "--version", "-v":
		fmt.Printf("genus version %s%s%s\n", colorCyan, version, colorReset)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "%sUnknown command: %s%s\n", colorBold, command, colorReset)
		if suggestion := suggestCommand(command); suggestion != "" {
			fmt.Fprintf(os.Stderr, "\nDid you mean %s%s%s?\n", colorGreen, suggestion, colorReset)
		}
		fmt.Fprintf(os.Stderr, "\nRun '%sgenus help%s' for usage.\n", colorCyan, colorReset)
		os.Exit(1)
	}
}

func runGenerate() error {
	args := os.Args[2:]

	// Parse flags
	var (
		outputDir = "."
		pkgName   = ""
		paths     []string
	)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-o" || arg == "--output" {
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", arg)
			}
			outputDir = args[i+1]
			i++
		} else if arg == "-p" || arg == "--package" {
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", arg)
			}
			pkgName = args[i+1]
			i++
		} else if arg == "-h" || arg == "--help" {
			printGenerateUsage()
			return nil
		} else {
			paths = append(paths, arg)
		}
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}

	generator := codegen.NewGenerator(codegen.Config{
		OutputDir:   outputDir,
		PackageName: pkgName,
	})

	fmt.Printf("%sGenerating field definitions...%s\n", colorCyan, colorReset)

	for _, path := range paths {
		fmt.Printf("  Processing: %s\n", path)
		if err := generator.GenerateFromPath(path); err != nil {
			return fmt.Errorf("failed to generate from %s: %w", path, err)
		}
	}

	fmt.Printf("\n%s[OK]%s Code generation completed successfully!\n", colorGreen, colorReset)
	return nil
}

func runGenerateScanners() error {
	args := os.Args[2:]

	if len(args) == 0 {
		args = []string{"."}
	}

	for _, path := range args {
		if path == "-h" || path == "--help" {
			printGenerateScannersUsage()
			return nil
		}

		fmt.Printf("%sGenerating scanners for:%s %s\n", colorCyan, colorReset, path)
		if err := codegen.GenerateScannersForDir(path); err != nil {
			return fmt.Errorf("failed to generate scanners: %w", err)
		}
	}

	fmt.Printf("\n%s[OK]%s Scanner generation completed!\n", colorGreen, colorReset)
	return nil
}

func printUsage() {
	fmt.Printf(`
%s%sGenus%s — Type-safe ORM for Go

%sUsage:%s
  genus <command> [arguments]

%sCode Generation:%s
  %sgenerate%s            Generate type-safe field definitions from Go structs
  %sgenerate-scanners%s   Generate optimized, reflection-free scanners

%sDatabase:%s
  %smigrate%s             Manage migrations (up, down, status, create, visualize)
  %sschema%s              Schema tools (diff)
  %srepl%s                Interactive query builder REPL
  %splayground%s          Start web-based query playground

%sProject:%s
  %sinit%s                Initialize a new Genus project with scaffolding

%sOther:%s
  %sversion%s             Print version information
  %shelp%s                Show this help message

Run '%sgenus <command> --help%s' for more information on a command.
`,
		colorBold, colorCyan, colorReset,
		colorBold, colorReset,
		colorYellow, colorReset,
		colorGreen, colorReset,
		colorGreen, colorReset,
		colorYellow, colorReset,
		colorGreen, colorReset,
		colorGreen, colorReset,
		colorGreen, colorReset,
		colorGreen, colorReset,
		colorYellow, colorReset,
		colorGreen, colorReset,
		colorYellow, colorReset,
		colorGreen, colorReset,
		colorGreen, colorReset,
		colorCyan, colorReset,
	)
}

func printGenerateUsage() {
	fmt.Println(`Generate type-safe field definitions from Go structs

Usage:
  genus generate [flags] [paths...]

Flags:
  -o, --output <dir>     Output directory for generated files (default: ".")
  -p, --package <name>   Package name for generated code (default: auto-detect)
  -h, --help             Show this help message

Examples:
  genus generate                    # Generate from current directory
  genus generate ./models           # Generate from models directory
  genus generate -o ./generated     # Generate to specific output directory
  genus generate -p mypackage ./models  # Generate with custom package name

The generator will:
1. Scan Go files for structs with 'db' tags
2. Generate type-safe field definitions (e.g., UserFields, ProductFields)
3. Save generated code to *_fields.gen.go files`)
}

func printGenerateScannersUsage() {
	fmt.Println(`Generate optimized, reflection-free scanners from Go structs

Usage:
  genus generate-scanners [paths...]

Examples:
  genus generate-scanners                # Generate from current directory
  genus generate-scanners ./models       # Generate from models directory

The generator will:
1. Scan Go files for structs with 'db' tags or embedded core.Model
2. Generate optimized ScanXxx() functions (10x faster than reflection)
3. Save generated code to scanners.gen.go

Generated functions:
  - ScanUser(rows) (User, error)           // Scan single row
  - ScanUsers(rows) ([]User, error)        // Scan all rows
  - ScanUsersWithCap(rows, cap)            // Scan with pre-allocated capacity
  - UserColumns() []string                 // Column names in order
  - UserColumnsString() string             // "id, name, email, ..."`)
}

// suggestCommand finds the closest matching command using Levenshtein distance.
func suggestCommand(input string) string {
	input = strings.ToLower(input)
	bestMatch := ""
	bestDist := 4 // max distance threshold

	for cmd := range commands {
		d := levenshtein(input, cmd)
		if d < bestDist {
			bestDist = d
			bestMatch = cmd
		}
		// Also check prefix match
		if strings.HasPrefix(cmd, input) && len(bestMatch) == 0 {
			bestMatch = cmd
		}
	}

	return bestMatch
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}
