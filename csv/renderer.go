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

// NewRenderer creates a new CSV Renderer with default settings.
//
// This constructor creates a CSV renderer with sensible defaults:
// - Semicolon (;) as field delimiter
// - Windows line endings (\r\n)
// - No header comments
// - Selective field quoting (only when necessary)
// - UTF-8 BOM output for Excel compatibility
//
// Parameters:
//   - config: Text formatting configuration for cell values
//
// Returns:
//   - A new Renderer instance ready for use
//
// Example:
//
//	renderer := csv.NewRenderer(strfmt.NewEnglishFormatConfig())
//	renderer = renderer.WithDelimiter(",").WithQuoteAllFields(true)
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

// WithFormat configures the renderer with settings from a Format struct.
//
// This method applies the separator and newline settings from the provided Format
// configuration to the renderer. It's useful when you have a detected or predefined
// format that you want to use for rendering.
//
// Parameters:
//   - format: The Format configuration containing separator and newline settings
//
// Returns:
//   - The renderer instance for method chaining
//
// Example:
//
//	format := csv.NewFormat(",")
//	renderer = renderer.WithFormat(format)
func (csv *Renderer) WithFormat(format *Format) *Renderer {
	csv.delimiter = []byte(format.Separator)
	csv.newLine = []byte(format.Newline)
	return csv
}

// WithDelimiter sets the field delimiter for CSV output.
//
// This method sets the character used to separate fields in the CSV output.
// Common delimiters include comma (,), semicolon (;), and tab (\t).
// Note: This method panics if an empty delimiter is provided, as CSV requires
// a field separator character.
//
// Parameters:
//   - delimiter: The field separator character (must not be empty)
//
// Returns:
//   - The renderer instance for method chaining
//
// Panics:
//   - If delimiter is empty (CSV requires a field separator)
//
// Example:
//
//	renderer = renderer.WithDelimiter(",")   // Comma-separated
//	renderer = renderer.WithDelimiter(";")   // Semicolon-separated
//	renderer = renderer.WithDelimiter("\t")  // Tab-separated
func (csv *Renderer) WithDelimiter(delimiter string) *Renderer {
	err := csv.SetDelimiter(delimiter)
	if err != nil {
		panic("error setting delimiter: " + err.Error())
	}
	return csv
}

// WithHeaderComment adds a comment line before the header row.
//
// This method sets a comment string that will be written before the header row
// in the CSV output. This is useful for adding metadata, instructions, or
// other information to the CSV file. If an empty string is provided, no
// header comment will be written.
//
// Parameters:
//   - headerSuffix: The comment text to write before the header (empty string to disable)
//
// Returns:
//   - The renderer instance for method chaining
//
// Example:
//
//	renderer = renderer.WithHeaderComment("# Generated on 2024-01-15")
//	renderer = renderer.WithHeaderComment("") // Disable header comment
func (csv *Renderer) WithHeaderComment(headerSuffix string) *Renderer {
	if headerSuffix == "" {
		csv.headerComment = nil
	} else {
		csv.headerComment = []byte(headerSuffix)
	}
	return csv
}

// WithQuoteAllFields controls whether all fields should be quoted.
//
// This method sets whether all fields in the CSV output should be enclosed in
// double quotes, regardless of their content. When enabled, every field will
// be quoted, which can be useful for ensuring consistent formatting or
// compatibility with certain CSV parsers.
//
// Parameters:
//   - quote: True to quote all fields, false for selective quoting
//
// Returns:
//   - The renderer instance for method chaining
//
// Example:
//
//	renderer = renderer.WithQuoteAllFields(true)  // Quote everything
//	renderer = renderer.WithQuoteAllFields(false) // Quote only when necessary
func (csv *Renderer) WithQuoteAllFields(quote bool) *Renderer {
	csv.quoteAllFields = quote
	return csv
}

// func (csv *Renderer) WithQuoteTextFields(quote bool) *Renderer {
// 	csv.quoteTextFields = quote
// 	return csv
// }

// WithQuoteEmptyFields controls whether empty fields should be quoted.
//
// This method sets whether empty fields (empty strings) should be enclosed in
// double quotes in the CSV output. When enabled, empty fields will be written
// as "" instead of being left unquoted. This can be useful for maintaining
// consistent field positioning or indicating intentional empty values.
//
// Parameters:
//   - quote: True to quote empty fields, false to leave them unquoted
//
// Returns:
//   - The renderer instance for method chaining
//
// Example:
//
//	renderer = renderer.WithQuoteEmptyFields(true)  // Empty fields as ""
//	renderer = renderer.WithQuoteEmptyFields(false) // Empty fields as empty
func (csv *Renderer) WithQuoteEmptyFields(quote bool) *Renderer {
	csv.quoteEmptyFields = quote
	return csv
}

