package structtable

import (
	"fmt"
	"go/token"
	"reflect"
	"strings"
	"unicode"
)

// DefaultReflectColumnTitles provides the default ReflectColumnTitles
// using "col" as Tag and the SpacePascalCase function for UntaggedFieldTitle.
// Implements ColumnMapper.
//
// This is the recommended default configuration for mapping struct fields
// to column titles. It uses the "col" struct tag for explicit column names,
// "-" as the ignore marker, and SpacePascalCase for formatting untagged fields.
//
// Example usage:
//
//	type Person struct {
//	    Name string `col:"Full Name"`
//	    Age  int    // Will be formatted as "Age"
//	    ID   string `col:"-"` // Will be ignored
//	}
var DefaultReflectColumnTitles = &ReflectColumnTitles{
	Tag:                "col",
	IgnoreTitle:        "-",
	UntaggedFieldTitle: SpacePascalCase,
}

// RowReflector is used to reflect column values from the fields of a struct
// representing a table row.
//
// This interface defines how to extract values from a struct instance
// and convert them into a slice of reflect.Value objects that can be
// used for rendering table rows.
type RowReflector interface {
	// ReflectRow returns reflection values for struct fields
	// of structValue representing a table row.
	//
	// The returned slice should contain reflect.Value objects for each
	// column in the same order as the column titles returned by
	// ColumnMapper.ColumnTitlesAndRowReflector.
	ReflectRow(structValue reflect.Value) (columnValues []reflect.Value)
}

// RowReflectorFunc implements RowReflector with a function.
//
// This allows you to use a simple function as a RowReflector without
// creating a custom type.
type RowReflectorFunc func(structValue reflect.Value) (columnValues []reflect.Value)

// ReflectRow calls the underlying function to reflect row values.
func (f RowReflectorFunc) ReflectRow(structValue reflect.Value) (columnValues []reflect.Value) {
	return f(structValue)
}

// ColumnMapper is used to map struct type fields to column names.
//
// This interface defines how to extract column titles and create a RowReflector
// from a struct type. It's the core abstraction for converting struct types
// into table column definitions.
type ColumnMapper interface {
	// ColumnTitlesAndRowReflector returns the column titles and indices for structFields.
	// The length of the titles and indices slices must be identical to the length of structFields.
	// The indices start at zero, the special index -1 filters removes the column
	// for the corresponding struct field.
	//
	// Parameters:
	//   - structType: The reflect.Type of the struct to map
	//
	// Returns:
	//   - titles: Slice of column titles for the table header
	//   - rowReflector: RowReflector that can extract values from struct instances
	ColumnTitlesAndRowReflector(structType reflect.Type) (titles []string, rowReflector RowReflector)
}

// ColumnMapperFunc implements the ColumnMapper interface with a function.
//
// This allows you to use a simple function as a ColumnMapper without
// creating a custom type.
type ColumnMapperFunc func(structType reflect.Type) (titles []string, rowReflector RowReflector)

// ColumnTitlesAndRowReflector calls the underlying function to map struct fields.
func (f ColumnMapperFunc) ColumnTitlesAndRowReflector(structType reflect.Type) (titles []string, rowReflector RowReflector) {
	return f(structType)
}

// ColumnTitles implements ColumnMapper by returning the underlying string slice as column titles
// and the StructFieldValues function of this package as RowReflector.
//
// This is a simple implementation that uses the provided string slice as column titles
// and maps them directly to struct fields in order. It does not check if the number
// of column titles and the reflected row values are identical, and re-mapping or
// ignoring of columns is not possible.
//
// Use this when you have a fixed set of column titles that correspond directly
// to struct fields in order.
type ColumnTitles []string

// ColumnTitlesAndRowReflector returns the column titles and a RowReflector that
// uses StructFieldValues to extract values from struct instances.
func (t ColumnTitles) ColumnTitlesAndRowReflector(structType reflect.Type) (titles []string, rowReflector RowReflector) {
	return t, RowReflectorFunc(StructFieldValues)
}

// NoColumnTitles returns a ColumnMapper that returns nil as column titles
// and the StructFieldValues function of this package as RowReflector.
//
// This is useful when you want to render table data without column headers,
// such as when generating data-only exports or when headers are handled
// separately.
func NoColumnTitles() ColumnMapper {
	return noColumnTitles{}
}

// noColumnTitles implements ColumnMapper by returning nil as column titles
// and the StructFieldValues function of this package as RowReflector.
//
// This is the internal implementation used by NoColumnTitles().
type noColumnTitles struct{}

