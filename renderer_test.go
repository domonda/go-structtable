package structtable_test

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"testing"

	"github.com/domonda/go-types/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fs "github.com/ungerik/go-fs"
	"github.com/ungerik/go-fs/fstest"

	structtable "github.com/domonda/go-structtable"
	"github.com/domonda/go-structtable/csv"
	"github.com/domonda/go-structtable/excel"
	"github.com/domonda/go-structtable/htmltable"
)

// TestRenderFile verifies that RenderFile writes exactly the bytes that
// RenderBytes produces to the target file, and that a write error is returned.
//
// Value: protects=RenderFile persisting rendered output via fs.File.WriteAll; fails_when=the go-fs WriteAll call is dropped, writes partial/empty data, or its error is swallowed; why_new=no test called RenderFile before; seam=none; no_claim=renderer formatting details (covered by csv renderer tests)
func TestRenderFile(t *testing.T) {
	type row struct {
		Name  string
		Count int
	}
	rows := []row{{"a", 1}, {"b", 2}}
	renderer := csv.NewRenderer(nil)

	t.Run("writes rendered bytes", func(t *testing.T) {
		file := fs.File(filepath.Join(t.TempDir(), "out.csv"))
		err := structtable.RenderFile(t.Context(), file, renderer, rows, true, structtable.DefaultReflectColumnTitles)
		require.NoError(t, err)

		expected, err := structtable.RenderBytes(csv.NewRenderer(nil), rows, true, structtable.DefaultReflectColumnTitles)
		require.NoError(t, err)
		require.NotEmpty(t, expected)

		written, err := file.ReadAll(t.Context())
		require.NoError(t, err)
		assert.Equal(t, string(expected), string(written))
	})

	t.Run("write error", func(t *testing.T) {
		file := fs.File(filepath.Join(t.TempDir(), "missing-dir", "out.csv"))
		err := structtable.RenderFile(t.Context(), file, csv.NewRenderer(nil), rows, true, structtable.DefaultReflectColumnTitles)
		assert.Error(t, err)
	})

	t.Run("cancelled context", func(t *testing.T) {
		// The ctx is passed through to the file write,
		// so a cancelled ctx must stop writing the file.
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		file := fs.File(filepath.Join(t.TempDir(), "out.csv"))
		err := structtable.RenderFile(ctx, file, csv.NewRenderer(nil), rows, true, structtable.DefaultReflectColumnTitles)
		assert.ErrorIs(t, err, context.Canceled)
		assert.False(t, file.Exists(), "no file written")
	})
}

// mockFileWriter records how a WriteResultFile implementation
// finishes a file writer: it fails Write with writeErr and Close with closeErr.
type mockFileWriter struct {
	writeErr error
	closeErr error
	closed   bool
}

func (w *mockFileWriter) Write(p []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	return len(p), nil
}

func (w *mockFileWriter) Close() error {
	w.closed = true
	return w.closeErr
}

// mockAbortingFileWriter is a mockFileWriter that implements fs.WriteAborter
// like the writer of an upload that commits the data only in Close.
type mockAbortingFileWriter struct {
	mockFileWriter
	aborted bool
}

func (w *mockAbortingFileWriter) Abort() error {
	w.aborted = true
	return nil
}

// TestWriteResultFile verifies how every WriteResultFile implementation
// finishes the file writer: the error of Close is returned, because uploads
// commit in Close, and after a failed write a writer implementing
// fs.WriteAborter is aborted instead of closed, so no partial data
// replaces the file.
//
// Value: protects=WriteResultFile reporting a failed commit in Close and never committing a partial write; fails_when=Close goes back to a deferred call whose error is ignored, the write error is dropped, or a failed write is closed instead of aborted; why_new=no test reached WriteResultFile with a failing writer; seam=none; no_claim=real remote file system behavior (mock file system stands in)
func TestWriteResultFile(t *testing.T) {
	var writer io.WriteCloser
	mockFS := &fstest.MockFileSystem{
		MockPrefix: "writeresultfile://",
		MockOpenWriter: func(filePath string, perm fs.Permissions) (io.WriteCloser, error) {
			return writer, nil
		},
	}
	require.NoError(t, fs.Register(mockFS))
	t.Cleanup(func() { _ = mockFS.Close() })

	type row struct {
		Name  string
		Count int
	}
	rows := []row{{"a", 1}, {"b", 2}}
	newRenderers := map[string]func(t *testing.T) structtable.Renderer{
		"text (csv)": func(t *testing.T) structtable.Renderer { return csv.NewRenderer(nil) },
		"html": func(t *testing.T) structtable.Renderer {
			return htmltable.NewRenderer("caption", strfmt.NewFormatConfig())
		},
		"excel": func(t *testing.T) structtable.Renderer {
			r, err := excel.NewRenderer("Data")
			require.NoError(t, err)
			return r
		},
	}
	writeErr := errors.New("write failed")
	closeErr := errors.New("upload failed on close")

	for name, newRenderer := range newRenderers {
		render := func(t *testing.T) structtable.Renderer {
			renderer := newRenderer(t)
			err := structtable.Render(renderer, rows, true, structtable.DefaultReflectColumnTitles)
			require.NoError(t, err)
			return renderer
		}

		t.Run(name+"/close error", func(t *testing.T) {
			w := &mockFileWriter{closeErr: closeErr}
			writer = w
			err := render(t).WriteResultFile(fs.File("writeresultfile://out"))
			assert.ErrorIs(t, err, closeErr)
			assert.True(t, w.closed, "writer closed")
		})

		t.Run(name+"/write error closes writer", func(t *testing.T) {
			w := &mockFileWriter{writeErr: writeErr, closeErr: closeErr}
			writer = w
			err := render(t).WriteResultFile(fs.File("writeresultfile://out"))
			assert.ErrorIs(t, err, writeErr)
			assert.ErrorIs(t, err, closeErr)
			assert.True(t, w.closed, "writer closed")
		})

		t.Run(name+"/write error aborts writer", func(t *testing.T) {
			w := &mockAbortingFileWriter{mockFileWriter: mockFileWriter{writeErr: writeErr}}
			writer = w
			err := render(t).WriteResultFile(fs.File("writeresultfile://out"))
			assert.ErrorIs(t, err, writeErr)
			assert.True(t, w.aborted, "writer aborted")
			assert.False(t, w.closed, "partial data not committed by Close")
		})
	}
}
