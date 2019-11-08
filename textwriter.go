package structtable

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	fs "github.com/ungerik/go-fs"
	reflection "github.com/ungerik/go-reflection"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/strfmt"
)

// TextWriterImpl has to be implemented for a format
// to be used by TextWriter.
type TextWriterImpl interface {
	WriteBeginTableText(writer io.Writer) error
	WriteHeaderRowText(writer io.Writer, columnTitles []string) error
	WriteRowText(writer io.Writer, fields []string) error
	WriteEndTableText(writer io.Writer) error
}

// TextWriter implements the Writer by using a TextWriterImpl
// for a specific text based table format.
type TextWriter struct {
	impl           TextWriterImpl
	config         *TextFormatConfig
	typeFormatters map[reflect.Type]TextFormatter
	buf            bytes.Buffer
	beginWritten   bool
}

func NewTextWriter(impl TextWriterImpl, config *TextFormatConfig) *TextWriter {
	tw := &TextWriter{
		impl:   impl,
		config: config,
	}
	return tw
}

// func (tw *TextWriter) SetTypeTextFormatter(columnType reflect.Type, formatter TextFormatter) {
// 	if formatter != nil {
// 		tw.typeFormatters[columnType] = formatter
// 	} else {
// 		delete(tw.typeFormatters, columnType)
// 	}
// }

func (tw *TextWriter) writeBeginIfMissing() error {
	if tw.beginWritten {
		return nil
	}
	err := tw.impl.WriteBeginTableText(&tw.buf)
	if err != nil {
		return err
	}
	tw.beginWritten = true
	return nil
}

func (tw *TextWriter) WriteHeaderRow(columnTitles []string) error {
	err := tw.writeBeginIfMissing()
	if err != nil {
		return err
	}
	return tw.impl.WriteHeaderRowText(&tw.buf, columnTitles)
}

func (tw *TextWriter) WriteRow(columnValues []reflect.Value) error {
	err := tw.writeBeginIfMissing()
	if err != nil {
		return err
	}
	fields := make([]string, len(columnValues))
	for i, val := range columnValues {
		fields[i] = tw.toString(val)
	}
	return tw.impl.WriteRowText(&tw.buf, fields)
}

func (tw *TextWriter) toString(val reflect.Value) string {
	valType := val.Type()
	derefVal, derefType := reflection.DerefValueAndType(val)

	if f, ok := tw.config.TypeFormatters[derefType]; ok && derefVal.IsValid() {
		// derefVal.IsValid() returns false for dereferenced nil pointer
		// so the following will only be called for non nil pointers:
		return f.FormatValue(derefVal, tw.config)
	}

	switch valType.Kind() {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return tw.config.Nil
		}
	}

	switch derefType.Kind() {
	case reflect.Bool:
		if derefVal.Bool() {
			return tw.config.True
		} else {
			return tw.config.False
		}

	case reflect.String:
		return derefVal.String()

	case reflect.Float32, reflect.Float64:
		return strfmt.FormatFloat(
			derefVal.Float(),
			tw.config.Float.ThousandsSep,
			tw.config.Float.DecimalSep,
			tw.config.Float.Precision,
			tw.config.Float.PadPrecision,
		)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(derefVal.Int(), 10)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(derefVal.Uint(), 10)
	}

	if s, ok := val.Interface().(fmt.Stringer); ok {
		return s.String()
	}
	if s, ok := val.Addr().Interface().(fmt.Stringer); ok {
		return s.String()
	}
	if s, ok := derefVal.Interface().(fmt.Stringer); ok {
		return s.String()
	}

	switch x := derefVal.Interface().(type) {
	case []byte:
		return string(x)
	}

	return fmt.Sprint(val.Interface())
}

func (tw *TextWriter) Result() ([]byte, error) {
	err := tw.impl.WriteEndTableText(&tw.buf)
	if err != nil {
		return nil, err
	}
	return tw.buf.Bytes(), nil
}

func (tw *TextWriter) WriteResultTo(writer io.Writer) error {
	_, err := tw.buf.WriteTo(writer)
	return err
}

func (tw *TextWriter) WriteResultFile(file fs.File, perm ...fs.Permissions) error {
	writer, err := file.OpenWriter(perm...)
	if err != nil {
		return err
	}
	defer writer.Close()

	return tw.WriteResultTo(writer)
}

func formatDateString(val reflect.Value, config *TextFormatConfig) string {
	return val.Interface().(date.Date).Format(config.Date)
}

func formatNullableDateString(val reflect.Value, config *TextFormatConfig) string {
	return val.Interface().(date.NullableDate).Format(config.Date)
}

func formatTimeString(val reflect.Value, config *TextFormatConfig) string {
	return val.Interface().(time.Time).Format(config.Time)
}

func formatDurationString(val reflect.Value, config *TextFormatConfig) string {
	return val.Interface().(time.Duration).String()
}

func formatMoneyAmountString(val reflect.Value, config *TextFormatConfig) string {
	return config.MoneyAmount.FormatAmount(val.Interface().(money.Amount))
}

func formatMoneyCurrencyAmountString(val reflect.Value, config *TextFormatConfig) string {
	return config.MoneyAmount.FormatCurrencyAmount(val.Interface().(money.CurrencyAmount))
}