// ColumnTitlesAndRowReflector returns nil titles and a RowReflector that
// uses StructFieldValues to extract values from struct instances.
func (noColumnTitles) ColumnTitlesAndRowReflector(structType reflect.Type) (titles []string, rowReflector RowReflector) {
	return nil, RowReflectorFunc(StructFieldValues)
}

// ReflectColumnTitles implements ColumnMapper with a struct field Tag
// to be used for naming and a UntaggedFieldTitle in case the Tag is not set.
//
// This is the most flexible and commonly used ColumnMapper implementation.
// It uses struct tags to determine column names and provides fallback
// formatting for untagged fields. It also supports field mapping and
// filtering through the MapIndices field.
//
// Example usage:
//
//	type Person struct {
//	    Name string `col:"Full Name"`
//	    Age  int    // Will use UntaggedFieldTitle function
//	    ID   string `col:"-"` // Will be ignored
//	}
//
//	mapper := &ReflectColumnTitles{
//	    Tag: "col",
//	    IgnoreTitle: "-",
//	    UntaggedFieldTitle: SpacePascalCase,
//	}
type ReflectColumnTitles struct {
	// Tag is the struct field tag to be used as column name.
	// If a field has this tag, its value will be used as the column title.
	Tag string
	// IgnoreTitle will result in a column index of -1.
	// Fields with this tag value will be excluded from the table.
	IgnoreTitle string
	// UntaggedFieldTitle will be called with the struct field name to
	// return a column name in case the struct field has no tag named Tag.
	// If UntaggedFieldTitle is nil, then the struct field name will be used unchanged.
	UntaggedFieldTitle func(fieldName string) (columnTitle string)
	// MapIndices is a map from the index of a field in struct
	// to the column index returned by ColumnTitlesAndRowReflector.
	// If MapIndices is nil, then no mapping will be performed.
	// Map to the index -1 to not create a column for a struct field.
	MapIndices map[int]int
}

// WithTag returns a copy of ReflectColumnTitles with the specified tag.
//
// This method creates a new instance with the modified tag, allowing
// for fluent configuration of the ColumnMapper.
func (n *ReflectColumnTitles) WithTag(tag string) *ReflectColumnTitles {
	mod := *n
	mod.Tag = tag
	return &mod
}

// WithIgnoreTitle returns a copy of ReflectColumnTitles with the specified ignore title.
//
// This method creates a new instance with the modified ignore title, allowing
// for fluent configuration of the ColumnMapper.
func (n *ReflectColumnTitles) WithIgnoreTitle(ignoreTitle string) *ReflectColumnTitles {
	mod := *n
	mod.IgnoreTitle = ignoreTitle
	return &mod
}

// WithMapIndex returns a copy of ReflectColumnTitles with a field-to-column mapping.
//
// This method creates a new instance with an additional mapping from fieldIndex
// to columnIndex, allowing for reordering or filtering of columns.
//
// Parameters:
//   - fieldIndex: The index of the struct field (0-based)
//   - columnIndex: The index of the column in the output table (0-based)
func (n *ReflectColumnTitles) WithMapIndex(fieldIndex, columnIndex int) *ReflectColumnTitles {
	mod := *n
	if mod.MapIndices == nil {
		mod.MapIndices = make(map[int]int)
	}
	mod.MapIndices[fieldIndex] = columnIndex
	return &mod
}

// WithIgnoreIndex returns a copy of ReflectColumnTitles that ignores the specified field.
//
// This method creates a new instance that will exclude the field at fieldIndex
// from the output table by mapping it to column index -1.
//
// Parameters:
//   - fieldIndex: The index of the struct field to ignore (0-based)
func (n *ReflectColumnTitles) WithIgnoreIndex(fieldIndex int) *ReflectColumnTitles {
	mod := *n
	if mod.MapIndices == nil {
		mod.MapIndices = make(map[int]int)
	}
	mod.MapIndices[fieldIndex] = -1
	return &mod
}

// WithMapIndices returns a copy of ReflectColumnTitles with the specified field mappings.
//
// This method creates a new instance with the complete map of field-to-column
// mappings, replacing any existing mappings.
//
// Parameters:
//   - mapIndices: Map from struct field indices to column indices
func (n *ReflectColumnTitles) WithMapIndices(mapIndices map[int]int) *ReflectColumnTitles {
	mod := *n
	mod.MapIndices = mapIndices
	return &mod
}

