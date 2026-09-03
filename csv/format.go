package csv

import (
	"github.com/domonda/go-errs"
)

// Format represents the configuration for CSV file parsing and writing.
//
// This struct contains the essential parameters needed to correctly parse
// and write CSV files, including encoding, separator, and newline settings.
type Format struct {
	// Encoding specifies the character encoding of the CSV file.
	// Common values include "UTF-8", "UTF-16LE", "ISO 8859-1", etc.
	Encoding string `json:"encoding"`
	// Separator is the field separator character (e.g., ",", ";", "\t").
	Separator string `json:"separator"`
	// Newline specifies the line ending format ("\n", "\r\n", "\n\r").
	Newline string `json:"newline"`
}

// NewFormat returns a Format with the passed separator,
// UTF-8 encoding, and "\r\n" newlines.
//
// This is a convenience constructor for creating a standard CSV format
// configuration with the most common settings.
//
// Parameters:
//   - separator: The field separator character (e.g., ",", ";", "\t")
//
// Returns:
//   - A new Format instance with UTF-8 encoding and Windows line endings
func NewFormat(separator string) *Format {
	return &Format{
		Encoding:  "UTF-8",
		Separator: separator,
		Newline:   "\r\n",
	}
}

// Validate returns an error in case of an invalid Format.
// Can be called on nil receiver.
//
// This method performs comprehensive validation of the Format configuration,
// checking for required fields and valid values:
//   - Encoding must not be empty
//   - Separator must be exactly one character long,
//     and must not be a quote or a control character other than tab
//   - Newline must be one of "\n", "\r\n" or "\n\r"
//
// Returns:
//   - err: An error describing any validation failures, or nil if valid
func (f *Format) Validate() error {
	switch {
	case f == nil:
		return errs.New("<nil> csv.Format")
	case f.Encoding == "":
		return errs.New("missing csv.Format.Encoding")
	case f.Separator == "":
		return errs.New("missing csv.Format.Separator")
	case len(f.Separator) > 1:
		return errs.Errorf("invalid csv.Format.Separator: %q", f.Separator)
	case !validSeparator(f.Separator[0]):
		return errs.Errorf("csv.Format.Separator must not be a quote or a control character other than tab: %q", f.Separator)
	case f.Newline == "":
		return errs.New("missing csv.Format.Newline")
	case f.Newline != "\n" && f.Newline != "\n\r" && f.Newline != "\r\n":
		return errs.Errorf("invalid csv.Format.Newline: %q", f.Newline)
	}
	return nil
}

// asciiDEL is the delete control character, the only one above the space.
const asciiDEL = 0x7f

// validSeparator reports whether c can be used as a CSV field separator.
// A quote can never be one because it would make all quote handling
// in readLines nonsensical, and control characters other than tab
// can never be one either.
func validSeparator(c byte) bool {
	return c != '"' && c != asciiDEL && (c >= ' ' || c == '\t')
}

// FormatDetectionConfig contains configuration for automatic CSV format detection.
//
// This struct provides settings for detecting CSV format parameters automatically
// from file content, including supported encodings and test strings for validation.
type FormatDetectionConfig struct {
	// Encodings is a list of character encodings to try during detection.
	Encodings []string `json:"encodings"`
	// EncodingTests contains test strings used to validate encoding detection.
	EncodingTests []string `json:"encodingTests"`
}

// NewFormatDetectionConfig creates a new FormatDetectionConfig with default settings.
//
// This constructor provides a sensible default configuration for CSV format detection,
// including common encodings and test strings for various languages and character sets.
//
// Returns:
//   - A new FormatDetectionConfig instance with default settings
func NewFormatDetectionConfig() *FormatDetectionConfig {
	return &FormatDetectionConfig{
		Encodings: []string{
			"UTF-8",
			"UTF-16LE",
			"ISO 8859-1",
			"Windows 1252", // like ANSI
			"Macintosh",
		},
		EncodingTests: []string{
			"ä",
			"Ä",
			"ö",
			"Ö",
			"ü",
			"Ü",
			"ß",
			"§",
			"€",
			"д",
			"Д",
			"ъ",
			"Ъ",
			"б",
			"Б",
			"л",
			"Л",
			"и",
			"И",
			"ж",
			// "ährung",
			// "mpfänger",
			// "rsprünglich",
			// "ückerstatt",
			// "übertrag",
			// "für",
			// "Jänner",
			// "März",
			// "cc§google.com",
		},
	}
}
