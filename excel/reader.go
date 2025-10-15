package excel

import (
	"archive/zip"
	"reflect"

	xlsx "github.com/tealeg/xlsx/v3"
	fs "github.com/ungerik/go-fs"

	"github.com/domonda/go-errs"
)

// Reader implements the structtable.Reader interface for Excel (.xlsx) files.
//
// This reader can parse Excel files and populate struct instances with the data.
// It supports reading from specific sheets and handles various data types.
type Reader struct {
	sheet *xlsx.Sheet
}

// NewReader creates a new structtable.Reader for the specified sheet in an Excel file.
//
// This constructor opens an Excel file and creates a Reader instance for the
// specified sheet. If sheetName is empty, the first sheet will be used.
//
// Note: Currently, this reader only populates string-type struct fields.
//
// Parameters:
//   - xlsxFile: The Excel file to read from
//   - sheetName: The name of the sheet to read (empty string for first sheet)
//
// Returns:
//   - A new Reader instance ready for use
//   - err: Any error that occurred during file opening or sheet access
func NewReader(xlsxFile fs.FileReader, sheetName string) (*Reader, error) {
	fileReader, err := xlsxFile.OpenReadSeeker()
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()

	zipReader, err := zip.NewReader(fileReader, xlsxFile.Size())
	if err != nil {
		return nil, err
	}

	file, err := xlsx.ReadZipReader(zipReader)
	if err != nil {
		return nil, err
	}

	reader := new(Reader)
	if sheetName != "" {
		reader.sheet = file.Sheet[sheetName]
		if reader.sheet == nil {
			return nil, errs.Errorf("excel file %s does not have a sheet called %q", xlsxFile, sheetName)
		}
	} else if len(file.Sheets) > 0 {
		reader.sheet = file.Sheets[0]
	} else {
		return nil, errs.New("excel file has no sheets")
	}

	return reader, nil
}

// NumRows returns the total number of rows in the Excel sheet.
//
// This method returns the maximum row index in the sheet, which represents
// the total number of rows available for reading. Note that this includes
// empty rows, so the actual number of rows with data may be less.
//
// Returns:
//   - int: The total number of rows in the sheet (1-based, so MaxRow is the count)
//
// Example:
//
//	reader, err := excel.NewReader(file, "Sheet1")
//	if err != nil {
//	    return err
//	}
//	totalRows := reader.NumRows()
//	fmt.Printf("Sheet has %d rows\n", totalRows)
func (r *Reader) NumRows() int {
	return r.sheet.MaxRow
}

// ReadRowStrings returns raw string values for a specific row in the Excel sheet.
//
// This method reads all cells from the specified row and returns them as a slice
// of strings. It reads from all columns up to MaxCol, filling empty cells with
// empty strings. This is useful when you need access to the raw data without
// struct mapping.
//
// Parameters:
//   - rowIndex: The zero-based index of the row to read
//
// Returns:
//   - []string: Slice of string values from the row (one per column)
//   - err: Any error that occurred during reading or bounds checking
//
// Bounds Checking:
//   - Returns error if rowIndex is negative or >= MaxRow
//   - Reads all columns from 0 to MaxCol-1
//
// Example:
//
//	rowData, err := reader.ReadRowStrings(0) // Read first row
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Row contains %d columns\n", len(rowData))
func (r *Reader) ReadRowStrings(rowIndex int) ([]string, error) {
	if rowIndex < 0 || rowIndex >= r.sheet.MaxRow {
		return nil, errs.Errorf("rowIndex %d out of bounds", rowIndex)
	}

	row, err := r.sheet.Row(rowIndex)
	if err != nil {
		return nil, err
	}
	strs := make([]string, r.sheet.MaxCol)
	for col := range strs {
		strs[col] = row.GetCell(col).String()
	}
	return strs, nil
}

// ReadRow populates a struct instance with data from the specified Excel row.
//
// This method reads data from the specified row and populates the fields of the
// provided struct instance. It maps Excel columns to struct fields by position
// (first column to first field, second column to second field, etc.).
//
// Important Limitations:
//   - Only populates string-type struct fields
//   - Field mapping is positional (column index = field index)
//   - Stops reading when either MaxCol or NumField() is reached
//   - All cell values are converted to strings using String() method
//
// Parameters:
//   - rowIndex: The zero-based index of the row to read
//   - destStruct: A reflect.Value pointing to the struct instance to populate
//
// Returns:
//   - err: Any error that occurred during reading or bounds checking
//
// Bounds Checking:
//   - Returns error if rowIndex is negative or >= MaxRow
//   - Only reads up to the minimum of MaxCol and destStruct.NumField()
//
// Example:
//
//	type Person struct {
//	    Name string
//	    Age  string  // Note: string type required
//	    City string
//	}
//	var person Person
//	err := reader.ReadRow(0, reflect.ValueOf(&person).Elem())
func (r *Reader) ReadRow(rowIndex int, destStruct reflect.Value) error {
	if rowIndex < 0 || rowIndex >= r.sheet.MaxRow {
		return errs.Errorf("rowIndex %d out of bounds", rowIndex)
	}

	row, err := r.sheet.Row(rowIndex)
	if err != nil {
		return err
	}
	for col := 0; col < r.sheet.MaxCol && col < destStruct.NumField(); col++ {
		destStruct.Field(col).SetString(row.GetCell(col).String())
	}
	return nil
}

// SheetName returns the name of the current Excel sheet.
//
// This method returns the name of the sheet that this reader is currently
// reading from. This is useful for debugging, logging, or when you need to
// identify which sheet the data came from.
//
// Returns:
//   - string: The name of the Excel sheet
//
// Example:
//
//	reader, err := excel.NewReader(file, "Sales Data")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Reading from sheet: %s\n", reader.SheetName())
func (r *Reader) SheetName() string {
	return r.sheet.Name
}