// ColumnTitlesAndRowReflector implements the ColumnMapper interface.
//
// This method analyzes the struct type and returns column titles and a RowReflector
// based on the configuration of this ReflectColumnTitles instance. It handles
// struct tags, field mapping, and filtering according to the configured rules.
func (n *ReflectColumnTitles) ColumnTitlesAndRowReflector(structType reflect.Type) (titles []string, rowReflector RowReflector) {
	structFields := StructFieldTypes(structType)
	indices := make([]int, len(structFields))

	columnIndexUsed := make(map[int]bool)
	getNextFreeColumnIndex := func() int {
		for i := range structFields {
			if !columnIndexUsed[i] {
				return i
			}
		}
		panic("getNextFreeColumnIndex should always find a free column index")
	}

	for i, structField := range structFields {
		title := n.titleFromStructField(structField)
		if title == n.IgnoreTitle {
			indices[i] = -1
			continue
		}

		index := getNextFreeColumnIndex()
		if n.MapIndices != nil {
			mappedIndex, ok := n.MapIndices[i]
			if ok && !columnIndexUsed[mappedIndex] {
				index = mappedIndex
			}
		}
		if index < 0 || index >= len(structFields) {
			indices[i] = -1
			continue
		}

		indices[i] = index
		columnIndexUsed[index] = true

		titles = append(titles, title)
	}

	rowReflector = RowReflectorFunc(func(structValue reflect.Value) []reflect.Value {
		columnValues := make([]reflect.Value, len(titles))
		structFields := StructFieldValues(structValue)
		for i, index := range indices {
			if index >= 0 && index < len(titles) {
				columnValues[index] = structFields[i]
			}
		}
		return columnValues
	})

	return titles, rowReflector
}

func (n *ReflectColumnTitles) titleFromStructField(structField reflect.StructField) string {
	if tag, ok := structField.Tag.Lookup(n.Tag); ok {
		if i := strings.IndexByte(tag, ','); i != -1 {
			tag = tag[:i]
		}
		if tag != "" {
			return tag
		}
	}
	if n.UntaggedFieldTitle == nil {
		return structField.Name
	}
	return n.UntaggedFieldTitle(structField.Name)
}

// String returns a string representation of the ReflectColumnTitles configuration.
//
// This is useful for debugging and logging purposes to see the current
// tag and ignore title configuration.
func (n *ReflectColumnTitles) String() string {
	return fmt.Sprintf("Tag: %q, Ignore: %q", n.Tag, n.IgnoreTitle)
}

// StructFieldTypes returns the exported fields of a struct type
// including the inlined fields of any anonymously embedded structs.
//
// This function recursively traverses a struct type and returns all
// exported fields, including those from embedded structs. It handles
// pointer types by dereferencing them to get the underlying struct type.
//
// Parameters:
//   - structType: The reflect.Type of the struct to analyze
//
// Returns:
//   - fields: Slice of reflect.StructField representing all exported fields
func StructFieldTypes(structType reflect.Type) (fields []reflect.StructField) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		switch {
		case field.Anonymous:
			fields = append(fields, StructFieldTypes(field.Type)...)
		case token.IsExported(field.Name):
			fields = append(fields, field)
		}
	}
	return fields
}

// StructFieldValues returns the reflect.Value of exported struct fields
// including the inlined fields of any anonymously embedded structs.
//
// This function recursively traverses a struct value and returns all
// exported field values, including those from embedded structs. It handles
// pointer types by dereferencing them to get the underlying struct value.
//
// Parameters:
//   - structValue: The reflect.Value of the struct instance to analyze
//
// Returns:
//   - values: Slice of reflect.Value representing all exported field values
func StructFieldValues(structValue reflect.Value) (values []reflect.Value) {
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		switch {
		case field.Anonymous:
			values = append(values, StructFieldValues(structValue.Field(i))...)
		case token.IsExported(field.Name):
			values = append(values, structValue.Field(i))
		}
	}
	return values
}

// SpacePascalCase inserts spaces before upper case
// characters within PascalCase like names.
// It also replaces underscore '_' characters with spaces.
//
// This function is commonly used as the UntaggedFieldTitle function
// for ReflectColumnTitles to create human-readable column names from
// struct field names.
//
// Examples:
//   - "FirstName" -> "First Name"
//   - "UserID" -> "User ID"
//   - "CreatedAt" -> "Created At"
//   - "user_name" -> "user name"
//
// Parameters:
//   - name: The field name to format
//
// Returns:
//   - The formatted name with spaces inserted appropriately
func SpacePascalCase(name string) string {
	var b strings.Builder
	b.Grow(len(name) + 4)
	lastWasUpper := true
	lastWasSpace := true
	for _, r := range name {
		if r == '_' {
			if !lastWasSpace {
				b.WriteByte(' ')
			}
			lastWasUpper = false
			lastWasSpace = true
			continue
		}
		isUpper := unicode.IsUpper(r)
		if isUpper && !lastWasUpper && !lastWasSpace {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
		lastWasUpper = isUpper
		lastWasSpace = unicode.IsSpace(r)
	}
	return strings.TrimSpace(b.String())
}
