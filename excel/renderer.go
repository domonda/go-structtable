package excel

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	xlsx "github.com/tealeg/xlsx/v3"
	fs "github.com/ungerik/go-fs"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
)

// ContentType is the MIME type for Excel files.
const ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

// ExcelFormatConfig contains configuration for Excel cell formatting.
//
// This struct provides settings for formatting different data types in Excel cells,
// including date/time formats, number formats, and null value representations.
type ExcelFormatConfig struct {
	// Time specifies the Excel format string for time values.
	// See https://exceljet.net/custom-number-formats for format options.
	Time string
	// Date specifies the Excel format string for date values.
	Date string
	// Location specifies the timezone for date/time formatting.
	Location *time.Location
	// Null specifies the string representation for null values.
	Null string
}

// ExcelCellWriter defines the interface for writing specific data types to Excel cells.
//
// This interface allows for custom formatting of different Go types when writing
// them to Excel cells, providing fine-grained control over cell appearance and data types.
type ExcelCellWriter interface {
	// WriteCell writes a value to an Excel cell with the given formatting configuration.
	WriteCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error
}

// ExcelCellWriterFunc implements ExcelCellWriter with a function.
//
// This allows you to use a simple function as an ExcelCellWriter without
// creating a custom type.
type ExcelCellWriterFunc func(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error

// WriteCell calls the underlying function to write a cell value.
func (f ExcelCellWriterFunc) WriteCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	return f(cell, val, config)
}

// Renderer implements the structtable.Renderer interface for Excel files.
//
// This renderer generates Excel (.xlsx) files from struct slices, with support
// for custom formatting, multiple sheets, and various data types.
type Renderer struct {
	file            *xlsx.File
	currentSheet    *xlsx.Sheet
	headerStyle     *xlsx.Style
	cellStyle       *xlsx.Style
	Config          ExcelFormatConfig
	TypeCellWriters map[reflect.Type]ExcelCellWriter
}

// NewRenderer creates a new Excel Renderer with default formatting and styling.
//
// This constructor sets up a new Excel file with header styles, cell writers for common types,
// and creates the initial sheet. The renderer is configured with sensible defaults for
// date/time formatting and includes built-in cell writers for common data types.
//
// Parameters:
//   - sheetName: Name for the initial sheet (will be sanitized to comply with Excel naming rules)
//
// Returns:
//   - A new Renderer instance ready for use
//   - err: Any error that occurred during initialization (e.g., invalid sheet name)
//
// Example:
//
//	renderer, err := excel.NewRenderer("Sales Data")
//	if err != nil {
//	    return err
//	}
func NewRenderer(sheetName string) (*Renderer, error) {
	headerStyle := xlsx.NewStyle()
	headerStyle.Font.Bold = true
	headerStyle.Font.Size = 10
	headerStyle.Font.Name = "Liberation Sans"
	headerStyle.ApplyFont = true

	excel := &Renderer{
		file:        xlsx.NewFile(),
		headerStyle: headerStyle,
		Config: ExcelFormatConfig{
			Time:     "dd.mm.yyyy hh:mm:ss", // xlsx.DefaultDateTimeFormat
			Date:     "dd.mm.yyyy",          // xlsx.DefaultDateFormat
			Location: time.UTC,
		},
		TypeCellWriters: map[reflect.Type]ExcelCellWriter{
			reflect.TypeOf((*date.Date)(nil)).Elem():            ExcelCellWriterFunc(writeDateExcelCell),
			reflect.TypeOf((*date.NullableDate)(nil)).Elem():    ExcelCellWriterFunc(writeNullableDateExcelCell),
			reflect.TypeOf((*time.Time)(nil)).Elem():            ExcelCellWriterFunc(writeTimeExcelCell),
			reflect.TypeOf((*time.Duration)(nil)).Elem():        ExcelCellWriterFunc(writeDurationExcelCell),
			reflect.TypeOf((*money.Amount)(nil)).Elem():         ExcelCellWriterFunc(writeMoneyAmountExcelCell),
			reflect.TypeOf((*money.CurrencyAmount)(nil)).Elem(): ExcelCellWriterFunc(writeMoneyCurrencyAmountExcelCell),
		},
	}

	excel.file.Date1904 = true

	err := excel.AddSheet(sanitizeSheetName(sheetName))
	if err != nil {
		return nil, err
	}

	return excel, nil
}

