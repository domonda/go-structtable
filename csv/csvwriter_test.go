package csv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/domonda/go-structtable"
	"github.com/domonda/go-structtable/test"
	"github.com/domonda/go-types/charset"
)

func Test_WriteCSV(t *testing.T) {
	writer := NewWriter(structtable.NewTextFormatConfig())
	err := structtable.WriteReflectColumnTitles(writer, test.NewTable(5), "title")
	assert.NoError(t, err, "WriteFile")

	result, err := writer.Result()
	assert.NoError(t, err, "Result")

	// fmt.Print(string(result))
	// t.Fail()

	const expectedCSV = `Bool;String;[]byte string;Int;IntPtr;Uint16;Float;Currency;MoneyAmount;CurrencyAmount;Time;TimePtr;Duration;Date
false;String 0;Bytes 0;0;0;0;604.6602879796196;;94,050.91;66,456.01;2012-12-12T12:12:12+01:00;2012-12-12T12:12:12+01:00;59m1s;2012-12-12
true;String 1;Bytes 1;1;;1;437.7141871869802;EUR;42,463.75;EUR 68,682.31;2012-12-12T12:12:12+01:00;;1h59m1s;
false;String 2;Bytes 2;2;2;2;65.63701921747622;USD;15,651.93;USD 9,696.95;2012-12-12T12:12:12+01:00;2012-12-12T12:12:12+01:00;2h59m1s;2012-12-12
true;String 3;Bytes 3;3;;3;300.91186058528706;;51,521.26;81,364.00;2012-12-12T12:12:12+01:00;;3h59m1s;
false;String 4;Bytes 4;4;4;4;214.26387258237492;EUR;38,065.72;EUR 31,805.82;2012-12-12T12:12:12+01:00;2012-12-12T12:12:12+01:00;4h59m1s;2012-12-12
`
	expected := append([]byte(charset.BOMUTF8), []byte(expectedCSV)...)
	result = bytes.Replace(result, []byte{'\r'}, []byte{}, -1)

	assert.Equal(t, string(expected), string(result), "Comparing CSV output")
}
