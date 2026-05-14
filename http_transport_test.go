package client

import (
	"context"
	"encoding/json"
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
