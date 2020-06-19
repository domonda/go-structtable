package structtable

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"reflect"
	"strconv"

	"github.com/domonda/go-types/strfmt"
	fs "github.com/ungerik/go-fs"
	reflection "github.com/ungerik/go-reflection"
)

// HTMLFormatRenderer is the renderer for the HTML format.
type HTMLFormatRenderer interface {
	RenderBeforeTable(writer io.Writer) error
	Caption() string
}

// HTMLRenderer implements Renderer by using a HTMLFormatRenderer
// for a specific text based table format.
type HTMLRenderer struct {
	format             HTMLFormatRenderer
	config             *TextFormatConfig
	typeFormatters     map[reflect.Type]TextFormatter
	buf                bytes.Buffer
}

func NewHTMLRenderer(format HTMLFormatRenderer, config *TextFormatConfig) *HTMLRenderer {
	return &HTMLRenderer{
		format: format,
		config: config,
	}
}

func (htm *HTMLRenderer) RenderHeaderRow(columnTitles []string) error {
	err := htm.format.RenderBeforeTable(&htm.buf)
	if err != nil {
		return err
	}

	err := htm.write("<table>\n")
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
		str, hasOwnFormatter := htm.toString(columnValue)
		// if the value does not have its own formatter, escape the resulting string
		if !hasOwnFormatter {
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

func (htm *HTMLRenderer) toString(val reflect.Value) (str string, hasOwnFormatter bool) {
	valType := val.Type()
	derefVal, derefType := reflection.DerefValueAndType(val)

	if f, ok := htm.config.TypeFormatters[derefType]; ok && derefVal.IsValid() {
		return f.FormatValue(derefVal, htm.config), true
	}

	switch valType.Kind() {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return htm.config.Nil, false
		}
	}

	switch derefType.Kind() {
	case reflect.Bool:
		if derefVal.Bool() {
			return htm.config.True, false
		}
		return htm.config.False, false

	case reflect.String:
		return derefVal.String(), false

	case reflect.Float32, reflect.Float64:
		return strfmt.FormatFloat(, false
			derefVal.Float(),
			htm.config.Float.ThousandsSep,
			htm.config.Float.DecimalSep,
			htm.config.Float.Precision,
			htm.config.Float.PadPrecision,
		)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(derefVal.Int(), 10), false

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(derefVal.Uint(), 10), false
	}

	if s, ok := val.Interface().(fmt.Stringer); ok {
		return s.String(), false
	}
	if s, ok := val.Addr().Interface().(fmt.Stringer); ok {
		return s.String(), false
	}
	if s, ok := derefVal.Interface().(fmt.Stringer); ok {
		return s.String(), false
	}

	switch x := derefVal.Interface().(type) {
	case []byte:
		return string(x), false
	}

	return fmt.Sprint(val.Interface()), false
}
