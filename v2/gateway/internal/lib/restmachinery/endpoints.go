package restmachinery

import "github.com/gorilla/mux"

// Endpoints is an interface to be implemented by all REST API endpoints.
type Endpoints interface {
	// Register is invoked during Server initialization, giving endpoint
	// implementations an opportunity to register path/function mappings with
	// the provided *mux.Router.
	Register(router *mux.Router)
}
