package ginx

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDocHeader(t *testing.T) {
	testStruct := struct {
		TestString  string                 `json:"str" doc:"required desc(test()) pattern(^[A-Z](.*))"`
		TestString2 string                 `json:"str2" doc:"required desc(test) pattern(^[A-Z].+)"`
		TestMap     map[string]interface{} `json:"map" doc:"required desc(service limit)"`
	}{
		TestString:  "test",
		TestString2: "t2",
		TestMap: map[string]interface{}{
			"number": 10,
			"string": "20",
			"bool":   true,
		},
	}
	sch := extractSchema(testStruct)
	b, err := sch.MarshalJSON()
	assertNil(t, err)
	assertNotNil(t, b)
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	assertNil(t, err)
	mapVal := func(mapDat map[string]interface{}, key string) interface{} {
		m := mapDat
		depths := strings.Split(key, ".")
		var k string
		for i, dep := range depths {
			if i == len(depths)-1 {
				k = dep
				break
			} else {
				m = m[dep].(map[string]interface{})
			}
		}
		return m[k]
	}
	assertEqual(t, "test", mapVal(m, "properties.str.example"))
	assertEqual(t, "string", mapVal(m, "properties.str.type"))
	assertEqual(t, "test()", mapVal(m, "properties.str.description"))
	assertEqual(t, "^[A-Z](.*)", mapVal(m, "properties.str.pattern"))
	assertEqual(t, "t2", mapVal(m, "properties.str2.example"))
	assertEqual(t, "string", mapVal(m, "properties.str2.type"))
	assertEqual(t, "test", mapVal(m, "properties.str2.description"))
	assertEqual(t, "^[A-Z].+", mapVal(m, "properties.str2.pattern"))
	required := mapVal(m, "required").([]interface{})
	assertEqual(t, "str", required[0])
}

type TypeString string

func TestExtractString(t *testing.T) {
	t1 := "test"
	out := extractString(t1)
	assertNotNil(t, out)
	assertEqual(t, out.Value.Interface(), "test")

	t2 := &t1
	out2 := extractString(t2)
	assertNotNil(t, out2)
	assertEqual(t, out2.Value.Interface(), "test")

	var t3 TypeString = "test"
	out3 := extractString(t3)
	assertNotNil(t, out3)
	assertEqual(t, out3.Value.Interface(), TypeString("test"))
}

type TypeInt int

func TestExtractInteger(t *testing.T) {
	var t1 int64 = 1
	out := extractInteger(t1)
	assertNotNil(t, out)
	assertEqual(t, out.Value.Interface(), int64(1))

	t2 := &t1
	out2 := extractInteger(t2)
	assertNotNil(t, out2)
	assertEqual(t, out2.Value.Interface(), int64(1))

	var t3 TypeInt = 1
	out3 := extractInteger(t3)
	assertNotNil(t, out3)
	assertEqual(t, out3.Value.Interface(), TypeInt(1))
}

type TypeBool bool

func TestExtractBoolean(t *testing.T) {
	t1 := true
	out := extractBoolean(t1)
	assertNotNil(t, out)
	assertEqual(t, out.Value.Interface(), true)

	t2 := &t1
	out2 := extractBoolean(t2)
	assertNotNil(t, out2)
	assertEqual(t, out2.Value.Interface(), true)

	var t3 TypeBool = true
	out3 := extractBoolean(t3)
	assertNotNil(t, out3)
	assertEqual(t, out3.Value.Interface(), TypeBool(true))
}

func TestExtractStruct(t *testing.T) {
	t1 := struct {
		Test string
	}{}
	out := extractStruct(t1)
	assertNotNil(t, out)

	t2 := &t1
	out2 := extractStruct(t2)
	assertNotNil(t, out2)
}

func TestParseDocTag(t *testing.T) {
	ret := parseDocTag(`required desc(#DescErrorCode)`)
	assertEqual(t, len(ret), 2)
	assertEqual(t, ret[0][0], "required")
	assertEqual(t, ret[0][1], "")
	assertEqual(t, ret[1][0], "desc")
	assertEqual(t, ret[1][1], "#DescErrorCode")
}
