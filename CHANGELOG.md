# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

This module has no `VERSION` file and is consumed through Go module versions,
so changes are listed under `[Unreleased]` until they are tagged.

> ⚠️ This package is deprecated in favour of
> [`github.com/domonda/go-retable`](https://github.com/domonda/go-retable).
> Entries here are fixes for consumers still on `go-structtable`.

## [Unreleased]

CSV parsing and format detection, ported from `go-retable`'s `csvtable` package.

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

### Documentation

- The README describes what format detection actually inspects: the encodings it
  tries, the separator candidates, the line endings it recognises, and how the
  separator is chosen when a table has single column header or trailer lines.

### Removed

- The `can't handle CSV field ... in line ...` error. Inputs that previously
  failed to parse now produce rows, so a pipeline that used this error to reject
  a malformed file no longer has that signal.
