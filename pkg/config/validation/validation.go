package validation

import (
	"fmt"
	"reflect"
	"strings"
)

// FindEmptyStringsInConfigStruct finds empty strings in a struct.
// In case the struct contains other nested structs, these struct are also checked.
func FindEmptyStringsInConfigStruct(v interface{}) ([]string, error) {
	emptyStrings := []string{}
	if reflect.ValueOf(v).Kind() != reflect.Struct {
		return emptyStrings, fmt.Errorf("'%v' (%T) is not a struct", v, v)
	}
	err := findNestedEmptyStrings(v, &emptyStrings, []string{})
	return emptyStrings, err
}

func findNestedEmptyStrings(v interface{}, emptyStrings *[]string, prefix []string) error {
	fields := reflect.TypeOf(v)
	values := reflect.ValueOf(v)
	for i := 0; i < fields.NumField(); i++ {
		switch values.Field(i).Kind() {
		case reflect.String:
			if len(values.Field(i).String()) == 0 {
				*emptyStrings = append(*emptyStrings, strings.Join(append(prefix, fields.Field(i).Name), "."))
			}
		case reflect.Struct:
			err := findNestedEmptyStrings(values.Field(i).Interface(), emptyStrings, append(prefix, fields.Field(i).Name))
			if err != nil {
				return err
			}
		case reflect.Int:
		case reflect.Int32:
		case reflect.Int64:
		case reflect.Bool:
		case reflect.Slice:
		// in case someone needs more types these types can be added here
		default:
			return fmt.Errorf("unexpected type '%v' of field: '%v', value: '%v'",
				values.Field(i).Kind(),
				fields.Field(i).Name,
				values.Field(i),
			)
		}
	}
	return nil
}
