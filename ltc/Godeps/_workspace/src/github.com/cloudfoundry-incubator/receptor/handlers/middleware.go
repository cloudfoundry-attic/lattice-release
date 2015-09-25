package handlers

import (
	"net/http"
	"strings"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/goji/httpauth"
	"github.com/pivotal-golang/lager"
)

func LogWrap(handler http.Handler, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestLog := logger.Session("request", lager.Data{
			"method":  r.Method,
			"request": r.URL.String(),
		})

		requestLog.Info("serving")
		handler.ServeHTTP(w, r)
		requestLog.Info("done")
	}
}

func CookieAuthWrap(handler http.Handler, cookieName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err == nil {
			r.Header.Set("Authorization", cookie.Value)
		}

		handler.ServeHTTP(w, r)
	}
}

func BasicAuthWrap(handler http.Handler, username, password string) http.Handler {
	opts := httpauth.AuthOptions{
		Realm:               "API Authentication",
		User:                username,
		Password:            password,
		UnauthorizedHandler: http.HandlerFunc(unauthorized),
	}

	return httpauth.BasicAuth(opts)(handler)
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	status := http.StatusUnauthorized
	writeJSONResponse(w, status, &receptor.Error{
		Type:    receptor.Unauthorized,
		Message: http.StatusText(status),
	})
}

// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS
func CORSWrapper(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isValidCORSRequest(r) {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if isCORSPreflightRequest(r) {
			w.Header().Set("Access-Control-Allow-Methods", r.Header.Get("Access-Control-Request-Method"))
			w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
			w.WriteHeader(http.StatusOK)
		} else {
			handler.ServeHTTP(w, r)
		}
	}
}

var invalidOriginHeaders = map[string]struct{}{
	"":  struct{}{},
	"*": struct{}{},
}

func isValidCORSRequest(r *http.Request) bool {
	_, isBlacklistedOrigin := invalidOriginHeaders[r.Header.Get("Origin")]
	return !isBlacklistedOrigin
}

func isCORSPreflightRequest(r *http.Request) bool {
	return strings.ToUpper(r.Method) == "OPTIONS"
}
