package validator

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
)

var Inst *validator.Validate
var decoder = schema.NewDecoder()

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

func ShouldBindQuery(r *http.Request, v interface{}) error {
	// TODO
	query := r.URL.Query()

	if err := decoder.Decode(v, query); err != nil {
		return err
	}

	// Optionally, you can add a validation step if your application requires it
	return Validate(v)
}
