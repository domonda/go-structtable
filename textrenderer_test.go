package structtable_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fs "github.com/ungerik/go-fs"

	structtable "github.com/domonda/go-structtable"
	"github.com/domonda/go-structtable/csv"
)

// TestTextRenderer_RepeatedWrite verifies that writing the result of a
// TextRenderer does not consume it: every WriteResultTo and WriteResultFile
// call writes the same complete table that Result returns,
// so a retry after a failed upload does not write an empty file.
//
// Value: protects=TextRenderer output staying complete across repeated writes; fails_when=WriteResultTo drains the buffer (bytes.Buffer.WriteTo) or Result renders the post-table content more than once; why_new=no test wrote a result twice; seam=none; no_claim=formats with post-table content (CSV renders none)
func TestTextRenderer_RepeatedWrite(t *testing.T) {
	type row struct {
		Name  string
		Count int
	}
	renderer := csv.NewRenderer(nil)
	err := structtable.Render(renderer, []row{{"a", 1}, {"b", 2}}, true, structtable.DefaultReflectColumnTitles)
	require.NoError(t, err)

	expected, err := renderer.Result()
	require.NoError(t, err)
	require.NotEmpty(t, expected)
	expected = bytes.Clone(expected)

	for i := range 2 {
		var buf bytes.Buffer
		require.NoError(t, renderer.WriteResultTo(&buf))
		assert.Equal(t, string(expected), buf.String(), "WriteResultTo call %d", i+1)
	}

	file := fs.File(filepath.Join(t.TempDir(), "out.csv"))
	for i := range 2 {
		require.NoError(t, renderer.WriteResultFile(file))
		written, err := file.ReadAll(t.Context())
		require.NoError(t, err)
		assert.Equal(t, string(expected), string(written), "WriteResultFile call %d", i+1)
	}

	result, err := renderer.Result()
	require.NoError(t, err)
	assert.Equal(t, string(expected), string(result), "Result after writing")
}
