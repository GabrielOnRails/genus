package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

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

	// Mapeia os nomes das colunas para os campos da struct
	fieldMap := buildFieldMap(destValue.Type())

	// Cria os ponteiros para os valores a serem escaneados
	scanValues := make([]interface{}, len(columns))
	for i, colName := range columns {
		if fieldIndex, ok := fieldMap[colName]; ok {
			field := destValue.Field(fieldIndex)
			scanValues[i] = field.Addr().Interface()
		} else {
			// Se a coluna não existe na struct, usa um placeholder
			var placeholder interface{}
			scanValues[i] = &placeholder
		}
	}

	return rows.Scan(scanValues...)
}

// buildFieldMap constrói um mapa de nome de coluna -> índice do campo.
func buildFieldMap(typ reflect.Type) map[string]int {
	fieldMap := make(map[string]int)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Se o campo é um embedded struct, processa recursivamente
		if field.Anonymous {
			embeddedMap := buildFieldMap(field.Type)
			for k := range embeddedMap {
				fieldMap[k] = i // Nota: simplificado, não lida com deep nesting
			}
			continue
		}

		// Obtém o nome da coluna da tag `db`
		colName := field.Tag.Get("db")
		if colName == "" {
			// Se não tem tag, usa o nome do campo em snake_case
			colName = toSnakeCase(field.Name)
		}

		fieldMap[colName] = i
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
