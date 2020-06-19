package htmltable

import (
	"fmt"
	"html"
	"io"
	"math/rand"

	"github.com/domonda/go-structtable"
	"github.com/domonda/go-types/txtfmt"
)

var (
	// EvenTableRowStyle used by WriteHTML
	EvenTableRowStyle = "background:#EEF"
	// OddTableRowStyle used by WriteHTML
	OddTableRowStyle = "background:#FFF"
)

// Render is a shortcut to render a HTML table with english text formating
func Render(writer io.Writer, structSlice interface{}, renderTitleRow bool, columnMapper structtable.ColumnMapper) error {
	renderer := NewRenderer(txtfmt.NewEnglishFormatConfig())
	return structtable.RenderTo(writer, renderer, structSlice, renderTitleRow, columnMapper)
}

type Renderer struct {
	*structtable.TextRenderer
	numRowsWritten int
	elemClass      string
}

func NewRenderer(config *txtfmt.FormatConfig) *Renderer {
	r := &Renderer{}
	r.TextRenderer = structtable.NewTextRenderer(r, config)
	return r
}

func (r *Renderer) RenderBeginTableText(writer io.Writer) error {
	r.elemClass = fmt.Sprintf("t%d", rand.Uint32())
	_, err := fmt.Fprintf(writer, `<style>table.%[1]s, td.%[1]s, th.%[1]s { border:1px solid black; padding: 4px; white-space: nowrap; font-family: "Lucida Console", Monaco, monospace; }</style>`, r.elemClass)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "<table class='%s' style='border-collapse:collapse'>\n", r.elemClass)
	return err
}

func (r *Renderer) RenderHeaderRowText(writer io.Writer, columnTitles []string) error {
	return r.writeRowText(writer, columnTitles, "th")
}

func (r *Renderer) RenderRowText(writer io.Writer, fields []string) error {
	return r.writeRowText(writer, fields, "td")
}

func (r *Renderer) writeRowText(writer io.Writer, fields []string, fieldTag string) error {
	var rowStyle string
	if r.numRowsWritten%2 == 0 {
		rowStyle = EvenTableRowStyle
	} else {
		rowStyle = OddTableRowStyle
	}
	_, err := fmt.Fprintf(writer, "<tr class='%s' style='%s'>", r.elemClass, rowStyle)
	if err != nil {
		return err
	}
	for _, field := range fields {
		_, err = fmt.Fprintf(writer, "<%s class='%s'>%s</%s>", fieldTag, r.elemClass, html.EscapeString(field), fieldTag)
		if err != nil {
			return err
		}
	}
	_, err = writer.Write([]byte("</tr>\n"))
	if err != nil {
		return err
	}
	r.numRowsWritten++
	return nil
}

func (*Renderer) RenderEndTableText(writer io.Writer) error {
	_, err := writer.Write([]byte("</table>\n"))
	return err
}

func (*Renderer) MIMEType() string {
	return "text/html; charset=UTF-8"
}
