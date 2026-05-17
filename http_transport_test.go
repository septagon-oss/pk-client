package client

// http_transport_test.go validates HTTP transport request construction and
// response handling.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (every Go file declares its purpose).

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPTransportListBuildsQueryAndHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/widgets" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" || r.URL.Query().Get("page_size") != "25" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing bearer token")
		}
		_ = json.NewEncoder(w).Encode(ListResponse[map[string]string]{
			Data:     []map[string]string{{"id": "one"}},
			Metadata: &ListMetadata{Page: 2, PageSize: 25, TotalCount: 1, TotalPages: 1},
		})
	}))
	defer server.Close()

	c, err := NewHTTP[map[string]string](NewHTTPConfig(server.URL, "/api/widgets", WithBearerToken("token")))
	if err != nil {
		t.Fatalf("NewHTTP failed: %v", err)
	}

	got, err := c.List(context.Background(), ListParams{Page: 2, PageSize: 25})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(got.Data) != 1 || got.Data[0]["id"] != "one" {
		t.Fatalf("unexpected response: %#v", got.Data)
	}
}

func TestHTTPTransportCreatePostsWrappedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input CreateInput[map[string]string]
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if input.Body["name"] != "demo" {
			t.Fatalf("unexpected body: %#v", input.Body)
		}
		_ = json.NewEncoder(w).Encode(ItemResponse[map[string]string]{Data: map[string]string{"id": "created"}})
	}))
	defer server.Close()

	c, err := NewHTTP[map[string]string](NewHTTPConfig(server.URL, "widgets"))
	if err != nil {
		t.Fatalf("NewHTTP failed: %v", err)
	}

	got, err := c.Create(context.Background(), CreateInput[map[string]string]{Body: map[string]string{"name": "demo"}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if got.Data["id"] != "created" {
		t.Fatalf("unexpected response: %#v", got.Data)
	}
}

func TestHTTPTransportReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"nope"}`, http.StatusTeapot)
	}))
	defer server.Close()

	c, err := NewHTTP[map[string]string](NewHTTPConfig(server.URL, "widgets"))
	if err != nil {
		t.Fatalf("NewHTTP failed: %v", err)
	}

	_, err = c.GetByID(context.Background(), "one")
	if err == nil {
		t.Fatal("expected API error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusTeapot {
		t.Fatalf("status = %d", apiErr.StatusCode)
	}
}

func TestHTTPTransportCopiesConfigAtConstruction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Mode"); got != "initial" {
			t.Fatalf("X-Mode = %q, want initial", got)
		}
		if got := r.URL.Query().Get("tenant"); got != "oss" {
			t.Fatalf("tenant query = %q, want oss", got)
		}
		_ = json.NewEncoder(w).Encode(ItemResponse[map[string]string]{Data: map[string]string{"id": "one"}})
	}))
	defer server.Close()

	config := NewHTTPConfig(server.URL+"/", "/api/widgets/", WithHeader("X-Mode", "initial"), WithQueryParam("tenant", "oss"))
	c, err := NewHTTP[map[string]string](config)
	if err != nil {
		t.Fatalf("NewHTTP failed: %v", err)
	}
	config.Headers["X-Mode"] = "mutated"
	config.QueryParams["tenant"] = "mutated"

	if _, err := c.GetByID(context.Background(), "one"); err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
}

func TestHTTPTransportImportSendsRawPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/widgets/import" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("format") != "csv" {
			t.Fatalf("format query = %q", r.URL.Query().Get("format"))
		}
		if got := r.Header.Get("Content-Type"); got != "text/csv" {
			t.Fatalf("Content-Type = %q, want text/csv", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if string(body) != "id,name\n1,demo\n" {
			t.Fatalf("body = %q", string(body))
		}
		_ = json.NewEncoder(w).Encode(ImportResponse{Imported: 1})
	}))
	defer server.Close()

	c, err := NewHTTP[map[string]string](NewHTTPConfig(server.URL, "api/widgets"))
	if err != nil {
		t.Fatalf("NewHTTP failed: %v", err)
	}
	result, err := c.Import(context.Background(), []byte("id,name\n1,demo\n"), "csv")
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if result.Imported != 1 {
		t.Fatalf("imported = %d, want 1", result.Imported)
	}
}

func TestHTTPConfigValidateRejectsInvalidBaseURL(t *testing.T) {
	_, err := NewHTTP[map[string]string](NewHTTPConfig("://bad", "widgets"))
	if err == nil {
		t.Fatal("expected invalid base URL to fail")
	}
}