// RenderBeginTableText writes the UTF-8 BOM at the beginning of the CSV file.
//
// This method writes the UTF-8 Byte Order Mark (BOM) at the start of the CSV output.
// The BOM helps Excel and other applications correctly identify the file as UTF-8
// encoded, especially when dealing with international characters.
//
// Parameters:
//   - writer: The io.Writer to write the BOM to
//
// Returns:
//   - err: Any error that occurred during writing
//
// Note:
//   - The BOM is only written once at the beginning of the file
//   - This helps with Excel compatibility for international characters
func (csv *Renderer) RenderBeginTableText(writer io.Writer) error {
	_, err := writer.Write([]byte(charset.BOMUTF8))
	return err
}

// SetDelimiter sets the field delimiter with proper error handling.
//
// This method sets the character used to separate fields in the CSV output,
// with validation to ensure the delimiter is not empty. Unlike WithDelimiter,
// this method returns an error instead of panicking, making it safer for
// programmatic use where errors should be handled gracefully.
//
// Parameters:
//   - delimiter: The field separator character (must not be empty)
//
// Returns:
//   - err: An error if the delimiter is empty or invalid
//
// Example:
//
//	err := renderer.SetDelimiter(",")
//	if err != nil {
//	    return err
//	}
func (csv *Renderer) SetDelimiter(delimiter string) error {
	if delimiter == "" {
		return errors.New("empty delimiter not possible for CSV")
	}

	csv.delimiter = []byte(delimiter)
	return nil
}

// RenderHeaderRowText renders a header row with optional header comment.
//
// This method renders the CSV header row, optionally preceded by a header comment
// if one was configured. The header comment is written first (if present), followed
// by the column titles formatted as a regular CSV row.
//
// Parameters:
//   - writer: The io.Writer to write the header to
//   - columnTitles: Slice of strings representing the column headers
//
// Returns:
//   - err: Any error that occurred during writing
//
// Example:
//
//	headers := []string{"Name", "Age", "City"}
//	err := renderer.RenderHeaderRowText(writer, headers)
func (csv *Renderer) RenderHeaderRowText(writer io.Writer, columnTitles []string) error {
	if len(csv.headerComment) > 0 {
		_, err := writer.Write(csv.headerComment)
		if err != nil {
			return err
		}
	}
	return csv.RenderRowText(writer, columnTitles)
}

// RenderRowText renders a data row with proper CSV formatting and quoting.
//
// This method renders a single CSV row with intelligent field quoting based on
// the configured quoting rules. Fields are quoted when:
// - quoteAllFields is enabled, OR
// - quoteEmptyFields is enabled and the field is empty, OR
// - the field contains quotes, newlines, or the delimiter character
//
// Parameters:
//   - writer: The io.Writer to write the row to
//   - fields: Slice of strings representing the field values
//
// Returns:
//   - err: Any error that occurred during writing
//
// Quoting Logic:
//   - Escaped quotes within fields are doubled ("" becomes """")
//   - Fields are separated by the configured delimiter
//   - Row ends with the configured newline sequence
//
// Example:
//
//	fields := []string{"John Doe", "25", "New York, NY"}
//	err := renderer.RenderRowText(writer, fields)
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

// RenderEndTableText performs cleanup after all rows have been rendered.
//
// This method is called after all CSV rows have been rendered to perform any
// necessary cleanup operations. For CSV rendering, no special cleanup is required,
// so this method simply returns nil.
//
// Parameters:
//   - writer: The io.Writer (unused for CSV cleanup)
//
// Returns:
//   - err: Always returns nil (no cleanup needed for CSV)
//
// Note:
//   - This method is part of the TextRenderer interface
//   - CSV files don't require special end-of-file markers
func (*Renderer) RenderEndTableText(writer io.Writer) error {
	return nil
}

// MIMEType returns the MIME type for CSV files.
//
// This method returns the standard MIME type for CSV files with UTF-8 charset
// specification. This is used for HTTP content-type headers and file type
// identification in web applications and file systems.
//
// Returns:
//   - string: The MIME type "text/csv; charset=UTF-8"
//
// Example:
//
//	contentType := renderer.MIMEType()
//	w.Header().Set("Content-Type", contentType)
func (*Renderer) MIMEType() string {
	return "text/csv; charset=UTF-8"
}
