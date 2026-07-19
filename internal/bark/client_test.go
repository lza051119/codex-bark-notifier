package bark

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestClientSendsEncodedCriticalNotification(t *testing.T) {
	requests := make(chan *http.Request, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", "device/key")
	if err := client.Send(context.Background(), "Codex / demo", "结果：完成 & 检查"); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	r := <-requests
	if !strings.HasPrefix(r.URL.Path, "/device%2Fkey/") && !strings.HasPrefix(r.RequestURI, "/device%2Fkey/") {
		t.Fatalf("device key was not safely encoded in request: %s", r.RequestURI)
	}
	if got := r.URL.Query().Get("level"); got != "critical" {
		t.Fatalf("level = %q, want critical", got)
	}
	if got := r.URL.Query().Get("group"); got != "codex" {
		t.Fatalf("group = %q, want codex", got)
	}
	if _, err := url.QueryUnescape(r.URL.Path); err != nil {
		t.Fatalf("request path is not valid URL encoding: %v", err)
	}
}

func TestClientReturnsErrorForNonSuccessResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL, "device-key")
	if err := client.Send(context.Background(), "title", "body"); err == nil {
		t.Fatal("Send returned nil for non-success response")
	}
}
