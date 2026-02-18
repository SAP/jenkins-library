package validation

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
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

func (v *validation) ValidateStruct(s any) error {
	var errStr strings.Builder
	errs := v.Validator.Struct(s)
	if errs != nil {
		if err, ok := errs.(*valid.InvalidValidationError); ok {
			return err
		}
		for _, err := range errs.(valid.ValidationErrors) {
			errStr.WriteString(err.Translate(v.Translator) + ". ")
		}
		return errors.New(errStr.String())
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

func contains(slice []string, str string) bool {
	return slices.Contains(slice, str)
}
