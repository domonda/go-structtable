package texttable

type Table interface {
	NumRows() int
	// NumCols() int

	// NumRowCells returns the number of cells in a row.
	// Rows may have less cells than the whole table has columns.
	// Returns zero for a non existing row.
	NumRowCells(row int) int

	// CellExists returns if there is a cell available at the given row and column position.
	CellExists(row, col int) bool

	// CellText returns the text of a cell
	// or an empty string if the cell does not exist.
	CellText(row, col int) string

	// HasCellBoundingBoxes returns if cell bounding boxes are available.
	// See CellBoundingBox
	HasCellBoundingBoxes() bool

	// CellBoundingBox returns the bounding box of a cell
	// or nil if the cell does not exist or has no bounding box information.
	// See HasCellBoundingBoxes
	CellBoundingBox(row, col int) *BoundingBox
}
