package helper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
)

// DecodeJSON is a method that in the case of response header/body valid for JSON format,
// attempts to decode the content into the given value.
func DecodeJSON(r *http.Request, value any) error {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return &HandlerError{
			Message: "not valid content-type",
			Code:    http.StatusBadRequest,
		}
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(value)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		var msg string
		code := http.StatusBadRequest
		switch {
		case errors.As(err, &syntaxError):
			msg = fmt.Sprintf("request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg = "request body contains badly-formed JSON"
		case errors.As(err, &unmarshalTypeError):
			msg = fmt.Sprintf("request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg = fmt.Sprintf("request body contains unknown field %s", fieldName)
		case errors.Is(err, io.EOF):
			msg = "request body must not be empty"
		default:
			return err
		}

		return &HandlerError{
			Message: msg,
			Code:    code,
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return &HandlerError{
			Message: "request body must only contain a single JSON object/array",
			Code:    http.StatusBadRequest,
		}
	}
	return nil
}
