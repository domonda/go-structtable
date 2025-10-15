package structtable

import (
	"bytes"
	"io"
	"reflect"

	"github.com/domonda/go-types/strfmt"
	fs "github.com/ungerik/go-fs"
)

// TextFormatRenderer has to be implemented for a format
// to be used by TextRenderer.
//
// This interface defines methods for rendering text-based table formats,
// such as CSV, TSV, or custom delimited formats.
type TextFormatRenderer interface {
	// RenderBeginTableText renders any content that should appear before the table.
	RenderBeginTableText(writer io.Writer) error
	// RenderHeaderRowText renders a header row with the given column titles.
	RenderHeaderRowText(writer io.Writer, columnTitles []string) error
	// RenderRowText renders a data row with the given field values.
	RenderRowText(writer io.Writer, fields []string) error
	// RenderEndTableText renders any content that should appear after the table.
	RenderEndTableText(writer io.Writer) error
}

// TextRenderer implements Renderer by using a TextFormatRenderer
// for a specific text based table format.
//
// This renderer generates text-based tables from struct slices, with support
// for custom formatting through the TextFormatRenderer interface.
type TextRenderer struct {
	format       TextFormatRenderer
	config       *strfmt.FormatConfig
	buf          bytes.Buffer
	beginWritten bool
}

// NewTextRenderer creates a new TextRenderer instance.
//
// This constructor initializes a TextRenderer with the provided format renderer
// and text formatting configuration.
//
// Parameters:
//   - format: The TextFormatRenderer for custom text formatting
//   - config: Text formatting configuration for cell values
//
// Returns:
//   - A new TextRenderer instance ready for use
func NewTextRenderer(format TextFormatRenderer, config *strfmt.FormatConfig) *TextRenderer {
	tw := &TextRenderer{
		format: format,
		config: config,
	}
	return tw
}

// func (txt *TextRenderer) SetTypeTextFormatter(columnType reflect.Type, formatter TextFormatter) {
// 	if formatter != nil {
// 		txt.typeFormatters[columnType] = formatter
// 	} else {
// 		delete(txt.typeFormatters, columnType)
// 	}
// }

func (txt *TextRenderer) writeBeginIfMissing() error {
	if txt.beginWritten {
		return nil
	}
	err := txt.format.RenderBeginTableText(&txt.buf)
	if err != nil {
		return err
	}
	txt.beginWritten = true
	return nil
}

// RenderHeaderRow renders the table header row with the given column titles.
//
// This method implements the Renderer interface and generates the text
// for the table header row, including any pre-table content.
func (txt *TextRenderer) RenderHeaderRow(columnTitles []string) error {
	err := txt.writeBeginIfMissing()
	if err != nil {
		return err
	}
	return txt.format.RenderHeaderRowText(&txt.buf, columnTitles)
}

// RenderRow renders a single data row with the given column values.
//
// This method implements the Renderer interface and generates the text
// for a single table row with the provided column values.
func (txt *TextRenderer) RenderRow(columnValues []reflect.Value) error {
	err := txt.writeBeginIfMissing()
	if err != nil {
		return err
	}
	fields := make([]string, len(columnValues))
	for i, val := range columnValues {
		fields[i] = strfmt.FormatValue(val, txt.config)
	}
	return txt.format.RenderRowText(&txt.buf, fields)
}

// Result returns the rendered table data as bytes.
//
// This method implements the Renderer interface and returns the complete
// text table as a byte slice, including any post-table content.
func (txt *TextRenderer) Result() ([]byte, error) {
	err := txt.format.RenderEndTableText(&txt.buf)
	if err != nil {
		return nil, err
	}
	return txt.buf.Bytes(), nil
}

// WriteResultTo writes the rendered table data to the given writer.
//
// This method implements the Renderer interface and writes the complete
// text table to the provided writer.
func (txt *TextRenderer) WriteResultTo(writer io.Writer) error {
	_, err := txt.buf.WriteTo(writer)
	return err
}

// WriteResultFile writes the rendered table data to the given file.
//
// This method implements the Renderer interface and writes the complete
// text table to the provided file with optional permissions.
func (txt *TextRenderer) WriteResultFile(file fs.File, perm ...fs.Permissions) error {
	writer, err := file.OpenWriter(perm...)
	if err != nil {
		return err
	}
	defer writer.Close()

	return txt.WriteResultTo(writer)
}