// AddSheet adds a new sheet to the Excel file and sets it as the current sheet.
//
// This method creates a new worksheet with the specified name and makes it the active
// sheet for subsequent rendering operations. The sheet name will be sanitized to comply
// with Excel naming rules (removing invalid characters, limiting length).
//
// Parameters:
//   - name: The name for the new sheet
//
// Returns:
//   - err: Any error that occurred during sheet creation
//
// Example:
//
//	err := renderer.AddSheet("Q1 Results")
//	if err != nil {
//	    return err
//	}
func (excel *Renderer) AddSheet(name string) error {
	newSheet, err := excel.file.AddSheet(sanitizeSheetName(name))
	if err != nil {
		return err
	}
	excel.currentSheet = newSheet
	return nil
}

// SetCurrentSheet sets the current sheet by name for subsequent rendering operations.
//
// This method switches the active sheet to an existing sheet with the specified name.
// All subsequent calls to RenderHeaderRow and RenderRow will operate on this sheet.
//
// Parameters:
//   - name: The name of the sheet to make current
//
// Returns:
//   - err: Any error that occurred (e.g., sheet not found)
//
// Example:
//
//	err := renderer.SetCurrentSheet("Summary")
//	if err != nil {
//	    return err
//	}
func (excel *Renderer) SetCurrentSheet(name string) error {
	for _, sheet := range excel.file.Sheets {
		if sheet.Name == name {
			excel.currentSheet = sheet
			return nil
		}
	}
	return fmt.Errorf("sheet with name '%s' not found", name)
}

// RenderHeaderRow renders a header row with bold styling to the current sheet.
//
// This method creates a new row in the current sheet and populates it with the
// provided column titles. The header row uses the configured header style (bold font,
// Liberation Sans, size 10) to distinguish it from data rows.
//
// Parameters:
//   - columnTitles: Slice of strings representing the column headers
//
// Returns:
//   - err: Any error that occurred during rendering
//
// Example:
//
//	headers := []string{"Name", "Age", "Department", "Salary"}
//	err := renderer.RenderHeaderRow(headers)
func (excel *Renderer) RenderHeaderRow(columnTitles []string) error {
	row := excel.currentSheet.AddRow()
	for _, title := range columnTitles {
		cell := row.AddCell()
		cell.SetStyle(excel.headerStyle)
		cell.SetString(title)
	}
	return nil
}

// ValueOf returns the argument casted to reflect.Value if it's already a reflect.Value,
// otherwise returns the standard result of reflect.ValueOf(val).
//
// This utility function is useful when working with interfaces that might contain
// either regular values or reflect.Value instances, ensuring consistent handling.
//
// Parameters:
//   - val: The value to convert to reflect.Value
//
// Returns:
//   - reflect.Value: The reflect.Value representation of the input
//
// Example:
//
//	value := excel.ValueOf(someInterface)
//	if value.IsValid() {
//	    // Process the value
//	}
func ValueOf(val any) reflect.Value {
	v, ok := val.(reflect.Value)
	if ok {
		return v
	}
	return reflect.ValueOf(val)
}

// DerefValueAndType dereferences pointers and returns the final value and type.
//
// This utility function is useful for handling pointer types in Excel rendering,
// ensuring that the actual underlying value and type are used for formatting decisions.
// It recursively dereferences pointers until it reaches a non-pointer value.
//
// Parameters:
//   - val: The value to dereference
//
// Returns:
//   - reflect.Value: The dereferenced value
//   - reflect.Type: The type of the dereferenced value
//
// Example:
//
//	value, valueType := excel.DerefValueAndType(somePointer)
//	if valueType.Kind() == reflect.String {
//	    // Handle string type
//	}
func DerefValueAndType(val any) (reflect.Value, reflect.Type) {
	v := ValueOf(val)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v, v.Type()
}

