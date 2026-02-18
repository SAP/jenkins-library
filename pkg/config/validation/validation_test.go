//go:build unit

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Connection struct {
	Endpoint string
	User     string
	Password string
}

type Dummy struct {
	Connection Connection
	Prop1      string
	Prop2      string
	Prop3      string
	Bool1      bool
	Int1       int64
	List       []string
}

func TestWeProvideNotAStruct(t *testing.T) {
	_, err := FindEmptyStringsInConfigStruct("Hello World")
	assert.EqualError(t, err, "'Hello World' (string) is not a struct")
}

func TestUnsupportedType(t *testing.T) {

	type DummyWithUnsupportedType struct {
		Dummy
		NotExpected float32
	}

	_, err := FindEmptyStringsInConfigStruct(DummyWithUnsupportedType{})
	assert.EqualError(t, err, "unexpected type 'float32' of field: 'NotExpected', value: '0'")

}

func TestFindEmptyStringsInConfig(t *testing.T) {
	myStruct := Dummy{
		Connection: Connection{
			Endpoint: "<set>",
			User:     "",
			Password: "<set>",
		},
		Prop1: "<set>",
		Prop2: "", // this is empty
		// Prop3: "this is missing intentionally"
		Bool1: false,
		Int1:  42,
		List:  []string{"1", "2"},
	}
	emptyStrings, err := FindEmptyStringsInConfigStruct(myStruct)
	if assert.NoError(t, err) {
		assert.Len(t, emptyStrings, 3)
		assert.Subset(t, emptyStrings, []string{
			"Connection.User", // empty value, nested
			"Prop2",           // empty value
			"Prop3",           // missing value
		})
	}
}
