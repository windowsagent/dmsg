package api

import (
	"net/http"

	"github.com/SkycoinProject/dmsg/disc"
)

var apiErrors = map[error]func() (int, string){

	disc.ErrKeyNotFound: func() (int, string) {
		return http.StatusNotFound, disc.ErrKeyNotFound.Error()
	},

	disc.ErrUnexpected: func() (int, string) {
		return http.StatusInternalServerError, disc.ErrUnexpected.Error()
	},

	disc.ErrUnauthorized: func() (int, string) {
		return http.StatusUnauthorized, disc.ErrUnauthorized.Error()
	},

	disc.ErrBadInput: func() (int, string) {
		return http.StatusBadRequest, disc.ErrBadInput.Error()
	},
}

func (a *API) handleError(w http.ResponseWriter, e error) {
	var code int
	var msg string

	if _, ok := e.(disc.EntryValidationError); ok {
		code = http.StatusUnprocessableEntity
		msg = e.Error()
	} else {
		f, ok := apiErrors[e]
		if !ok {
			f = func() (int, string) { return http.StatusInternalServerError, disc.ErrUnexpected.Error() }
		}
		code, msg = f()
	}

	if a.logger != nil && code != http.StatusNotFound {
		a.logger.Warnf("%d: %s", code, e)
	}

	a.writeJSON(w, code, disc.HTTPMessage{Code: code, Message: msg})
}
