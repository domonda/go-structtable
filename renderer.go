package structtable

import (
	"io"
	"reflect"

	fs "github.com/ungerik/go-fs"
	reflection "github.com/ungerik/go-reflection"

	"github.com/domonda/go-wraperr"
)

type Renderer interface {
	RenderHeaderRow(columnTitles []string) error
	RenderRow(columnValues []reflect.Value) error

	Result() ([]byte, error)
	WriteResultTo(w io.Writer) error
	WriteResultFile(file fs.File, perm ...fs.Permissions) error

	// MIMEType returns the MIME-Type of the rendered content
	MIMEType() string
}

func Render(renderer Renderer, structSlice interface{}, columnTitles ...string) error {
	rows := reflect.ValueOf(structSlice)
	if rows.Kind() != reflect.Slice {
		return wraperr.Errorf("passed value is not a slice, but %T", structSlice)
	}

	if len(columnTitles) > 0 {
		err := renderer.RenderHeaderRow(columnTitles)
		if err != nil {
			return err
		}
	}
	for i := 0; i < rows.Len(); i++ {
		columnValues := reflection.FlatStructFieldValues(rows.Index(i))
		err := renderer.RenderRow(columnValues)
		if err != nil {
			return err
		}
	}

	return nil
}

func RenderReflectColumnTitles(renderer Renderer, structSlice interface{}, columnTitleTag string) error {
	columnTitles, err := reflectColumnTitles(structSlice, columnTitleTag)
	if err != nil {
		return err
	}

	return Render(renderer, structSlice, columnTitles...)
}

func RenderTo(writer io.Writer, renderer Renderer, structSlice interface{}, columnTitles ...string) error {
	err := Render(renderer, structSlice, columnTitles...)
	if err != nil {
		return err
	}
	data, err := renderer.Result()
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

func RenderToReflectColumnTitles(writer io.Writer, renderer Renderer, structSlice interface{}, columnTitleTag string) error {
	columnTitles, err := reflectColumnTitles(structSlice, columnTitleTag)
	if err != nil {
		return err
	}

	return RenderTo(writer, renderer, structSlice, columnTitles...)
}

func RenderFile(file fs.File, renderer Renderer, structSlice interface{}, columnTitles ...string) error {
	err := Render(renderer, structSlice, columnTitles...)
	if err != nil {
		return err
	}
	data, err := renderer.Result()
	if err != nil {
		return err
	}
	return file.WriteAll(data)
}

func RenderFileReflectColumnTitles(file fs.File, renderer Renderer, structSlice interface{}, columnTitleTag string) error {
	columnTitles, err := reflectColumnTitles(structSlice, columnTitleTag)
	if err != nil {
		return err
	}

	return RenderFile(file, renderer, structSlice, columnTitles...)
}

func reflectColumnTitles(structSlice interface{}, columnTitleTag string) ([]string, error) {
	rows := reflect.ValueOf(structSlice)
	if rows.Kind() != reflect.Slice {
		return nil, wraperr.Errorf("passed value is not a slice, but %T", structSlice)
	}
	columnTitles := reflection.FlatStructFieldTagsOrNames(rows.Type().Elem(), columnTitleTag)
	return columnTitles, nil
}
