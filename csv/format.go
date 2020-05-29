package csv

type Format struct {
	Encoding  string `json:"encoding"`
	Separator string `json:"separator"`
	Newline   string `json:"newline"`
}

func (f *Format) Valid() bool {
	return f != nil && f.Encoding != "" && f.Separator != "" && f.Newline != ""
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
			"Macintosh",
		},
		EncodingTests: []string{
			"ä",
			"ö",
			"ü",
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
