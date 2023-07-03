package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

var _ Service = new(mockService)

type mockService struct{}

func (m *mockService) CreateArtifactInNameContainer(runID string) (string, error) {
	return "testContainerID", nil
}

func (m *mockService) PatchArtifactSize(runID string) {
	// do nothing
}

func (m *mockService) ListArtifacts(runID string) (string, []string, error) {
	return "testContainerID", []string{"file1.txt", "file2.txt"}, nil
}

func (m *mockService) UploadArtifactToFileContainer(containerID string, itemPath string, content string) error {
	return nil
}

func (m *mockService) GetContainerItems(containerID string, itemPath string) ([]string, error) {
	return []string{"foo", "bar"}, nil
}

func (m *mockService) DownloadSingleArtifact(path string) (string, error) {
	return "test content", nil
}

func TestHandler_HandleCreateArtifactInNameContainer(t *testing.T) {
	handler := &handler{srv: &mockService{}}
	router := httprouter.New()
	router.POST("/_apis/pipelines/workflows/:runID/artifacts", handler.HandleCreateArtifactInNameContainer)

	req, err := http.NewRequest("POST", "/_apis/pipelines/workflows/123/artifacts", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the host to test the url in the response body
	req.Host = "example.com"

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rr.Code)
	}

	expectedResponse := `{"fileContainerResourceUrl":"http://example.com/upload/testContainerID","name":""}`
	actualResponse := strings.TrimSpace(rr.Body.String())

	if actualResponse != expectedResponse {
		t.Errorf("Expected response body %q, but got %q", expectedResponse, actualResponse)
	}
}

func TestHandler_HandleListArtifacts(t *testing.T) {
	handler := &handler{srv: &mockService{}}
	router := httprouter.New()
	router.GET("/_apis/pipelines/workflows/:runID/artifacts", handler.HandleListArtifacts)

	req, err := http.NewRequest("GET", "/_apis/pipelines/workflows/123/artifacts", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the host to test the url in the response body
	req.Host = "example.com"

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rr.Code)
	}

	expectedResponse := `{"count":2,"value":[{"fileContainerResourceUrl":"http://example.com/download/testContainerID","name":"file1.txt"},{"fileContainerResourceUrl":"http://example.com/download/testContainerID","name":"file2.txt"}]}`
	actualResponse := strings.TrimSpace(rr.Body.String())

	if actualResponse != expectedResponse {
		t.Errorf("Unexpected response body. Expected %s, but got %s", expectedResponse, actualResponse)
	}
}

func TestHandler_HandleUploadArtifactToFileContainer(t *testing.T) {
	handler := &handler{srv: &mockService{}}
	router := httprouter.New()
	router.PUT("/upload/:containerID", handler.HandleUploadArtifactToFileContainer)

	body := bytes.NewBufferString("test content")
	req, err := http.NewRequest("PUT", "/upload/testContainerID?itemPath=test.txt", body)
	if err != nil {
		t.Fatal(err)
	}

	// Set the host to test the url in the response body
	req.Host = "example.com"

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rr.Code)
	}
}

func TestHandler_HandleGetContainerItems(t *testing.T) {
	handler := &handler{srv: &mockService{}}
	router := httprouter.New()
	router.GET("/download/:containerID", handler.HandleGetContainerItems)

	req, err := http.NewRequest("GET", "/download/testContainerID?itemPath=foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the host to test the url in the response body
	req.Host = "example.com"

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rr.Code)
	}

	expectedResponse := `{"count":2,"value":[{"path":"foo","itemType":"file","contentLocation":"http://example.com/artifact/testContainerID/foo"},{"path":"bar","itemType":"file","contentLocation":"http://example.com/artifact/testContainerID/bar"}]}`
	actualResponse := strings.TrimSpace(rr.Body.String())

	if actualResponse != expectedResponse {
		t.Errorf("Expected response body %q, but got %q", expectedResponse, actualResponse)
	}
}

func TestHandler_HandleDownloadSingleArtifact(t *testing.T) {
	handler := &handler{srv: &mockService{}}
	router := httprouter.New()
	router.GET("/artifact/*path", handler.HandleDownloadSingleArtifact)

	req, err := http.NewRequest("GET", "/artifact/file1.txt", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the host to test the url in the response body
	req.Host = "example.com"

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rr.Code)
	}

	expectedResponse := "test content"
	actualResponse := strings.TrimSpace(rr.Body.String())

	if actualResponse != expectedResponse {
		t.Errorf("Expected response body %q, but got %q", expectedResponse, actualResponse)
	}
}
