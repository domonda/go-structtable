package excel

import (
	"archive/zip"
	"reflect"

	xlsx "github.com/tealeg/xlsx/v3"
	fs "github.com/ungerik/go-fs"

	"github.com/domonda/go-errs"
)

type Reader struct {
	sheet *xlsx.Sheet
}

// NewReader creates a new structtable.Reader for the sheet sheetName in xlsxFile.
// If sheetName is "", then the first sheet will be used.
// Note: Reader only reads into string kind struct fields so far.
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
	} else {
		reader.sheet = file.Sheets[0]
	}

	return reader, nil
}

func (r *Reader) NumRows() int {
	return r.sheet.MaxRow
}

func (r *Reader) ReadRowStrings(rowIndex int) ([]string, error) {
	if rowIndex < 0 || rowIndex >= r.sheet.MaxRow {
		return nil, errs.Errorf("row index %d out of bounds", rowIndex)
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

func (r *Reader) ReadRow(rowIndex int, destStruct reflect.Value) error {
	if rowIndex < 0 || rowIndex >= r.sheet.MaxRow {
		return errs.Errorf("row index %d out of bounds", rowIndex)
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

func (r *Reader) SheetName() string {
	return r.sheet.Name
}
