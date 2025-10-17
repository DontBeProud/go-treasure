package form

import (
	"net/url"
	"reflect"

	"github.com/go-playground/form/v4"
	"google.golang.org/protobuf/proto"
)

const (
	// Name is form Codec name
	Name = "x-www-form-urlencoded"
	// Null value string
	nullStr = "null"
)

var (
	DefaultEncoder = form.NewEncoder()
	DefaultDecoder = form.NewDecoder()
)

func init() {
	DefaultDecoder.SetTagName("json")
	DefaultEncoder.SetTagName("json")
}

type Codec struct {
	Encoder *form.Encoder
	Decoder *form.Decoder
}

func (c Codec) Marshal(v interface{}) ([]byte, error) {
	var vs url.Values
	var err error
	if m, ok := v.(proto.Message); ok {
		vs, err = EncodeValues(m)
		if err != nil {
			return nil, err
		}
	} else {
		vs, err = c.Encoder.Encode(v)
		if err != nil {
			return nil, err
		}
	}
	for k, v := range vs {
		if len(v) == 0 {
			delete(vs, k)
		}
	}
	return []byte(vs.Encode()), nil
}

func (c Codec) Unmarshal(data []byte, v interface{}) error {
	vs, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if m, ok := v.(proto.Message); ok {
		return DecodeValues(m, vs)
	}
	if m, ok := rv.Interface().(proto.Message); ok {
		return DecodeValues(m, vs)
	}

	return c.Decoder.Decode(v, vs)
}

func (Codec) Name() string {
	return Name
}
