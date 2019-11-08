package texttable

// StringsTable implements Table for a 2D slice of strings.
type StringsTable [][]string

func (t StringsTable) NumRows() int {
	return len(t)
}

// func (t StringsTable) NumCols() int {
// 	cols := 0
// 	for _, row := range t {
// 		if n := len(row); n > cols {
// 			cols = n
// 		}
// 	}
// 	return cols
// }

func (t StringsTable) NumRowCells(row int) int {
	if row < 0 || row >= len(t) {
		return 0
	}
	return len(t[row])
}

func (t StringsTable) CellExists(row, col int) bool {
	return row >= 0 && row < len(t) && col >= 0 && col < len(t[row])
}

func (t StringsTable) CellText(row, col int) string {
	if !t.CellExists(row, col) {
		return ""
	}
	return t[row][col]
}

func (t StringsTable) HasCellBoundingBoxes() bool {
	return false
}

func (t StringsTable) CellBoundingBox(row, col int) *BoundingBox {
	return nil
}
