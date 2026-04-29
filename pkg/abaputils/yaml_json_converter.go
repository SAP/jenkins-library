package abaputils

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

const (
	caseMask     = ^byte(0x20)
	kelvin       = '\u212a'
	smallLongEss = '\u017f'
)

var fieldCache struct {
	sync.RWMutex
	m map[reflect.Type][]field
}

// field represents a struct field.
type field struct {
	name      string
	nameBytes []byte
	equalFold func(s, t []byte) bool
	tag       bool
	index     []int
	typ       reflect.Type
	omitEmpty bool
	quoted    bool
}

type tagOptions string

// byName sorts fields by name, then by index length, then by tag, then by index.
type byName []field

func (x byName) Len() int      { return len(x) }
func (x byName) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x byName) Less(i, j int) bool {
	if x[i].name != x[j].name {
		return x[i].name < x[j].name
	}
	if len(x[i].index) != len(x[j].index) {
		return len(x[i].index) < len(x[j].index)
	}
	if x[i].tag != x[j].tag {
		return x[i].tag
	}
	return byIndex(x).Less(i, j)
}

// byIndex sorts fields by index sequence.
type byIndex []field

func (x byIndex) Len() int      { return len(x) }
func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x byIndex) Less(i, j int) bool {
	for k, xik := range x[i].index {
		if k >= len(x[j].index) {
			return false
		}
		if xik != x[j].index[k] {
			return xik < x[j].index[k]
		}
	}
	return len(x[i].index) < len(x[j].index)
}

// YAMLToJSON converts YAML to JSON.
func YAMLToJSON(y []byte) ([]byte, error) {
	return yamlToJSON(y, nil, yaml.Unmarshal)
}

func yamlToJSON(y []byte, jsonTarget *reflect.Value, yamlUnmarshal func([]byte, interface{}) error) ([]byte, error) {
	var yamlObj interface{}
	if err := yamlUnmarshal(y, &yamlObj); err != nil {
		return nil, err
	}
	jsonObj, err := convertToJSONableObject(yamlObj, jsonTarget)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jsonObj)
}

func convertToJSONableObject(yamlObj interface{}, jsonTarget *reflect.Value) (interface{}, error) {
	// Handle pointer and interface indirection for jsonTarget
	if jsonTarget != nil {
		ju, tu, pv := indirect(*jsonTarget, false)
		if ju != nil || tu != nil {
			jsonTarget = nil
		} else {
			jsonTarget = &pv
		}
	}

	switch v := yamlObj.(type) {
	case map[interface{}]interface{}:
		strMap := make(map[string]interface{})
		for k, val := range v {
			keyString, err := yamlKeyToString(k)
			if err != nil {
				return nil, fmt.Errorf("unsupported map key of type: %T, key: %+#v, value: %+#v", k, k, val)
			}
			if jsonTarget != nil {
				t := *jsonTarget
				switch t.Kind() {
				case reflect.Struct:
					keyBytes := []byte(keyString)
					var f *field
					fields := cachedTypeFields(t.Type())
					for i := range fields {
						ff := &fields[i]
						if bytes.Equal(ff.nameBytes, keyBytes) || ff.equalFold(ff.nameBytes, keyBytes) {
							f = ff
							break
						}
					}
					if f != nil {
						jtf := t.Field(f.index[0])
						strMap[keyString], err = convertToJSONableObject(val, &jtf)
						if err != nil {
							return nil, err
						}
						continue
					}
				case reflect.Map:
					jtv := reflect.Zero(t.Type().Elem())
					strMap[keyString], err = convertToJSONableObject(val, &jtv)
					if err != nil {
						return nil, err
					}
					continue
				}
			}
			strMap[keyString], err = convertToJSONableObject(val, nil)
			if err != nil {
				return nil, err
			}
		}
		return strMap, nil
	case []interface{}:
		var jsonSliceElemValue *reflect.Value
		if jsonTarget != nil {
			t := *jsonTarget
			if t.Kind() == reflect.Slice {
				ev := reflect.Indirect(reflect.New(t.Type().Elem()))
				jsonSliceElemValue = &ev
			}
		}
		arr := make([]interface{}, len(v))
		for i, elem := range v {
			var err error
			arr[i], err = convertToJSONableObject(elem, jsonSliceElemValue)
			if err != nil {
				return nil, err
			}
		}
		return arr, nil
	default:
		if jsonTarget != nil && (*jsonTarget).Kind() == reflect.String {
			if s, ok := yamlPrimitiveToString(v); ok {
				yamlObj = s
			}
		}
		return yamlObj, nil
	}
}

func yamlKeyToString(k interface{}) (string, error) {
	switch key := k.(type) {
	case string:
		return key, nil
	case int:
		return strconv.Itoa(key), nil
	case int64:
		return strconv.FormatInt(key, 10), nil
	case float64:
		s := strconv.FormatFloat(key, 'g', -1, 32)
		switch s {
		case "+Inf":
			s = ".inf"
		case "-Inf":
			s = "-.inf"
		case "NaN":
			s = ".nan"
		}
		return s, nil
	case bool:
		if key {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("unsupported key type: %T", k)
	}
}

func yamlPrimitiveToString(v interface{}) (string, bool) {
	switch val := v.(type) {
	case int:
		return strconv.FormatInt(int64(val), 10), true
	case int64:
		return strconv.FormatInt(val, 10), true
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 32), true
	case uint64:
		return strconv.FormatUint(val, 10), true
	case bool:
		if val {
			return "true", true
		}
		return "false", true
	default:
		return "", false
	}
}

