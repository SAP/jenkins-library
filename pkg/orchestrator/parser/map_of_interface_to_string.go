package parser

import (
	"fmt"
	"sort"
	"strings"
)

// targetImagesFlag implements pflag.Value to parse repeated key/value pairs
// into a map[string]string. Accepts "key=value", "key = value", "key: value".
type targetImagesFlag map[string]string

func (m *targetImagesFlag) String() string {
	if m == nil || len(*m) == 0 {
		return ""
	}
	// Stable output (sorted keys)
	keys := make([]string, 0, len(*m))
	for k := range *m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString((*m)[k])
	}
	return b.String()
}

func (m *targetImagesFlag) Type() string {
	return "mapStringString"
}

func (m *targetImagesFlag) Set(s string) error {
	// Allow comma-separated pairs in a single invocation:
	// --targetImages "a=1,b=2"
	parts := splitCSVLike(s)
	if *m == nil {
		*m = make(map[string]string, len(parts))
	}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Support "=", ":"
		var key, val string
		if i := strings.IndexAny(p, "=:"); i >= 0 {
			key = strings.TrimSpace(p[:i])
			val = strings.TrimSpace(p[i+1:])
		} else {
			return fmt.Errorf("invalid targetImages entry %q (expected key=value or key: value)", p)
		}
		if key == "" {
			return fmt.Errorf("empty key in %q", p)
		}
		if val == "" {
			return fmt.Errorf("empty value for key %q", key)
		}
		(*m)[key] = val
	}
	return nil
}

// splitCSVLike splits on commas, but ignores commas inside quotes.
// Minimal handling for simple inputs.
func splitCSVLike(s string) []string {
	// For most CLI inputs, a simple Split is fine:
	// return strings.Split(s, ",")
	// If you want quoted values with commas, use a small scanner:
	var out []string
	var b strings.Builder
	inQuotes := false
	for _, r := range s {
		switch r {
		case '"':
			inQuotes = !inQuotes
			b.WriteRune(r)
		case ',':
			if inQuotes {
				b.WriteRune(r)
			} else {
				out = append(out, b.String())
				b.Reset()
			}
		default:
			b.WriteRune(r)
		}
	}
	out = append(out, b.String())
	return out
}

//package parser
//
//import (
//	"bytes"
//	"fmt"
//	"reflect"
//	"strconv"
//	"strings"
//)
//
//type interfaceToString struct {
//	value   *map[string]string
//	changed bool
//}
//
//func newInterfaceToString(val map[string]string, p *map[string]string) *interfaceToString {
//	ssv := new(interfaceToString)
//	ssv.value = p
//	*ssv.value = val
//	return ssv
//}
//
//// Format: a=1,b=2
//func (s *interfaceToString) Set(val string) error {
//	ss := strings.Split(val, ",")
//	out := make(map[string]string, len(ss))
//	for _, pair := range ss {
//		kv := strings.SplitN(pair, ":", 2) //todo: do we support only : as separator or also = ???
//		if len(kv) != 2 {
//			return fmt.Errorf("%s must be formatted as key:value", pair)
//		}
//		out[kv[0]] = kv[1]
//	}
//	if !s.changed {
//		*s.value = out
//	} else {
//		for k, v := range out {
//			(*s.value)[k] = v
//		}
//	}
//	s.changed = true
//	return nil
//}
//
//func (s *interfaceToString) Type() string {
//	return "interfaceToString"
//}
//
//func (s *interfaceToString) String() string {
//	var buf bytes.Buffer
//	i := 0
//	for k, v := range *s.value {
//		if i > 0 {
//			buf.WriteRune(',')
//		}
//		buf.WriteString(k)
//		buf.WriteRune(':')
//		buf.WriteString(v)
//		i++
//	}
//	return "[" + buf.String() + "]"
//}
//
//func interfaceToStringConv(val any, sep string) (string, error) {
//	// Normalize separator: default to ": " if empty
//	if sep == "" {
//		sep = ": "
//	}
//
//	return joinMapStringString(val, sep), nil
//}
//
//func joinMapStringString(m map[string]string, sep string) string {
//	var b strings.Builder
//	first := true
//	for k, v := range m {
//		if !first {
//			b.WriteByte(',')
//		}
//		first = false
//		b.WriteString(k)
//		b.WriteString(sep)
//		b.WriteString(v)
//	}
//	return b.String()
//}
//
//// GetStringToInt return the map[string]int value of a flag with the given name
//func (f *FlagSet) GetStringToInt(name string) (map[string]int, error) {
//	val, err := f.getFlagType(name, "stringToInt", interfaceToStringConv)
//	if err != nil {
//		return map[string]int{}, err
//	}
//	return val.(map[string]int), nil
//}
//
//// StringToIntVar defines a string flag with specified name, default value, and usage string.
//// The argument p points to a map[string]int variable in which to store the values of the multiple flags.
//// The value of each argument will not try to be separated by comma
//func (f *FlagSet) StringToIntVar(p *map[string]int, name string, value map[string]int, usage string) {
//	f.VarP(newInterfaceToString(value, p), name, "", usage)
//}
//
//// StringToIntVarP is like StringToIntVar, but accepts a shorthand letter that can be used after a single dash.
//func (f *FlagSet) StringToIntVarP(p *map[string]int, name, shorthand string, value map[string]int, usage string) {
//	f.VarP(newInterfaceToString(value, p), name, shorthand, usage)
//}
//
//// StringToIntVar defines a string flag with specified name, default value, and usage string.
//// The argument p points to a map[string]int variable in which to store the value of the flag.
//// The value of each argument will not try to be separated by comma
//func StringToIntVar(p *map[string]int, name string, value map[string]int, usage string) {
//	CommandLine.VarP(newInterfaceToString(value, p), name, "", usage)
//}
//
//// StringToIntVarP is like StringToIntVar, but accepts a shorthand letter that can be used after a single dash.
//func StringToIntVarP(p *map[string]int, name, shorthand string, value map[string]int, usage string) {
//	CommandLine.VarP(newInterfaceToString(value, p), name, shorthand, usage)
//}
//
//// StringToInt defines a string flag with specified name, default value, and usage string.
//// The return value is the address of a map[string]int variable that stores the value of the flag.
//// The value of each argument will not try to be separated by comma
//func (f *FlagSet) StringToInt(name string, value map[string]int, usage string) *map[string]int {
//	p := map[string]int{}
//	f.StringToIntVarP(&p, name, "", value, usage)
//	return &p
//}
//
//// StringToIntP is like StringToInt, but accepts a shorthand letter that can be used after a single dash.
//func (f *FlagSet) StringToIntP(name, shorthand string, value map[string]int, usage string) *map[string]int {
//	p := map[string]int{}
//	f.StringToIntVarP(&p, name, shorthand, value, usage)
//	return &p
//}
//
//// StringToInt defines a string flag with specified name, default value, and usage string.
//// The return value is the address of a map[string]int variable that stores the value of the flag.
//// The value of each argument will not try to be separated by comma
//func StringToInt(name string, value map[string]int, usage string) *map[string]int {
//	return CommandLine.StringToIntP(name, "", value, usage)
//}
//
//// StringToIntP is like StringToInt, but accepts a shorthand letter that can be used after a single dash.
//func StringToIntP(name, shorthand string, value map[string]int, usage string) *map[string]int {
//	return CommandLine.StringToIntP(name, shorthand, value, usage)
//}
