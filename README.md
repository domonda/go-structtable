# go-structtable

⚠️ **DEPRECATED** ⚠️

This package is deprecated and no longer maintained. Please use [`github.com/domonda/go-retable`](https://github.com/domonda/go-retable) instead, which provides improved functionality, better performance, and active maintenance.

---

Read and write data-table formats as slices of Go structs

## Overview

The `go-structtable` package provides a comprehensive solution for converting between Go struct slices and various table formats including CSV, Excel, HTML, and custom text formats. It offers both reading (parsing) and writing (rendering) capabilities with flexible column mapping and formatting options.

## Key Features

- **Multiple Format Support**: CSV, Excel (.xlsx), HTML, and custom text formats
- **Flexible Column Mapping**: Map struct fields to table columns using tags or custom logic
- **Type-Safe Operations**: Full reflection-based type handling with custom formatters
- **Reader Interface**: Unified interface for reading tabular data into struct slices
- **Renderer Interface**: Unified interface for rendering struct slices to various formats
- **Format Detection**: Automatic detection of CSV format parameters
- **Data Modification**: Built-in modifiers for cleaning and transforming data

## Core Concepts

### Column Mapping

The package uses `ColumnMapper` implementations to define how struct fields map to table columns:

- **ReflectColumnTitles**: Uses struct tags and field names for column mapping
- **ColumnTitles**: Simple slice-based column title mapping
- **NoColumnTitles**: Renders data without column headers

### Readers

Readers implement the `Reader` interface to parse tabular data into struct instances:

- **CSV Reader**: Parses CSV files with format detection and data modification
- **Excel Reader**: Reads Excel (.xlsx) files from specific sheets
- **Text Reader**: Works with pre-parsed 2D string slices

### Renderers

Renderers implement the `Renderer` interface to convert struct slices into various formats:

- **HTML Renderer**: Generates HTML tables with custom styling
- **CSV Renderer**: Creates CSV files with configurable delimiters and quoting
- **Excel Renderer**: Produces Excel files with formatting and multiple sheets
- **Text Renderer**: Base renderer for custom text-based formats

## Quick Start

### Reading CSV Data

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/domonda/go-structtable"
    "github.com/domonda/go-structtable/csv"
)

type Person struct {
    Name string `col:"Full Name"`
    Age  int    `col:"Age"`
    City string `col:"City"`
}

func main() {
    // Open CSV file
    file, err := os.Open("people.csv")
    if err != nil {
        panic(err)
    }
    defer file.Close()
    
    // Create CSV reader
    reader, err := csv.NewReader(file, csv.NewFormat(","), "\n", nil, []csv.ColumnMapping{
        {Index: 0, StructField: "Name"},
        {Index: 1, StructField: "Age"},
        {Index: 2, StructField: "City"},
    })
    if err != nil {
        panic(err)
    }
    
    // Read data into struct slice
    var people []Person
    headers, err := structtable.Read(reader, &people, 1) // Skip 1 header row
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Headers: %v\n", headers)
    fmt.Printf("People: %+v\n", people)
}
```

### Rendering HTML Tables

```go
package main

import (
    "os"
    
    "github.com/domonda/go-structtable"
    "github.com/domonda/go-structtable/htmltable"
)

type Product struct {
    Name  string  `col:"Product Name"`
    Price float64 `col:"Price"`
    Stock int     `col:"In Stock"`
}

func main() {
    products := []Product{
        {Name: "Laptop", Price: 999.99, Stock: 5},
        {Name: "Mouse", Price: 29.99, Stock: 50},
        {Name: "Keyboard", Price: 79.99, Stock: 25},
    }
    
    // Render to HTML
    err := htmltable.Render(os.Stdout, products, "Product Catalog", true, structtable.DefaultReflectColumnTitles)
    if err != nil {
        panic(err)
    }
}
```

### Custom Column Mapping

```go
package main

import (
    "fmt"
    "reflect"
    
    "github.com/domonda/go-structtable"
)

