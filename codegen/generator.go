package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Config contém as configurações para o gerador de código.
type Config struct {
	// OutputDir é o diretório onde os arquivos gerados serão salvos
	OutputDir string
	// PackageName é o nome do pacote para o código gerado (auto-detectado se vazio)
	PackageName string
}

// Generator gera código de campos tipados a partir de structs Go.
type Generator struct {
	config Config
}

// NewGenerator cria um novo gerador.
func NewGenerator(config Config) *Generator {
	return &Generator{config: config}
}

// FieldInfo contém informações sobre um campo de struct.
type FieldInfo struct {
	Name       string // Nome do campo Go (ex: "Name")
	ColumnName string // Nome da coluna no DB (ex: "name")
	Type       string // Tipo Go (ex: "string", "int64", "core.Optional[string]")
	FieldType  string // Tipo do campo query (ex: "StringField", "OptionalStringField")
}

// StructInfo contém informações sobre uma struct.
type StructInfo struct {
	Name       string      // Nome da struct (ex: "User")
	Fields     []FieldInfo // Campos da struct
	Package    string      // Nome do pacote
	ImportCore bool        // Se precisa importar o pacote core
}

// GenerateFromPath gera código a partir de um caminho (arquivo ou diretório).
func (g *Generator) GenerateFromPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return g.generateFromDir(path)
	}

	return g.generateFromFile(path)
}

// generateFromDir gera código para todos os arquivos Go em um diretório.
func (g *Generator) generateFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		// Ignora arquivos gerados e de teste
		if strings.HasSuffix(entry.Name(), "_test.go") || strings.HasSuffix(entry.Name(), ".gen.go") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if err := g.generateFromFile(path); err != nil {
			return err
		}
	}

	return nil
}

// generateFromFile gera código a partir de um arquivo Go.
func (g *Generator) generateFromFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	structs := g.extractStructs(node)
	if len(structs) == 0 {
		return nil
	}

	// Detecta o nome do pacote se não foi fornecido
	pkgName := g.config.PackageName
	if pkgName == "" {
		pkgName = node.Name.Name
	}

	// Gera código para cada struct
	for _, structInfo := range structs {
		structInfo.Package = pkgName

		if err := g.generateFieldsFile(structInfo, filename); err != nil {
			return err
		}
	}

	return nil
}

// extractStructs extrai informações de structs do AST.
func (g *Generator) extractStructs(node *ast.File) []StructInfo {
	var structs []StructInfo

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structInfo := StructInfo{
			Name:   typeSpec.Name.Name,
			Fields: []FieldInfo{},
		}

		for _, field := range structType.Fields.List {
			// Ignora campos sem tag
			if field.Tag == nil {
				continue
			}

			tag := field.Tag.Value
			if !strings.Contains(tag, "db:") {
				continue
			}

			// Extrai o nome da coluna da tag db
			columnName := extractDBTag(tag)
			if columnName == "" || columnName == "-" {
				continue
			}

			// Extrai informações do campo
			for _, name := range field.Names {
				fieldType := g.getFieldType(field.Type)
				queryFieldType := g.getQueryFieldType(fieldType)

				fieldInfo := FieldInfo{
					Name:       name.Name,
					ColumnName: columnName,
					Type:       fieldType,
					FieldType:  queryFieldType,
				}

				structInfo.Fields = append(structInfo.Fields, fieldInfo)

				// Marca se precisa importar core
				if strings.Contains(fieldType, "Optional") {
					structInfo.ImportCore = true
				}
			}
		}

		// Só adiciona struct se tiver campos
		if len(structInfo.Fields) > 0 {
			structs = append(structs, structInfo)
		}

		return true
	})

	return structs
}

// getFieldType extrai o tipo Go de um campo.
func (g *Generator) getFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", g.getFieldType(t.X), t.Sel.Name)
	case *ast.IndexExpr:
		// Generic type (ex: Optional[string])
		return fmt.Sprintf("%s[%s]", g.getFieldType(t.X), g.getFieldType(t.Index))
	case *ast.StarExpr:
		return "*" + g.getFieldType(t.X)
	default:
		return "unknown"
	}
}

// getQueryFieldType mapeia tipos Go para tipos de campo query.
func (g *Generator) getQueryFieldType(goType string) string {
	// Remove ponteiros
	goType = strings.TrimPrefix(goType, "*")

	// Mapeia tipos primitivos
	switch goType {
	case "string":
		return "query.StringField"
	case "int":
		return "query.IntField"
	case "int64":
		return "query.Int64Field"
	case "bool":
		return "query.BoolField"
	case "float64":
		return "query.Float64Field"
	}

	// Mapeia tipos Optional
	if strings.HasPrefix(goType, "Optional[") || strings.HasPrefix(goType, "core.Optional[") {
		innerType := extractGenericType(goType)
		switch innerType {
		case "string":
			return "query.OptionalStringField"
		case "int":
			return "query.OptionalIntField"
		case "int64":
			return "query.OptionalInt64Field"
		case "bool":
			return "query.OptionalBoolField"
		case "float64":
			return "query.OptionalFloat64Field"
		}
	}

	// Fallback para StringField
	return "query.StringField"
}

// generateFieldsFile gera o arquivo de campos para uma struct.
func (g *Generator) generateFieldsFile(structInfo StructInfo, sourceFile string) error {
	// Gera o código
	code, err := g.generateCode(structInfo)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Determina o nome do arquivo de saída
	outputDir := g.config.OutputDir
	if outputDir == "" {
		outputDir = filepath.Dir(sourceFile)
	}

	baseName := strings.TrimSuffix(filepath.Base(sourceFile), ".go")
	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_fields.gen.go", baseName))

	// Escreve o arquivo
	if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("  Generated: %s\n", outputFile)
	return nil
}

// generateCode gera o código Go para os campos.
func (g *Generator) generateCode(structInfo StructInfo) (string, error) {
	// Cria o template com a função helper
	tmpl := template.New("fields").Funcs(template.FuncMap{
		"fieldConstructor": fieldConstructor,
	})

	tmpl, err := tmpl.Parse(fieldsTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, structInfo); err != nil {
		return "", err
	}

	// Formata o código gerado
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Se houver erro na formatação, retorna o código não formatado
		// para facilitar debugging
		return buf.String(), nil
	}

	return string(formatted), nil
}

// extractDBTag extrai o valor da tag db.
func extractDBTag(tag string) string {
	// Remove as aspas da tag
	tag = strings.Trim(tag, "`")

	// Procura por db:"value"
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "db:") {
			value := strings.TrimPrefix(part, "db:")
			value = strings.Trim(value, `"`)
			return value
		}
	}

	return ""
}

// extractGenericType extrai o tipo interno de um tipo genérico.
// Ex: "Optional[string]" -> "string"
func extractGenericType(genericType string) string {
	start := strings.Index(genericType, "[")
	end := strings.Index(genericType, "]")

	if start == -1 || end == -1 || start >= end {
		return ""
	}

	return genericType[start+1 : end]
}
