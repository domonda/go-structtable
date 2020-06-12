package csv

import (
	"io"
	"io/ioutil"
	"reflect"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/strfmt"
	"github.com/ungerik/go-fs"
)

// type TableDetectionConfig struct {
// 	Format  *FormatDetectionConfig
// 	Columns []TableDetectionConfigColumn
// }

// type TableDetectionConfigColumn struct {
// 	StructField string
// 	HeaderNames []string
// }

type ColumnMapping struct {
	Index       int
	StructField string
}

type Reader struct {
	Format          *Format                `json:"format,omitempty"`
	FormatDetection *FormatDetectionConfig `json:"formatDetection,omitempty"`
	ScanConfig      *strfmt.ScanConfig     `json:"config"`
	Modifiers       ModifierList           `json:"modifiers"`
	Columns         []ColumnMapping        `json:"columns"`

	rows [][]string
}

// NewReader reads from an io.Reader
func NewReader(reader io.Reader, format *Format, newlineReplacement string, modifiers ModifierList, columns []ColumnMapping, scanConfig ...*strfmt.ScanConfig) (r *Reader, err error) {
	defer errs.WrapWithFuncParams(&err, reader, format, newlineReplacement, modifiers, columns, scanConfig)

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	rows, err := ParseStringsWithFormat(data, format)
	if err != nil {
		return nil, err
	}

	return NewReaderFromRows(rows, format, newlineReplacement, modifiers, columns, scanConfig...)
}

// NewReaderFromRows returns a Reader that uses pre-parsed rows
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

// NewReaderFromFile reads from a fs.FileReader
func NewReaderFromFile(file fs.FileReader, format *Format, newlineReplacement string, modifiers ModifierList, columns []ColumnMapping, scanConfig ...*strfmt.ScanConfig) (r *Reader, err error) {
	defer errs.WrapWithFuncParams(&err, file, format, newlineReplacement, modifiers, columns, scanConfig)

	reader, err := file.OpenReader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return NewReader(reader, format, newlineReplacement, modifiers, columns, scanConfig...)
}

func (r *Reader) NumRows() int {
	return len(r.rows)
}

func (r *Reader) ReadRowStrings(index int) ([]string, error) {
	if index < 0 || index > len(r.rows) {
		return nil, errs.Errorf("row index %d out of bounds [0..%d)", index, len(r.rows))
	}
	return r.rows[index], nil
}

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

// 	data, err := ioutil.ReadAll(reader)
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
