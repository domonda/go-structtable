package csv

import (
	"io"
	"io/ioutil"

	fs "github.com/ungerik/go-fs"

	"github.com/domonda/go-wraperr"
)

type TableDetectionConfig struct {
	Format  *FormatDetectionConfig
	Columns []TableDetectionConfigColumn
}

type TableDetectionConfigColumn struct {
	StructField string
	HeaderNames []string
}

type ReadConfig struct {
	Format                         *Format
	NewlineReplacement             string
	CleanSpacedStrings             bool
	EmptyRowsWithNonUniformColumns bool
	EmptyEmptyRows                 bool
	IgnoreTopRows                  uint
	HasHeaderRow                   bool
	IgnoreBottomRows               uint
	Columns                        []ColumnMapping
}

type ColumnMapping struct {
	Index       int
	StructField string
}

func Read(r io.Reader, config *ReadConfig, structSlicePtr interface{}) (err error) {
	defer wraperr.WithFuncParams(&err, r, config, structSlicePtr)

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	readRows, err := ParseStringsWithFormat(data, config.Format, config.NewlineReplacement)
	if err != nil {
		return err
	}

	if config.CleanSpacedStrings {
		CleanSpacedStrings(readRows)
	}

	cleanedRows := readRows
	if config.EmptyRowsWithNonUniformColumns {
		cleanedRows = EmptyRowsWithNonUniformColumns(cleanedRows)
	}
	if config.EmptyEmptyRows {
		cleanedRows = EmptyEmptyRows(cleanedRows)
	}

	ignoreTop := int(config.IgnoreTopRows)
	if config.HasHeaderRow {
		ignoreTop++
	}
	for i := 0; i < ignoreTop; i++ {
		cleanedRows[i] = nil
	}
	for i := len(cleanedRows) - int(config.IgnoreBottomRows); i < len(cleanedRows); i++ {
		cleanedRows[i] = nil
	}

	return mapStrings(cleanedRows, config.Columns, structSlicePtr)
}

func mapStrings(rows [][]string, colMapping []ColumnMapping, structSlicePtr interface{}) (err error) {
	return nil
}

func ReadFile(file fs.FileReader, config *ReadConfig, structSlicePtr interface{}) (err error) {
	defer wraperr.WithFuncParams(&err, file, config, structSlicePtr)

	reader, err := file.OpenReader()
	if err != nil {
		return err
	}
	defer reader.Close()

	return Read(reader, config, structSlicePtr)
}
