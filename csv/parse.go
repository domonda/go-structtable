package csv

import (
	"bytes"
	"context"

	"github.com/ungerik/go-fs"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/charset"
)

// ParseDetectFormat parses CSV data with automatic format detection.
//
// This function analyzes the input data to automatically detect the CSV format parameters
// including character encoding, field separator, and line endings. It uses the provided
// FormatDetectionConfig to determine which encodings and separators to test.
//
// Format detection algorithm:
//  1. Encoding detection: the configured encodings are tested against the
//     configured test strings to find the one that decodes special characters
//  2. Line ending detection: counts \r\n, \n\r and bare \n outside of quoted
//     fields and takes the most frequent one
//  3. Separator detection: scores comma, semicolon, tab and pipe by how uniform
//     the resulting column count is instead of by how often they occur, counting
//     only outside of quoted fields
//  4. Header line detection: an Excel style "sep=X" first line declares the
//     separator explicitly and wins over the detection
//
// Parameters:
//   - data: Raw CSV data bytes to parse
//   - configOrNil: Format detection configuration (uses default if nil)
//
// Returns:
//   - rows: Parsed CSV rows as a 2D slice of strings
//   - format: The detected format configuration when err is nil, otherwise nil.
//     A returned format is always valid, so callers can re-use it for parsing
//     and writing further data
//   - err: Any error that occurred during parsing or format detection
//
// Example:
//
//	data, err := ioutil.ReadFile("data.csv")
//	if err != nil {
//	    return err
//	}
//	rows, format, err := csv.ParseDetectFormat(data, nil)
func ParseDetectFormat(data []byte, configOrNil *FormatDetectionConfig) (rows [][]string, format *Format, err error) {
	defer errs.WrapWithFuncParams(&err, data, configOrNil)
	defer errs.RecoverPanicAsError(&err)

	config := configOrNil
	if config == nil {
		config = NewFormatDetectionConfig()
	}

	format, lines, err := detectFormatAndSplitLines(data, config)
	if err != nil {
		return nil, format, err
	}

	rows, err = readLines(lines, []byte(format.Separator), "\n")
	return rows, format, err
}

// ParseFileDetectFormat parses a CSV file with automatic format detection using context support.
//
// This function reads a CSV file from a fs.FileReader and automatically detects the format
// parameters including character encoding, field separator, and line endings. It supports
// context cancellation for long-running operations.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - csvFile: The file reader containing CSV data
//   - configOrNil: Format detection configuration (uses default if nil)
//
// Returns:
//   - rows: Parsed CSV rows as a 2D slice of strings
//   - format: The detected format configuration
//   - err: Any error that occurred during file reading or parsing
//
// Example:
//
//	file := fs.NewFile("data.csv")
//	rows, format, err := csv.ParseFileDetectFormat(ctx, file, nil)
func ParseFileDetectFormat(ctx context.Context, csvFile fs.FileReader, configOrNil *FormatDetectionConfig) (rows [][]string, format *Format, err error) {
	defer errs.WrapWithFuncParams(&err, ctx, csvFile, configOrNil)

	data, err := csvFile.ReadAllContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	return ParseDetectFormat(data, configOrNil)
}

// ParseWithFormat parses CSV data using a specific format configuration.
//
// This function parses CSV data according to the provided Format configuration,
// including character encoding, field separator, and line endings. It validates
// the format before parsing and handles BOM removal for UTF-8 files.
//
// Parameters:
//   - data: Raw CSV data bytes to parse
//   - format: The format configuration to use for parsing
//
// Returns:
//   - rows: Parsed CSV rows as a 2D slice of strings
//   - err: Any error that occurred during parsing or format validation
//
// Example:
//
//	format := &csv.Format{
//	    Encoding:  "UTF-8",
//	    Separator: ",",
//	    Newline:   "\n",
//	}
//	rows, err := csv.ParseWithFormat(data, format)
func ParseWithFormat(data []byte, format *Format) (rows [][]string, err error) {
	defer errs.WrapWithFuncParams(&err, data, format)
	defer errs.RecoverPanicAsError(&err)

	err = format.Validate()
	if err != nil {
		return nil, err
	}

	if format.Encoding == "UTF-8" {
		data = charset.TrimBOM(data, charset.BOMUTF8)
	} else {
		enc, err := charset.GetEncoding(format.Encoding)
		if err != nil {
			return nil, err
		}
		data, err = enc.Decode(data)
		if err != nil {
			return nil, err
		}
	}

	data = sanitizeUTF8(data)

	lines := splitLines(data, format.Newline)
	if len(lines) > 0 {
		if headerSep := parseSepHeaderLine(lines[0]); headerSep != "" {
			if headerSep != format.Separator {
				return nil, errs.Errorf("separator '%s' in header line is different from format.Separator '%s'", headerSep, format.Separator)
			}
			lines = lines[1:]
		}
	}

	return readLines(lines, []byte(format.Separator), "\n")
}

