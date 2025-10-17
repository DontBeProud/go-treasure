package json

//type testEmbed struct {
//	Level1a int `json:"a"`
//	Level1b int `json:"b"`
//	Level1c int `json:"c"`
//}
//
//type testMessage struct {
//	Field1 string     `json:"a"`
//	Field2 string     `json:"b"`
//	Field3 string     `json:"c"`
//	Embed  *testEmbed `json:"embed,omitempty"`
//}
//
//type mock struct {
//	value int
//}
//
//const (
//	Unknown = iota
//	Gopher
//	Zebra
//)
//
//func (a *mock) UnmarshalJSON(b []byte) error {
//	var s string
//	if err := json.Unmarshal(b, &s); err != nil {
//		return err
//	}
//	switch strings.ToLower(s) {
//	default:
//		a.value = Unknown
//	case "gopher":
//		a.value = Gopher
//	case "zebra":
//		a.value = Zebra
//	}
//
//	return nil
//}
//
//func (a *mock) MarshalJSON() ([]byte, error) {
//	var s string
//	switch a.value {
//	default:
//		s = "unknown"
//	case Gopher:
//		s = "gopher"
//	case Zebra:
//		s = "zebra"
//	}
//
//	return json.Marshal(s)
//}
