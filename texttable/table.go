package texttable

// Table defines the interface for accessing tabular data.
//
// This interface provides methods for reading table data, including cell values,
// dimensions, and optional bounding box information for each cell.
type Table interface {
	// NumRows returns the total number of rows in the table.
	NumRows() int
	// NumRowCells returns the number of cells in a specific row.
	// Rows may have fewer cells than the table's maximum column count.
	// Returns zero for non-existent rows.
	NumRowCells(row int) int

	// CellExists returns true if there is a cell at the given row and column position.
	CellExists(row, col int) bool

	// CellText returns the text content of a cell.
	// Returns an empty string if the cell does not exist.
	CellText(row, col int) string

	// HasCellBoundingBoxes returns true if cell bounding box information is available.
	// See CellBoundingBox for more details.
	HasCellBoundingBoxes() bool

	// CellBoundingBox returns the bounding box of a cell.
	// Returns nil if the cell does not exist or has no bounding box information.
	// See HasCellBoundingBoxes for availability check.
	CellBoundingBox(row, col int) *BoundingBox
}
