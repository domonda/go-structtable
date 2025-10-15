package csv

import (
	"strings"
)

type Modifier interface {
	Name() string
	Modify(rows [][]string) [][]string
}

var ModifiersByName = map[string]Modifier{
	SetRowsWithNonUniformColumnsNilModifier{}.Name(): SetRowsWithNonUniformColumnsNilModifier{},
	SetEmptyRowsNilModifier{}.Name():                 SetEmptyRowsNilModifier{},
	RemoveEmptyRowsModifier{}.Name():                 RemoveEmptyRowsModifier{},
	CompactSpacedStringsModifier{}.Name():            CompactSpacedStringsModifier{},
	RemoveTopRowModifier{}.Name():                    RemoveTopRowModifier{},
	RemoveBottomRowModifier{}.Name():                 RemoveBottomRowModifier{},
	SetTopRowNilModifier{}.Name():                    SetTopRowNilModifier{},
	SetBottomRowNilModifier{}.Name():                 SetBottomRowNilModifier{},
	ReplaceNewlineWithSpaceModifier{}.Name():         ReplaceNewlineWithSpaceModifier{},
}

type SetRowsWithNonUniformColumnsNilModifier struct{}

func (m SetRowsWithNonUniformColumnsNilModifier) Name() string {
	return "SetRowsWithNonUniformColumnsNil"
}

func (m SetRowsWithNonUniformColumnsNilModifier) Modify(rows [][]string) [][]string {
	return SetRowsWithNonUniformColumnsNil(rows)
}

// SetRowsWithNonUniformColumnsNil set rows to nil that don't have the same field count as the majority of rows,
// so every rows is either nil or has the same number of fields.
func SetRowsWithNonUniformColumnsNil(rows [][]string) [][]string {
	if len(rows) == 0 {
		return nil
	}

	result := make([][]string, len(rows))

	// map from number of columns to number of rows with that column
	rowColumnsCount := make(map[int]int)
	for _, row := range rows {
		if rowColumns := len(row); rowColumns > 1 {
			rowColumnsCount[rowColumns]++
		}
	}
	majorityRowColumns := 0
	highestRowCount := 0
	for rowColumns, rowCount := range rowColumnsCount {
		if rowCount > highestRowCount || (rowCount == highestRowCount && rowColumns > majorityRowColumns) {
			majorityRowColumns = rowColumns
			highestRowCount = rowCount
		}
	}
	for i, row := range rows {
		if len(row) == majorityRowColumns {
			result[i] = row
		}
	}

	return result
}

type SetEmptyRowsNilModifier struct{}

func (m SetEmptyRowsNilModifier) Name() string {
	return "SetEmptyRowsNil"
}

func (m SetEmptyRowsNilModifier) Modify(rows [][]string) [][]string {
	return SetEmptyRowsNil(rows)
}

// SetEmptyRowsNil sets rows to nil,
// where all columns are empty strings.
func SetEmptyRowsNil(rows [][]string) [][]string {
	if len(rows) == 0 {
		return nil
	}

	result := make([][]string, len(rows))
	for i, row := range rows {
		rowIsEmpty := true
		for _, field := range row {
			if field != "" {
				rowIsEmpty = false
				break
			}
		}
		if !rowIsEmpty {
			result[i] = row
		}
	}

	return result
}

type RemoveEmptyRowsModifier struct{}

func (m RemoveEmptyRowsModifier) Name() string {
	return "RemoveEmptyRows"
}

func (m RemoveEmptyRowsModifier) Modify(rows [][]string) [][]string {
	return RemoveEmptyRows(rows)
}

// RemoveEmptyRows removes rows without columns,
// or rows where all columns are empty strings.
func RemoveEmptyRows(rows [][]string) [][]string {
	if len(rows) == 0 {
		return nil
	}
	var (
		hasEmptyRows bool
		nonEmptyRows [][]string
	)
	for i, row := range rows {
		rowIsEmpty := true
		for _, field := range row {
			if field != "" {
				rowIsEmpty = false
				break
			}
		}
		if rowIsEmpty {
			if !hasEmptyRows {
				if i > 0 {
					nonEmptyRows = append(nonEmptyRows, rows[:i]...)
				}
				hasEmptyRows = true
			}
		} else {
			if hasEmptyRows {
				nonEmptyRows = append(nonEmptyRows, row)
			}
		}
	}
	if !hasEmptyRows {
		// Nothing removed, return original rows
		return rows
	}
	return nonEmptyRows
}

