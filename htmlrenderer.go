package structtable

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"reflect"
	"strings"

	"github.com/domonda/go-types/txtfmt"
	fs "github.com/ungerik/go-fs"
	"github.com/ungerik/go-reflection"
)

// HTMLFormatRenderer is the renderer for the HTML format.
type HTMLFormatRenderer interface {
	// RenderBeforeTable is useful when you want to add custom styles or render anything before the table element.
	RenderBeforeTable(writer io.Writer) error
}

// HTMLTableConfig is the config for the actual, visual, resulting HTML table.
type HTMLTableConfig struct {
	Caption         string
	TableClass      string
	CaptionClass    string
	RowClass        string
	CellClass       string
	HeaderRowClass  string
	HeaderCellClass string
	DataRowClass    string
	DataCellClass   string
}

// HTMLRenderer implements Renderer by using a HTMLFormatRenderer
// for a specific text based table format.
type HTMLRenderer struct {
	format      HTMLFormatRenderer
	TableConfig *HTMLTableConfig
	config      *txtfmt.FormatConfig
	buf         bytes.Buffer
}

func NewHTMLRenderer(format HTMLFormatRenderer, TableConfig *HTMLTableConfig, config *txtfmt.FormatConfig) *HTMLRenderer {
	return &HTMLRenderer{format: format, TableConfig: TableConfig, config: config}
}

func (htm *HTMLRenderer) RenderHeaderRow(columnTitles []string) error {
	err := htm.format.RenderBeforeTable(&htm.buf)
	if err != nil {
		return err
	}

	if htm.TableConfig.TableClass != "" {
		err = htm.write("<table class='%s'><tbody>\n", html.EscapeString(htm.TableConfig.TableClass))
	} else {
		err = htm.write("<table><tbody>\n")
	}
	if err != nil {
		return err
	}
	caption := htm.TableConfig.Caption
	if caption != "" {
		if htm.TableConfig.CaptionClass != "" {
			err = htm.write("<caption class='%s'>%s</caption>\n", caption, html.EscapeString(htm.TableConfig.CaptionClass))
		} else {
			err = htm.write("<caption>%s</caption>\n", caption)
		}
		if err != nil {
			return err
		}
	}
	if htm.TableConfig.HeaderRowClass != "" || htm.TableConfig.RowClass != "" {
		err = htm.write("<tr class='%s'>\n", strings.TrimSpace(htm.TableConfig.HeaderRowClass+" "+htm.TableConfig.RowClass))
	} else {
		err = htm.write("<tr>\n")
	}
	if err != nil {
		return err
	}
	for _, columnTitle := range columnTitles {
		if htm.TableConfig.HeaderCellClass != "" || htm.TableConfig.CellClass != "" {
			err = htm.write("<th class='%s'>%s</th>", columnTitle, strings.TrimSpace(htm.TableConfig.HeaderCellClass+" "+htm.TableConfig.CellClass))
		} else {
			err = htm.write("<th>%s</th>", columnTitle)
		}
		if err != nil {
			return err
		}
	}

	return htm.write("</tr>\n")
}

func (htm *HTMLRenderer) RenderRow(columnValues []reflect.Value) error {
	var err error
	if htm.TableConfig.DataRowClass != "" || htm.TableConfig.RowClass != "" {
		err = htm.write("<tr class='%s'>\n", strings.TrimSpace(htm.TableConfig.DataRowClass+" "+htm.TableConfig.RowClass))
	} else {
		err = htm.write("<tr>\n")
	}
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

		if htm.TableConfig.DataCellClass != "" || htm.TableConfig.CellClass != "" {
			err = htm.write("<td class='%s'>%s</td>", str, strings.TrimSpace(htm.TableConfig.DataCellClass+" "+htm.TableConfig.CellClass))
		} else {
			err = htm.write("<td>%s</td>", str)
		}
		if err != nil {
			return err
		}
	}

	return htm.write("</tr>\n")
}

func (htm *HTMLRenderer) Result() ([]byte, error) {
	_, err := htm.buf.WriteString("</tbody></table>\n")
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
