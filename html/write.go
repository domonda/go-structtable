package html

import (
	"io"
	"strings"

	"github.com/domonda/go-structtable"
)

// WriteTable is a shortcut to write a HTML table with english text formating and reflected column titles.
// The optional columnTitleTag strings will be merged into one string,
// where an empty string means using the struct field names.
func WriteTable(destination io.Writer, structSlice interface{}, columnTitleTag ...string) error {
	writer := NewWriter(structtable.NewEnglishTextFormatConfig())
	return structtable.WriteToReflectColumnTitles(destination, writer, structSlice, strings.Join(columnTitleTag, ""))
}
