package handlers

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// validationErrors transformă erorile de validare în mesaje clare pentru client
func validationErrors(err error) map[string]string {
	errs := make(map[string]string)

	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		errs["error"] = err.Error()
		return errs
	}

	for _, fe := range ve {
		field := fe.Field()
		switch fe.Tag() {
		case "required":
			errs[field] = fmt.Sprintf("%s este obligatoriu", field)
		case "email":
			errs[field] = fmt.Sprintf("%s trebuie să fie o adresă de email validă", field)
		case "min":
			errs[field] = fmt.Sprintf("%s trebuie să aibă minim %s caractere", field, fe.Param())
		case "max":
			errs[field] = fmt.Sprintf("%s poate avea maxim %s caractere", field, fe.Param())
		case "oneof":
			errs[field] = fmt.Sprintf("%s trebuie să fie unul din: %s", field, fe.Param())
		default:
			errs[field] = fmt.Sprintf("%s este invalid (%s)", field, fe.Tag())
		}
	}

	return errs
}
