package form

import (
	"testing"
)

func TestJsonSnakeCase(t *testing.T) {
	tests := []struct {
		camelCase string
		snakeCase string
	}{
		{
			"userId", "user_id",
		},
		{
			"user", "user",
		},
		{
			"userIdAndUsername", "user_id_and_username",
		},
		{
			"", "",
		},
	}
	for _, test := range tests {
		t.Run(test.camelCase, func(t *testing.T) {
			snake := jsonSnakeCase(test.camelCase)
			if snake != test.snakeCase {
				t.Errorf("want: %s, got: %s", test.snakeCase, snake)
			}
		})
	}
}

func TestIsASCIIUpper(t *testing.T) {
	tests := []struct {
		b     byte
		upper bool
	}{
		{
			'A', true,
		},
		{
			'a', false,
		},
		{
			',', false,
		},
		{
			'1', false,
		},
		{
			' ', false,
		},
	}
	for _, test := range tests {
		t.Run(string(test.b), func(t *testing.T) {
			upper := isASCIIUpper(test.b)
			if test.upper != upper {
				t.Errorf("'%s' is not ascii upper", string(test.b))
			}
		})
	}
}

func TestParseURLQueryMapKey(t *testing.T) {
	tests := []struct {
		fieldName string
		field     string
		fieldKey  string
		err       error
	}{
		{
			fieldName: "map[kratos]", field: "map", fieldKey: "kratos", err: nil,
		},
		{
			fieldName: "map[]", field: "map", fieldKey: "", err: nil,
		},
		{
			fieldName: "", field: "", fieldKey: "", err: errInvalidFormatMapKey,
		},
		{
			fieldName: "[[]", field: "", fieldKey: "", err: errInvalidFormatMapKey,
		},
		{
			fieldName: "map[kratos]=", field: "", fieldKey: "", err: errInvalidFormatMapKey,
		},
		{
			fieldName: "[kratos]", field: "", fieldKey: "", err: errInvalidFormatMapKey,
		},
		{
			fieldName: "map", field: "", fieldKey: "", err: errInvalidFormatMapKey,
		},
		{
			fieldName: "map[", field: "", fieldKey: "", err: errInvalidFormatMapKey,
		},
		{
			fieldName: "]kratos[", field: "", fieldKey: "", err: errInvalidFormatMapKey,
		},
		{
			fieldName: "[kratos", field: "", fieldKey: "", err: errInvalidFormatMapKey,
		},
		{
			fieldName: "kratos]", field: "", fieldKey: "", err: errInvalidFormatMapKey,
		},
	}
	for _, test := range tests {
		t.Run(test.fieldName, func(t *testing.T) {
			fieldName, fieldKey, err := parseURLQueryMapKey(test.fieldName)
			if test.err != err {
				t.Fatalf("want: %s, got: %s", test.err, err)
			}
			if test.field != fieldName {
				t.Errorf("want: %s, got: %s", test.field, fieldName)
			}
			if test.fieldKey != fieldKey {
				t.Errorf("want: %s, got: %s", test.fieldKey, fieldKey)
			}
		})
	}
}
