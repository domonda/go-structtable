package excel

import (
	"archive/zip"
	"reflect"

	"github.com/tealeg/xlsx"
	fs "github.com/ungerik/go-fs"

	"github.com/domonda/go-wraperr"
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
			return nil, wraperr.Errorf("excel file %s does not have a sheet called %q", xlsxFile, sheetName)
		}
	} else {
		reader.sheet = file.Sheets[0]
	}

	return reader, nil
}

func (r *Reader) NumRows() int {
	return len(r.sheet.Rows)
}

func (r *Reader) ReadRowStrings(index int) ([]string, error) {
	if index < 0 || index >= len(r.sheet.Rows) {
		return nil, wraperr.Errorf("row index %d out of bounds", index)
	}

	row := make([]string, len(r.sheet.Rows[index].Cells))
	for i, cell := range r.sheet.Rows[index].Cells {
		row[i] = cell.String()
	}
	return row, nil
}

func (r *Reader) ReadRow(index int, destStruct reflect.Value) error {
	if index < 0 || index >= len(r.sheet.Rows) {
		return wraperr.Errorf("row index %d out of bounds", index)
	}

	cells := r.sheet.Rows[index].Cells
	for col := 0; col < len(cells) && col < destStruct.NumField(); col++ {
		destStruct.Field(col).SetString(cells[col].String())
	}

	return nil
}

func (r *Reader) SheetName() string {
	return r.sheet.Name
}