// ParseFileWithFormat parses a CSV file using a specific format configuration with context support.
//
// This function reads a CSV file from a fs.FileReader and parses it according to the
// provided Format configuration. It supports context cancellation for long-running operations.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - csvFile: The file reader containing CSV data
//   - format: The format configuration to use for parsing
//
// Returns:
//   - rows: Parsed CSV rows as a 2D slice of strings
//   - err: Any error that occurred during file reading or parsing
//
// Example:
//
//	format := csv.NewFormat(",")
//	file := fs.NewFile("data.csv")
//	rows, err := csv.ParseFileWithFormat(ctx, file, format)
func ParseFileWithFormat(ctx context.Context, csvFile fs.FileReader, format *Format) (rows [][]string, err error) {
	defer errs.WrapWithFuncParams(&err, ctx, csvFile, format)

	data, err := csvFile.ReadAllContext(ctx)
	if err != nil {
		return nil, err
	}

	return ParseWithFormat(data, format)
}

// detectFormatAndSplitLines detects CSV format parameters and splits data into lines.
//
// This internal function analyzes the input data to automatically detect the CSV format
// parameters including character encoding, field separator, and line endings. It then
// splits the data into individual lines for further processing.
//
// Parameters:
//   - data: Raw CSV data bytes to analyze
//   - config: Format detection configuration (must not be nil)
//
// Returns:
//   - format: The detected format configuration
//   - lines: Data split into individual lines as byte slices, with the
//     "sep=X" header line removed if present, or nil when the data holds
//     no non-empty line
//   - err: Any error that occurred during format detection
//
// The function performs the following detection steps:
//  1. Character encoding detection using charset.AutoDecode
//  2. Scanning the structural bytes outside of quoted fields
//  3. Line ending detection from the counted newlines
//  4. Data splitting into individual lines
//  5. An Excel style "sep=X" header line, which declares the separator
//     explicitly and returns without the separator detection below
//  6. Field separator detection from the uniformity of the column count
//
// Separators and newlines inside quoted fields are excluded from both detections,
// so a quoted field cannot outvote the real structure of the data.
func detectFormatAndSplitLines(data []byte, config *FormatDetectionConfig) (format *Format, lines [][]byte, err error) {
	defer errs.WrapWithFuncParams(&err, data, config)

	if config == nil {
		return nil, nil, errs.New("FormatDetectionConfig must not be nil")
	}

	format = new(Format)

	///////////////////////////////////////////////////////////////////////////
	// Detect charset encoding

	var encodings []charset.Encoding
	for _, name := range config.Encodings {
		enc, err := charset.GetEncoding(name)
		if err != nil {
			return nil, nil, err
		}
		encodings = append(encodings, enc)
	}

	data, format.Encoding, err = charset.AutoDecode(data, encodings, config.EncodingTests)
	if err != nil {
		return nil, nil, err
	}
	if format.Encoding == "" {
		format.Encoding = "UTF-8"
	}

	data = sanitizeUTF8(data)

	///////////////////////////////////////////////////////////////////////////
	// Scan the structure outside of quoted fields for the detections below

	structure := scanStructure(data, true)
	if structure.endedQuoted {
		// The data ends within a quoted field, so its quoting is unbalanced
		// and everything after the offending quote was skipped. Scan again
		// without quoting instead of guessing which quote is the wrong one.
		structure = scanStructure(data, false)
	}

	///////////////////////////////////////////////////////////////////////////
	// Detect line endings

	// Newlines within a quoted field are part of its value and were not
	// counted, so a single quoted \r\n can't switch a whole \n separated
	// file to \r\n line endings. A wider line ending wins a tie because a
	// file using one has no bare \n of its own to count, and \r\n wins over
	// \n\r because it is the standard.
	numBareLF := structure.numLF - structure.numCRLF - structure.numLFCR
	switch {
	case structure.numCRLF > 0 && structure.numCRLF >= structure.numLFCR && structure.numCRLF >= numBareLF:
		format.Newline = "\r\n"
	case structure.numLFCR > 0 && structure.numLFCR >= numBareLF:
		format.Newline = "\n\r"
	default:
		format.Newline = "\n"
	}

	///////////////////////////////////////////////////////////////////////////
	// Detect separator

	lines = splitLines(data, format.Newline)

	if len(lines) > 0 {
		format.Separator = parseSepHeaderLine(lines[0])
		if format.Separator != "" {
			return format, lines[1:], nil
		}
	}

	// Default separator, also used when there is no line to detect one from,
	// because the returned Format is used by callers for parsing and writing
	// further data and has to be valid in any case.
	format.Separator = ","

	numNonEmptyLines := 0
	for _, line := range lines {
		if len(line) > 0 {
			numNonEmptyLines++
		}
	}
	if numNonEmptyLines == 0 {
		return format, nil, nil
	}

	if separator, ok := structure.bestSeparator(); ok {
		format.Separator = separator
	}

	return format, lines, nil
}

