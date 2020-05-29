package excel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	fs "github.com/ungerik/go-fs"

	"github.com/domonda/go-structtable"
	"github.com/domonda/go-structtable/test"
)

func Test_RenderExcel(t *testing.T) {
	outputFile := fs.TempDir().Joinf("Test_%s.xlsx", time.Now())

	renderer, err := NewRenderer("Sheet 1")
	assert.NoError(t, err, "Sheet 1")

	err = structtable.Render(renderer, test.NewTable(30), true, structtable.DefaultReflectColumnTitles)
	assert.NoError(t, err, "WriteFile")

	err = renderer.AddSheet("Sheet 2")
	assert.NoError(t, err, "Sheet 2")

	renderer.WriteResultFile(outputFile)
	assert.True(t, outputFile.Exists(), "WriteFile")

	// t.Log("Written file:", outputFile)
}
