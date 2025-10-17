package form

import (
	"github.com/DontBeProud/go-treasure/tr-encode/encoding"
	"reflect"
	"testing"
)

type LoginRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func TestFormCodecMarshal(t *testing.T) {
	req := &LoginRequest{
		Username: "kratos",
		Password: "kratos_pwd",
	}
	encoding.RegisterCodec(Codec{Encoder: DefaultEncoder, Decoder: DefaultDecoder})
	content, err := encoding.GetCodec(Name).Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]byte("password=kratos_pwd&username=kratos"), content) {
		t.Errorf("expect %s, got %s", "password=kratos_pwd&username=kratos", content)
	}

	req = &LoginRequest{
		Username: "kratos",
		Password: "",
	}
	content, err = encoding.GetCodec(Name).Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]byte("username=kratos"), content) {
		t.Errorf("expect %s, got %s", "username=kratos", content)
	}

	m := struct {
		ID   int32  `json:"id"`
		Name string `json:"name"`
	}{
		ID:   1,
		Name: "kratos",
	}
	content, err = encoding.GetCodec(Name).Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]byte("id=1&name=kratos"), content) {
		t.Errorf("expect %s, got %s", "id=1&name=kratos", content)
	}
}

func TestFormCodecUnmarshal(t *testing.T) {
	req := &LoginRequest{
		Username: "kratos",
		Password: "kratos_pwd",
	}
	encoding.RegisterCodec(Codec{Encoder: DefaultEncoder, Decoder: DefaultDecoder})
	content, err := encoding.GetCodec(Name).Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	bindReq := new(LoginRequest)
	err = encoding.GetCodec(Name).Unmarshal(content, bindReq)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual("kratos", bindReq.Username) {
		t.Errorf("expect %v, got %v", "kratos", bindReq.Username)
	}
	if !reflect.DeepEqual("kratos_pwd", bindReq.Password) {
		t.Errorf("expect %v, got %v", "kratos_pwd", bindReq.Password)
	}
}