// splitLines splits data into lines separated by newline and removes
// stray newline characters from the end of every line.
//
// Trimming is part of splitting because a file can use a line ending wider
// than the newline it is split by, which would otherwise leak into the last
// field of every line.
//
// Only the end of a line is trimmed. A \r at the start of a line is not
// residue of a \n\r line ending, because those are detected and split by,
// but a carriage return within a quoted field that has to be preserved.
//
// Trimming the end still loses a \r that directly precedes the newline the
// lines are split by, like the one in A;"x\r\ny";B within a file with \n
// line endings. There the \r can't be told apart from the residue of a file
// with mixed line endings, which is what is trimmed here. Telling the two
// apart needs the quoted state while splitting, which this parser only has
// after the lines are split.
func splitLines(data []byte, newline string) [][]byte {
	lines := bytes.Split(data, []byte(newline))
	for i := range lines {
		lines[i] = bytes.TrimRight(lines[i], "\r\n")
	}
	return lines
}

// separatorCandidates are the separators that can be detected from the data,
// ordered by preference so that the earlier one wins a tie in bestSeparator.
var separatorCandidates = []byte{',', ';', '\t', '|'}

// csvStructure is what scanStructure counted outside of quoted fields.
type csvStructure struct {
	numLF      int // Newlines, of which numCRLF and numLFCR are part of a wider line ending
	numCRLF    int
	numLFCR    int
	numRecords int // Records with any content, a record can span several lines

	// columnsPerRecord maps for every separatorCandidates index
	// the number of columns to the number of records with that many columns
	columnsPerRecord []map[int]int

	// endedQuoted reports that the data ends within a quoted field, meaning
	// that its quoting is unbalanced and everything after the offending quote
	// was skipped
	endedQuoted bool
}

