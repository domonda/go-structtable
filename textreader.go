package structtable

import (
	"reflect"

	"github.com/domonda/go-types/assign"
	"github.com/domonda/go-wraperr"
)

type TextReader struct {
	rows           [][]string
	columnMapping  map[int]string
	columnTitleTag string
	config         *assign.StringConfig
}

func NewTextReader(rows [][]string, columnMapping map[int]string, columnTitleTag string, config *assign.StringConfig) *TextReader {
	return &TextReader{
		rows:           rows,
		columnMapping:  columnMapping,
		columnTitleTag: columnTitleTag,
		config:         config,
	}
}

func (tr *TextReader) NumRows() int {
	return len(tr.rows)
}

func (tr *TextReader) ReadRow(index int, destStruct reflect.Value) error {
	if index < 0 || index >= len(tr.rows) {
		return wraperr.Errorf("row index %d out of range [0..%d)", index, len(tr.rows))
	}
	row := tr.rows[index]

	for col, name := range tr.columnMapping {
		if col < 0 || col >= len(row) {
			return wraperr.Errorf("row %d column index %d out of range [0..%d)", index, col, len(row))
		}

		// Find struct field with name
		var destVal reflect.Value
		for i := 0; i < destStruct.NumField(); i++ {
			fieldType := destStruct.Type().Field(i)
			fieldName := fieldType.Name
			if tag := fieldType.Tag.Get(tr.columnTitleTag); tag != "" {
				fieldName = tag
			}
			if fieldName == name {
				destVal = destStruct.Field(i)
				break
			}
		}
		if !destVal.IsValid() {
			return wraperr.Errorf("no struct field %q found in %s using tag %q", name, destStruct.Type(), tr.columnTitleTag)
		}

		err := assign.String(row[col], tr.config, destVal)
		if err != nil {
			return wraperr.Errorf("error reading row %d, column %d: %w", index, col, err)
		}
	}

	return nil
}
