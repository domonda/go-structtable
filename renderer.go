package structtable

import (
	"io"
	"reflect"

	fs "github.com/ungerik/go-fs"

	"github.com/domonda/go-errs"
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

func Render(renderer Renderer, structSlice interface{}, renderTitleRow bool, columnMapper ColumnMapper) error {
	rows := reflect.ValueOf(structSlice)
	if rows.Kind() != reflect.Slice {
		return errs.Errorf("passed value is not a slice, but %T", structSlice)
	}

	columnTitles, rowReflector := columnMapper.ColumnTitlesAndRowReflector(rows.Type().Elem())

	if renderTitleRow {
		err := renderer.RenderHeaderRow(columnTitles)
		if err != nil {
			return err
		}
	}

	for i := 0; i < rows.Len(); i++ {
		err := renderer.RenderRow(rowReflector.ReflectRow(rows.Index(i)))
		if err != nil {
			return err
		}
	}

	return nil
}

func RenderTo(writer io.Writer, renderer Renderer, structSlice interface{}, renderTitleRow bool, columnMapper ColumnMapper) error {
	err := Render(renderer, structSlice, renderTitleRow, columnMapper)
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

func RenderBytes(renderer Renderer, structSlice interface{}, renderTitleRow bool, columnMapper ColumnMapper) ([]byte, error) {
	err := Render(renderer, structSlice, renderTitleRow, columnMapper)
	if err != nil {
		return nil, err
	}
	return renderer.Result()
}

func RenderFile(file fs.File, renderer Renderer, structSlice interface{}, renderTitleRow bool, columnMapper ColumnMapper) error {
	err := Render(renderer, structSlice, renderTitleRow, columnMapper)
	if err != nil {
		return err
	}
	data, err := renderer.Result()
	if err != nil {
		return err
	}
	return file.WriteAll(data)
}
