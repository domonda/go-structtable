package csv

import (
	"io"
	"io/ioutil"
	"reflect"

	"github.com/domonda/go-types/assign"
	"github.com/domonda/go-wraperr"
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
	StringParser    *assign.StringParser   `json:"stringParser"`
	Modifiers       ModifierList           `json:"modifiers"`
	Columns         []ColumnMapping        `json:"columns"`

	rows [][]string
}

// NewReader reads from an io.Reader
func NewReader(reader io.Reader, format *Format, newlineReplacement string, modifiers ModifierList, columns []ColumnMapping, stringParser ...*assign.StringParser) (r *Reader, err error) {
	defer wraperr.WithFuncParams(&err, reader)

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// if format

	rows, err := ParseStringsWithFormat(data, format)
	if err != nil {
		return nil, err
	}

	r = &Reader{
		Format:       format,
		StringParser: assign.DefaultStringParser,
		Modifiers:    modifiers,
		Columns:      columns,
		rows:         rows,
	}
	if len(stringParser) > 0 && stringParser[0] != nil {
		r.StringParser = stringParser[0]
	}
	return r, nil
}

// // NewReaderFromFile reads from a fs.FileReader
// func NewReaderFromFile(file fs.FileReader, structSlicePtr interface{}) (err error) {
// 	defer wraperr.WithFuncParams(&err, file, structSlicePtr)

// 	reader, err := file.OpenReader()
// 	if err != nil {
// 		return err
// 	}
// 	defer reader.Close()

// 	return r.Read(reader, structSlicePtr)
// }

func (r *Reader) NumRows() int {
	return len(r.rows)
}

func (r *Reader) ReadRowStrings(index int) ([]string, error) {
	if index < 0 || index > len(r.rows) {
		return nil, wraperr.Errorf("row index %d out of bounds [0..%d)", index, len(r.rows))
	}
	return r.rows[index], nil
}

func (r *Reader) ReadRow(index int, destStruct reflect.Value) error {
	if index < 0 || index >= len(r.rows) {
		return wraperr.Errorf("row index %d out of bounds [0..%d)", index, len(r.rows))
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
		err := assign.String(destStructField, row[col.Index], r.StringParser)
		if err != nil {
			return wraperr.Errorf("error parsing row %d, column %d string %q: %w", index, col.Index, row[col.Index], err)
		}
	}

	return nil
}

// // Read reads from an io.Reader to a structSlicePtr
// func (r *Reader) Read(reader io.Reader, structSlicePtr interface{}) (err error) {
// 	defer wraperr.WithFuncParams(&err, reader, structSlicePtr)

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
// 	defer wraperr.WithFuncParams(&err, file, structSlicePtr)

// 	reader, err := file.OpenReader()
// 	if err != nil {
// 		return err
// 	}
// 	defer reader.Close()

// 	return r.Read(reader, structSlicePtr)
// }
