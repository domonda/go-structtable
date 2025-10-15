package structtable

import (
	"reflect"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/strfmt"
)

// TextReader implements the Reader interface for reading tabular data from
// a 2D slice of strings into struct instances.
//
// This reader is useful when you have pre-parsed tabular data as strings
// and want to populate struct instances with the data. It supports column
// mapping and custom scanning configuration.
type TextReader struct {
	rows           [][]string
	columnMapping  map[int]string
	columnTitleTag string
	scanConfig     *strfmt.ScanConfig
}

// NewTextReader creates a new TextReader instance.
//
// This constructor initializes a TextReader with the provided data and configuration.
// The columnMapping maps column indices to struct field names, and the columnTitleTag
// specifies which struct tag to use for field name resolution.
//
// Parameters:
//   - rows: The 2D slice of strings containing the tabular data
//   - columnMapping: Map from column index to struct field name
//   - columnTitleTag: The struct tag to use for field name resolution
//   - scanConfig: Optional scanning configuration (uses default if nil)
//
// Returns:
//   - A new TextReader instance ready for use
func NewTextReader(rows [][]string, columnMapping map[int]string, columnTitleTag string, scanConfig ...*strfmt.ScanConfig) *TextReader {
	tr := &TextReader{
		rows:           rows,
		columnMapping:  columnMapping,
		columnTitleTag: columnTitleTag,
		scanConfig:     strfmt.DefaultScanConfig,
	}
	if len(scanConfig) > 0 && scanConfig[0] != nil {
		tr.scanConfig = scanConfig[0]
	}
	return tr
}

// NumRows returns the total number of rows available for reading.
//
// This method implements the Reader interface and returns the count
// of rows in the underlying data.
func (tr *TextReader) NumRows() int {
	return len(tr.rows)
}

// ReadRow populates a struct instance with data from the specified row.
//
// This method implements the Reader interface and populates the destStruct
// with data from the row at the given index. It uses the columnMapping to
// determine which columns correspond to which struct fields, and uses the
// columnTitleTag to resolve field names from struct tags.
//
// Parameters:
//   - index: The row index to read (0-based)
//   - destStruct: The reflect.Value of the struct to populate
//
// Returns:
//   - err: Any error that occurred during reading or field population
func (tr *TextReader) ReadRow(index int, destStruct reflect.Value) error {
	if index < 0 || index >= len(tr.rows) {
		return errs.Errorf("row index %d out of range [0..%d)", index, len(tr.rows))
	}
	row := tr.rows[index]

	for col, name := range tr.columnMapping {
		if col < 0 || col >= len(row) {
			return errs.Errorf("row %d column index %d out of range [0..%d)", index, col, len(row))
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
			return errs.Errorf("no struct field %q found in %s using tag %q", name, destStruct.Type(), tr.columnTitleTag)
		}

		err := strfmt.Scan(destVal, row[col], tr.scanConfig)
		if err != nil {
			return errs.Errorf("error reading row %d, column %d: %w", index, col, err)
		}
	}

	return nil
}
