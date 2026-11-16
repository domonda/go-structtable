package structtable

import (
	"context"
	"errors"
	"io"
	"reflect"

	fs "github.com/ungerik/go-fs"

	"github.com/domonda/go-errs"
)

// Renderer defines the interface for rendering struct slices into various table formats.
//
// This interface provides methods to render tabular data from Go struct slices
// into different output formats such as HTML, CSV, Excel, etc.
type Renderer interface {
	// RenderHeaderRow renders the table header row with the given column titles.
	RenderHeaderRow(columnTitles []string) error
	// RenderRow renders a single data row with the given column values.
	RenderRow(columnValues []reflect.Value) error

	// Result returns the rendered table data as bytes.
	Result() ([]byte, error)
	// WriteResultTo writes the rendered table data to the given writer.
	WriteResultTo(w io.Writer) error
	// WriteResultFile writes the rendered table data to the given file.
	WriteResultFile(file fs.File, perm ...fs.Permissions) error

	// MIMEType returns the MIME-Type of the rendered content.
	MIMEType() string
}

// Render renders a slice of structs into a table using the given Renderer.
//
// This function takes a slice of structs and renders them as a table using
// the provided Renderer implementation. It uses the ColumnMapper to determine
// column titles and extract values from the struct instances.
//
// Parameters:
//   - renderer: The Renderer implementation to use for output formatting
//   - structSlice: The slice of structs to render
//   - renderTitleRow: Whether to include a header row with column titles
//   - columnMapper: The ColumnMapper to use for field-to-column mapping
//
// Returns:
//   - err: Any error that occurred during rendering
func Render(renderer Renderer, structSlice any, renderTitleRow bool, columnMapper ColumnMapper) error {
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

// RenderTo renders a slice of structs into a table and writes the result to a writer.
//
// This is a convenience function that combines Render and WriteResultTo.
// It renders the struct slice and immediately writes the result to the provided writer.
//
// Parameters:
//   - writer: The io.Writer to write the rendered table to
//   - renderer: The Renderer implementation to use for output formatting
//   - structSlice: The slice of structs to render
//   - renderTitleRow: Whether to include a header row with column titles
//   - columnMapper: The ColumnMapper to use for field-to-column mapping
//
// Returns:
//   - err: Any error that occurred during rendering or writing
func RenderTo(writer io.Writer, renderer Renderer, structSlice any, renderTitleRow bool, columnMapper ColumnMapper) error {
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

// RenderBytes renders a slice of structs into a table and returns the result as bytes.
//
// This is a convenience function that combines Render and Result.
// It renders the struct slice and returns the result as a byte slice.
//
// Parameters:
//   - renderer: The Renderer implementation to use for output formatting
//   - structSlice: The slice of structs to render
//   - renderTitleRow: Whether to include a header row with column titles
//   - columnMapper: The ColumnMapper to use for field-to-column mapping
//
// Returns:
//   - data: The rendered table data as bytes
//   - err: Any error that occurred during rendering
func RenderBytes(renderer Renderer, structSlice any, renderTitleRow bool, columnMapper ColumnMapper) ([]byte, error) {
	err := Render(renderer, structSlice, renderTitleRow, columnMapper)
	if err != nil {
		return nil, err
	}
	return renderer.Result()
}

// RenderFile renders a slice of structs into a table and writes the result to a file.
//
// This is a convenience function that combines Render and WriteResultFile.
// It renders the struct slice and immediately writes the result to the provided file.
//
// Parameters:
//   - ctx: Cancels or times out writing the file
//   - file: The fs.File to write the rendered table to
//   - renderer: The Renderer implementation to use for output formatting
//   - structSlice: The slice of structs to render
//   - renderTitleRow: Whether to include a header row with column titles
//   - columnMapper: The ColumnMapper to use for field-to-column mapping
//
// Returns:
//   - err: Any error that occurred during rendering or writing
func RenderFile(ctx context.Context, file fs.File, renderer Renderer, structSlice any, renderTitleRow bool, columnMapper ColumnMapper) error {
	err := Render(renderer, structSlice, renderTitleRow, columnMapper)
	if err != nil {
		return err
	}
	data, err := renderer.Result()
	if err != nil {
		return err
	}
	return file.WriteAll(ctx, data)
}

// closeOrAbortWriter finishes writing to a writer opened with fs.File.OpenWriter
// after the write returned writeErr.
//
// Without writeErr the writer is closed and the error of Close is returned,
// because writers of uploads commit the data in Close.
// After a failed write a writer implementing fs.WriteAborter is aborted
// instead of closed, so no partial data replaces the file,
// and writeErr is returned joined with the error of Abort or Close.
func closeOrAbortWriter(writer io.WriteCloser, writeErr error) error {
	if writeErr == nil {
		return writer.Close()
	}
	if aborter, ok := writer.(fs.WriteAborter); ok {
		return errors.Join(writeErr, aborter.Abort())
	}
	return errors.Join(writeErr, writer.Close())
}
