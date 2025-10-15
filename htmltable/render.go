package htmltable

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/domonda/go-structtable"
	"github.com/domonda/go-types/strfmt"
)

// Renderer implements HTML table rendering with embedded CSS styling.
//
// This renderer extends the base HTMLRenderer with predefined CSS styles for
// borders, colors, and layout. It generates random CSS class names to avoid
// conflicts when multiple tables are rendered on the same page.
type Renderer struct {
	*structtable.HTMLRenderer
}

// NewRenderer creates a new HTML table renderer with random CSS class names and a caption.
//
// This constructor creates an HTML table renderer with embedded CSS styling and
// generates random CSS class names to prevent conflicts when multiple tables
// are rendered on the same page. The renderer includes predefined styles for
// borders, alternating row colors, and caption formatting.
//
// Parameters:
//   - caption: Table caption text that will be displayed above the table
//   - config: Text formatting configuration for cell values
//
// Returns:
//   - A new Renderer instance with embedded CSS styling
//
// CSS Features:
//   - Collapsed borders with black 1px solid lines
//   - 6px horizontal and 12px vertical padding
//   - Header row with dark background (#00000052)
//   - Alternating row colors (light gray and white)
//   - Large caption font (1.4em)
//
// Example:
//
//	renderer := htmltable.NewRenderer("Employee Data", strfmt.NewEnglishFormatConfig())
func NewRenderer(caption string, config *strfmt.FormatConfig) *Renderer {
	r := &Renderer{}
	//#nosec G404 -- weak random number OK
	table := &structtable.HTMLTableConfig{
		Caption:        caption,
		CaptionClass:   fmt.Sprintf("c%d", rand.Uint32()),
		TableClass:     fmt.Sprintf("c%d", rand.Uint32()),
		CellClass:      fmt.Sprintf("c%d", rand.Uint32()),
		HeaderRowClass: fmt.Sprintf("c%d", rand.Uint32()),
		DataRowClass:   fmt.Sprintf("c%d", rand.Uint32()),
	}
	r.HTMLRenderer = structtable.NewHTMLRenderer(r, table, config)
	return r
}

// Render is a convenience function to render an HTML table with English text formatting in one call.
//
// This function creates a renderer, configures it with English text formatting,
// and renders the data to the writer in a single operation. It's useful for
// quick HTML table generation without needing to manage the renderer lifecycle.
//
// Parameters:
//   - writer: The io.Writer to write the HTML table to
//   - structSlice: The data to render (slice of structs)
//   - caption: Table caption text
//   - renderHeaderRow: Whether to include a header row
//   - columnMapper: Column mapping configuration
//
// Returns:
//   - err: Any error that occurred during rendering
//
// Example:
//
//	var employees []Employee
//	err := htmltable.Render(writer, employees, "Employee List", true, columnMapper)
func Render(writer io.Writer, structSlice any, caption string, renderHeaderRow bool, columnMapper structtable.ColumnMapper) error {
	renderer := NewRenderer(caption, strfmt.NewEnglishFormatConfig())
	return structtable.RenderTo(writer, renderer, structSlice, renderHeaderRow, columnMapper)
}

// RenderBeforeTable writes embedded CSS styles before the HTML table.
//
// This method generates and writes CSS styles that define the appearance of the
// HTML table. The styles include borders, padding, alternating row colors, and
// caption formatting. The CSS uses randomly generated class names to avoid
// conflicts with other tables on the same page.
//
// Parameters:
//   - writer: The io.Writer to write the CSS styles to
//
// Returns:
//   - err: Any error that occurred during writing
//
// CSS Styles Applied:
//   - Table borders: 1px solid black with collapsed borders
//   - Cell padding: 6px horizontal, 12px vertical
//   - Caption: Large font (1.4em), left-aligned with bottom margin
//   - Header row: Dark background (#00000052)
//   - Data rows: Alternating colors (light gray #00000014 and white #ffffff)
//
// Example:
//
//	// This method is called automatically by the HTMLRenderer
//	err := renderer.RenderBeforeTable(writer)
func (r *Renderer) RenderBeforeTable(writer io.Writer) error {
	_, err := fmt.Fprintf(
		writer,
		`<style type='text/css'>
			table.%s, th.%s, td.%s, th.%s {
				border-collapse: collapse;
				border: 1px solid black;
				padding: 6px 12px;
			}
			caption.%s {
				font-size: 1.4em;
				text-align: left;
				margin-bottom: 8px;
			}
			tr.%s {
				background-color: #00000052
			}
			tr.%s:nth-child(odd) {
				background-color: #00000014
			}
			tr.%s:nth-child(even) {
				background-color: #ffffff
			}
		</style>`,
		r.TableConfig.TableClass, r.TableConfig.CellClass, r.TableConfig.CellClass, r.TableConfig.HeaderRowClass,
		r.TableConfig.CaptionClass,
		r.TableConfig.HeaderRowClass,
		r.TableConfig.DataRowClass,
		r.TableConfig.DataRowClass,
	)
	return err
}
