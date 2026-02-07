package squad

import "encoding/json"

// PropertyType represents a JSON Schema type
type PropertyType string

const (
	TypeString  PropertyType = "string"
	TypeNumber  PropertyType = "number"
	TypeInteger PropertyType = "integer"
	TypeBoolean PropertyType = "boolean"
	TypeArray   PropertyType = "array"
	TypeObject  PropertyType = "object"
)

// Property defines a single property in a JSON Schema
type Property struct {
	Type        PropertyType `json:"type"`
	Description string       `json:"description,omitempty"`
	Items       *Property    `json:"items,omitempty"`      // For array types
	Properties  PropertyMap  `json:"properties,omitempty"` // For nested objects
	Required    []string     `json:"required,omitempty"`   // For nested objects
}

// PropertyMap is a map of property names to their definitions
type PropertyMap map[string]Property

// Schema represents a JSON Schema for tool parameters
type Schema struct {
	Type       PropertyType `json:"type"`
	Properties PropertyMap  `json:"properties"`
	Required   []string     `json:"required,omitempty"`
}

// String returns the JSON representation of the schema
func (s Schema) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}
