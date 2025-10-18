package gtencode

import (
	"fmt"
	"github.com/DontBeProud/go-treasure/gt-encode/encoding"
	"github.com/DontBeProud/go-treasure/gt-encode/encoding/form"
	"github.com/DontBeProud/go-treasure/gt-encode/encoding/json"
	encodingProto "github.com/DontBeProud/go-treasure/gt-encode/encoding/proto"
	"github.com/DontBeProud/go-treasure/gt-encode/encoding/xml"
	"github.com/DontBeProud/go-treasure/gt-encode/encoding/yaml"
	"github.com/bytedance/sonic"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// 预注册默认编码器
// 可通过再次调用encoding.RegisterCodec，注册新的编码器，或重置默认编码器
func init() {
	encoding.RegisterCodec(form.Codec{Encoder: form.DefaultEncoder, Decoder: form.DefaultDecoder})
	encoding.RegisterCodec(json.Codec{})
	encoding.RegisterCodec(encodingProto.Codec{})
	encoding.RegisterCodec(xml.Codec{})
	encoding.RegisterCodec(yaml.Codec{})
}

// Unmarshal2Map 反序列化至map[string]interface{}
func Unmarshal2Map(contentType string, data []byte) (map[string]interface{}, error) {
	t := encoding.GetCodec(contentType)
	if t == nil {
		return nil, fmt.Errorf("unknown contentType: %s", contentType)
	}
	dataMap := make(map[string]interface{})
	err := t.Unmarshal(data, &dataMap)
	return dataMap, err
}

// Unmarshal2StructExt 反序列化操作额外配置
type Unmarshal2StructExt struct {
	Resolve      func(map[string]interface{}) // map[string]interface{}追加处理。例如需要对特定数据进行解密，即可在Resolve中实现
	MarshalOpt   *protojson.MarshalOptions    // 序列化时的json选项，默认为MarshalOptions{EmitUnpopulated: true}
	UnmarshalOpt *protojson.UnmarshalOptions  // 反序列化时的json选项，默认为UnmarshalOptions{DiscardUnknown: true}
}

// Unmarshal2Struct 反序列化至结构体，兼容PB结构体
// 1. data -> map[string]interface{}
// 2. map[string]interface{} -> json
// 3. json -> struct
func Unmarshal2Struct(contentType string, data []byte, v interface{}, ext *Unmarshal2StructExt) error {
	dataMap, err := Unmarshal2Map(contentType, data)
	if err != nil {
		return fmt.Errorf("unmarshal to map: %w", err)
	}

	var resolve func(map[string]interface{})
	var marshalOpt *protojson.MarshalOptions
	var unmarshalOpt *protojson.UnmarshalOptions
	if ext != nil {
		resolve = ext.Resolve
		if ext.MarshalOpt != nil {
			marshalOpt = ext.MarshalOpt
		}
		if ext.UnmarshalOpt != nil {
			unmarshalOpt = ext.UnmarshalOpt
		}
	}

	if resolve != nil {
		resolve(dataMap)
	}

	targetBytes, err := marshalJSON(convertMap(dataMap), marshalOpt)
	if err != nil {
		return fmt.Errorf("marshal json fail: %s", err.Error())
	}

	err = unmarshalJSON(targetBytes, v, unmarshalOpt)
	if err != nil {
		return fmt.Errorf("secondary unmarshal fail: %s", err.Error())
	}

	return nil
}

func marshalJSON(v interface{}, opt *protojson.MarshalOptions) ([]byte, error) {
	if m, ok := v.(proto.Message); ok {
		if opt == nil {
			return protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(m)
		} else {
			return opt.Marshal(m)
		}
	}
	return sonic.Marshal(v)
}

func unmarshalJSON(data []byte, v interface{}, opt *protojson.UnmarshalOptions) error {
	if m, ok := v.(proto.Message); ok {
		if opt == nil {
			return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, m)
		} else {
			return opt.Unmarshal(data, m)
		}
	}
	return sonic.Unmarshal(data, v)
}

// 递归转换
func convertMap(src interface{}) interface{} {
	switch m := src.(type) {
	case map[string]interface{}:
		dst := make(map[string]interface{}, len(m))
		for k, v := range m {
			dst[k] = convertMap(v)
		}
		return dst
	case map[interface{}]interface{}:
		dst := make(map[string]interface{}, len(m))
		for k, v := range m {
			dst[fmt.Sprint(k)] = convertMap(v)
		}
		return dst
	case []interface{}:
		dst := make([]interface{}, len(m))
		for k, v := range m {
			dst[k] = convertMap(v)
		}
		return dst
	case []byte:
		return string(m)
	default:
		return src
	}
}
