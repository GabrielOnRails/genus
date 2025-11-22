package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// fieldPath representa o caminho para um campo através de embedded structs
type fieldPath []int

// scanStruct faz o scan de uma row para uma struct.
// Esta é uma das poucas funções que usa reflection, mas é controlada e isolada.
func scanStruct(rows *sql.Rows, dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Mapeia os nomes das colunas para os caminhos dos campos da struct
	fieldMap := buildFieldMap(destValue.Type())

	// Cria os ponteiros para os valores a serem escaneados
	scanValues := make([]interface{}, len(columns))
	for i, colName := range columns {
		if path, ok := fieldMap[colName]; ok {
			field := getFieldByPath(destValue, path)
			if field.IsValid() && field.CanAddr() {
				scanValues[i] = field.Addr().Interface()
			} else {
				// Se o campo não é válido, usa um placeholder
				var placeholder interface{}
				scanValues[i] = &placeholder
			}
		} else {
			// Se a coluna não existe na struct, usa um placeholder
			var placeholder interface{}
			scanValues[i] = &placeholder
		}
	}

	return rows.Scan(scanValues...)
}

// getFieldByPath navega até um campo usando um caminho de índices
func getFieldByPath(value reflect.Value, path fieldPath) reflect.Value {
	current := value
	for _, index := range path {
		if current.Kind() == reflect.Ptr {
			current = current.Elem()
		}
		if index >= current.NumField() {
			return reflect.Value{}
		}
		current = current.Field(index)
	}
	return current
}

// buildFieldMap constrói um mapa de nome de coluna -> caminho do campo.
func buildFieldMap(typ reflect.Type) map[string]fieldPath {
	return buildFieldMapWithPrefix(typ, nil)
}

// buildFieldMapWithPrefix constrói um mapa de nome de coluna -> caminho do campo com suporte para embedded structs.
func buildFieldMapWithPrefix(typ reflect.Type, parentPath fieldPath) map[string]fieldPath {
	fieldMap := make(map[string]fieldPath)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		currentPath := append(fieldPath(nil), parentPath...)
		currentPath = append(currentPath, i)

		// Se o campo é um embedded struct, processa recursivamente
		if field.Anonymous {
			embeddedMap := buildFieldMapWithPrefix(field.Type, currentPath)

			// Mescla o mapa do embedded struct com o mapa atual
			for k, v := range embeddedMap {
				fieldMap[k] = v
			}
			continue
		}

		// Obtém o nome da coluna da tag `db`
		colName := field.Tag.Get("db")
		if colName == "" || colName == "-" {
			// Se não tem tag ou é "-", pula o campo
			continue
		}

		fieldMap[colName] = currentPath
	}

	return fieldMap
}

// toSnakeCase converte CamelCase para snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// GetFieldIndices retorna os índices dos campos de uma struct para scanning.
// Usado internamente pelo scanner.
func GetFieldIndices(dest interface{}) ([]interface{}, error) {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("dest must be a pointer")
	}

	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("dest must be a pointer to struct")
	}

	var indices []interface{}
	for i := 0; i < destValue.NumField(); i++ {
		field := destValue.Field(i)
		if field.CanAddr() {
			indices = append(indices, field.Addr().Interface())
		}
	}

	return indices, nil
}
