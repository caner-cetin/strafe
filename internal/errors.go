package internal

import (
	"encoding/json"
	"net/http"
	"runtime"

	log "github.com/sirupsen/logrus"
)

type BaseError struct {
	Code    int
	Message string
	Error   error
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var (
	MalformedJSONBody = BaseError{
		Code:    http.StatusBadRequest,
		Message: "cannot decode the malformed json body",
	}
	BrokenTransport = BaseError{
		Code:    http.StatusBadGateway,
		Message: "cannot encode data, http transport is broken",
	}
	ServerErrorBase = BaseError{
		Code: http.StatusInternalServerError,
		Message: "internal server error",
	}
)
func WriteError(w http.ResponseWriter, err BaseError) {
	pc, file, no, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	var entry *log.Entry
	fields := make(log.Fields)
	fields["message"] = err.Message
	fields["error"] = err.Error.Error()
	if ok {
		fields["function"] = details.Name()
		fields["file"] = file
		fields["line"] = no
	}
	entry = entry.WithFields(fields)
	if err.Code == http.StatusInternalServerError {
		entry.Error("internal error")
	} else {
		entry.Warn("bad request")
	}
	w.WriteHeader(err.Code)
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(ErrorResponse{Message: err.Message})
}

func ServerError(err error) BaseError {
	var base = ServerErrorBase
	base.Error = err
	return base
}
func WriteServerError(w http.ResponseWriter, err error) {
	var base = ServerErrorBase
	base.Error = err
	WriteError(w, base)
}