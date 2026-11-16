package csv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fs "github.com/ungerik/go-fs"
)

func TestNewReaderFromFile(t *testing.T) {
	file := fs.NewMemFile("test.csv", []byte("A;B\r\nC;D"))

	t.Run("reads rows", func(t *testing.T) {
		reader, err := NewReaderFromFile(t.Context(), file, NewFormat(";"), "\n", nil, nil)
		require.NoError(t, err)
		require.Equal(t, 2, reader.NumRows())
		row, err := reader.ReadRowStrings(1)
		require.NoError(t, err)
		assert.Equal(t, []string{"C", "D"}, row)
	})

	t.Run("cancelled context", func(t *testing.T) {
		// The ctx is passed through to the file read,
		// so a cancelled ctx must stop reading the file.
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		_, err := NewReaderFromFile(ctx, file, NewFormat(";"), "\n", nil, nil)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
