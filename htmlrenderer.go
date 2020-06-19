package structtable

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"reflect"

	"github.com/domonda/go-types/txtfmt"
	fs "github.com/ungerik/go-fs"
	"github.com/ungerik/go-reflection"
)

// HTMLFormatRenderer is the renderer for the HTML format.
type HTMLFormatRenderer interface {
	RenderBeforeTable(writer io.Writer) error
	Caption() string
}

// HTMLRenderer implements Renderer by using a HTMLFormatRenderer
// for a specific text based table format.
type HTMLRenderer struct {
	format HTMLFormatRenderer
	config *txtfmt.FormatConfig
	buf    bytes.Buffer
}

func NewHTMLRenderer(format HTMLFormatRenderer, config *txtfmt.FormatConfig) *HTMLRenderer {
	return &HTMLRenderer{format: format, config: config}
}

func (htm *HTMLRenderer) RenderHeaderRow(columnTitles []string) error {
	err := htm.format.RenderBeforeTable(&htm.buf)
	if err != nil {
		return err
	}

	err = htm.write("<table>\n")
	if err != nil {
		return err
	}
	caption := htm.format.Caption()
	if caption != "" {
		err = htm.write("<caption>%s</caption>\n", caption)
		if err != nil {
			return err
		}
	}
	err = htm.write("<tr>\n")
	if err != nil {
		return err
	}
	for _, columnTitle := range columnTitles {
		err = htm.write("<th>%s</th>", columnTitle)
		if err != nil {
			return err
		}
	}

	return htm.write("</tr>\n")
}

func (htm *HTMLRenderer) RenderRow(columnValues []reflect.Value) error {
	err := htm.write("<tr>\n")
	if err != nil {
		return err
	}

	for _, columnValue := range columnValues {
		str := txtfmt.FormatValue(columnValue, htm.config)

		// if the value does not have its own formatter, escape the resulting string
		_, derefType := reflection.DerefValueAndType(columnValue)
		if _, ok := htm.config.TypeFormatters[derefType]; !ok {
			str = html.EscapeString(str)
		}

		err := htm.write("<td>%s</td>", str)
		if err != nil {
			return err
		}
	}

	return htm.write("</tr>\n")
}

func (htm *HTMLRenderer) Result() ([]byte, error) {
	_, err := htm.buf.WriteString("</table>\n")
	if err != nil {
		return nil, err
	}
	return htm.buf.Bytes(), nil
}

func (htm *HTMLRenderer) WriteResultTo(writer io.Writer) error {
	_, err := htm.buf.WriteTo(writer)
	return err
}

func (htm *HTMLRenderer) WriteResultFile(file fs.File, perm ...fs.Permissions) error {
	writer, err := file.OpenWriter(perm...)
	if err != nil {
		return err
	}
	defer writer.Close()

	return htm.WriteResultTo(writer)
}

func (*HTMLRenderer) MIMEType() string {
	return "text/html; charset=UTF-8"
}

func (htm *HTMLRenderer) write(format string, a ...interface{}) error {
	_, err := fmt.Fprintf(&htm.buf, format, a...)
	return err
}
