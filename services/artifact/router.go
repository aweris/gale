package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

// ArtifactResponse represents the response of the artifact creation. Most of the fields are not used, Only used fields are
// FileContainerResourceURL and Name. Rest of the fields are ignored.
// Source: https://github.com/actions/toolkit/blob/main/packages/artifact/src/internal/contracts.ts#L1
type ArtifactResponse struct {
	FileContainerResourceURL string `json:"fileContainerResourceUrl"`
	Name                     string `json:"name"`
}

// ListArtifactsResponse represents the response of the artifact listing.
// Source: https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/artifact/src/internal/contracts.ts#L49C18-L52
type ListArtifactsResponse struct {
	Count int                `json:"count"`
	Value []ArtifactResponse `json:"value"`
}

// ContainerEntry represents a single entry in the container. Only used fields are Path and ItemType. Rest of the fields
// are ignored.
// Source: https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/artifact/src/internal/contracts.ts#L59C18-L76
type ContainerEntry struct {
	Path            string `json:"path"`
	ItemType        string `json:"itemType"`
	ContentLocation string `json:"contentLocation"`
}

// QueryArtifactResponse represents the response of the artifact query.
// Source: https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/artifact/src/internal/contracts.ts#L54C18-L57
type QueryArtifactResponse struct {
	Count int              `json:"count"`
	Value []ContainerEntry `json:"value"`
}

// Serve starts the artifact service router on the given port
func Serve(port string, srv Service) error {
	router := httprouter.New()

	handler := &handler{srv: srv}

	router.POST("/_apis/pipelines/workflows/:runID/artifacts", handler.HandleCreateArtifactInNameContainer)
	router.PATCH("/_apis/pipelines/workflows/:runID/artifacts", handler.HandlePatchArtifactSize)
	router.GET("/_apis/pipelines/workflows/:runID/artifacts", handler.HandleListArtifacts)
	router.PUT("/upload/:containerID", handler.HandleUploadArtifactToFileContainer)
	router.GET("/download/:containerID", handler.HandleGetContainerItems)
	router.GET("/artifact/*path", handler.HandleDownloadSingleArtifact)
	router.GET("/healthz", handler.HandleHealthz)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server.ListenAndServe()
}

const (
	EncodingGzip = "gzip"
	ExtGzip      = ".gz"
)

type handler struct {
	srv Service
}

func (h *handler) HandleCreateArtifactInNameContainer(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	runID := params.ByName("runID")

	containerID, err := h.srv.CreateArtifactInNameContainer(runID)
	if err != nil {
		fmt.Printf("Error creating artifact container: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	containerResourceURL := fmt.Sprintf("http://%s/upload/%s", r.Host, containerID)

	// Upload client only interested in this field, others are not used
	// https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/artifact/src/internal/artifact-client.ts#L99-L121
	h.sendJSON(w, http.StatusOK, ArtifactResponse{FileContainerResourceURL: containerResourceURL})
}

func (h *handler) HandlePatchArtifactSize(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	runID := params.ByName("runID")

	h.srv.PatchArtifactSize(runID)

	w.WriteHeader(http.StatusOK)
}

func (h *handler) HandleListArtifacts(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	runID := params.ByName("runID")

	containerID, entries, err := h.srv.ListArtifacts(runID)
	if err != nil {
		fmt.Printf("Error listing artifacts: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// pre-allocate the slice to avoid reallocation
	artifacts := make([]ArtifactResponse, 0, len(entries))

	// Download client only interested in these fields, others are not used
	// source: https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/artifact/src/internal/artifact-client.ts#L164-L181

	for _, entry := range entries {
		resource := ArtifactResponse{
			Name:                     entry,
			FileContainerResourceURL: fmt.Sprintf("http://%s/download/%s", r.Host, containerID),
		}

		artifacts = append(artifacts, resource)
	}

	h.sendJSON(w, http.StatusOK, ListArtifactsResponse{Count: len(artifacts), Value: artifacts})
}

func (h *handler) HandleUploadArtifactToFileContainer(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	containerID := params.ByName("containerID")
	itemPath := r.URL.Query().Get("itemPath")

	// If the request is gzipped, add the .gz extension to the item path
	if r.Header.Get("Content-Encoding") == EncodingGzip {
		itemPath += ExtGzip
	}

	offset := 0

	contentRange := r.Header.Get("Content-Range")
	if contentRange != "" && !strings.HasPrefix(contentRange, "bytes 0-") {
		rangeStart, err := strconv.Atoi(strings.Split(contentRange, "-")[0][6:])
		if err != nil {
			fmt.Printf("Error parsing content range: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		offset = rangeStart
	}

	h.srv.UploadArtifactToFileContainer(containerID, itemPath, offset, r.Body)

	w.WriteHeader(http.StatusOK)
}

func (h *handler) HandleGetContainerItems(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	containerID := params.ByName("containerID")
	itemPath := r.URL.Query().Get("itemPath")

	items, err := h.srv.GetContainerItems(containerID, itemPath)
	if err != nil {
		fmt.Printf("Error getting container items: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// pre-allocate the slice to avoid reallocation
	files := make([]ContainerEntry, 0, len(items))

	for _, item := range items {
		// Remove the .gz extension from the item path while listing container items
		item = strings.TrimSuffix(item, ExtGzip)

		files = append(files, ContainerEntry{
			Path:            item,
			ItemType:        "file",
			ContentLocation: fmt.Sprintf("http://%s/artifact/%s/%s", r.Host, containerID, item),
		})
	}

	h.sendJSON(w, http.StatusOK, QueryArtifactResponse{Count: len(files), Value: files})
}

func (h *handler) HandleDownloadSingleArtifact(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	path := params.ByName("path")[1:]

	// look for the non-gzipped version first
	content, err := h.srv.DownloadSingleArtifact(path)
	if err != nil {
		// If the file is not found, try to download the gzipped version
		content, err = h.srv.DownloadSingleArtifact(path + ExtGzip)
		if err != nil {
			fmt.Printf("Error downloading single artifact: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Encoding", EncodingGzip)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

func (h *handler) sendJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func (h *handler) HandleHealthz(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}
