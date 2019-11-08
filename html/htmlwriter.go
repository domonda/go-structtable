package html

import (
	"fmt"
	"html"
	"io"
	"math/rand"

	"github.com/domonda/go-structtable"
)

var (
	// EvenTableRowStyle used by WriteHTML
	EvenTableRowStyle = "background:#EEF"
	// OddTableRowStyle used by WriteHTML
	OddTableRowStyle = "background:#FFF"
)

type Writer struct {
	*structtable.TextWriter
	numRowsWritten int
	elemClass      string
}

func NewWriter(config *structtable.TextFormatConfig) *Writer {
	h := &Writer{}
	h.TextWriter = structtable.NewTextWriter(h, config)
	return h
}

func (h *Writer) WriteBeginTableText(writer io.Writer) error {
	h.elemClass = fmt.Sprintf("t%d", rand.Uint32())
	_, err := fmt.Fprintf(writer, `<style>table.%[1]s, td.%[1]s, th.%[1]s { border:1px solid black; padding: 4px; white-space: nowrap; font-family: "Lucida Console", Monaco, monospace; }</style>`, h.elemClass)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "<table class='%s' style='border-collapse:collapse'>\n", h.elemClass)
	return err
}

func (h *Writer) WriteHeaderRowText(writer io.Writer, columnTitles []string) error {
	return h.writeRowText(writer, columnTitles, "th")
}

func (h *Writer) WriteRowText(writer io.Writer, fields []string) error {
	return h.writeRowText(writer, fields, "td")
}

func (h *Writer) writeRowText(writer io.Writer, fields []string, fieldTag string) error {
	var rowStyle string
	if h.numRowsWritten%2 == 0 {
		rowStyle = EvenTableRowStyle
	} else {
		rowStyle = OddTableRowStyle
	}
	_, err := fmt.Fprintf(writer, "<tr class='%s' style='%s'>", h.elemClass, rowStyle)
	if err != nil {
		return err
	}
	for _, field := range fields {
		_, err = fmt.Fprintf(writer, "<%s class='%s'>%s</%s>", fieldTag, h.elemClass, html.EscapeString(field), fieldTag)
		if err != nil {
			return err
		}
	}
	_, err = writer.Write([]byte("</tr>\n"))
	if err != nil {
		return err
	}
	h.numRowsWritten++
	return nil
}

func (*Writer) WriteEndTableText(writer io.Writer) error {
	_, err := writer.Write([]byte("</table>\n"))
	return err
}
