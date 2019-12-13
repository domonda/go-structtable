package csv

import (
	"bytes"

	"github.com/ungerik/go-fs"

	"github.com/domonda/go-types/charset"
	"github.com/domonda/go-wraperr"
)

// FileParseStringsDetectFormat returns a slice of strings per row with the format detected via the FormatDetectionConfig.
func FileParseStringsDetectFormat(csvFile fs.File, config *FormatDetectionConfig) (rows [][]string, format *Format, err error) {
	defer wraperr.WithFuncParams(&err, csvFile, config)

	data, err := csvFile.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	return ParseStringsDetectFormat(data, config)
}

// ParseStringsDetectFormat returns a slice of strings per row with the format detected via the FormatDetectionConfig.
func ParseStringsDetectFormat(data []byte, config *FormatDetectionConfig) (rows [][]string, format *Format, err error) {
	defer wraperr.WithFuncParams(&err, data, config)

	format, lines, err := detectFormat(data, config)
	if err != nil {
		return nil, format, err
	}

	rows, err = readLines(lines, []byte(format.Separator), "\n")
	return rows, format, err
}

func ParseStringsWithFormat(data []byte, format *Format) (rows [][]string, err error) {
	defer wraperr.WithFuncParams(&err, data, format)

	lines := bytes.Split(data, []byte(format.Newline))
	return readLines(lines, []byte(format.Separator), "\n")
}

func detectFormat(data []byte, config *FormatDetectionConfig) (format *Format, lines [][]byte, err error) {
	defer wraperr.WithFuncParams(&err, len(data), config)

	var encodings []charset.Encoding
	for _, name := range config.Encodings {
		enc, err := charset.GetEncoding(name)
		if err != nil {
			return nil, nil, err
		}
		encodings = append(encodings, enc)
	}

	data, encoding, err := charset.AutoDecode(data, encodings, config.EncodingTests)
	if err != nil {
		return nil, nil, err
	}
	if encoding == "" {
		encoding = "UTF-8"
	}

	format = &Format{
		Encoding: encoding,
	}

	// data = strutil.SanitizeLineEndingsBytes(data)

	type sepCounts struct {
		commas     int
		semicolons int
		tabs       int
	}

	var (
		sep sepCounts
		// lineSepCounts  []sepCounts
		numNonEmptyLines int

		numLinesR  = bytes.Count(data, []byte{'\r'})
		numLinesN  = bytes.Count(data, []byte{'\n'})
		numLinesRN = bytes.Count(data, []byte{'\r', '\n'})
	)

	// fmt.Println("n:", numLinesN, "rn:", numLinesRN, "r:", numLinesR)

	if numLinesR > numLinesN {
		format.Newline = "\r"
	} else if numLinesN > numLinesRN {
		format.Newline = "\n"
	} else {
		format.Newline = "\r\n"
	}

	lines = bytes.Split(data, []byte(format.Newline))

	for i := range lines {
		lines[i] = bytes.Trim(lines[i], "\r\n")
		line := lines[i]

		if len(line) == 0 {
			continue
		}

		numNonEmptyLines++

		commas := bytes.Count(line, []byte{','})
		semicolons := bytes.Count(line, []byte{';'})
		tabs := bytes.Count(line, []byte{'\t'})

		sep.commas += commas
		sep.semicolons += semicolons
		sep.tabs += tabs
		// lineSepCounts = append(lineSepCounts, sepCounts{
		// 	commas:     commas,
		// 	semicolons: semicolons,
		// 	tabs:       tabs,
		// })
	}

	if numNonEmptyLines == 0 {
		return format, nil, nil
	}

	switch {
	case sep.commas > sep.semicolons && sep.commas > sep.tabs:
		format.Separator = ","
	case sep.semicolons > sep.commas && sep.semicolons > sep.tabs:
		format.Separator = ";"
	case sep.tabs > sep.commas && sep.tabs > sep.semicolons:
		format.Separator = "\t"
	default:
		format.Separator = ","
	}

	return format, lines, nil
}