// scanStructure counts the structural bytes of the CSV data that are not part
// of a quoted field value: the line endings, and for every separator candidate
// how many records have how many columns.
//
// A quote toggles the quoted state, except for a doubled quote within a quoted
// field which is an escaped quote that does not end it. Toggling on quotes
// alone keeps the scan independent of the separator, which is not known yet
// while it is detected. Any newline outside of a quoted field ends a record,
// so the record boundaries don't depend on the detected line ending either,
// and a newline within a quoted field does not split a record in two.
//
// With skipQuoted false the quoting is ignored, which is the fallback for data
// whose quoting is unbalanced.
func scanStructure(data []byte, skipQuoted bool) *csvStructure {
	s := &csvStructure{columnsPerRecord: make([]map[int]int, len(separatorCandidates))}
	for i := range s.columnsPerRecord {
		s.columnsPerRecord[i] = make(map[int]int)
	}

	separators := make([]int, len(separatorCandidates))
	recordHasContent := false
	endRecord := func() {
		if recordHasContent {
			s.numRecords++
			for i, numSeparators := range separators {
				s.columnsPerRecord[i][numSeparators+1]++
			}
		}
		clear(separators)
		recordHasContent = false
	}

	quoted := false
	for i := 0; i < len(data); i++ {
		c := data[i]
		if skipQuoted && c == '"' {
			if quoted && i+1 < len(data) && data[i+1] == '"' {
				i++ // Escaped quote within a quoted field
				continue
			}
			quoted = !quoted
			recordHasContent = true
			continue
		}
		if quoted {
			continue
		}
		switch c {
		case '\r':
			if i+1 < len(data) && data[i+1] == '\n' {
				i++
				s.numLF++
				s.numCRLF++
			}
			endRecord()
		case '\n':
			s.numLF++
			if i+1 < len(data) && data[i+1] == '\r' {
				i++
				s.numLFCR++
			}
			endRecord()
		default:
			recordHasContent = true
			for candidate, separator := range separatorCandidates {
				if c == separator {
					separators[candidate]++
				}
			}
		}
	}
	endRecord()

	s.endedQuoted = quoted
	return s
}

// bestSeparator returns the candidate that splits the records into the most
// uniform number of columns, because the right separator is the one that makes
// the data rectangular. Counting occurrences alone is not enough: unquoted text
// containing commas can hold more of them than a semicolon separated file has
// semicolons. More columns win a tie, then the candidate order.
//
// ok is false when no candidate separates the records into more than one
// column, so the caller keeps its default separator.
func (s *csvStructure) bestSeparator() (separator string, ok bool) {
	if s.numRecords == 0 {
		return "", false
	}
	var (
		bestUniformity float64
		bestColumns    int
	)
	for candidate, sep := range separatorCandidates {
		// The most common number of columns of this candidate,
		// more columns win a tie.
		//
		// Records that the candidate does not separate at all are left out of
		// that vote unless it separates no record, because the header and
		// trailer lines of a table are single column records that must not
		// outvote the actual table rows even when there are more of them.
		// A bank statement with more preamble and trailer lines than data rows
		// is the common case, and counting those made the most common column
		// count one, which discarded the real separator below.
		// [SetRowsWithNonUniformColumnsNil] applies the same rule to the rows.
		var columns, records int
		for c, r := range s.columnsPerRecord[candidate] {
			if c < 2 && len(s.columnsPerRecord[candidate]) > 1 {
				continue
			}
			if r > records || (r == records && c > columns) {
				columns, records = c, r
			}
		}
		if columns < 2 {
			// The candidate doesn't separate anything in the typical record
			continue
		}
		uniformity := float64(records) / float64(s.numRecords)
		if uniformity > bestUniformity || (uniformity == bestUniformity && columns > bestColumns) {
			bestUniformity, bestColumns = uniformity, columns
			separator, ok = string(sep), true
		}
	}
	return separator, ok
}

// parseSepHeaderLine parses Excel-style separator header lines and extracts the separator character.
//
// This function handles Excel-style separator declaration lines like "sep=," or "SEP=;"
// that are sometimes included at the beginning of CSV files to indicate the field separator.
// It supports both quoted and unquoted formats and case-insensitive "sep" keyword.
//
// Parameters:
//   - line: The header line bytes to parse
//
// Returns:
//   - sep: The declared separator character, or empty string if the line is no
//     header line or declares a character that can never be a separator
//
// Supported formats:
//   - "sep=," (unquoted)
//   - "SEP=;" (case-insensitive)
//   - "\"sep=,\"" (quoted)
//
// Example:
//
//	separator := parseSepHeaderLine([]byte("sep=;"))
//	// Returns ";"
func parseSepHeaderLine(line []byte) (sep string) {
	if len(line) < 5 {
		return ""
	}
	if line[0] == '"' && line[len(line)-1] == '"' {
		line = line[1 : len(line)-1]
	}
	if len(line) != 5 {
		return ""
	}
	if !bytes.HasPrefix(line, []byte("sep=")) && !bytes.HasPrefix(line, []byte("SEP=")) {
		return ""
	}
	if !validSeparator(line[4]) {
		return ""
	}
	return string(line[4:5])
}

