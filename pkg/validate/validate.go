package validate

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var ErrRequest = errors.New("request error")

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func BindRequestBody(req *http.Request, param any) error {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return validate.StructCtx(req.Context(), param)
	}
	err = json.Unmarshal(body, param)
	if err != nil {
		return err
	}
	return validate.StructCtx(req.Context(), param)
}
