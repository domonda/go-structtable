package structtable

import (
	"bytes"
	"io"
	"reflect"

	"github.com/domonda/go-types/strfmt"
	fs "github.com/ungerik/go-fs"
)

// TextFormatRenderer has to be formatemented for a format
// to be used by TextRenderer.
type TextFormatRenderer interface {
	RenderBeginTableText(writer io.Writer) error
	RenderHeaderRowText(writer io.Writer, columnTitles []string) error
	RenderRowText(writer io.Writer, fields []string) error
	RenderEndTableText(writer io.Writer) error
}

// TextRenderer implements Renderer by using a TextFormatRenderer
// for a specific text based table format.
type TextRenderer struct {
	format       TextFormatRenderer
	config       *strfmt.FormatConfig
	buf          bytes.Buffer
	beginWritten bool
}

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

func (txt *TextRenderer) RenderHeaderRow(columnTitles []string) error {
	err := txt.writeBeginIfMissing()
	if err != nil {
		return err
	}
	return txt.format.RenderHeaderRowText(&txt.buf, columnTitles)
}

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

func (txt *TextRenderer) Result() ([]byte, error) {
	err := txt.format.RenderEndTableText(&txt.buf)
	if err != nil {
		return nil, err
	}
	return txt.buf.Bytes(), nil
}

func (txt *TextRenderer) WriteResultTo(writer io.Writer) error {
	_, err := txt.buf.WriteTo(writer)
	return err
}

func (txt *TextRenderer) WriteResultFile(file fs.File, perm ...fs.Permissions) error {
	writer, err := file.OpenWriter(perm...)
	if err != nil {
		return err
	}
	defer writer.Close()

	return txt.WriteResultTo(writer)
}
