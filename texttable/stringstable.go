package texttable

// StringsTable implements Table for a 2D slice of strings.
//
// This is a simple implementation of the Table interface that works directly
// with a 2D slice of strings. It does not provide bounding box information.
type StringsTable [][]string

// NumRows returns the total number of rows in the table.
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

// NumRowCells returns the number of cells in the specified row.
func (t StringsTable) NumRowCells(row int) int {
	if row < 0 || row >= len(t) {
		return 0
	}
	return len(t[row])
}

// CellExists returns true if there is a cell at the given row and column position.
func (t StringsTable) CellExists(row, col int) bool {
	return row >= 0 && row < len(t) && col >= 0 && col < len(t[row])
}

// CellText returns the text content of a cell.
// Returns an empty string if the cell does not exist.
func (t StringsTable) CellText(row, col int) string {
	if !t.CellExists(row, col) {
		return ""
	}
	return t[row][col]
}

// HasCellBoundingBoxes returns false since StringsTable does not provide bounding box information.
func (t StringsTable) HasCellBoundingBoxes() bool {
	return false
}

// CellBoundingBox returns nil since StringsTable does not provide bounding box information.
func (t StringsTable) CellBoundingBox(row, col int) *BoundingBox {
	return nil
}
