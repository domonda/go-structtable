package csv

import (
	"encoding/json"

	"github.com/domonda/go-errs"
)

type ModifierList []Modifier

func (l ModifierList) Modify(rows [][]string) [][]string {
	for _, m := range l {
		rows = m.Modify(rows)
	}
	return rows
}

// MarshalJSON implements encoding/json.Marshaler
func (l ModifierList) MarshalJSON() ([]byte, error) {
	names := make([]string, len(l))
	for i, modifier := range l {
		names[i] = modifier.Name()
	}
	return json.Marshal(names)
}

// UnmarshalJSON implements encoding/json.Unmarshaler
func (l *ModifierList) UnmarshalJSON(data []byte) error {
	var names []string
	err := json.Unmarshal(data, &names)
	if err != nil {
		return err
	}
	list := make(ModifierList, len(names))
	for i, name := range names {
		modifier, ok := ModifiersByName[name]
		if !ok {
			return errs.Errorf("can't find csv.Modifier with name %q", name)
		}
		list[i] = modifier
	}
	*l = list
	return nil
}
