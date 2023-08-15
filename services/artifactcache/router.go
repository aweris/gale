package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aweris/gale/internal/log"
	"github.com/julienschmidt/httprouter"
)

// Serve starts the artifact service router on the given port
func Serve(port string, srv Service) error {
	router := httprouter.New()

	handler := &handler{srv: srv}

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
	srv Service
}

func (h *handler) HandleGetCacheEntry(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var keys []string

	// parse keys from query string and normalize them
	for _, key := range strings.Split(r.URL.Query().Get("keys"), ",") {
		keys = append(keys, strings.TrimSpace(strings.ToLower(key)))
	}

	version := strings.TrimSpace(r.URL.Query().Get("version"))

	// find cache entry
	ok, entry, err := h.srv.Find(keys, version)
	if err != nil {
		h.sendJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// return 204 if cache entry not found
	if !ok {
		h.sendJSON(w, http.StatusNoContent, nil)
		return
	}

	h.sendJSON(w, http.StatusOK, entry)
}

func (h *handler) HandleReserveCache(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req ReserveCacheRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Errorf("Error decoding request: %s", "error", err)
		h.sendJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// key, version combination must be unique, check if cache already exists
	ok, err := h.srv.Exist(req.Key, req.Version)
	if err != nil {
		h.sendJSON(w, http.StatusInternalServerError, err)
		return
	}

	// return 400 if cache already exists
	if ok {
		h.sendJSON(w, http.StatusBadRequest, err)
		return
	}

	// reserve cache
	id, err := h.srv.Reserve(req.Key, req.Version, req.CacheSize)
	if err != nil {
		log.Errorf("Failed to reserve cache", "error", err)
		h.sendJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendJSON(w, http.StatusOK, &ReserveCacheResponse{CacheID: id})
}

func (h *handler) HandleUploadCache(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	cacheID := params.ByName("cacheID")

	id, err := strconv.Atoi(cacheID)
	if err != nil {
		log.Errorf("Invalid cache ID", "error", err, "cacheID", cacheID)
		h.sendJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	offset := 0

	contentRange := r.Header.Get("Content-Range")
	if contentRange != "" && !strings.HasPrefix(contentRange, "bytes 0-") {
		rangeStart, err := strconv.Atoi(strings.Split(contentRange, "-")[0][6:])
		if err != nil {
			log.Errorf("Invalid content range", "error", err, "contentRange", contentRange)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		offset = rangeStart
	}

	err = h.srv.Upload(id, offset, r.Body)
	if err != nil {
		h.sendJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) HandleDownloadArtifact(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	artifactID := params.ByName("artifactID")

	id, err := strconv.Atoi(artifactID)
	if err != nil {
		log.Errorf("Invalid artifact ID", "error", err, "artifactID", artifactID)
		h.sendJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	path, err := h.srv.GetFilePath(id)
	if err != nil {
		h.sendJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.ServeFile(w, r, path)
}

func (h *handler) HandleCommitCache(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	cacheID := params.ByName("cacheID")

	id, err := strconv.Atoi(cacheID)
	if err != nil {
		log.Errorf("Invalid cache ID", "error", err, "cacheID", cacheID)
		h.sendJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.srv.Commit(id)
	if err != nil {
		h.sendJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
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

		log.Debugf("Http request", keyvals...)

		next(w, r, params)
	}
}
