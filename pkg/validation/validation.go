package validation

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	valid "github.com/go-playground/validator/v10"
)

type Translation struct {
	Tag           string
	TranslationFn func(valid.FieldError) string
}

type validation struct {
	Validator    *valid.Validate
	translations map[string]func(valid.FieldError) string
}

type validationOption func(*validation) error

func New(opts ...validationOption) (*validation, error) {
	validator := valid.New()
	validator.RegisterValidation("possible-values", isPossibleValues)

	v := &validation{
		Validator:    validator,
		translations: map[string]func(valid.FieldError) string{},
	}

	for _, opt := range opts {
		if err := opt(v); err != nil {
			return nil, err
		}
	}

	return v, nil
}

func WithJSONNamesForStructFields() validationOption {
	return func(v *validation) error {
		v.Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
			return name
		})
		return nil
	}
}

func WithPredefinedErrorMessages() validationOption {
	translations := []Translation{
		{
			Tag: "possible-values",
			TranslationFn: func(fe valid.FieldError) string {
				return fmt.Sprintf("The %s must use the following values: %s", fe.Field(), fe.Param())
			},
		},
		{
			Tag: "required_if",
			// TODO: Improve the message for condition required_if for several fields
			TranslationFn: func(fe valid.FieldError) string {
				params := strings.Split(fe.Param(), " ")
				if len(params) >= 2 {
					return fmt.Sprintf("The %s is required since the %s is %s", fe.Field(), params[0], params[1])
				}
				return fmt.Sprintf("The %s is required", fe.Field())
			},
		},
	}
	return WithCustomErrorMessages(translations)
}

func WithCustomErrorMessages(translations []Translation) validationOption {
	return func(v *validation) error {
		for _, t := range translations {
			v.translations[t.Tag] = t.TranslationFn
		}
		return nil
	}
}

func (v *validation) ValidateStruct(s any) error {
	errs := v.Validator.Struct(s)
	if errs != nil {
		if err, ok := errs.(*valid.InvalidValidationError); ok {
			return err
		}
		var errStr strings.Builder
		for _, fe := range errs.(valid.ValidationErrors) {
			if fn, ok := v.translations[fe.Tag()]; ok {
				errStr.WriteString(fn(fe) + ". ")
			} else {
				errStr.WriteString(fe.Error() + ". ")
			}
		}
		return errors.New(errStr.String())
	}
	return nil
}

func isPossibleValues(fl valid.FieldLevel) bool {
	vals := strings.Split(strings.TrimSpace(fl.Param()), " ")

	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		val := field.String()
		// Empty value can be used
		vals = append(vals, "")
		return contains(vals, val)
	case reflect.Int:
		val := strconv.FormatInt(field.Int(), 10)
		return contains(vals, val)
	case reflect.Slice:
		slice, ok := field.Interface().([]string)
		if !ok {
			panic("Only []string can be used as slice type")
		}
		for _, val := range slice {
			if !contains(vals, val) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}
}

func contains(slice []string, str string) bool {
	return slices.Contains(slice, str)
}
