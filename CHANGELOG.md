# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

Releases are tagged with the Go module version in the `VERSION` file;
headings here leave out its `v` prefix.

> ⚠️ This package is deprecated in favour of
> [`github.com/domonda/go-retable`](https://github.com/domonda/go-retable).
> Entries here are fixes for consumers still on `go-structtable`.

## [0.1.0] - 2026-11-16

The first tagged release. Earlier commits were only available as Go
pseudo-versions. It brings the CSV parsing and format detection fixes ported
from `go-retable`'s `csvtable` package and moves to go-fs v1.

### Fixed

- Quoted fields that end with an escaped quote before the separator are read
  correctly, so a CSV column holding a JSON object as its last value no longer
  breaks the whole row.
- A field split across lines and across separators at the same time is joined
  back into one value. Previously a value containing both a newline and the
  separator inside its quotes was returned as several broken fields.
- An unterminated quote no longer swallows the rows after it. Only the field
  that opens it is affected; every following row is parsed normally.
- A closing quote followed by more characters no longer ends the field early,
  and a mis-quoted field keeps its last character.
- Two values containing newlines in the same row are read correctly. The second
  one used to gain one extra newline for every line the first one spanned.
- The field separator is detected from the column count it produces rather than
  from how often the character occurs, so text containing commas — address
  lists, decimal amounts — no longer outvotes the real separator.
- Header and trailer lines of a table no longer hide the real separator. A bank
  statement with more preamble and trailer lines than data rows used to be split
  on the decimal commas of its amounts instead of on its actual separator.
- Separators and newlines inside quoted values are ignored while detecting the
  format, so a quoted `\r\n` no longer switches a whole file to CRLF.
- A `sep=` header line no longer leaks a carriage return into the last field of
  every line, and a carriage return inside a quoted value survives parsing.
- A table whose rows all have a single column is kept instead of being emptied
  by `SetRowsWithNonUniformColumnsNil`.
- Files whose line endings are not recognised no longer take quadratic time to
  parse. Such a file is read as one very long row, and both the search for the
  end of a quoted value and the joining of one grew with the square of the
  number of fields. A 3 MB file of this shape went from 10.6 s to 42 ms when no
  value closes its quote, and from 3.5 s to 42 ms when they do.
- `WriteResultFile` of the text, HTML and Excel renderers returns the error from
  closing the file. A go-fs file system that uploads or flushes on close used to
  report success even when that final write failed. When writing fails, a file
  writer that can abort, like an upload, is aborted instead of closed, so the
  partial output no longer replaces the existing file.
- Writing the result of a text or HTML renderer no longer consumes it. A second
  `WriteResultTo` or `WriteResultFile` call, like a retry after a failed upload,
  used to write an empty file and report success, and `Result` returned nothing
  afterwards.
- `WriteResultTo` and `WriteResultFile` of the HTML renderer write the closing
  `</tbody></table>` tags. Files written this way used to end without them, and
  every `Result` call appended them once more.

### Changed

These change results for input that parsed before. Check them against your data
before upgrading:

- `ParseDetectFormat` can now return `Newline: "\n\r"`, which it never returned
  before. `Renderer` writes that value back out, so a detect-then-write round
  trip can produce line endings Excel does not read.
- `|` is now a candidate separator and can win detection. A two-column comma
  file whose second column always contains two pipes is now split into four.
- `Format.Validate` rejects a separator that is a quote or a control character
  other than tab. A stored configuration using an ASCII unit or record separator
  now fails instead of parsing.
- `SetRowsWithNonUniformColumnsNil` keeps single-column tables. Reader
  configurations that select this modifier by name change behaviour on upgrade.
- Detection on empty input returns `Separator: ","` instead of `""`, so an empty
  separator no longer signals "no table found" — check the number of rows.
- The text and HTML renderers keep their rendered data when it is written.
  Code that renders rows, calls `WriteResultTo`, renders more rows and calls
  `WriteResultTo` again now writes the first rows twice; render all rows before
  writing, or use a new renderer for each batch.
- The text and HTML renderers render their closing content, like the HTML
  `</tbody></table>`, on the first `Result`, `WriteResultTo` or
  `WriteResultFile` call only. A second table rendered on the same HTML renderer
  after that call is left without closing tags; use a new renderer per table.

### Breaking

- `github.com/ungerik/go-fs` goes from `v0.1.0` to `v1.0.0-rc.8`, a release
  candidate. v0 and v1 share one import path, so Go's minimal version selection
  raises every module that depends on `go-structtable` to `v1.0.0-rc.8` as well,
  and its own go-fs calls stop compiling where v1 changed them. The changes most
  code hits: `ReadAll`, `ReadAllString`, `ContentHash`, `WriteAll`,
  `WriteAllString` and the `ListDir*` methods take a `context.Context` as first
  parameter, and their `*Context` twins (`ReadAllContext`, `WriteAllContext`,
  `ListDirContext` …) are removed. `fs.ReadSeekCloser` becomes
  `fs.ReadAtSeekCloser` and `File.URL()` becomes `File.URI()`. The full list is
  in go-fs's
  [`docs/MIGRATION_v1.md`](https://github.com/ungerik/go-fs/blob/master/docs/MIGRATION_v1.md).
  This module will move to go-fs `v1.0.0` once it is tagged.
- `RenderFile`, `csv.NewReaderFromFile` and `excel.NewReader` take a
  `context.Context` as first parameter that cancels or times out the file read
  or write, like `csv.ParseFileDetectFormat` and `csv.ParseFileWithFormat`.
  `excel.NewReader` reads the whole file into memory before unzipping it.
- `github.com/domonda/go-errs` goes from `v1.0.4` to `v1.3.1`. Errors returned by
  this module cache their message on the first `Error()` call, so two otherwise
  identical errors are no longer `reflect.DeepEqual` (testify's `assert.Equal`)
  when only one of them has rendered its message; compare with `errors.Is` or
  by message instead. A panic recovered by `csv.ParseDetectFormat` or
  `csv.ParseWithFormat` comes back as an error of an unexported type with the
  same `Error()` text and `Unwrap()` result as before.

### Dependencies

- Indirect dependencies are updated to their latest versions.

### Documentation

- The README describes what format detection actually inspects: the encodings it
  tries, the separator candidates, the line endings it recognises, and how the
  separator is chosen when a table has single column header or trailer lines.

### Removed

- The `can't handle CSV field ... in line ...` error. Inputs that previously
  failed to parse now produce rows, so a pipeline that used this error to reject
  a malformed file no longer has that signal.
