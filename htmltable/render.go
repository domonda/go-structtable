package htmltable

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/domonda/go-structtable"
	"github.com/domonda/go-types/txtfmt"
)

type Renderer struct {
	*structtable.HTMLRenderer
}

func NewRenderer(caption string, config *txtfmt.FormatConfig) *Renderer {
	r := &Renderer{}
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

// Render is a shortcut to render a HTML table with english text formating
func Render(writer io.Writer, structSlice interface{}, caption string, renderHeaderRow bool, columnMapper structtable.ColumnMapper) error {
	renderer := NewRenderer(caption, txtfmt.NewEnglishFormatConfig())
	return structtable.RenderTo(writer, renderer, structSlice, renderHeaderRow, columnMapper)
}

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
