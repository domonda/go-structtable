package excel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	fs "github.com/ungerik/go-fs"

	"github.com/domonda/go-structtable"
	"github.com/domonda/go-structtable/test"
)

func Test_WriteExcel(t *testing.T) {
	outputFile := fs.TempDir().Joinf("Test_%s.xlsx", time.Now())

	writer, err := NewWriter("Sheet 1")
	assert.NoError(t, err, "Sheet 1")

	err = structtable.WriteReflectColumnTitles(writer, test.NewTable(30), "title")
	assert.NoError(t, err, "WriteFile")

	err = writer.AddSheet("Sheet 2")
	assert.NoError(t, err, "Sheet 2")

	writer.WriteResultFile(outputFile)
	assert.True(t, outputFile.Exists(), "WriteFile")

	// t.Log("Written file:", outputFile)
}
