package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aweris/gale/internal/log"
	"github.com/julienschmidt/httprouter"
)

// Serve starts the artifact service router on the given port
func Serve(port string) error {
	router := httprouter.New()

	handler := &handler{}

	router.GET("/_apis/artifactcache/cache", handler.loggingMiddleware(handler.HandleGetCacheEntry))
	router.POST("/_apis/artifactcache/caches", handler.loggingMiddleware(handler.HandleReserveCache))
	router.PATCH("/_apis/artifactcache/caches/:cacheID", handler.loggingMiddleware(handler.HandleUploadCache))
	router.POST("/_apis/artifactcache/caches/:cacheID", handler.loggingMiddleware(handler.HandleCommitCache))
	router.GET("/_apis/artifactcache/artifacts/:artifactID", handler.loggingMiddleware(handler.HandleDownloadArtifact))
	router.GET("/healthz", handler.HandleHealthz)

	fmt.Printf("Starting server on port %s\n", port)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server.ListenAndServe()
}

type handler struct {
	// TBD
}

func (h *handler) HandleGetCacheEntry(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

func (h *handler) HandleReserveCache(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

func (h *handler) HandleUploadCache(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

func (h *handler) HandleDownloadArtifact(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

func (h *handler) HandleCommitCache(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

// HandleHealthz handles health check requests
func (h *handler) HandleHealthz(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}

// sendJSON sends a JSON response with the given status code and data. If data is nil, skips the body.
func (h *handler) sendJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// loggingMiddleware logs the request method, path and query string
func (h *handler) loggingMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

		var keyvals []interface{}

		keyvals = append(keyvals, "method", r.Method)
		keyvals = append(keyvals, "path", r.URL.Path)
		keyvals = append(keyvals, "raw_query", r.URL.RawQuery)

		log.Debugf("http request", keyvals...)

		next(w, r, params)
	}
}
