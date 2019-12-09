package csv

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EmptyRowsWithNonUniformColumns(t *testing.T) {
	testCases := []struct {
		source   [][]string
		expected [][]string
	}{
		{
			source:   nil,
			expected: nil,
		},
		{
			source:   [][]string{nil, []string{"", "", ""}, nil},
			expected: [][]string{nil, []string{"", "", ""}, nil}, // nil rows can't dominate
		},
		{
			source:   [][]string{[]string{"1", "2", "3"}, []string{"0"}, []string{"4", "5", "6"}},
			expected: [][]string{[]string{"1", "2", "3"}, nil, []string{"4", "5", "6"}},
		},
		{
			source:   [][]string{[]string{"0"}, []string{"1", "2", "3"}, []string{"4", "5", "6"}},
			expected: [][]string{nil, []string{"1", "2", "3"}, []string{"4", "5", "6"}},
		},
		{
			source:   [][]string{[]string{"1", "2", "3"}, []string{"0"}, []string{"0", "0"}, []string{"4", "5", "6"}},
			expected: [][]string{[]string{"1", "2", "3"}, nil, nil, []string{"4", "5", "6"}}, // take longer row if count of columns is identical
		},
		{
			source:   [][]string{[]string{"0", "0"}, []string{"1", "2", "3"}},
			expected: [][]string{nil, []string{"1", "2", "3"}}, // take longer row if count of columns is identical
		},
		{
			source:   [][]string{[]string{"1"}, []string{"2", "2"}, []string{"3", "3", "3"}},
			expected: [][]string{nil, nil, []string{"3", "3", "3"}}, // take longer row if count of columns is identical
		},
	}

	for i, test := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			result := SetRowsWithNonUniformColumnsNil(test.source)
			assert.Equal(t, test.expected, result, "EmptyRowsWithNonUniformColumns")
		})
	}
}

func Test_RemoveEmptyRows(t *testing.T) {
	testCases := []struct {
		source   [][]string
		expected [][]string
	}{
		{
			source:   nil,
			expected: nil,
		},
		{
			source:   [][]string{},
			expected: nil,
		},
		{
			source:   [][]string{nil, []string{}, nil},
			expected: nil,
		},
		{
			source:   [][]string{nil, []string{"", "", ""}, nil},
			expected: nil,
		},
		{
			source:   [][]string{nil, []string{"1", "2", "3"}, nil},
			expected: [][]string{[]string{"1", "2", "3"}},
		},
		{
			source:   [][]string{[]string{"1", "2", "3"}, nil, nil},
			expected: [][]string{[]string{"1", "2", "3"}},
		},
		{
			source:   [][]string{nil, nil, []string{"1", "2", "3"}},
			expected: [][]string{[]string{"1", "2", "3"}},
		},
		{
			source:   [][]string{[]string{"1", "2", "3"}, nil, []string{"4", "5", "6"}},
			expected: [][]string{[]string{"1", "2", "3"}, []string{"4", "5", "6"}},
		},
	}

	for i, test := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			result := RemoveEmptyRows(test.source)
			assert.Equal(t, test.expected, result, "RemoveEmptyRows")
		})
	}
}

func Test_CleanSpacedString(t *testing.T) {
	// Also see http://localhost:5006/payment-import/20e66223-f7ab-4e1b-a59a-d15c104c9562-doc.csv.html
	testCases := map[string]string{
		"":                                      "",
		" ":                                     " ",
		"  ":                                    "  ",
		"1 2":                                   "12",
		"1 2 3":                                 "123",
		"1 2 3 ":                                "123", // do we want this?
		"Hello World!":                          "Hello World!",
		"S h i n e r g y   S c h ö n b r u n n": "Shinergy Schönbrunn",
		"S a l z b u r g e r   T e n n i s c o u r t s   S ü d": "Salzburger Tenniscourts Süd",
	}
	for source, expected := range testCases {
		t.Run(source, func(t *testing.T) {
			cleaned, modified := compactSpacedString(source)
			assert.Equal(t, expected, cleaned, "cleaned string")
			assert.True(t, modified == (cleaned != source), "modified")
		})
	}
}
