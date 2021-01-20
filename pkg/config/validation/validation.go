package validation

import (
	"fmt"
	"reflect"
	"strings"
)

// FindEmptyStringsInConfigStruct finds empty strings in a struct.
// in case the struct contains another struct, also this struct is checked.
func FindEmptyStringsInConfigStruct(v interface{}) ([]string, error) {
	emptyStrings := []string{}
	if reflect.ValueOf(v).Kind() != reflect.Struct {
		return emptyStrings, fmt.Errorf("%v (%T) was not a struct", v, v)
	}
	findNestedEmptyStrings(v, &emptyStrings, []string{})
	return emptyStrings, nil
}

func findNestedEmptyStrings(v interface{}, emptyStrings *[]string, prefix []string) (bool, error) {
	fields := reflect.TypeOf(v)
	values := reflect.ValueOf(v)
	for i := 0; i < fields.NumField(); i++ {
		switch values.Field(i).Kind() {
		case reflect.String:
			if len(values.Field(i).String()) == 0 {
				*emptyStrings = append(*emptyStrings, strings.Join(append(prefix, fields.Field(i).Name), "."))
			}
		case reflect.Struct:
			_, err := findNestedEmptyStrings(values.Field(i).Interface(), emptyStrings, append(prefix, fields.Field(i).Name))
			if err != nil {
				return false, err
			}
		case reflect.Int:
		case reflect.Int32:
		case reflect.Int64:
		case reflect.Bool:
		case reflect.Slice:
		default:
			return false, fmt.Errorf("unexpected field: %v, value: %v", fields.Field(i), values.Field(i))
		}
	}
	return false, nil
}
