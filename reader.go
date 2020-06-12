package structtable

import (
	"reflect"

	"github.com/domonda/go-errs"
)

type Reader interface {
	NumRows() int
	ReadRowStrings(index int) ([]string, error)
	ReadRow(index int, destStruct reflect.Value) error
}

func Read(reader Reader, structSlicePtr interface{}, numHeaderRows int) (headerRows [][]string, err error) {
	if numHeaderRows < 0 {
		return nil, errs.New("numHeaderRows can't be negative")
	}
	destVal := reflect.ValueOf(structSlicePtr)
	if destVal.Kind() != reflect.Ptr {
		return nil, errs.Errorf("structSlicePtr must be pointer to a struct slice, but is %T", structSlicePtr)
	}
	if destVal.IsNil() {
		return nil, errs.Errorf("structSlicePtr must not be nil")
	}
	sliceType := destVal.Elem().Type()
	if sliceType.Kind() != reflect.Slice {
		return nil, errs.Errorf("structSlicePtr must be pointer to a struct slice, but is %T", structSlicePtr)
	}
	structType := sliceType.Elem()
	isSliceOfPtr := structType.Kind() == reflect.Ptr
	if isSliceOfPtr {
		structType = structType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return nil, errs.Errorf("structSlicePtr must be pointer to a struct slice, but is %T", structSlicePtr)
	}

	for i := 0; i < numHeaderRows && i < reader.NumRows(); i++ {
		row, err := reader.ReadRowStrings(i)
		if err != nil {
			return nil, err
		}
		headerRows = append(headerRows, row)
	}

	numRows := reader.NumRows() - numHeaderRows
	sliceVal := reflect.MakeSlice(sliceType, numRows, numRows)
	for i := 0; i < numRows; i++ {
		var destStruct reflect.Value
		if isSliceOfPtr {
			// Allocate new struct pointer
			ptrVal := reflect.New(structType)
			// and assign to slice delement
			sliceVal.Index(i).Set(ptrVal)
			// Don't need pointer for ReadRow, reflect.Value is writeable
			destStruct = ptrVal.Elem()
		} else {
			destStruct = sliceVal.Index(i)
		}
		err := reader.ReadRow(int(numHeaderRows)+i, destStruct)
		if err != nil {
			return nil, err
		}
	}

	// Assign result only if there was no error
	destVal.Elem().Set(sliceVal)
	return headerRows, nil
}
