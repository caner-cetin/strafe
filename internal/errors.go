package internal

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
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
	MissingJSONBody = WrapErr(BaseError{
		Code:    http.StatusBadRequest,
		Message: "missing json body",
	})
	MalformedJSONBody = WrapErr(BaseError{
		Code:    http.StatusBadRequest,
		Message: "cannot decode the malformed json body",
	})
	BrokenTransport = WrapErr(BaseError{
		Code:    http.StatusBadGateway,
		Message: "cannot encode data, http transport is broken",
	})
	ServerErrorBase = WrapErr(BaseError{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
	})
)

func WriteError(w http.ResponseWriter, err BaseError) {
	var pc, file, no, ok = runtime.Caller(1)
	var details = runtime.FuncForPC(pc)
	if strings.Contains(strings.ToLower(details.Name()), "servererror") {
		// then caller is the internal.ServerError, we need to skip two
		pc, file, no, ok = runtime.Caller(2)
		details = runtime.FuncForPC(pc)
	}
	var entry = log.With().Err(err.Error)
	if ok {
		entry.Str("function", details.Name())
		entry.Str("file", file)
		entry.Int("line", no)
	}
	logger := entry.Logger()
	if err.Code == http.StatusInternalServerError {
		logger.Error().Msg("internal error")
	} else {
		logger.Warn().Msg("bad request")
	}
	w.WriteHeader(err.Code)
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(ErrorResponse{Message: err.Message})
}

func WrapErr(base BaseError) func(err error) BaseError {
	return func(err error) BaseError {
		base.Error = err
		return base
	}
}

// same as
//
//	internal.WriteError(w, internal.ServerError(err)) // 49 characters
//
// but you can
// write
//
//	internal.ServerError(w, err) // 28 characters
//
// to save 21 characters
//
// you cant shave off the extra unnecessary `return `line.
// so we are currently doing this:
//
//	if err != nil {
//		internal.ServerError(w, err)
//		return
//	}
//
// and you cannot do this:
//
//	if err != nil {
//		return internal.ServerError(w, err)
//	}
//
// because for returning from a function, that function also needs to return *something*
//
// https://pkg.go.dev/golang.org/x/tools/internal/typesinternal#TooManyValues
//
// a simple *return* from this function wont suffice, for doing `return ...` from the main function, the wrapper needs to return *something*
// and when wrapper returns something, main function returns that same thing, as we are calling this from handlers, handlers cannot return anything.
//
// magna carta, or more accurately, this pdf is https://www.archives.gov/files/press/press-kits/magna-carta/magna-carta-translation.pdf
// 19270 characters long
//
// so if you write 917 err != nil blocks, you can save enough characters to write the magna carta
//
// considering how `GetRandomTrack`, a function with 60 line of code is containing 6 calls to this function
// i am pretty sure you can hit that 917 soon enough.
//
// fun to read threads while contemplating the life choices that led to writing golang
// https://www.reddit.com/r/programmingcirclejerk/comments/1giuuxu/go_really_blew_me_away_with_its_explicit_error/
// https://www.reddit.com/r/programmingcirclejerk/comments/pomrnj/i_have_found_beauty_in_the_explicit_nature_of/
// https://www.reddit.com/r/programmingcirclejerk/comments/bmjatd/go_libraries_dont_have_bugs_because_the_authors/
// https://www.reddit.com/r/programmingcirclejerk/comments/6v3ykh/the_elimination_of_the_boilerplate_error_handling/
// https://www.reddit.com/r/programmingcirclejerk/comments/hkrx1g/many_functions_have_more_if_err_nil_return_err/
// https://www.reddit.com/r/programmingcirclejerk/comments/1icqfdt/no_go_will_never_support_this_as_it_doesnt_make/
// https://www.reddit.com/r/programmingcirclejerk/comments/lr8gr0/we_have_a_language_with_linear_types_dependent/
// https://www.reddit.com/r/programmingcirclejerk/comments/1ggiltj/its_go_25_of_the_code_is_just_basic_error/
// https://www.reddit.com/r/programmingcirclejerk/comments/12oiiuc/how_go_fixed_everything_that_was_wrong_with/
//
// more serious threads
// https://www.reddit.com/r/programming/comments/pnzgj5/going_insane_endless_error_handling/
// https://www.reddit.com/r/golang/comments/134k3u8/will_error_handling_ever_changeimprove/
// https://www.reddit.com/r/programming/comments/1ee3qod/gos_error_handling_a_grave_error/
//
// github search link so that you can look at the beauty of err != nil blocks
// https://github.com/search?q=%22if+err+%21%3D+nil+%7B+return+err+%7D%22&type=code
//
// if you are delusional you might enjoy this
// https://www.reddit.com/r/programming/comments/18wndv9/error_handling_in_go_is_awesome_here_is_why/
//
// god bless rob pike
func ServerError(w http.ResponseWriter, err error) {
	WriteError(w, ServerErrorBase(err))
}