type CompactSpacedStringsModifier struct{}

func (m CompactSpacedStringsModifier) Name() string {
	return "CompactSpacedStrings"
}

func (m CompactSpacedStringsModifier) Modify(rows [][]string) [][]string {
	CompactSpacedStrings(rows)
	return rows
}

// CompactSpacedStrings removes spaces if they are between every other character,
// meaning that every odd character index is a space.
func CompactSpacedStrings(rows [][]string) (numModified int) {
	for _, row := range rows {
		for col, field := range row {
			cleaned, modified := compactSpacedString(field)
			if modified {
				row[col] = cleaned
				numModified++
			}
		}
	}
	return numModified
}

// compactSpacedString removes spaces if they are between every other character,
// meaning that every odd character index is a space.
func compactSpacedString(str string) (cleaned string, modified bool) {
	if len(str) < 3 {
		return str, false
	}

	// First check if every odd indexed rune is a space.
	numSpaces := 0
	i := 0 // Don't use index from range over string because it counts bytes not UTF-8 runes
	for _, r := range str {
		if i&1 == 1 {
			if r != ' ' {
				return str, false
			}
			numSpaces++
		}
		i++
	}

	b := strings.Builder{}
	b.Grow(len(str) - numSpaces)
	i = 0
	for _, r := range str {
		if i&1 == 0 {
			b.WriteRune(r)
		}
		i++
	}
	return b.String(), true
}

// RemoveTopRowModifier removes the given number of rows at the top
type RemoveTopRowModifier struct{}

func (m RemoveTopRowModifier) Name() string {
	return "RemoveTopRow"
}

func (m RemoveTopRowModifier) Modify(rows [][]string) [][]string {
	if len(rows) < 2 {
		return nil
	}
	return rows[1:]
}

// RemoveBottomRowModifier removes the given number of rows at the bottom
type RemoveBottomRowModifier struct{}

func (m RemoveBottomRowModifier) Name() string {
	return "RemoveBottomRow"
}

func (m RemoveBottomRowModifier) Modify(rows [][]string) [][]string {
	if len(rows) < 2 {
		return nil
	}
	return rows[:len(rows)-1]
}

// SetTopRowNilModifier removes the given number of rows at the top
type SetTopRowNilModifier struct{}

func (m SetTopRowNilModifier) Name() string {
	return "SetTopRowNil"
}

func (m SetTopRowNilModifier) Modify(rows [][]string) [][]string {
	if len(rows) > 0 {
		rows[0] = nil
	}
	return rows
}

// SetBottomRowNilModifier removes the given number of rows at the bottom
type SetBottomRowNilModifier struct{}

func (m SetBottomRowNilModifier) Name() string {
	return "SetBottomRowNil"
}

func (m SetBottomRowNilModifier) Modify(rows [][]string) [][]string {
	if len(rows) > 0 {
		rows[len(rows)-1] = nil
	}
	return rows
}

// // RemoveTopRowsModifier removes the given number of rows at the top
// type RemoveTopRowsModifier uint

// func (m RemoveTopRowsModifier) Name() string {
// 	return "RemoveTopRows"
// }

// func (m RemoveTopRowsModifier) Modify(rows [][]string) [][]string {
// 	if len(rows) <= int(m) {
// 		return nil
// 	}
// 	return rows[int(m):]
// }

// // RemoveBottomRowsModifier removes the given number of rows at the bottom
// type RemoveBottomRowsModifier uint

// func (m RemoveBottomRowsModifier) Name() string {
// 	return "RemoveBottomRows"
// }

// func (m RemoveBottomRowsModifier) Modify(rows [][]string) [][]string {
// 	if len(rows) <= int(m) {
// 		return nil
// 	}
// 	return rows[:len(rows)-int(m)]
// }

func ReplaceNewlineWithSpacefunc(rows [][]string) {
	for _, row := range rows {
		for col, field := range row {
			row[col] = strings.ReplaceAll(field, "\n", " ")
		}
	}
}

type ReplaceNewlineWithSpaceModifier struct{}

func (m ReplaceNewlineWithSpaceModifier) Name() string {
	return "ReplaceNewlineWithSpace"
}

func (m ReplaceNewlineWithSpaceModifier) Modify(rows [][]string) [][]string {
	ReplaceNewlineWithSpacefunc(rows)
	return rows
}
