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
	RenderBeforeTable(writer io.Writer) error
	Caption() string
}

// HTMLTableConfig is the config for the actual, visual, resulting HTML table.
type HTMLTableConfig struct {
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
	config      *txtfmt.FormatConfig
	tableConfig *HTMLTableConfig
	buf         bytes.Buffer
}

func NewHTMLRenderer(format HTMLFormatRenderer, config *txtfmt.FormatConfig, tableConfig *HTMLTableConfig) *HTMLRenderer {
	return &HTMLRenderer{format: format, config: config, tableConfig: tableConfig}
}

func (htm *HTMLRenderer) RenderHeaderRow(columnTitles []string) error {
	err := htm.format.RenderBeforeTable(&htm.buf)
	if err != nil {
		return err
	}

	if htm.tableConfig.TableClass != "" {
		err = htm.write("<table class='%s'>\n", html.EscapeString(htm.tableConfig.TableClass))
	} else {
		err = htm.write("<table>\n")
	}
	if err != nil {
		return err
	}
	caption := htm.format.Caption()
	if caption != "" {
		if htm.tableConfig.CaptionClass != "" {
			err = htm.write("<caption class='%s'>%s</caption>\n", caption, html.EscapeString(htm.tableConfig.CaptionClass))
		} else {
			err = htm.write("<caption>%s</caption>\n", caption)
		}
		if err != nil {
			return err
		}
	}
	if htm.tableConfig.HeaderRowClass != "" || htm.tableConfig.RowClass != "" {
		err = htm.write("<tr class='%s'>\n", strings.TrimSpace(htm.tableConfig.HeaderRowClass+" "+htm.tableConfig.RowClass))
	} else {
		err = htm.write("<tr>\n")
	}
	if err != nil {
		return err
	}
	for _, columnTitle := range columnTitles {
		if htm.tableConfig.HeaderCellClass != "" || htm.tableConfig.CellClass != "" {
			err = htm.write("<th class='%s'>%s</th>", columnTitle, strings.TrimSpace(htm.tableConfig.HeaderCellClass+" "+htm.tableConfig.CellClass))
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
	if htm.tableConfig.DataRowClass != "" || htm.tableConfig.RowClass != "" {
		err = htm.write("<tr class='%s'>\n", strings.TrimSpace(htm.tableConfig.DataRowClass+" "+htm.tableConfig.RowClass))
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

		if htm.tableConfig.DataCellClass != "" || htm.tableConfig.CellClass != "" {
			err = htm.write("<td class='%s'>%s</td>", str, strings.TrimSpace(htm.tableConfig.DataCellClass+" "+htm.tableConfig.CellClass))
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