// readLines parses CSV lines into fields, handling quoted fields and multi-line fields properly.
//
// This function processes a slice of byte lines and converts them into CSV fields,
// properly handling quoted fields that may contain separators, newlines, or escaped quotes.
// It supports multi-line fields where quoted content spans multiple lines.
//
// Quoting and escaping rules (RFC 4180):
//   - Fields containing separator, newline, or quotes must be quoted
//   - Quotes within quoted fields are escaped by doubling: "" represents "
//   - A field beginning with an odd number of quotes and holding an odd total
//     number of quotes is not closed within itself, so it was split by a
//     separator or a newline within the quotes and is joined together again
//   - The closing part of such a field may only begin with escaped quotes and
//     must end with an unescaped quote, so an ordinary quoted field on a later
//     line is not mistaken for it
//   - A closing quote followed by unquoted characters does not end the field,
//     its quotes are literal and are only unescaped
//
// Parameters:
//   - lines: Slice of byte lines to parse
//   - separator: Field separator as byte slice
//   - newlineReplacement: String to replace newlines in quoted fields
//
// Returns:
//   - rows: Parsed CSV rows as 2D slice of strings, lines joined into a
//     multi-line field become nil entries so the line indices stay correct
//   - err: Any error that occurred during parsing
//
// Example:
//
//	lines := [][]byte{
//	    []byte("Name,Age,City"),
//	    []byte("\"John Doe\",25,\"New York\""),
//	}
//	rows, err := readLines(lines, []byte(","), "\n")
func readLines(lines [][]byte, separator []byte, newlineReplacement string) (rows [][]string, err error) {
	defer errs.WrapWithFuncParams(&err, lines, separator, newlineReplacement)

	// noClosingFieldInLaterLines caches that no line after the last searched
	// one holds a field closing an unterminated quoted field. The searches
	// begin at ever later lines and lines are only emptied while parsing, so
	// a search that reached the end without a match can never find one in a
	// later line again. Without the cache every field of a file whose lines
	// all begin an unterminated quoted field re-scans the whole rest.
	noClosingFieldInLaterLines := false

	rows = make([][]string, len(lines))
	for lineIndex, line := range lines {
		if len(line) == 0 {
			continue
		}

		fields := bytes.Split(line, separator)
		// Line that the current field begins on. Joining a field across lines
		// continues this line with the fields of the closing line, so every
		// following field of the same row begins on that line and not on
		// lineIndex anymore.
		curLine := lineIndex
		// noClosingFieldInRow caches the same for the fields of the current
		// row that follow the searched one. A later field searches a subset
		// of the fields already searched, and the fields are only rebuilt by
		// a search that found a closing field, so once a search failed every
		// later search of the same row fails too. Without the cache a single
		// line whose fields all open an unterminated quoted field re-scans
		// the rest of the line for every one of them, which is quadratic in
		// the number of fields. A file with bare carriage return line
		// endings is one such line, because those are not detected as a
		// newline and the whole file stays a single line.
		noClosingFieldInRow := false
		// The row is built forward instead of the joined fields being cut out
		// of the split line, because compacting in place moves the whole tail
		// of the row per join, which is quadratic in the number of fields.
		row := make([]string, 0, len(fields))
		for i := 0; i < len(fields); i++ {
			field := fields[i]

			// Only a field beginning with a quote needs quote handling.
			// An empty field and every other field's quotes are literal and
			// are just unescaped below, so counting them would be wasted work.
			if leftQuotes := countQuotesLeft(field); len(field) > 0 && leftQuotes > 0 {
				totalQuotes := bytes.Count(field, []byte{'"'})
				switch {
				case totalQuotes == len(field) && len(field)%2 == 0:
					// Field consists only of an even number of quotes, which is an escaped
					// empty field `""`, an escaped quote `""""`, and so on.
					// An odd number of quotes leaves one quote unescaped that opens a field
					// continued after a separator or newline, which is handled by the case below.
					// Remove outermost quotes
					field = field[1 : len(field)-1]

				case leftQuotes%2 == 1 && totalQuotes%2 == 1:
					// An odd number of leading quotes opens a quoted field
					// and an odd total number of quotes means that the field
					// is not closed again within itself, so it was wrongly split
					// by a separator or a newline inside the quoted field
					// and has to be joined together again.
					//
					// Search for the field that closes the quoted field, first in
					// the remaining fields of this line which were split off by a
					// separator within the quotes, then in the fields of the
					// following lines which were split off by a newline within the
					// quotes. Newlines are allowed in quoted CSV fields.
					// A field can be split by both, so the search must neither stop
					// at the end of this line nor at the first field of a line.
					var (
						closeLine   = -1
						closeField  = -1
						closeFields [][]byte
					)
				findClosingField:
					for l := curLine; l < len(lines); l++ {
						lineFields, r := fields, i+1
						if l > curLine {
							if noClosingFieldInLaterLines {
								break
							}
							lineFields, r = bytes.Split(lines[l], separator), 0
						} else if noClosingFieldInRow {
							continue
						}
						for ; r < len(lineFields); r++ {
							if closesQuotedField(lineFields[r]) {
								closeLine, closeField, closeFields = l, r, lineFields
								break findClosingField
							}
						}
					}
					if closeLine == -1 {
						// The fields following this one and every later line
						// were searched without a match, so neither can hold
						// a closing field for a later field of this row.
						noClosingFieldInRow = true
						noClosingFieldInLaterLines = true
					}

					switch {
					case closeLine == curLine:
						// Only fields of this line were split off by a separator,
						// so join the fields [i..closeField] back together
						field = bytes.Join(fields[i:closeField+1], separator)
						// Remove quotes
						field = field[1 : len(field)-1]
						// Continue after the fields joined into this one. They are
						// skipped rather than cut out of the slice, so joining costs
						// nothing beyond the join itself.
						i = closeField

					case closeLine > curLine:
						// The field was also split off by a newline, so join the
						// remaining fields of this line, the lines in between and
						// the fields of the closing line up to closeField
						joined := bytes.Join(fields[i:], separator)
						for l := curLine + 1; l < closeLine; l++ {
							joined = append(joined, newlineReplacement...)
							joined = append(joined, lines[l]...)
						}
						joined = append(joined, newlineReplacement...)
						joined = append(joined, bytes.Join(closeFields[:closeField+1], separator)...)

						// Remove quotes of joined field
						if joined[0] != '"' || joined[len(joined)-1] != '"' {
							return nil, errs.New("should never happen: csv.Read is broken")
						}
						field = joined[1 : len(joined)-1]

						// Continue this row with the fields of the closing
						// line that follow the closing field
						fields, i = closeFields, closeField

						// Empty lines that have been joined
						// so line indices are still correct
						for l := curLine + 1; l <= closeLine; l++ {
							lines[l] = nil
						}

						// The fields following the closing field begin
						// on the closing line, not on lineIndex
						curLine = closeLine

					case totalQuotes == len(field):
						// Nothing to join the unterminated field with,
						// so only its opening quote is removed and the
						// remaining quotes are unescaped further down.
						field = field[1:]
					}

				case leftQuotes%2 == 1 && field[len(field)-1] == '"':
					// Quoted field that is closed again within itself.
					// Remove outermost quotes
					field = field[1 : len(field)-1]

				default:
					// Field is not quoted, or its closing quote is followed by
					// unquoted characters, so all its quotes are literal
					// and only have to be unescaped further down
				}
			}

			// bytes.ReplaceAll allocates a copy of the field even when
			// there is nothing to replace, so only call it when there is.
			if bytes.Contains(field, []byte(`""`)) {
				field = bytes.ReplaceAll(field, []byte(`""`), []byte{'"'})
			}
			row = append(row, string(field))
		}

		rows[lineIndex] = row
	}

	return rows, nil
}

