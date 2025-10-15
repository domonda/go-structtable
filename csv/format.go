package csv

import (
	"errors"
	"fmt"
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
	// Newline specifies the line ending format ("\n", "\r\n", "\r").
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
// checking for required fields and valid values.
//
// Returns:
//   - err: An error describing any validation failures, or nil if valid
func (f *Format) Validate() error {
	switch {
	case f == nil:
		return errors.New("<nil> csv.Format")
	case f.Encoding == "":
		return errors.New("missing csv.Format.Encoding")
	case f.Separator == "":
		return errors.New("missing csv.Format.Separator")
	case len(f.Separator) > 1:
		return fmt.Errorf("invalid csv.Format.Separator: %q", f.Separator)
	case f.Newline == "":
		return errors.New("missing csv.Format.Newline")
	case f.Newline != "\n" && f.Newline != "\n\r" && f.Newline != "\r\n":
		return fmt.Errorf("invalid csv.Format.Newline: %q", f.Newline)
	}
	return nil
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
