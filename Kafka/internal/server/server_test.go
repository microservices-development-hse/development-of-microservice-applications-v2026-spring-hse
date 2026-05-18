package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockProducer struct {
	err error
}

func (m *mockProducer) SendImportRequest(ctx context.Context, projectKey string) error {
	return m.err
}

func newTestServer(p producerIface) *Server {
	return New(0, p)
}

func TestHandleImport_MethodNotAllowed(t *testing.T) {
	s := newTestServer(&mockProducer{})

	req := httptest.NewRequest(http.MethodGet, "/import", nil)
	w := httptest.NewRecorder()

	s.handleImport(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleImport_InvalidJSON(t *testing.T) {
	s := newTestServer(&mockProducer{})

	req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer([]byte("bad json")))
	w := httptest.NewRecorder()

	s.handleImport(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleImport_EmptyKey(t *testing.T) {
	s := newTestServer(&mockProducer{})

	body := `{"project_key":""}`
	req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer([]byte(body)))
	w := httptest.NewRecorder()

	s.handleImport(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleImport_ProducerError(t *testing.T) {
	s := newTestServer(&mockProducer{err: errors.New("fail")})

	body := `{"project_key":"TEST"}`
	req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer([]byte(body)))
	w := httptest.NewRecorder()

	s.handleImport(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandleImport_Success(t *testing.T) {
	s := newTestServer(&mockProducer{})

	body := `{"project_key":"TEST"}`
	req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer([]byte(body)))
	w := httptest.NewRecorder()

	s.handleImport(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if resp["status"] != "import queued" {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestHandleHealth(t *testing.T) {
	s := newTestServer(&mockProducer{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