// countQuotesLeft counts consecutive quote characters from the beginning of a string.
//
// This utility function counts the number of consecutive double-quote characters
// starting from the beginning of the byte slice until it encounters a non-quote character.
//
// Parameters:
//   - str: The byte slice to analyze
//
// Returns:
//   - int: The number of consecutive quotes from the start,
//     or the length if the entire slice is quotes
//
// Example:
//
//	count := countQuotesLeft([]byte("\"\"\"text"))
//	// Returns 3
func countQuotesLeft(str []byte) int {
	for i, c := range str {
		if c != '"' {
			return i
		}
	}
	return len(str)
}

// countQuotesRight counts consecutive quote characters from the end of a string.
//
// This utility function counts the number of consecutive double-quote characters
// starting from the end of the byte slice until it encounters a non-quote character.
//
// Parameters:
//   - str: The byte slice to analyze
//
// Returns:
//   - int: The number of consecutive quotes from the end,
//     or the length if the entire slice is quotes
//
// Example:
//
//	count := countQuotesRight([]byte("text\"\"\""))
//	// Returns 3
func countQuotesRight(str []byte) int {
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] != '"' {
			return len(str) - 1 - i
		}
	}
	return len(str)
}

// closesQuotedField reports whether field is the closing part of a quoted
// field that was split by a separator or newline inside the quotes.
//
// The closing part may only begin with escaped quotes and must end with an
// unescaped closing quote. Requiring both is what distinguishes it from an
// ordinary quoted field like `"value"`, which must not be mistaken for the
// closing part of an unterminated field further up.
//
// Example:
//
//	closesQuotedField([]byte(`value"`))    // Returns: true
//	closesQuotedField([]byte(`""value"`))  // Returns: true
//	closesQuotedField([]byte(`"`))         // Returns: true
//	closesQuotedField([]byte(`"value"`))   // Returns: false (a complete field)
//	closesQuotedField([]byte(`value`))     // Returns: false (no closing quote)
func closesQuotedField(field []byte) bool {
	leftQuotes := countQuotesLeft(field)
	if leftQuotes == len(field) {
		// Field consists only of quotes, so it closes the
		// quoted field if one quote is left unescaped
		return leftQuotes%2 == 1
	}
	// A single leading quote can never occur inside a quoted field
	return leftQuotes%2 == 0 && countQuotesRight(field)%2 == 1
}

