package structtable_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/domonda/go-types/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fs "github.com/ungerik/go-fs"

	structtable "github.com/domonda/go-structtable"
	"github.com/domonda/go-structtable/htmltable"
)

// TestHTMLRenderer_CompleteRepeatedWrite verifies that WriteResultTo and
// WriteResultFile write the complete HTML table including its closing tags,
// exactly once, and that writing does not consume the result,
// so a retry after a failed upload does not write an empty file.
//
// Value: protects=HTML files ending with exactly one </tbody></table> and staying complete across repeated writes and Result calls; fails_when=WriteResultTo skips the closing tags, drains the buffer, or Result appends the closing tags on every call; why_new=no test wrote an HTML result to a writer or file; seam=none; no_claim=HTML formatting of cells (covered by htmltable tests)
func TestHTMLRenderer_CompleteRepeatedWrite(t *testing.T) {
	type row struct {
		Name  string
		Count int
	}
	const closingTags = "</tbody></table>\n"
	renderer := htmltable.NewRenderer("caption", strfmt.NewFormatConfig())
	err := structtable.Render(renderer, []row{{"a", 1}, {"b", 2}}, true, structtable.DefaultReflectColumnTitles)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, renderer.WriteResultTo(&buf))
	expected := buf.String()
	assert.True(t, strings.HasSuffix(expected, closingTags), "WriteResultTo ends with closing tags")
	assert.Equal(t, 1, strings.Count(expected, closingTags), "closing tags written once")

	for i := range 2 {
		result, err := renderer.Result()
		require.NoError(t, err)
		assert.Equal(t, expected, string(result), "Result call %d", i+1)
	}

	buf.Reset()
	require.NoError(t, renderer.WriteResultTo(&buf))
	assert.Equal(t, expected, buf.String(), "second WriteResultTo")

	file := fs.File(filepath.Join(t.TempDir(), "out.html"))
	for i := range 2 {
		require.NoError(t, renderer.WriteResultFile(file))
		written, err := file.ReadAll(t.Context())
		require.NoError(t, err)
		assert.Equal(t, expected, string(written), "WriteResultFile call %d", i+1)
	}
}
