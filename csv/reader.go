package csv

import (
	"io"
	"reflect"

	"github.com/ungerik/go-fs"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/strfmt"
)

// type TableDetectionConfig struct {
// 	Format  *FormatDetectionConfig
// 	Columns []TableDetectionConfigColumn
// }

// type TableDetectionConfigColumn struct {
// 	StructField string
// 	HeaderNames []string
// }

// ColumnMapping represents the mapping between a CSV column and a struct field.
//
// This struct defines how a specific column in the CSV data should be mapped
// to a field in the target struct.
type ColumnMapping struct {
	// Index is the zero-based index of the column in the CSV data.
	Index int
	// StructField is the name of the struct field to populate.
	StructField string
}

// Reader implements the structtable.Reader interface for CSV data.
//
// This reader can parse CSV files and populate struct instances with the data.
// It supports format detection, data modification, and flexible column mapping.
type Reader struct {
	// Format contains the CSV format configuration.
	Format *Format `json:"format,omitempty"`
	// FormatDetection contains configuration for automatic format detection.
	FormatDetection *FormatDetectionConfig `json:"formatDetection,omitempty"`
	// ScanConfig contains configuration for parsing string values into Go types.
	ScanConfig *strfmt.ScanConfig `json:"config"`
	// Modifiers contains a list of data modification functions to apply.
	Modifiers ModifierList `json:"modifiers"`
	// Columns defines the mapping between CSV columns and struct fields.
	Columns []ColumnMapping `json:"columns"`

	rows [][]string
}

// NewReader creates a new CSV Reader from an io.Reader.
//
// This constructor reads CSV data from the provided io.Reader and creates
// a Reader instance ready for use. It parses the data according to the
// specified format and applies any modifiers.
//
// Parameters:
//   - reader: The io.Reader containing CSV data
//   - format: The CSV format configuration
//   - newlineReplacement: String to replace newlines in quoted fields
//   - modifiers: List of data modification functions to apply
//   - columns: Mapping between CSV columns and struct fields
//   - scanConfig: Optional scanning configuration (uses default if nil)
//
// Returns:
//   - r: A new Reader instance ready for use
//   - err: Any error that occurred during parsing
func NewReader(reader io.Reader, format *Format, newlineReplacement string, modifiers ModifierList, columns []ColumnMapping, scanConfig ...*strfmt.ScanConfig) (r *Reader, err error) {
	defer errs.WrapWithFuncParams(&err, reader, format, newlineReplacement, modifiers, columns, scanConfig)

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	rows, err := ParseWithFormat(data, format)
	if err != nil {
		return nil, err
	}

	return NewReaderFromRows(rows, format, newlineReplacement, modifiers, columns, scanConfig...)
}

// NewReaderFromRows creates a Reader that uses pre-parsed CSV rows.
//
// This constructor creates a Reader instance from already parsed CSV data,
// allowing you to reuse parsed data or work with data from other sources.
// The modifiers are applied to the provided rows during construction.
//
// Parameters:
//   - rows: Pre-parsed CSV data as a 2D slice of strings
//   - format: The CSV format configuration (for reference)
//   - newlineReplacement: String to replace newlines in quoted fields (unused in this constructor)
//   - modifiers: List of data modification functions to apply to the rows
//   - columns: Mapping between CSV columns and struct fields
//   - scanConfig: Optional scanning configuration (uses default if nil)
//
// Returns:
//   - r: A new Reader instance ready for use
//   - err: Any error that occurred during construction
//
// Example:
//
//	rows := [][]string{
//	    {"Name", "Age", "City"},
//	    {"John", "25", "New York"},
//	}
//	reader, err := csv.NewReaderFromRows(rows, format, "", modifiers, columns)
func NewReaderFromRows(rows [][]string, format *Format, newlineReplacement string, modifiers ModifierList, columns []ColumnMapping, scanConfig ...*strfmt.ScanConfig) (r *Reader, err error) {
	defer errs.WrapWithFuncParams(&err, rows, format, newlineReplacement, modifiers, columns, scanConfig)

	r = &Reader{
		Format:     format,
		ScanConfig: strfmt.DefaultScanConfig,
		Modifiers:  modifiers,
		Columns:    columns,
		rows:       modifiers.Modify(rows),
	}
	if len(scanConfig) > 0 && scanConfig[0] != nil {
		r.ScanConfig = scanConfig[0]
	}
	return r, nil
}

// NewReaderFromFile creates a CSV Reader from a file using fs.FileReader interface.
//
// This constructor reads CSV data from a file and creates a Reader instance ready for use.
// It handles file opening, reading, and closing automatically. The data is parsed according
// to the specified format and modifiers are applied.
//
// Parameters:
//   - file: The file reader containing CSV data
//   - format: The CSV format configuration
//   - newlineReplacement: String to replace newlines in quoted fields
//   - modifiers: List of data modification functions to apply
//   - columns: Mapping between CSV columns and struct fields
//   - scanConfig: Optional scanning configuration (uses default if nil)
//
// Returns:
//   - r: A new Reader instance ready for use
//   - err: Any error that occurred during file operations or parsing
//
// Example:
//
//	file := fs.NewFile("data.csv")
//	reader, err := csv.NewReaderFromFile(file, format, "\n", modifiers, columns)
func NewReaderFromFile(file fs.FileReader, format *Format, newlineReplacement string, modifiers ModifierList, columns []ColumnMapping, scanConfig ...*strfmt.ScanConfig) (r *Reader, err error) {
	defer errs.WrapWithFuncParams(&err, file, format, newlineReplacement, modifiers, columns, scanConfig)

	reader, err := file.OpenReader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return NewReader(reader, format, newlineReplacement, modifiers, columns, scanConfig...)
}

