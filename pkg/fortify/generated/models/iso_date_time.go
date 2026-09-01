package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	strfmt "github.com/go-openapi/strfmt"
)

// IsDateTime returns true when the string is a valid date-time
func IsDateTime(str string) bool {
	if len(str) < 4 {
		return false
	}
	s := strings.Split(strings.ToLower(str), "t")
	if len(s) < 2 || !strfmt.IsDate(s[0]) {
		return false
	}

	matches := rxIso8601MilliDateTime.FindAllStringSubmatch(s[1], -1)
	if len(matches) == 0 || len(matches[0]) == 0 {
		return false
	}
	m := matches[0]
	res := m[1] <= "23" && m[2] <= "59" && m[3] <= "59"
	return res
}

const (
	// ISO8601Milli represents a ISO8601 format to ISO8601 with timezone
	ISO8601Milli = "2006-01-02T15:04:05.000-0700"
	// Iso8601MilliDateTimePattern pattern to match for the date-time format from http://tools.ietf.org/html/rfc3339#section-5.6
	Iso8601MilliDateTimePattern = `^([0-9]{2}):([0-9]{2}):([0-9]{2})(.[0-9]+)?([+-][0-9]{4})$`
)

var (
	dateTimeFormats        = []string{ISO8601Milli}
	rxIso8601MilliDateTime = regexp.MustCompile(Iso8601MilliDateTimePattern)
	// MarshalFormat sets the time resolution format used for marshaling time (set to milliseconds)
	MarshalFormat = ISO8601Milli
)

// ParseIso8601MilliDateTime parses a string that represents an ISO8601 time or a unix epoch
func ParseIso8601MilliDateTime(data string) (Iso8601MilliDateTime, error) {
	if data == "" {
		return NewIso8601MilliDateTime(), nil
	}
	var lastError error
	for _, layout := range dateTimeFormats {
		dd, err := time.Parse(layout, data)
		if err != nil {
			lastError = err
			continue
		}
		return Iso8601MilliDateTime(dd), nil
	}
	return Iso8601MilliDateTime{}, lastError
}

// Iso8601MilliDateTime is a time but it serializes to ISO8601 format with millis
// It knows how to read 3 different variations of a RFC3339 date time.
// Most APIs we encounter want either millisecond or second precision times.
// This just tries to make it worry-free.
//
// swagger:strfmt date-time
type Iso8601MilliDateTime time.Time

// NewIso8601MilliDateTime is a representation of zero value for Iso8601MilliDateTime type
func NewIso8601MilliDateTime() Iso8601MilliDateTime {
	return Iso8601MilliDateTime(time.Unix(0, 0).UTC())
}

// String converts this time to a string
func (t Iso8601MilliDateTime) String() string {
	return time.Time(t).Format(MarshalFormat)
}

// MarshalText implements the text marshaller interface
func (t Iso8601MilliDateTime) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the text unmarshaller interface
func (t *Iso8601MilliDateTime) UnmarshalText(text []byte) error {
	tt, err := ParseIso8601MilliDateTime(string(text))
	if err != nil {
		return err
	}
	*t = tt
	return nil
}

// Scan scans a Iso8601MilliDateTime value from database driver type.
func (t *Iso8601MilliDateTime) Scan(raw interface{}) error {
	// TODO: case int64: and case float64: ?
	switch v := raw.(type) {
	case []byte:
		return t.UnmarshalText(v)
	case string:
		return t.UnmarshalText([]byte(v))
	case time.Time:
		*t = Iso8601MilliDateTime(v)
	case nil:
		*t = Iso8601MilliDateTime{}
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.Iso8601MilliDateTime from: %#v", v)
	}

	return nil
}

// Value converts Iso8601MilliDateTime to a primitive value ready to written to a database.
func (t Iso8601MilliDateTime) Value() (driver.Value, error) {
	return driver.Value(t.String()), nil
}

// MarshalJSON returns the Iso8601MilliDateTime as JSON
func (t Iso8601MilliDateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Format(MarshalFormat))
}

// UnmarshalJSON sets the Iso8601MilliDateTime from JSON
func (t *Iso8601MilliDateTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var tstr string
	if err := json.Unmarshal(data, &tstr); err != nil {
		return err
	}
	tt, err := ParseIso8601MilliDateTime(tstr)
	if err != nil {
		return err
	}
	*t = tt
	return nil
}

// DeepCopyInto copies the receiver and writes its value into out.
func (t *Iso8601MilliDateTime) DeepCopyInto(out *Iso8601MilliDateTime) {
	*out = *t
}

// DeepCopy copies the receiver into a new Iso8601MilliDateTime.
func (t *Iso8601MilliDateTime) DeepCopy() *Iso8601MilliDateTime {
	if t == nil {
		return nil
	}
	out := new(Iso8601MilliDateTime)
	t.DeepCopyInto(out)
	return out
}
