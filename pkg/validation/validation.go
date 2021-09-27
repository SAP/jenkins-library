package validation

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	valid "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type validation struct {
	Validator  *valid.Validate
	Translator ut.Translator
}

func New() (*validation, error) {
	validator := valid.New()
	enTranslator := en.New()
	universalTranslator := ut.New(enTranslator, enTranslator)
	translator, found := universalTranslator.GetTranslator("en")
	if !found {
		return nil, errors.New("translator for en locale is not found")
	}
	validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		return name
	})

	if err := registerTranslations(validator, translator); err != nil {
		return nil, err
	}
	return &validation{
		Validator:  validator,
		Translator: translator,
	}, nil
}

func (v *validation) ValidateStruct(s interface{}) error {
	var errStr string
	errs := v.Validator.Struct(s)
	if errs != nil {
		for _, err := range errs.(valid.ValidationErrors) {
			errStr += err.Translate(v.Translator)
		}
		return errors.New(errStr)
	}
	return nil
}

func registerTranslations(validator *valid.Validate, translator ut.Translator) error {
	if err := en_translations.RegisterDefaultTranslations(validator, translator); err != nil {
		return err
	}

	err := validator.RegisterTranslation("oneof", translator, func(ut ut.Translator) error {
		return ut.Add("oneof", "The {0} must use the folowing values: {1}. ", true)
	}, func(ut ut.Translator, fe valid.FieldError) string {
		t, _ := ut.T("oneof", fe.Field(), fe.Param())
		return t
	})
	if err != nil {
		return err
	}
	err = validator.RegisterTranslation("required_if", translator, func(ut ut.Translator) error {
		return ut.Add("required_if", "The {0} is required since the {1} is {2}. ", true)
	}, func(ut ut.Translator, fe valid.FieldError) string {
		params := strings.Split(fe.Param(), " ")
		t, _ := ut.T("required_if", fe.Field(), params[0], params[1])
		return t
	})
	if err != nil {
		return err
	}
	return nil
}