// NumRows returns the total number of rows in the CSV data.
//
// This method returns the number of rows that were parsed from the CSV data
// and are available for reading. This count includes all rows after modifiers
// have been applied (e.g., empty rows may have been removed).
//
// Returns:
//   - int: The total number of rows available for reading
//
// Example:
//
//	reader, err := csv.NewReader(file, format, "\n", modifiers, columns)
//	if err != nil {
//	    return err
//	}
//	totalRows := reader.NumRows()
//	fmt.Printf("CSV has %d rows\n", totalRows)
func (r *Reader) NumRows() int {
	return len(r.rows)
}

// ReadRowStrings returns raw string values for a specific row in the CSV data.
//
// This method reads all fields from the specified row and returns them as a slice
// of strings. This is useful when you need access to the raw CSV data without
// struct mapping or when debugging the parsing process.
//
// Parameters:
//   - index: The zero-based index of the row to read
//
// Returns:
//   - []string: Slice of string values from the row (one per field)
//   - err: Any error that occurred during bounds checking
//
// Bounds Checking:
//   - Returns error if index is negative or >= NumRows()
//   - Returns the actual row data as parsed (after modifiers applied)
//
// Example:
//
//	rowData, err := reader.ReadRowStrings(0) // Read first row
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Row contains %d fields: %v\n", len(rowData), rowData)
func (r *Reader) ReadRowStrings(index int) ([]string, error) {
	if index < 0 || index > len(r.rows) {
		return nil, errs.Errorf("row index %d out of bounds [0..%d)", index, len(r.rows))
	}
	return r.rows[index], nil
}

// ReadRow populates a struct instance with data from the specified CSV row.
//
// This method reads data from the specified row and populates the fields of the
// provided struct instance using the configured column mapping. It uses the
// strfmt.Scan function to convert string values to appropriate Go types based
// on the struct field types and scan configuration.
//
// Column Mapping:
//   - Uses the Columns field to map CSV column indices to struct field names
//   - Skips columns with invalid indices or non-existent struct fields
//   - Applies type conversion using strfmt.Scan with the configured ScanConfig
//
// Parameters:
//   - index: The zero-based index of the row to read
//   - destStruct: A reflect.Value pointing to the struct instance to populate
//
// Returns:
//   - err: Any error that occurred during reading, bounds checking, or type conversion
//
// Bounds Checking:
//   - Returns error if index is negative or >= NumRows()
//   - Skips columns with invalid indices (negative or >= row length)
//   - Skips non-existent struct fields gracefully
//
// Type Conversion:
//   - Uses strfmt.Scan for intelligent type conversion
//   - Supports various Go types (int, float, bool, time, etc.)
//   - Respects the configured ScanConfig for parsing rules
//
// Example:
//
//	type Person struct {
//	    Name string
//	    Age  int
//	    City string
//	}
//	var person Person
//	err := reader.ReadRow(0, reflect.ValueOf(&person).Elem())
func (r *Reader) ReadRow(index int, destStruct reflect.Value) error {
	if index < 0 || index >= len(r.rows) {
		return errs.Errorf("row index %d out of bounds [0..%d)", index, len(r.rows))
	}

	row := r.rows[index]
	for _, col := range r.Columns {
		if col.Index < 0 || col.Index >= len(row) {
			continue
		}
		destStructField := destStruct.FieldByName(col.StructField)
		if !destStructField.IsValid() {
			continue
		}
		err := strfmt.Scan(destStructField, row[col.Index], r.ScanConfig)
		if err != nil {
			return errs.Errorf("error parsing row %d, column %d string %q: %w", index, col.Index, row[col.Index], err)
		}
	}

	return nil
}

// // Read reads from an io.Reader to a structSlicePtr
// func (r *Reader) Read(reader io.Reader, structSlicePtr interface{}) (err error) {
// 	defer errs.WrapWithFuncParams(&err, reader, structSlicePtr)

// 	data, err := io.ReadAll(reader)
// 	if err != nil {
// 		return err
// 	}

// 	readRows, err := ParseStringsWithFormat(data, r.Format, r.NewlineReplacement)
// 	if err != nil {
// 		return err
// 	}

// 	return mapStrings(cleanedRows, r.Columns, structSlicePtr)
// }

// // ReadFile reads from a fs.FileReader to a structSlicePtr
// func (r *Reader) ReadFile(file fs.FileReader, structSlicePtr interface{}) (err error) {
// 	defer errs.WrapWithFuncParams(&err, file, structSlicePtr)

// 	reader, err := file.OpenReader()
// 	if err != nil {
// 		return err
// 	}
// 	defer reader.Close()

// 	return r.Read(reader, structSlicePtr)
// }