// RenderRow renders a data row to the current sheet with appropriate formatting.
//
// This method creates a new row in the current sheet and populates it with the
// provided column values. It applies intelligent formatting based on the data type:
// - Uses custom cell writers for registered types (dates, money, etc.)
// - Handles nullable values with configurable null representation
// - Applies appropriate Excel data types (numbers, strings, booleans)
// - Sets right alignment for numeric values
// - Falls back to String() method or fmt.Sprint for unknown types
//
// Parameters:
//   - columnValues: Slice of reflect.Value instances representing the row data
//
// Returns:
//   - err: Any error that occurred during rendering
//
// Example:
//
//	values := []reflect.Value{
//	    reflect.ValueOf("John Doe"),
//	    reflect.ValueOf(25),
//	    reflect.ValueOf(time.Now()),
//	}
//	err := renderer.RenderRow(values)
func (excel *Renderer) RenderRow(columnValues []reflect.Value) error {
	row := excel.currentSheet.AddRow()
	for _, val := range columnValues {
		cell := row.AddCell()
		cell.SetStyle(excel.cellStyle)

		derefVal := val
		for derefVal.Kind() == reflect.Ptr && !derefVal.IsNil() {
			derefVal = derefVal.Elem()
		}
		derefType := derefVal.Type()

		if w, ok := excel.TypeCellWriters[derefType]; ok && derefVal.IsValid() {
			// derefVal.IsValid() returns false for dereferenced nil pointer
			// so the following will only be called for non nil pointers:
			err := w.WriteCell(cell, derefVal, &excel.Config)
			if err != nil {
				return err
			}
			continue
		}

		if nullable.ReflectIsNull(val) {
			if excel.Config.Null != "" {
				cell.SetString(excel.Config.Null)
			}
			continue
		}

		switch derefType.Kind() {
		case reflect.Bool:
			cell.SetBool(derefVal.Bool())
			continue

		case reflect.String:
			cell.SetString(derefVal.String())
			continue

		case reflect.Float32, reflect.Float64:
			cell.SetFloat(derefVal.Float())
			cell.GetStyle().Alignment.Horizontal = "right"
			cell.GetStyle().ApplyAlignment = true
			continue

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			cell.SetInt64(derefVal.Int())
			cell.GetStyle().Alignment.Horizontal = "right"
			cell.GetStyle().ApplyAlignment = true
			continue

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			cell.SetInt64(int64(derefVal.Uint()))
			cell.GetStyle().Alignment.Horizontal = "right"
			cell.GetStyle().ApplyAlignment = true
			continue
		}

		if s, ok := val.Interface().(fmt.Stringer); ok {
			cell.SetString(s.String())
			continue
		}
		if val.CanAddr() {
			if s, ok := val.Addr().Interface().(fmt.Stringer); ok {
				cell.SetString(s.String())
				continue
			}
		}
		if s, ok := derefVal.Interface().(fmt.Stringer); ok {
			cell.SetString(s.String())
			continue
		}

		switch x := derefVal.Interface().(type) {
		case []byte:
			cell.SetString(string(x))
			continue
		}

		cell.SetString(fmt.Sprint(val.Interface()))
	}
	return nil
}

