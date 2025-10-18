package xml

import (
	"encoding/xml"
)

// Name is the name registered for the xml Codec.
const Name = "xml"

// Codec is a Codec implementation with xml.
type Codec struct{}

func (Codec) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

func (Codec) Unmarshal(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

func (Codec) Name() string {
	return Name
}
