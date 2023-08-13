package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Serve starts the artifact service router on the given port
func Serve(port string) error {
	router := httprouter.New()

	handler := &handler{}

	router.GET("/healthz", handler.HandleHealthz)

	return http.ListenAndServe(fmt.Sprintf(":%s", port), router)
}

type handler struct {
	// TBD
}

// HandleHealthz handles health check requests
func (h *handler) HandleHealthz(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}
