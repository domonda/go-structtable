package htmltable

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/domonda/go-structtable"
)

func TestRenderPreservesEnglishFormatting(t *testing.T) {
	type row struct {
		When   time.Time
		Active bool
	}

	rows := []row{
		{When: time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC), Active: true},
		{Active: false},
	}
	var output bytes.Buffer
	if err := Render(&output, rows, "", true, &structtable.ReflectColumnTitles{}); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"02/01/2024 03:04:05 UTC", ">yes</td>", ">no</td>"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("rendered HTML does not contain %q: %s", expected, output.String())
		}
	}
}