// indirect walks down v allocating pointers as needed, until it gets to a non-pointer.
func indirect(v reflect.Value, decodingNull bool) (json.Unmarshaler, encoding.TextUnmarshaler, reflect.Value) {
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
				v = e
				continue
			}
		}
		if v.Kind() != reflect.Ptr {
			break
		}
		if v.Elem().Kind() != reflect.Ptr && decodingNull && v.CanSet() {
			break
		}
		if v.IsNil() {
			if v.CanSet() {
				v.Set(reflect.New(v.Type().Elem()))
			} else {
				v = reflect.New(v.Type().Elem())
			}
		}
		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(json.Unmarshaler); ok {
				return u, nil, reflect.Value{}
			}
			if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
				return nil, u, reflect.Value{}
			}
		}
		v = v.Elem()
	}
	return nil, nil, v
}

// cachedTypeFields returns cached fields for a type.
func cachedTypeFields(t reflect.Type) []field {
	fieldCache.RLock()
	f := fieldCache.m[t]
	fieldCache.RUnlock()
	if f != nil {
		return f
	}
	f = typeFields(t)
	if f == nil {
		f = []field{}
	}
	fieldCache.Lock()
	if fieldCache.m == nil {
		fieldCache.m = map[reflect.Type][]field{}
	}
	fieldCache.m[t] = f
	fieldCache.Unlock()
	return f
}

// typeFields returns a list of fields that JSON should recognize for the given type.
func typeFields(t reflect.Type) []field {
	current := []field{}
	next := []field{{typ: t}}
	count := map[reflect.Type]int{}
	nextCount := map[reflect.Type]int{}
	visited := map[reflect.Type]bool{}
	var fields []field

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}
		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				if sf.PkgPath != "" || sf.Tag.Get("json") == "-" {
					continue
				}
				name, opts := parseTag(sf.Tag.Get("json"))
				if !isValidTag(name) {
					name = ""
				}
				index := append(append([]int(nil), f.index...), i)
				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {

					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					fields = append(fields, fillField(field{
						name:      name,
						tag:       tagged,
						index:     index,
						typ:       ft,
						omitEmpty: opts.Contains("omitempty"),
						quoted:    opts.Contains("string"),
					}))
					if count[f.typ] > 1 {
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, fillField(field{name: ft.Name(), index: index, typ: ft}))
				}
			}
		}
	}

	sort.Sort(byName(fields))

	// Remove hidden fields, keeping only the dominant one for each name.
	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			if fields[i+advance].name != name {
				break
			}
		}
		if advance == 1 {
			out = append(out, fi)
			continue
		}
		if dominant, ok := dominantField(fields[i : i+advance]); ok {
			out = append(out, dominant)
		}
	}
	fields = out
	sort.Sort(byIndex(fields))
	return fields
}

// parseTag splits a struct field's json tag into its name and options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && !strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c) {
			return false
		}
	}
	return true
}

func fillField(f field) field {
	f.nameBytes = []byte(f.name)
	f.equalFold = foldFunc(f.nameBytes)
	return f
}

// foldFunc returns a case folding equivalence function for s.
func foldFunc(s []byte) func(s, t []byte) bool {
	nonLetter, special := false, false
	for _, b := range s {
		if b >= utf8.RuneSelf {
			return bytes.EqualFold
		}
		upper := b & caseMask
		if upper < 'A' || upper > 'Z' {
			nonLetter = true
		} else if upper == 'K' || upper == 'S' {
			special = true
		}
	}
	if special {
		return equalFoldRight
	}
	if nonLetter {
		return asciiEqualFold
	}
	return simpleLetterEqualFold
}

func equalFoldRight(s, t []byte) bool {
	for _, sb := range s {
		if len(t) == 0 {
			return false
		}
		tb := t[0]
		if tb < utf8.RuneSelf {
			if sb != tb {
				sbUpper := sb & caseMask
				if 'A' <= sbUpper && sbUpper <= 'Z' {
					if sbUpper != tb&caseMask {
						return false
					}
				} else {
					return false
				}
			}
			t = t[1:]
			continue
		}
		tr, size := utf8.DecodeRune(t)
		switch sb {
		case 's', 'S':
			if tr != smallLongEss {
				return false
			}
		case 'k', 'K':
			if tr != kelvin {
				return false
			}
		default:
			return false
		}
		t = t[size:]
	}
	return len(t) == 0
}

func asciiEqualFold(s, t []byte) bool {
	if len(s) != len(t) {
		return false
	}
	for i, sb := range s {
		tb := t[i]
		if sb == tb {
			continue
		}
		if ('a' <= sb && sb <= 'z') || ('A' <= sb && sb <= 'Z') {
			if sb&caseMask != tb&caseMask {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func simpleLetterEqualFold(s, t []byte) bool {
	if len(s) != len(t) {
		return false
	}
	for i, b := range s {
		if b&caseMask != t[i]&caseMask {
			return false
		}
	}
	return true
}

// Contains reports whether a comma-separated list of options contains a particular flag.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

// dominantField returns the dominant field among fields with the same name.
func dominantField(fields []field) (field, bool) {
	length := len(fields[0].index)
	tagged := -1
	for i, f := range fields {
		if len(f.index) > length {
			fields = fields[:i]
			break
		}
		if f.tag {
			if tagged >= 0 {
				return field{}, false
			}
			tagged = i
		}
	}
	if tagged >= 0 {
		return fields[tagged], true
	}
	if len(fields) > 1 {
		return field{}, false
	}
	return fields[0], true
}
