package excel

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/tealeg/xlsx"
	fs "github.com/ungerik/go-fs"
	reflection "github.com/ungerik/go-reflection"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
)

const ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

type ExcelFormatConfig struct {
	// https://exceljet.net/custom-number-formats
	Time     string
	Date     string
	Location *time.Location
	Nil      string
}

type ExcelCellWriter interface {
	WriteCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error
}

type ExcelCellWriterFunc func(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error

func (f ExcelCellWriterFunc) WriteCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	return f(cell, val, config)
}

type Writer struct {
	file            *xlsx.File
	currentSheet    *xlsx.Sheet
	headerStyle     *xlsx.Style
	cellStyle       *xlsx.Style
	Config          ExcelFormatConfig
	TypeCellWriters map[reflect.Type]ExcelCellWriter
}

func NewWriter(sheetName string) (*Writer, error) {
	headerStyle := xlsx.NewStyle()
	headerStyle.Font.Bold = true
	headerStyle.Font.Size = 10
	headerStyle.Font.Name = "Liberation Sans"
	headerStyle.ApplyFont = true

	excel := &Writer{
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

	err := excel.AddSheet(sheetName)
	if err != nil {
		return nil, err
	}

	return excel, nil
}

func (excel *Writer) AddSheet(name string) error {
	newSheet, err := excel.file.AddSheet(name)
	if err != nil {
		return err
	}
	excel.currentSheet = newSheet
	return nil
}

func (excel *Writer) SetCurrentSheet(name string) error {
	for _, sheet := range excel.file.Sheets {
		if sheet.Name == name {
			excel.currentSheet = sheet
			return nil
		}
	}
	return errors.Errorf("sheet with name '%s' not found", name)
}

func (excel *Writer) WriteHeaderRow(columnTitles []string) error {
	row := excel.currentSheet.AddRow()
	for _, title := range columnTitles {
		cell := row.AddCell()
		cell.SetStyle(excel.headerStyle)
		cell.SetString(title)
	}
	return nil
}

func (excel *Writer) WriteRow(columnValues []reflect.Value) error {
	row := excel.currentSheet.AddRow()
	for _, val := range columnValues {
		cell := row.AddCell()
		cell.SetStyle(excel.cellStyle)

		valType := val.Type()
		derefVal, derefType := reflection.DerefValueAndType(val)

		if w, ok := excel.TypeCellWriters[derefType]; ok && derefVal.IsValid() {
			// derefVal.IsValid() returns false for dereferenced nil pointer
			// so the following will only be called for non nil pointers:
			err := w.WriteCell(cell, derefVal, &excel.Config)
			if err != nil {
				return err
			}
			continue
		}

		switch valType.Kind() {
		case reflect.Ptr, reflect.Interface:
			if val.IsNil() {
				if excel.Config.Nil != "" {
					cell.SetString(excel.Config.Nil)
				}
				continue
			}
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
		if s, ok := val.Addr().Interface().(fmt.Stringer); ok {
			cell.SetString(s.String())
			continue
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

func (excel *Writer) Result() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := excel.file.Write(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (excel *Writer) WriteResultTo(writer io.Writer) error {
	return excel.file.Write(writer)
}

func (excel *Writer) WriteResultFile(file fs.File, perm ...fs.Permissions) error {
	writer, err := file.OpenWriter(perm...)
	if err != nil {
		return err
	}
	defer writer.Close()

	return excel.file.Write(writer)
}

func writeDateExcelCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	if d := val.Interface().(date.Date); !d.IsZero() {
		cell.SetDateWithOptions(
			d.MidnightTimeInLocation(config.Location),
			xlsx.DateTimeOptions{
				Location:        config.Location,
				ExcelTimeFormat: config.Date,
			},
		)
	}
	return nil
}

func writeNullableDateExcelCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	if d := val.Interface().(date.NullableDate); !d.IsZero() {
		cell.SetDateWithOptions(
			d.MidnightTimeInLocation(config.Location),
			xlsx.DateTimeOptions{
				Location:        config.Location,
				ExcelTimeFormat: config.Date,
			},
		)
	}
	return nil
}

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

func writeDurationExcelCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	duration := val.Interface().(time.Duration)
	excel1904Epoc := time.Date(1904, time.January, 1, 0, 0, 0, 0, time.UTC)
	cell.SetFloatWithFormat(xlsx.TimeToExcelTime(excel1904Epoc.Add(duration), true), "[h]:mm:ss")
	return nil
}

func writeMoneyAmountExcelCell(cell *xlsx.Cell, val reflect.Value, config *ExcelFormatConfig) error {
	cell.SetFloatWithFormat(val.Float(), "#,##0.00")
	return nil
}

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
