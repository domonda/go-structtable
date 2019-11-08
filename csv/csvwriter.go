package csv

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/domonda/go-structtable"
	"github.com/domonda/go-types/charset"
)

var (
	doubleQuote       = []byte{'"'}
	doubleDoubleQuote = []byte{'"', '"'}
)

type Writer struct {
	*structtable.TextWriter
	headerComment []byte
	delimiter     []byte
	quoteFields   bool
	newLine       []byte
}

func NewWriter(config *structtable.TextFormatConfig) *Writer {
	csv := &Writer{
		headerComment: nil,
		delimiter:     []byte{';'},
		quoteFields:   false,
		newLine:       []byte{'\r', '\n'},
	}
	csv.TextWriter = structtable.NewTextWriter(csv, config)
	return csv
}

func (csv *Writer) SetDelimiter(delimiter string) error {
	if delimiter == "" {
		return errors.New("empty delimiter not possible for CSV")
	}

	csv.delimiter = []byte(delimiter)
	return nil
}

func (csv *Writer) SetHeaderComment(headerSuffix string) {
	if headerSuffix == "" {
		csv.headerComment = nil
	} else {
		csv.headerComment = []byte(headerSuffix)
	}
}

func (csv *Writer) SetQuoteFields(quoteFields bool) {
	csv.quoteFields = quoteFields
}

func (*Writer) WriteBeginTableText(writer io.Writer) error {
	_, err := writer.Write([]byte(charset.BOMUTF8))
	return err
}

func (csv *Writer) WriteHeaderRowText(writer io.Writer, columnTitles []string) error {
	if len(csv.headerComment) > 0 {
		_, err := writer.Write(csv.headerComment)
		if err != nil {
			return err
		}
	}
	return csv.WriteRowText(writer, columnTitles)
}

func (csv *Writer) WriteRowText(writer io.Writer, fields []string) error {
	for i, field := range fields {
		if i > 0 {
			_, err := writer.Write(csv.delimiter)
			if err != nil {
				return err
			}
		}

		mustQuote := csv.quoteFields || strings.ContainsAny(field, "\"\n"+string(csv.delimiter))

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

func (*Writer) WriteEndTableText(writer io.Writer) error {
	return nil
}
