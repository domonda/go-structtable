package structtable

import (
	"reflect"
	"testing"
)

func TestReflectColumnTitles_ColumnTitlesAndRowReflector(t *testing.T) {
	type fields struct {
		Tag                string
		IgnoreTitle        string
		UntaggedFieldTitle func(fieldName string) (columnTitle string)
		MapIndices         map[int]int
	}
	type args struct {
		structType reflect.Type
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantTitles       []string
		wantRowReflector RowReflector
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &ReflectColumnTitles{
				Tag:                tt.fields.Tag,
				IgnoreTitle:        tt.fields.IgnoreTitle,
				UntaggedFieldTitle: tt.fields.UntaggedFieldTitle,
				MapIndices:         tt.fields.MapIndices,
			}
			gotTitles, gotRowReflector := n.ColumnTitlesAndRowReflector(tt.args.structType)
			if !reflect.DeepEqual(gotTitles, tt.wantTitles) {
				t.Errorf("ReflectColumnTitles.ColumnTitlesAndRowReflector() gotTitles = %v, want %v", gotTitles, tt.wantTitles)
			}
			if !reflect.DeepEqual(gotRowReflector, tt.wantRowReflector) {
				t.Errorf("ReflectColumnTitles.ColumnTitlesAndRowReflector() gotRowReflector = %v, want %v", gotRowReflector, tt.wantRowReflector)
			}
		})
	}
}
