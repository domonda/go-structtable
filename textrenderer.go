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

// TextFormatRenderer has to be formatemented for a format
// to be used by TextRenderer.
type TextFormatRenderer interface {
	RenderBeginTableText(writer io.Writer) error
	RenderHeaderRowText(writer io.Writer, columnTitles []string) error
	RenderRowText(writer io.Writer, fields []string) error
	RenderEndTableText(writer io.Writer) error
}

// TextRenderer implements Renderer by using a TextFormatRenderer
// for a specific text based table format.
type TextRenderer struct {
	format         TextFormatRenderer
	config         *TextFormatConfig
	typeFormatters map[reflect.Type]TextFormatter
	buf            bytes.Buffer
	beginWritten   bool
}

func NewTextRenderer(format TextFormatRenderer, config *TextFormatConfig) *TextRenderer {
	tw := &TextRenderer{
		format: format,
		config: config,
	}
	return tw
}

// func (txt *TextRenderer) SetTypeTextFormatter(columnType reflect.Type, formatter TextFormatter) {
// 	if formatter != nil {
// 		txt.typeFormatters[columnType] = formatter
// 	} else {
// 		delete(txt.typeFormatters, columnType)
// 	}
// }

func (txt *TextRenderer) writeBeginIfMissing() error {
	if txt.beginWritten {
		return nil
	}
	err := txt.format.RenderBeginTableText(&txt.buf)
	if err != nil {
		return err
	}
	txt.beginWritten = true
	return nil
}

func (txt *TextRenderer) RenderHeaderRow(columnTitles []string) error {
	err := txt.writeBeginIfMissing()
	if err != nil {
		return err
	}
	return txt.format.RenderHeaderRowText(&txt.buf, columnTitles)
}

func (txt *TextRenderer) RenderRow(columnValues []reflect.Value) error {
	err := txt.writeBeginIfMissing()
	if err != nil {
		return err
	}
	fields := make([]string, len(columnValues))
	for i, val := range columnValues {
		fields[i] = txt.toString(val)
	}
	return txt.format.RenderRowText(&txt.buf, fields)
}

func (txt *TextRenderer) toString(val reflect.Value) string {
	valType := val.Type()
	derefVal, derefType := reflection.DerefValueAndType(val)

	if f, ok := txt.config.TypeFormatters[derefType]; ok && derefVal.IsValid() {
		// derefVal.IsValid() returns false for dereferenced nil pointer
		// so the following will only be called for non nil pointers:
		return f.FormatValue(derefVal, txt.config)
	}

	switch valType.Kind() {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return txt.config.Nil
		}
	}

	switch derefType.Kind() {
	case reflect.Bool:
		if derefVal.Bool() {
			return txt.config.True
		} else {
			return txt.config.False
		}

	case reflect.String:
		return derefVal.String()

	case reflect.Float32, reflect.Float64:
		return strfmt.FormatFloat(
			derefVal.Float(),
			txt.config.Float.ThousandsSep,
			txt.config.Float.DecimalSep,
			txt.config.Float.Precision,
			txt.config.Float.PadPrecision,
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

func (txt *TextRenderer) Result() ([]byte, error) {
	err := txt.format.RenderEndTableText(&txt.buf)
	if err != nil {
		return nil, err
	}
	return txt.buf.Bytes(), nil
}

func (txt *TextRenderer) WriteResultTo(writer io.Writer) error {
	_, err := txt.buf.WriteTo(writer)
	return err
}

func (txt *TextRenderer) WriteResultFile(file fs.File, perm ...fs.Permissions) error {
	writer, err := file.OpenWriter(perm...)
	if err != nil {
		return err
	}
	defer writer.Close()

	return txt.WriteResultTo(writer)
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
