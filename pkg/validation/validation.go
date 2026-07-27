package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	valid "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type Translation struct {
	Tag           string
	RegisterFn    valid.RegisterTranslationsFunc
	TranslationFn valid.TranslationFunc
}

type validation struct {
	Validator  *valid.Validate
	Translator ut.Translator
}

type validationOption func(*validation) error

func New(opts ...validationOption) (*validation, error) {
	validator := valid.New()
	validator.RegisterValidation("possible-values", isPossibleValues)
	validator.RegisterValidation("required_if_oneof", isRequiredIfOneOf)
	enTranslator := en.New()
	universalTranslator := ut.New(enTranslator, enTranslator)
	translator, found := universalTranslator.GetTranslator("en")
	if !found {
		return nil, errors.New("translator for en locale is not found")
	}

	validation := &validation{
		Validator:  validator,
		Translator: translator,
	}

	for _, opt := range opts {
		if err := opt(validation); err != nil {
			return nil, err
		}
	}

	return validation, nil
}

func WithJSONNamesForStructFields() validationOption {
	return func(v *validation) error {
		v.Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			return name
		})
		return nil
	}
}

func WithPredefinedErrorMessages() validationOption {
	translations := []Translation{
		{
			Tag: "possible-values",
			RegisterFn: func(ut ut.Translator) error {
				return ut.Add("possible-values", "The {0} must use the following values: {1}", true)
			},
			TranslationFn: func(ut ut.Translator, fe valid.FieldError) string {
				t, _ := ut.T("possible-values", fe.Field(), fe.Param())
				return t
			},
		}, {
			Tag: "required_if",
			RegisterFn: func(ut ut.Translator) error {
				// TODO: Improve the message for condition required_if for several fields
				return ut.Add("required_if", "The {0} is required since the {1} is {2}", true)
			},
			TranslationFn: func(ut ut.Translator, fe valid.FieldError) string {
				params := []string{fe.Field()}
				params = append(params, strings.Split(fe.Param(), " ")...)
				t, _ := ut.T("required_if", params...)
				return t
			},
		}, {
			Tag: "required_if_oneof",
			RegisterFn: func(ut ut.Translator) error {
				return ut.Add("required_if_oneof", "The {0} is required since the {1} is one of {2}", true)
			},
			TranslationFn: func(ut ut.Translator, fe valid.FieldError) string {
				params := strings.Fields(fe.Param())
				var conditionField, conditionValues string
				if len(params) > 0 {
					conditionField = params[0]
				}
				if len(params) > 1 {
					conditionValues = strings.Join(params[1:], ", ")
				}
				t, _ := ut.T("required_if_oneof", fe.Field(), conditionField, conditionValues)
				return t
			},
		},
	}
	return func(v *validation) error {
		if err := registerTranslations(translations, v.Validator, v.Translator); err != nil {
			return err
		}
		return nil
	}
}

func WithCustomErrorMessages(translations []Translation) validationOption {
	return func(v *validation) error {
		if err := registerTranslations(translations, v.Validator, v.Translator); err != nil {
			return err
		}
		return nil
	}
}

func (v *validation) ValidateStruct(s interface{}) error {
	var errStr string
	errs := v.Validator.Struct(s)
	if errs != nil {
		if err, ok := errs.(*valid.InvalidValidationError); ok {
			return err
		}
		for _, err := range errs.(valid.ValidationErrors) {
			errStr += err.Translate(v.Translator) + ". "
		}
		return errors.New(errStr)
	}
	return nil
}

func registerTranslations(translations []Translation, validator *valid.Validate, translator ut.Translator) error {
	if err := en_translations.RegisterDefaultTranslations(validator, translator); err != nil {
		return err
	}

	for _, t := range translations {
		if err := validator.RegisterTranslation(t.Tag, translator, t.RegisterFn, t.TranslationFn); err != nil {
			return err
		}
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

// isRequiredIfOneOf implements a "required_if_oneof=FieldName val1 val2 ..." validator.
// The annotated field is required if the referenced field (identified by its Go struct
// field name) has one of the given values.
//
// This is needed in addition to the built-in "required_if" validator because that one
// only supports combining multiple field/value pairs with AND semantics and, since
// go-playground/validator v10.15.0, panics with "Duplicate param ... for required_if"
// if the same field is referenced more than once (which would be needed to express an
// OR condition, e.g. "mandatory if field is A or B or C").
func isRequiredIfOneOf(fl valid.FieldLevel) bool {
	params := strings.Fields(fl.Param())
	if len(params) < 2 {
		return true
	}

	conditionFieldName := params[0]
	allowedValues := params[1:]

	conditionField, kind, _, found := fl.GetStructFieldOKAdvanced2(fl.Parent(), conditionFieldName)
	if !found {
		return true
	}

	if kind != reflect.String {
		// only string typed condition fields are supported
		return true
	}

	if !contains(allowedValues, conditionField.String()) {
		return true
	}

	return hasFieldValue(fl.Field())
}

// hasFieldValue reports whether field holds a non-zero/non-nil value.
func hasFieldValue(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		return field.IsValid() && !field.IsZero()
	}
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
