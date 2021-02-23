package rest

import (
	"net/http"

	"github.com/brigadecore/brigade-github-app/v2/gateway/internal/lib/restmachinery"
	"github.com/google/go-github/github"
)

type AuthFilterConfig struct {
	AppID        string
	SharedSecret string
}

type authFilter struct {
	config AuthFilterConfig
}

func NewAuthFilter(config AuthFilterConfig) restmachinery.Filter {
	return &authFilter{
		config: config,
	}
}

func (a *authFilter) Decorate(handle http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err :=
			github.ValidatePayload(r, []byte(a.config.SharedSecret)); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(emptyResponse)
			return
		}
		// TODO: Should also check app id
		handle(w, r)
	}
}
