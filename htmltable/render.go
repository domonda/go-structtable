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
		TableClass:     fmt.Sprintf("%d", rand.Uint32()),
		HeaderRowClass: fmt.Sprintf("%d", rand.Uint32()),
		DataRowClass:   fmt.Sprintf("%d", rand.Uint32()),
		DataCellClass:  fmt.Sprintf("%d", rand.Uint32()),
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
	_, err := fmt.Fprintf(writer, `<style type='text/css'>
	table.%s, td.%s, th.%s {
		border: 1px solid black;
		padding: 4px;
		white-space: nowrap;
	}
	tr.%s {
		background-color: #00000052
	}
	tr.%s:nth-child(odd) {
		background-color: #00000014
	}
</style>`,
		r.TableConfig.TableClass, r.TableConfig.DataCellClass, r.TableConfig.HeaderRowClass,
		r.TableConfig.HeaderRowClass,
		r.TableConfig.DataRowClass,
	)
	return err
}