func readLines(lines [][]byte, separator []byte, newlineReplacement string) (rows [][]string, err error) {
	defer wraperr.WithFuncParams(&err, lines, separator, newlineReplacement)

	rows = make([][]string, len(lines))
	for lineIndex := range lines {
		if len(lines[lineIndex]) == 0 {
			continue
		}

		fields := bytes.Split(lines[lineIndex], separator)
		for i := 0; i < len(fields); i++ {
			field := fields[i]
			if len(field) < 2 {
				continue
			}

			var (
				left  = field[0]
				right = field[len(field)-1]
			)
			switch {
			case left == '"' && right == '"':
				// Quoted field
				field = field[1 : len(field)-1]

			case left != '"' && right != '"':
				// Unquoted field

			case left == '"' && right != '"':
				// Field begins with quote but does not end with one
				if field[1] == '"' && (len(field) <= 2 || field[2] != '"') {
					// Begins with two quotes wich is an escaped quote,
					// but not with a tripple quote.
					// No special handling needed, will be unescaped futher down
				} else {

					joinLineIndex := -1
					if i == len(fields)-1 {
						// When last field of the line begins with a quote but does not end with one
						// then search following lines for a first field that ends with a quote
						// which will be the right side of this field wrongly splitted into more
						// lines because it contained newline characters.
						// Newlines are allowed in quoted CSV fields.
						for joinLineIndex = lineIndex + 1; joinLineIndex < len(lines); joinLineIndex++ {
							joinLine := lines[joinLineIndex]
							joinLineFields := bytes.Split(joinLine, separator)
							if len(joinLineFields) > 0 && joinLineFields[0][len(joinLineFields[0])-1] == '"' {
								// Found the line where the first field holds the closing quote for the multi line field
								break
							}
						}
					}

					if joinLineIndex > lineIndex && joinLineIndex < len(lines) {
						// Join lines until including joinLineIndex as multi line field
						// then empty those lines so line indices are still correct

						joinLine := lines[joinLineIndex]
						joinLineFields := bytes.Split(joinLine, separator)

						// Join lines between lineIndex and joinLineIndex
						for index := lineIndex + 1; index < joinLineIndex; index++ {
							field = append(field, []byte(newlineReplacement)...)
							field = append(field, lines[index]...)
						}

						// Join first field of line joinLineIndex
						field = append(field, []byte(newlineReplacement)...)
						field = append(field, joinLineFields[0]...)

						// Remove quotes of joined field
						if field[0] != '"' || field[len(field)-1] != '"' {
							panic("csv.Read is broken")
						}
						field = field[1 : len(field)-1]

						// Append following fields after first joined field of line joinLineIndex
						fields = append(fields, joinLineFields[1:]...)

						// Empty lines that have been joined
						for i := lineIndex + 1; i <= joinLineIndex; i++ {
							lines[i] = nil
						}

					} else {

						// Begins with quote but does not end with one
						// means that a separator was in a quoted field
						// that has been wrongly splitted into multiple fields.
						// Needs merging of fields:
						for j := i + 1; j < len(fields); j++ {
							// Find following field that does not begin
							// with a quote, but ends with exactly one
							fj := fields[j]
							if len(fj) < 2 {
								continue
							}
							var (
								leftOK  = fj[0] != '"' || (fj[0] == '"' && fj[1] == '"')
								rightOK = (fj[len(fj)-2] != '"' && fj[len(fj)-1] == '"') // || bytes.HasSuffix(fj, []byte(`"""`))
							)
							if leftOK && rightOK {
								// Join fields [i..j]
								field = bytes.Join(fields[i:j+1], separator)
								// Remove quotes
								field = field[1 : len(field)-1]
								// Shift remaining slice fields over the ones joined into fields[i]
								copy(fields[i+1:], fields[j+1:])
								fields = fields[:len(fields)-(j-i)]
								break
							}
						}
					}
				}

			default:
				return nil, wraperr.Errorf("can't handle CSV field %q in line %q", field, lines[lineIndex])
				// /var/domonda-data/documents/39/d20/301/65394733/b7e967e7f98ec1e8/2019-01-03_09-46-50.435/doc.csv
			}

			fields[i] = bytes.ReplaceAll(field, []byte(`""`), []byte{'"'})
		}

		row := make([]string, len(fields))
		for i := range fields {
			row[i] = string(fields[i])
		}
		rows[lineIndex] = row
	}

	return rows, nil
}
