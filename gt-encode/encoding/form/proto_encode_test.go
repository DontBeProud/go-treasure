package form

import (
	"testing"
)

func TestJsonCamelCase(t *testing.T) {
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
		t.Run(test.snakeCase, func(t *testing.T) {
			camel := jsonCamelCase(test.snakeCase)
			if camel != test.camelCase {
				t.Errorf("want: %s, got: %s", test.camelCase, camel)
			}
		})
	}
}

func TestIsASCIILower(t *testing.T) {
	tests := []struct {
		b     byte
		lower bool
	}{
		{
			'A', false,
		},
		{
			'a', true,
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
			lower := isASCIILower(test.b)
			if test.lower != lower {
				t.Errorf("'%s' is not ascii lower", string(test.b))
			}
		})
	}
}