// Result returns the Excel file as a byte slice.
//
// This method generates the complete Excel (.xlsx) file in memory and returns it
// as a byte slice. This is useful when you need to store the file in memory,
// send it over a network, or process it further before writing to disk.
//
// Returns:
//   - []byte: The complete Excel file as bytes
//   - err: Any error that occurred during file generation
//
// Example:
//
//	data, err := renderer.Result()
//	if err != nil {
//	    return err
//	}
//	// Use data bytes for further processing
func (excel *Renderer) Result() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := excel.file.Write(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteResultTo writes the Excel file to an io.Writer.
//
// This method generates the complete Excel (.xlsx) file and writes it directly
// to the provided io.Writer. This is more memory-efficient than Result() when
// writing to files or network streams, as it doesn't need to hold the entire
// file in memory.
//
// Parameters:
//   - writer: The io.Writer to write the Excel file to
//
// Returns:
//   - err: Any error that occurred during writing
//
// Example:
//
//	file, err := os.Create("output.xlsx")
//	if err != nil {
//	    return err
//	}
//	defer file.Close()
//	err = renderer.WriteResultTo(file)
func (excel *Renderer) WriteResultTo(writer io.Writer) error {
	return excel.file.Write(writer)
}

// WriteResultFile writes the Excel file to a file using fs.File interface.
//
// This method generates the complete Excel (.xlsx) file and writes it to the
// specified file using the fs.File interface. It handles file opening, writing,
// and closing automatically. Optional file permissions can be specified.
//
// Parameters:
//   - file: The fs.File to write the Excel file to
//   - perm: Optional file permissions (uses default if not provided)
//
// Returns:
//   - err: Any error that occurred during file operations or writing
//
// Example:
//
//	outputFile := fs.NewFile("reports/sales.xlsx")
//	err := renderer.WriteResultFile(outputFile, fs.Permissions(0644))
func (excel *Renderer) WriteResultFile(file fs.File, perm ...fs.Permissions) error {
	writer, err := file.OpenWriter(perm...)
	if err != nil {
		return err
	}
	defer writer.Close()

	return excel.file.Write(writer)
}

// MIMEType returns the MIME type for Excel files.
//
// This method returns the standard MIME type for Excel (.xlsx) files,
// which is used for HTTP content-type headers and file type identification.
//
// Returns:
//   - string: The MIME type "vnd.openxmlformats-officedocument.spreadsheetml.sheet"
//
// Example:
//
//	contentType := renderer.MIMEType()
//	w.Header().Set("Content-Type", contentType)
func (*Renderer) MIMEType() string {
	return "vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

// writeDateExcelCell writes date.Date values to Excel cells with proper date formatting.
//
// This function handles date.Date values by converting them to Excel date format
// using the configured date format string and timezone. Only non-zero dates are
// written to avoid Excel date calculation issues.
//
// Parameters:
//   - cell: The Excel cell to write to
//   - val: The reflect.Value containing a date.Date
//   - config: The Excel formatting configuration
//
// Returns:
//   - err: Any error that occurred during cell writing
func writeDateExcelCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	if d := val.Interface().(date.Date); !d.IsZero() {
		cell.SetDateWithOptions(
			d.MidnightInLocation(config.Location),
			xlsx.DateTimeOptions{
				Location:        config.Location,
				ExcelTimeFormat: config.Date,
			},
		)
	}
	return nil
}

// writeNullableDateExcelCell writes date.NullableDate values to Excel cells with proper date formatting.
//
// This function handles nullable date values by converting them to Excel date format
// using the configured date format string and timezone. Only non-zero dates are
// written to avoid Excel date calculation issues.
//
// Parameters:
//   - cell: The Excel cell to write to
//   - val: The reflect.Value containing a date.NullableDate
//   - config: The Excel formatting configuration
//
// Returns:
//   - err: Any error that occurred during cell writing
func writeNullableDateExcelCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	if d := val.Interface().(date.NullableDate); !d.IsZero() {
		cell.SetDateWithOptions(
			d.MidnightInLocation(config.Location).Time,
			xlsx.DateTimeOptions{
				Location:        config.Location,
				ExcelTimeFormat: config.Date,
			},
		)
	}
	return nil
}

// writeTimeExcelCell writes time.Time values to Excel cells with proper time formatting.
//
// This function handles time.Time values by converting them to Excel datetime format
// using the configured time format string and the time's original timezone.
// Only non-zero times are written to avoid Excel date calculation issues.
//
// Parameters:
//   - cell: The Excel cell to write to
//   - val: The reflect.Value containing a time.Time
//   - config: The Excel formatting configuration
//
// Returns:
//   - err: Any error that occurred during cell writing
func writeTimeExcelCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	if t := val.Interface().(time.Time); !t.IsZero() {
		cell.SetDateWithOptions(
			t,
			xlsx.DateTimeOptions{
				Location:        t.Location(),
				ExcelTimeFormat: config.Time,
			},
		)
	}
	return nil
}