// sanitizeUTF8 replaces every byte that is not valid UTF-8, every U+FFFD
// replacement character, and every no-break space with a plain space,
// and returns the result as a newly allocated slice.
//
// Invalid bytes are replaced one by one, so a two byte sequence becomes two
// spaces. The result is always valid UTF-8, so the parser and everything
// downstream can treat the data as text without checking it again.
//
// A no-break space is replaced because it reads as a space but is not one to
// code that trims or compares field values, and spreadsheet exports are full
// of them. Note that this also changes field values that legitimately contain
// one.
//
// Sanitizing hides a failed encoding detection instead of reporting it: data
// decoded with the wrong encoding loses its undecodable bytes to spaces rather
// than raising an error, so `Müller` in Windows 1252 read as UTF-8 becomes
// `M ller`. Both callers sanitize directly after decoding, so a caller that
// has to know whether the encoding was right must check the decoded data
// itself.
//
// Parameters:
//   - str: The byte slice to sanitize
//
// Returns:
//   - []byte: The sanitized and always valid UTF-8 byte slice
//
// Example:
//
//	sanitizeUTF8([]byte("Jänner"))        // Returns: "Jänner"
//	sanitizeUTF8([]byte("a\u00a0b"))      // Returns: "a b"
//	sanitizeUTF8([]byte{'M', 0xfc, 'l'})  // Returns: "M l"
func sanitizeUTF8(str []byte) []byte {
	return bytes.Map(
		func(r rune) rune {
			switch r {
			// \u00a0 is No-Break Space (NBSP)
			case '�', '\u00a0':
				return ' '
			default:
				return r
			}
		},
		str,
	)
}
