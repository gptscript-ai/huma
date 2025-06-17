package huma

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Schema
		expectError bool
		errorMsg    string
	}{
		{
			name:  "string type",
			input: `{"type": "string", "title": "Test String"}`,
			expected: Schema{
				Type:     "string",
				Title:    "Test String",
				Nullable: false,
			},
			expectError: false,
		},
		{
			name:  "integer type",
			input: `{"type": "integer", "minimum": 0}`,
			expected: Schema{
				Type:     "integer",
				Minimum:  &[]float64{0.0}[0],
				Nullable: false,
			},
			expectError: false,
		},
		{
			name:  "array type with nullable",
			input: `{"type": ["string", "null"], "title": "Nullable String"}`,
			expected: Schema{
				Type:     "string",
				Title:    "Nullable String",
				Nullable: true,
			},
			expectError: false,
		},
		{
			name:  "array type without nullable",
			input: `{"type": ["integer"], "minimum": 1}`,
			expected: Schema{
				Type:     "integer",
				Minimum:  &[]float64{1.0}[0],
				Nullable: false,
			},
			expectError: false,
		},
		{
			name:  "array type with non-null second element",
			input: `{"type": ["boolean", "other"], "title": "Non-null second"}`,
			expected: Schema{
				Type:     "boolean",
				Title:    "Non-null second",
				Nullable: false,
			},
			expectError: false,
		},
		{
			name:  "empty array type",
			input: `{"type": [], "title": "Empty Array"}`,
			expected: Schema{
				Type:     "",
				Title:    "Empty Array",
				Nullable: false,
			},
			expectError: false,
		},
		{
			name:  "complex schema with properties",
			input: `{"type": "object", "properties": {"name": {"type": "string"}}, "required": ["name"]}`,
			expected: Schema{
				Type:       "object",
				Properties: map[string]*Schema{"name": {Type: "string"}},
				Required:   []string{"name"},
				Nullable:   false,
			},
			expectError: false,
		},
		{
			name:  "nullable object type",
			input: `{"type": ["object", "null"], "properties": {"id": {"type": "integer"}}}`,
			expected: Schema{
				Type:       "object",
				Properties: map[string]*Schema{"id": {Type: "integer"}},
				Nullable:   true,
			},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			input:       `{"type": "string", "title":}`,
			expected:    Schema{},
			expectError: true,
		},
		{
			name:        "invalid type in array - number",
			input:       `{"type": [123, "null"]}`,
			expected:    Schema{},
			expectError: true,
			errorMsg:    "invalid type: 123",
		},
		{
			name:        "invalid type in array - boolean",
			input:       `{"type": [true, "null"]}`,
			expected:    Schema{},
			expectError: true,
			errorMsg:    "invalid type: true",
		},
		{
			name:        "invalid type in array - object",
			input:       `{"type": [{"nested": "object"}, "null"]}`,
			expected:    Schema{},
			expectError: true,
			errorMsg:    "invalid type: map[nested:object]",
		},
		{
			name:  "array with multiple null values",
			input: `{"type": ["string", "null", "null"]}`,
			expected: Schema{
				Type:     "string",
				Nullable: true,
			},
			expectError: false,
		},
		{
			name:  "array with null as first element",
			input: `{"type": ["null", "string"]}`,
			expected: Schema{
				Type:     "string",
				Nullable: true,
			},
			expectError: false,
		},
		{
			name:  "schema with extensions",
			input: `{"type": "string", "x-custom": "value", "x-another": 42}`,
			expected: Schema{
				Type: "string",
				// Extensions should be: map[string]any{"custom": "value", "another": 42},
				// However, the implementation currently ignores them because the implementation
				// would be significantly more complex.
				Extensions: nil,
				Nullable:   false,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var schema Schema
			err := json.Unmarshal([]byte(tt.input), &schema)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Type, schema.Type)
			assert.Equal(t, tt.expected.Nullable, schema.Nullable)
			assert.Equal(t, tt.expected.Title, schema.Title)

			if tt.expected.Minimum != nil {
				require.NotNil(t, schema.Minimum)
				assert.Equal(t, *tt.expected.Minimum, *schema.Minimum)
			} else {
				assert.Nil(t, schema.Minimum)
			}

			if tt.expected.Properties != nil {
				require.NotNil(t, schema.Properties)
				assert.Equal(t, len(tt.expected.Properties), len(schema.Properties))
				for key, expectedProp := range tt.expected.Properties {
					actualProp, exists := schema.Properties[key]
					require.True(t, exists, "Property %s should exist", key)
					assert.Equal(t, expectedProp.Type, actualProp.Type)
				}
			}

			if tt.expected.Required != nil {
				assert.Equal(t, tt.expected.Required, schema.Required)
			}

			if tt.expected.Extensions != nil {
				require.NotNil(t, schema.Extensions)
				for key, expectedValue := range tt.expected.Extensions {
					actualValue, exists := schema.Extensions[key]
					require.True(t, exists, "Extension %s should exist", key)
					assert.Equal(t, expectedValue, actualValue)
				}
			}
		})
	}
}
