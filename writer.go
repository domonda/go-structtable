package structtable

import (
	"io"
	"reflect"

	fs "github.com/ungerik/go-fs"
	reflection "github.com/ungerik/go-reflection"

	"github.com/domonda/go-wraperr"
)

type Writer interface {
	WriteHeaderRow(columnTitles []string) error
	WriteRow(columnValues []reflect.Value) error

	Result() ([]byte, error)
	WriteResultTo(w io.Writer) error
	WriteResultFile(file fs.File, perm ...fs.Permissions) error
}

func Write(writer Writer, structSlice interface{}, columnTitles ...string) error {
	rows := reflect.ValueOf(structSlice)
	if rows.Kind() != reflect.Slice {
		return wraperr.Errorf("passed value is not a slice, but %T", structSlice)
	}

	if len(columnTitles) > 0 {
		err := writer.WriteHeaderRow(columnTitles)
		if err != nil {
			return err
		}
	}
	for i := 0; i < rows.Len(); i++ {
		columnValues := reflection.FlatStructFieldValues(rows.Index(i))
		err := writer.WriteRow(columnValues)
		if err != nil {
			return err
		}
	}

	return nil
}

func WriteReflectColumnTitles(writer Writer, structSlice interface{}, columnTitleTag string) error {
	columnTitles, err := reflectColumnTitles(structSlice, columnTitleTag)
	if err != nil {
		return err
	}

	return Write(writer, structSlice, columnTitles...)
}

func WriteTo(destination io.Writer, writer Writer, structSlice interface{}, columnTitles ...string) error {
	err := Write(writer, structSlice, columnTitles...)
	if err != nil {
		return err
	}
	data, err := writer.Result()
	if err != nil {
		return err
	}
	_, err = destination.Write(data)
	return err
}

func WriteToReflectColumnTitles(destination io.Writer, writer Writer, structSlice interface{}, columnTitleTag string) error {
	columnTitles, err := reflectColumnTitles(structSlice, columnTitleTag)
	if err != nil {
		return err
	}

	return WriteTo(destination, writer, structSlice, columnTitles...)
}

func WriteFile(file fs.File, writer Writer, structSlice interface{}, columnTitles ...string) error {
	err := Write(writer, structSlice, columnTitles...)
	if err != nil {
		return err
	}
	data, err := writer.Result()
	if err != nil {
		return err
	}
	return file.WriteAll(data)
}

func WriteFileReflectColumnTitles(file fs.File, writer Writer, structSlice interface{}, columnTitleTag string) error {
	columnTitles, err := reflectColumnTitles(structSlice, columnTitleTag)
	if err != nil {
		return err
	}

	return WriteFile(file, writer, structSlice, columnTitles...)
}

func reflectColumnTitles(structSlice interface{}, columnTitleTag string) ([]string, error) {
	rows := reflect.ValueOf(structSlice)
	if rows.Kind() != reflect.Slice {
		return nil, wraperr.Errorf("passed value is not a slice, but %T", structSlice)
	}
	columnTitles := reflection.FlatStructFieldTagsOrNames(rows.Type().Elem(), columnTitleTag)
	return columnTitles, nil
}
