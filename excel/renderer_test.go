package excel

import (
	"testing"
	"time"

	"github.com/domonda/go-structtable"
	"github.com/domonda/go-structtable/test"
	"github.com/stretchr/testify/assert"
	fs "github.com/ungerik/go-fs"
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

func Test_sanitizeSheetName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "", want: "UNNAMED"},
		{name: " ", want: "UNNAMED"},
		{name: "*[X]*", want: "__X__"},
		{name: " 123456789*123456789\\123456789/123456789 ", want: "123456789_123456789_123456789_…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeSheetName(tt.name); got != tt.want {
				t.Errorf("sanitizeSheetName() = %v, want %v", got, tt.want)
			}
		})
	}
}
