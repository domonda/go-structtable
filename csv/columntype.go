package csv

import (
	"strconv"
	"strings"

	"github.com/domonda/go-types/bank"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/money"
)

type DataType string

const (
	DataTypeString              DataType = "STRING"
	DataTypeNullableString      DataType = "NULL_STRING"
	DataTypeInt                 DataType = "INT"
	DataTypeNullableInt         DataType = "NULL_INT"
	DataTypeFloat               DataType = "FLOAT"
	DataTypeNullableFloat       DataType = "NULL_FLOAT"
	DataTypeMoneyAmount         DataType = "MONEY_AMOUNT"
	DataTypeNullableMoneyAmount DataType = "NULL_MONEY_AMOUNT"
	DataTypeCurrency            DataType = "CURRENCY"
	DataTypeNullableCurrency    DataType = "NULL_CURRENCY"
	DataTypeDate                DataType = "DATE"
	DataTypeNullableDate        DataType = "NULL_DATE"
	DataTypeTime                DataType = "TIME"
	DataTypeNullableTime        DataType = "NULL_TIME"
	DataTypeIBAN                DataType = "IBAN"
	DataTypeNullableIBAN        DataType = "NULL_IBAN"
	DataTypeBIC                 DataType = "BIC"
	DataTypeNullableBIC         DataType = "NULL_BIC"
)

func (t DataType) Valid() bool {
	switch t {
	case DataTypeString,
		DataTypeNullableString,
		DataTypeInt,
		DataTypeNullableInt,
		DataTypeFloat,
		DataTypeNullableFloat,
		DataTypeMoneyAmount,
		DataTypeNullableMoneyAmount,
		DataTypeCurrency,
		DataTypeNullableCurrency,
		DataTypeDate,
		DataTypeNullableDate,
		DataTypeTime,
		DataTypeNullableTime,
		DataTypeIBAN,
		DataTypeNullableIBAN,
		DataTypeBIC,
		DataTypeNullableBIC:
		return true
	}
	return false
}

func (t DataType) Nullable() bool {
	return strings.HasPrefix(string(t), "NULL_")
}

// StringDataTypes returns valid non nullable data types for
// the passed string.
// DataTypeString is not returned because it's always valid.
func StringDataTypes(str string) []DataType {
	var types []DataType
	if _, err := strconv.ParseInt(str, 10, 64); err == nil {
		types = append(types, DataTypeInt)
	}
	if _, err := float.Parse(str); err == nil {
		types = append(types, DataTypeFloat)
	}
	if _, err := money.ParseAmount(str); err == nil {
		types = append(types, DataTypeMoneyAmount)
	}
	if money.StringIsCurrency(str) {
		types = append(types, DataTypeCurrency)
	}
	if date.StringIsDate(str) {
		types = append(types, DataTypeDate)
	}
	if _, ok := date.ParseTime(str); ok {
		types = append(types, DataTypeTime)
	}
	if bank.StringIsIBAN(str) {
		types = append(types, DataTypeIBAN)
	}
	if bank.StringIsBIC(str) {
		types = append(types, DataTypeBIC)
	}
	return types
}
