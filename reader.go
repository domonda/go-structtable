package structtable

import (
	"reflect"

	"github.com/domonda/go-errs"
)

// Reader defines the interface for reading table data into struct slices.
//
// This interface provides methods to read tabular data and convert it into
// Go struct instances. It supports both string-based access and direct
// struct field population.
type Reader interface {
	// NumRows returns the total number of rows available for reading.
	NumRows() int
	// ReadRowStrings returns the raw string values for a specific row.
	// This is useful for debugging or when you need access to the raw data.
	ReadRowStrings(index int) ([]string, error)
	// ReadRow populates a struct instance with data from the specified row.
	// The destStruct parameter should be a reflect.Value of the struct to populate.
	ReadRow(index int, destStruct reflect.Value) error
}

// Read reads table data from a Reader into a slice of structs.
//
// This function reads all rows from the Reader and populates a slice of structs
// with the data. It can optionally skip header rows and return them separately.
//
// Parameters:
//   - reader: The Reader implementation to read data from
//   - structSlicePtr: A pointer to a slice of structs to populate
//   - numHeaderRows: Number of header rows to skip (returned separately)
//
// Returns:
//   - headerRows: The header rows that were skipped (if any)
//   - err: Any error that occurred during reading
//
// Example:
//
//	var people []Person
//	headers, err := Read(csvReader, &people, 1)
func Read(reader Reader, structSlicePtr any, numHeaderRows int) (headerRows [][]string, err error) {
	if numHeaderRows < 0 {
		return nil, errs.New("numHeaderRows can't be negative")
	}
	destVal := reflect.ValueOf(structSlicePtr)
	if destVal.Kind() != reflect.Ptr {
		return nil, errs.Errorf("structSlicePtr must be pointer to a struct slice, but is %T", structSlicePtr)
	}
	if destVal.IsNil() {
		return nil, errs.Errorf("structSlicePtr must not be nil")
	}
	sliceType := destVal.Elem().Type()
	if sliceType.Kind() != reflect.Slice {
		return nil, errs.Errorf("structSlicePtr must be pointer to a struct slice, but is %T", structSlicePtr)
	}
	structType := sliceType.Elem()
	isSliceOfPtr := structType.Kind() == reflect.Ptr
	if isSliceOfPtr {
		structType = structType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return nil, errs.Errorf("structSlicePtr must be pointer to a struct slice, but is %T", structSlicePtr)
	}

	for i := 0; i < numHeaderRows && i < reader.NumRows(); i++ {
		row, err := reader.ReadRowStrings(i)
		if err != nil {
			return nil, err
		}
		headerRows = append(headerRows, row)
	}

	numRows := reader.NumRows() - numHeaderRows
	sliceVal := reflect.MakeSlice(sliceType, numRows, numRows)
	for i := 0; i < numRows; i++ {
		var destStruct reflect.Value
		if isSliceOfPtr {
			// Allocate new struct pointer
			ptrVal := reflect.New(structType)
			// and assign to slice delement
			sliceVal.Index(i).Set(ptrVal)
			// Don't need pointer for ReadRow, reflect.Value is writeable
			destStruct = ptrVal.Elem()
		} else {
			destStruct = sliceVal.Index(i)
		}
		err := reader.ReadRow(int(numHeaderRows)+i, destStruct)
		if err != nil {
			return nil, err
		}
	}

	// Assign result only if there was no error
	destVal.Elem().Set(sliceVal)
	return headerRows, nil
}
