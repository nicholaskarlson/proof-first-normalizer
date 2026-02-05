package normalizer

import (
	"encoding/json"
	"fmt"
	"os"
)

type Schema struct {
	Columns []Column `json:"columns"`
}

type Column struct {
	Name     string `json:"name"`
	Type     string `json:"type"`     // "string" | "date" | "decimal"
	Required bool   `json:"required"` // v0.1.0: required fields only
}

func LoadSchema(path string) (*Schema, []byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var s Schema
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, nil, fmt.Errorf("schema parse: %w", err)
	}
	if len(s.Columns) == 0 {
		return nil, nil, fmt.Errorf("schema: columns must be non-empty")
	}
	seen := make(map[string]bool, len(s.Columns))
	for i := range s.Columns {
		if s.Columns[i].Name == "" {
			return nil, nil, fmt.Errorf("schema: column[%d] name is empty", i)
		}
		if seen[s.Columns[i].Name] {
			return nil, nil, fmt.Errorf("schema: duplicate column name %q", s.Columns[i].Name)
		}
		seen[s.Columns[i].Name] = true
		switch s.Columns[i].Type {
		case "string", "date", "decimal":
		default:
			return nil, nil, fmt.Errorf("schema: column[%s] has invalid type %q", s.Columns[i].Name, s.Columns[i].Type)
		}
	}
	return &s, b, nil
}