// writeDurationExcelCell writes time.Duration values to Excel cells with time format.
//
// This function handles duration values by converting them to Excel time format
// using the Excel 1904 epoch and the "[h]:mm:ss" format string, which allows
// durations longer than 24 hours to be displayed correctly.
//
// Parameters:
//   - cell: The Excel cell to write to
//   - val: The reflect.Value containing a time.Duration
//   - config: The Excel formatting configuration (not used for durations)
//
// Returns:
//   - err: Any error that occurred during cell writing
func writeDurationExcelCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	duration := val.Interface().(time.Duration)
	excel1904Epoc := time.Date(1904, time.January, 1, 0, 0, 0, 0, time.UTC)
	cell.SetFloatWithFormat(xlsx.TimeToExcelTime(excel1904Epoc.Add(duration), true), "[h]:mm:ss")
	return nil
}

// writeMoneyAmountExcelCell writes money.Amount values to Excel cells with currency formatting.
//
// This function handles money amount values by formatting them as numbers with
// the "#,##0.00" format string, which displays numbers with thousands separators
// and two decimal places, suitable for currency amounts.
//
// Parameters:
//   - cell: The Excel cell to write to
//   - val: The reflect.Value containing a money.Amount
//   - config: The Excel formatting configuration (not used for amounts)
//
// Returns:
//   - err: Any error that occurred during cell writing
func writeMoneyAmountExcelCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	cell.SetFloatWithFormat(val.Float(), "#,##0.00")
	return nil
}

// writeMoneyCurrencyAmountExcelCell writes money.CurrencyAmount values to Excel cells with currency-specific formatting.
//
// This function handles currency amount values by formatting them with currency-specific
// format strings. If no currency is specified, it uses the standard "#,##0.00" format.
// For currencies, it uses a format like "#,##0.00 [$EUR];-#,##0.00 [$EUR]" to display
// the currency symbol alongside the amount.
//
// Parameters:
//   - cell: The Excel cell to write to
//   - val: The reflect.Value containing a money.CurrencyAmount
//   - config: The Excel formatting configuration (not used for currency amounts)
//
// Returns:
//   - err: Any error that occurred during cell writing
func writeMoneyCurrencyAmountExcelCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	ca := val.Interface().(money.CurrencyAmount)
	if ca.Currency == "" {
		cell.SetFloatWithFormat(float64(ca.Amount), "#,##0.00")
		return nil
	}
	// #.##0,00 [$€-407];[ROT]-#.##0,00 [$€-407]
	// format := fmt.Sprintf("[$%[1]s] #,##0.00;[$%[1]s] -#,##0.00", ca.Currency.Symbol())
	format := fmt.Sprintf("#,##0.00 [$%[1]s];-#,##0.00 [$%[1]s]", ca.Currency)
	cell.SetFloatWithFormat(float64(ca.Amount), format)
	return nil
}

// sanitizeSheetName sanitizes sheet names to comply with Excel naming rules.
//
// This function ensures that sheet names are valid for Excel by:
// - Trimming whitespace
// - Using "UNNAMED" for empty names
// - Limiting length to 31 characters (with ellipsis for truncation)
// - Replacing invalid characters (\, /, ?, *, [, ]) with underscores
//
// Parameters:
//   - name: The original sheet name to sanitize
//
// Returns:
//   - string: The sanitized sheet name safe for Excel
//
// Example:
//
//	safeName := sanitizeSheetName("Sales/Q1 2024") // Returns "Sales_Q1 2024"
func sanitizeSheetName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "UNNAMED"
	}
	b := strings.Builder{}
	b.Grow(len(name))
	runeCount := 0
	for _, r := range name {
		if runeCount == 30 {
			// Only 31 runes allowed, write ellipsis as 31st and last rune
			b.WriteRune('…')
			break
		}
		switch r {
		case '\\', '/', '?', '*', '[', ']':
			// Disallowed characters, write placeholder
			b.WriteByte('_')
		default:
			b.WriteRune(r)
		}
		runeCount++
	}
	return b.String()
}
