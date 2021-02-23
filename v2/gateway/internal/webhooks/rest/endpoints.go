package rest

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/brigadecore/brigade-github-app/v2/gateway/internal/lib/restmachinery" // nolint: lll
	"github.com/brigadecore/brigade-github-app/v2/gateway/internal/webhooks"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
)

var emptyResponse = []byte("{}")

type WebhookEndpoints struct {
	AuthFilter restmachinery.Filter
	Service    webhooks.Service
}

func (w *WebhookEndpoints) Register(router *mux.Router) {
	router.HandleFunc(
		"/events",
		w.AuthFilter.Decorate(w.handle),
	).Methods(http.MethodPost)
}

func (w *WebhookEndpoints) handle(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(emptyResponse)
		return
	}

	var payload []byte
	switch contentType := r.Header.Get("Content-Type"); contentType {

	case "application/json":
		payload = body

	case "application/x-www-form-urlencoded":
		form, err := url.ParseQuery(string(body))
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write(emptyResponse)
			return
		}
		payload = []byte(form.Get("payload"))

	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(emptyResponse)
		return

	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(emptyResponse)
		return
	}

	if err = w.Service.Handle(r.Context(), event, payload); err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(emptyResponse)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(emptyResponse)
}
