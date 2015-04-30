package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/pivotal-golang/lager"
)

type AuthCookieHandler struct {
	logger lager.Logger
}

func NewAuthCookieHandler(logger lager.Logger) *AuthCookieHandler {
	return &AuthCookieHandler{
		logger: logger.Session("auth-cookie-handler"),
	}
}

func (h *AuthCookieHandler) GenerateCookie(w http.ResponseWriter, req *http.Request) {
	authorization := req.Header.Get("Authorization")

	if authorization != "" {
		cookie := http.Cookie{
			Name:     receptor.AuthorizationCookieName,
			Value:    authorization,
			MaxAge:   0,
			HttpOnly: true,
		}
		w.Header().Set("Set-Cookie", cookie.String())
	}

	w.WriteHeader(http.StatusNoContent)
}
