package csv

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/domonda/go-structtable"
	"github.com/domonda/go-types/charset"
	"github.com/domonda/go-types/strfmt"
)

var (
	doubleQuote       = []byte{'"'}
	doubleDoubleQuote = []byte{'"', '"'}
)

type Renderer struct {
	*structtable.TextRenderer

	headerComment  []byte
	delimiter      []byte
	quoteAllFields bool
	// quoteTextFields  bool
	quoteEmptyFields bool
	newLine          []byte
}

func NewRenderer(config *strfmt.FormatConfig) *Renderer {
	csv := &Renderer{
		headerComment:  nil,
		delimiter:      []byte{';'},
		quoteAllFields: false,
		// quoteTextFields:  false,
		quoteEmptyFields: false,
		newLine:          []byte{'\r', '\n'},
	}
	csv.TextRenderer = structtable.NewTextRenderer(csv, config)
	return csv
}

func (csv *Renderer) WithFormat(format *Format) *Renderer {
	csv.delimiter = []byte(format.Separator)
	csv.newLine = []byte(format.Newline)
	return csv
}

func (csv *Renderer) WithDelimiter(delimiter string) *Renderer {
	err := csv.SetDelimiter(delimiter)
	if err != nil {
		panic("err")
	}
	return csv
}

func (csv *Renderer) WithHeaderComment(headerSuffix string) *Renderer {
	if headerSuffix == "" {
		csv.headerComment = nil
	} else {
		csv.headerComment = []byte(headerSuffix)
	}
	return csv
}

func (csv *Renderer) WithQuoteAllFields(quote bool) *Renderer {
	csv.quoteAllFields = quote
	return csv
}

// func (csv *Renderer) WithQuoteTextFields(quote bool) *Renderer {
// 	csv.quoteTextFields = quote
// 	return csv
// }

func (csv *Renderer) WithQuoteEmptyFields(quote bool) *Renderer {
	csv.quoteEmptyFields = quote
	return csv
}

func (csv *Renderer) RenderBeginTableText(writer io.Writer) error {
	_, err := writer.Write([]byte(charset.BOMUTF8))
	return err
}

func (csv *Renderer) SetDelimiter(delimiter string) error {
	if delimiter == "" {
		return errors.New("empty delimiter not possible for CSV")
	}

	csv.delimiter = []byte(delimiter)
	return nil
}

func (csv *Renderer) RenderHeaderRowText(writer io.Writer, columnTitles []string) error {
	if len(csv.headerComment) > 0 {
		_, err := writer.Write(csv.headerComment)
		if err != nil {
			return err
		}
	}
	return csv.RenderRowText(writer, columnTitles)
}

func (csv *Renderer) RenderRowText(writer io.Writer, fields []string) error {
	for i, field := range fields {
		if i > 0 {
			_, err := writer.Write(csv.delimiter)
			if err != nil {
				return err
			}
		}

		mustQuote := csv.quoteAllFields || (csv.quoteEmptyFields && field == "") || strings.ContainsAny(field, "\"\n"+string(csv.delimiter))

		if mustQuote {
			_, err := writer.Write(doubleQuote)
			if err != nil {
				return err
			}
		}

		_, err := writer.Write(bytes.Replace([]byte(field), doubleQuote, doubleDoubleQuote, -1))
		if err != nil {
			return err
		}

		if mustQuote {
			_, err := writer.Write(doubleQuote)
			if err != nil {
				return err
			}
		}
	}

	_, err := writer.Write(csv.newLine)
	return err
}

func (*Renderer) RenderEndTableText(writer io.Writer) error {
	return nil
}

func (*Renderer) MIMEType() string {
	return "text/csv; charset=UTF-8"
}
