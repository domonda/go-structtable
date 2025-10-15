package structtable

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"reflect"
	"strings"

	"github.com/ungerik/go-fs"

	"github.com/domonda/go-types/strfmt"
)

// HTMLFormatRenderer is the renderer for the HTML format.
//
// This interface defines methods for customizing HTML table rendering,
// particularly for adding custom styles or content before the table element.
type HTMLFormatRenderer interface {
	// RenderBeforeTable is useful when you want to add custom styles or render anything before the table element.
	//
	// This method is called before the table element is rendered, allowing
	// you to add custom CSS styles, scripts, or other HTML content.
	RenderBeforeTable(writer io.Writer) error
}

// HTMLTableConfig is the config for the actual, visual, resulting HTML table.
//
// This struct contains configuration options for customizing the appearance
// and structure of the generated HTML table, including CSS classes and captions.
type HTMLTableConfig struct {
	// Caption is the table caption text.
	Caption string
	// TableClass is the CSS class for the table element.
	TableClass string
	// CaptionClass is the CSS class for the caption element.
	CaptionClass string
	// RowClass is the CSS class applied to all table rows.
	RowClass string
	// CellClass is the CSS class applied to all table cells.
	CellClass string
	// HeaderRowClass is the CSS class for header rows.
	HeaderRowClass string
	// HeaderCellClass is the CSS class for header cells.
	HeaderCellClass string
	// DataRowClass is the CSS class for data rows.
	DataRowClass string
	// DataCellClass is the CSS class for data cells.
	DataCellClass string
}

// HTMLRenderer implements Renderer by using a HTMLFormatRenderer
// for a specific text based table format.
//
// This renderer generates HTML tables from struct slices, with support
// for custom formatting, CSS classes, and pre-table content rendering.
type HTMLRenderer struct {
	format      HTMLFormatRenderer
	TableConfig *HTMLTableConfig
	txtConfig   *strfmt.FormatConfig
	buf         bytes.Buffer
}

// NewHTMLRenderer creates a new HTMLRenderer instance.
//
// This constructor initializes an HTMLRenderer with the provided format renderer,
// table configuration, and text formatting configuration.
//
// Parameters:
//   - format: The HTMLFormatRenderer for custom pre-table content
//   - TableConfig: Configuration for the HTML table appearance
//   - config: Text formatting configuration for cell values
//
// Returns:
//   - A new HTMLRenderer instance ready for use
func NewHTMLRenderer(format HTMLFormatRenderer, TableConfig *HTMLTableConfig, config *strfmt.FormatConfig) *HTMLRenderer {
	return &HTMLRenderer{format: format, TableConfig: TableConfig, txtConfig: config}
}

// RenderHeaderRow renders the table header row with the given column titles.
//
// This method implements the Renderer interface and generates the HTML
// for the table header row, including the opening table tag and caption.
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
			err = htm.write("<caption class='%s'>%s</caption>\n", html.EscapeString(htm.TableConfig.CaptionClass), caption)
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
			err = htm.write("<th class='%s'>%s</th>", strings.TrimSpace(htm.TableConfig.HeaderCellClass+" "+htm.TableConfig.CellClass), columnTitle)
		} else {
			err = htm.write("<th>%s</th>", columnTitle)
		}
		if err != nil {
			return err
		}
	}

	return htm.write("</tr>\n")
}

// RenderRow renders a single data row with the given column values.
//
// This method implements the Renderer interface and generates the HTML
// for a single table row with the provided column values.
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
		str := strfmt.FormatValue(columnValue, htm.txtConfig)

		// if the value does not have its own formatter, escape the resulting string
		derefType := columnValue.Type()
		for derefType.Kind() == reflect.Ptr {
			derefType = derefType.Elem()
		}
		if htm.txtConfig.TypeFormatters[derefType] == nil {
			str = html.EscapeString(str)
		}

		if htm.TableConfig.DataCellClass != "" || htm.TableConfig.CellClass != "" {
			err = htm.write("<td class='%s'>%s</td>", strings.TrimSpace(htm.TableConfig.DataCellClass+" "+htm.TableConfig.CellClass), str)
		} else {
			err = htm.write("<td>%s</td>", str)
		}
		if err != nil {
			return err
		}
	}

	return htm.write("</tr>\n")
}

// Result returns the rendered table data as bytes.
//
// This method implements the Renderer interface and returns the complete
// HTML table as a byte slice, including the closing table tag.
func (htm *HTMLRenderer) Result() ([]byte, error) {
	_, err := htm.buf.WriteString("</tbody></table>\n")
	if err != nil {
		return nil, err
	}
	return htm.buf.Bytes(), nil
}

// WriteResultTo writes the rendered table data to the given writer.
//
// This method implements the Renderer interface and writes the complete
// HTML table to the provided writer.
func (htm *HTMLRenderer) WriteResultTo(writer io.Writer) error {
	_, err := htm.buf.WriteTo(writer)
	return err
}

// WriteResultFile writes the rendered table data to the given file.
//
// This method implements the Renderer interface and writes the complete
// HTML table to the provided file with optional permissions.
func (htm *HTMLRenderer) WriteResultFile(file fs.File, perm ...fs.Permissions) error {
	writer, err := file.OpenWriter(perm...)
	if err != nil {
		return err
	}
	defer writer.Close()

	return htm.WriteResultTo(writer)
}

// MIMEType returns the MIME-Type of the rendered content.
//
// This method implements the Renderer interface and returns the MIME type
// for HTML content with UTF-8 encoding.
func (*HTMLRenderer) MIMEType() string {
	return "text/html; charset=UTF-8"
}

func (htm *HTMLRenderer) write(format string, a ...any) error {
	_, err := fmt.Fprintf(&htm.buf, format, a...)
	return err
}
