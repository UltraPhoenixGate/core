package validator

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Inst *validator.Validate

func init() {
	Inst = validator.New(validator.WithRequiredStructEnabled())
}

func Validate(v interface{}) error {
	return Inst.Struct(v)
}

func ShouldBind(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return Validate(v)
}
