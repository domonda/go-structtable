package csv

import (
	"errors"
	"fmt"
)

type Format struct {
	Encoding  string `json:"encoding"`
	Separator string `json:"separator"`
	Newline   string `json:"newline"`
}

// Validate returns an error in case of an invalid Format.
// Can be called on nil receiver.
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

type FormatDetectionConfig struct {
	Encodings     []string `json:"encodings"`
	EncodingTests []string `json:"encodingTests"`
}

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
			"ö",
			"ü",
			"ß",
			"Ä",
			"Ö",
			"Ü",
			"§",
			"€",
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
