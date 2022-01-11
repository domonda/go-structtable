package test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
)

type Struct struct {
	Bool           bool
	String         string
	Bytes          []byte `col:"[]byte string,IGNORE_AFTER_COMMA"`
	Ignore         string `col:"-"`
	Int            int
	IntPtr         *int
	Uint16         uint16
	Float          float64
	Currency       money.Currency
	MoneyAmount    money.Amount
	CurrencyAmount money.CurrencyAmount
	Time           time.Time
	TimePtr        *time.Time
	Duration       time.Duration
	Date           date.Date
}

// NewTable returns a new test table.
func NewTable(numRows int) []Struct {
	rows := make([]Struct, numRows)
	//#nosec G404 -- weak random numbers OK
	for i := range rows {
		rows[i].Bool = i%2 > 0
		rows[i].String = fmt.Sprintf("String %d", i)
		rows[i].Bytes = []byte(fmt.Sprintf("Bytes %d", i))
		rows[i].Int = i
		if i%2 == 0 {
			rows[i].IntPtr = &rows[i].Int
		}
		rows[i].Uint16 = uint16(i)
		rows[i].Float = rand.Float64() * 1000
		rows[i].Currency = []money.Currency{"", money.EUR, money.USD}[i%3]
		rows[i].MoneyAmount = money.Amount(rand.Float64() * 100000)
		rows[i].CurrencyAmount = money.CurrencyAmount{
			Currency: []money.Currency{"", money.EUR, money.USD}[i%3],
			Amount:   money.Amount(rand.Float64() * 100000),
		}
		rows[i].Duration = time.Hour*time.Duration(i) + time.Minute*59 + time.Second
		rows[i].Time = time.Date(2012, 12, 12, 12, 12, 12, 0, time.Local)
		if i%2 == 0 {
			rows[i].TimePtr = &rows[i].Time
			rows[i].Date = "2012-12-12"
		}
	}
	return rows
}
