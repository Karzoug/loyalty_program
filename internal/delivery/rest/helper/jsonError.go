package helper

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type jsonError struct {
	Error string `json:"error"`
}

func WriteJSONError(w http.ResponseWriter, msg string, code int, logger *zap.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errStruct := jsonError{
		Error: msg,
	}
	b, _ := json.Marshal(errStruct)
	_, err := w.Write(b)
	if err != nil {
		logger.Error("Write the data to response error", zap.Error(err))
	}
}
