package excel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fs "github.com/ungerik/go-fs"

	"github.com/domonda/go-structtable"
)

func TestNewReader(t *testing.T) {
	type row struct {
		Name  string
		Count int
	}
	renderer, err := NewRenderer("Data")
	require.NoError(t, err)
	err = structtable.Render(renderer, []row{{"a", 1}, {"b", 2}}, true, structtable.DefaultReflectColumnTitles)
	require.NoError(t, err)
	data, err := renderer.Result()
	require.NoError(t, err)
	file := fs.NewMemFile("test.xlsx", data)

	t.Run("reads rendered sheet", func(t *testing.T) {
		reader, err := NewReader(t.Context(), file, "Data")
		require.NoError(t, err)
		assert.Equal(t, "Data", reader.SheetName())
		assert.Equal(t, 3, reader.NumRows())
	})

	t.Run("missing sheet", func(t *testing.T) {
		_, err := NewReader(t.Context(), file, "Missing")
		assert.Error(t, err)
	})

	t.Run("cancelled context", func(t *testing.T) {
		// The ctx is passed through to the file read,
		// so a cancelled ctx must stop reading the file.
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		_, err := NewReader(ctx, file, "Data")
		assert.ErrorIs(t, err, context.Canceled)
	})
}
