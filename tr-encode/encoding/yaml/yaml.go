package yaml

import (
	"gopkg.in/yaml.v3"
)

// Name is the name registered for the yaml Codec.
const Name = "yaml"

// Codec is a Codec implementation with yaml.
type Codec struct{}

func (Codec) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (Codec) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func (Codec) Name() string {
	return Name
}