type User struct {
    ID       int    `col:"-"`           // Ignored
    Username string `col:"User Name"`    // Custom title
    Email    string                     // Uses SpacePascalCase: "Email"
    Active   bool   `col:"Is Active"`   // Custom title
}

func main() {
    users := []User{
        {ID: 1, Username: "john", Email: "john@example.com", Active: true},
        {ID: 2, Username: "jane", Email: "jane@example.com", Active: false},
    }
    
    // Use default column mapper
    mapper := structtable.DefaultReflectColumnTitles
    
    // Get column titles and row reflector
    titles, reflector := mapper.ColumnTitlesAndRowReflector(reflect.TypeOf(User{}))
    fmt.Printf("Column titles: %v\n", titles)
    // Output: [User Name Email Is Active]
    
    // Reflect values from first user
    values := reflector.ReflectRow(reflect.ValueOf(users[0]))
    fmt.Printf("First user values: %v\n", values)
}
```

## Package Structure

### Main Package (`structtable`)

- **Interfaces**: `Reader`, `Renderer`, `ColumnMapper`, `RowReflector`
- **Core Functions**: `Read`, `Render`, `RenderTo`, `RenderBytes`, `RenderFile`
- **Column Mappers**: `ReflectColumnTitles`, `ColumnTitles`, `NoColumnTitles`
- **Utilities**: `StructFieldTypes`, `StructFieldValues`, `SpacePascalCase`

### CSV Package (`csv`)

- **Reader**: `csv.Reader` for parsing CSV files
- **Renderer**: `csv.Renderer` for generating CSV files
- **Format**: `csv.Format` and `csv.FormatDetectionConfig`
- **Modifiers**: Data cleaning and transformation functions
- **Types**: `csv.DataType` for type detection

### Excel Package (`excel`)

- **Reader**: `excel.Reader` for reading Excel files
- **Renderer**: `excel.Renderer` for generating Excel files
- **Configuration**: `excel.ExcelFormatConfig` for formatting options
- **Cell Writers**: Custom formatting for different data types

### HTML Package (`htmltable`)

- **Renderer**: `htmltable.Renderer` for generating HTML tables
- **Configuration**: `structtable.HTMLTableConfig` for styling options

### Text Table Package (`texttable`)

- **Interface**: `texttable.Table` for accessing tabular data
- **Implementation**: `texttable.StringsTable` for 2D string slices
- **Utilities**: `texttable.BoundingBox` for spatial information

## Advanced Usage

### Custom Format Renderer

```go
type CustomRenderer struct {
    *structtable.TextRenderer
}

func (r *CustomRenderer) RenderBeginTableText(w io.Writer) error {
    _, err := fmt.Fprintf(w, "=== TABLE START ===\n")
    return err
}

func (r *CustomRenderer) RenderHeaderRowText(w io.Writer, titles []string) error {
    _, err := fmt.Fprintf(w, "HEADER: %s\n", strings.Join(titles, " | "))
    return err
}

func (r *CustomRenderer) RenderRowText(w io.Writer, fields []string) error {
    _, err := fmt.Fprintf(w, "ROW: %s\n", strings.Join(fields, " | "))
    return err
}

func (r *CustomRenderer) RenderEndTableText(w io.Writer) error {
    _, err := fmt.Fprintf(w, "=== TABLE END ===\n")
    return err
}
```

### Data Modification

```go
// Apply modifiers to clean CSV data
modifiers := csv.ModifierList{
    csv.RemoveEmptyRowsModifier{},
    csv.CompactSpacedStringsModifier{},
    csv.SetRowsWithNonUniformColumnsNilModifier{},
}

reader, err := csv.NewReader(file, format, "\n", modifiers, columnMappings)
```

### Format Detection

```go
// Automatically detect CSV format
config := csv.NewFormatDetectionConfig()
rows, format, err := csv.ParseDetectFormat(data, config)
```

## Migration to go-retable

If you're currently using `go-structtable`, consider migrating to `go-retable` for:

- Better performance and memory efficiency
- More flexible column mapping options
- Enhanced error handling and validation
- Active maintenance and bug fixes
- Additional format support
- Improved API design

## License

This package is part of the Domonda project and follows the same licensing terms.
