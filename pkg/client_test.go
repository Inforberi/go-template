package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientDoJSON(t *testing.T) {
	t.Run("sends JSON headers and decodes response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != "/items" {
				t.Fatalf("request = %s %s", r.Method, r.URL.Path)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer token" {
				t.Fatalf("Authorization = %q", got)
			}
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q", got)
			}

			var request struct{ Name string }
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}
			if request.Name != "item" {
				t.Fatalf("request name = %q", request.Name)
			}

			_ = json.NewEncoder(w).Encode(struct{ ID int }{ID: 42})
		}))
		defer server.Close()

		headers := make(http.Header)
		headers.Set("Authorization", "Bearer token")
		client, err := NewClient(server.URL, headers, nil)
		if err != nil {
			t.Fatal(err)
		}

		var response struct{ ID int }
		if err := client.Post(context.Background(), "/items", struct{ Name string }{Name: "item"}, &response); err != nil {
			t.Fatal(err)
		}
		if response.ID != 42 {
			t.Fatalf("response ID = %d", response.ID)
		}
	})

	t.Run("accepts no content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, err := NewClient(server.URL, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := client.Delete(context.Background(), "/items/42", &struct{}{}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("returns HTTP errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		}))
		defer server.Close()

		client, err := NewClient(server.URL, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		err = client.Get(context.Background(), "/missing", nil)
		if err == nil || !strings.Contains(err.Error(), "status=404") {
			t.Fatalf("Get() error = %v", err)
		}
	})

	t.Run("returns JSON decoding errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("not JSON"))
		}))
		defer server.Close()

		client, err := NewClient(server.URL, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		err = client.Get(context.Background(), "/items", &struct{}{})
		if err == nil || !strings.Contains(err.Error(), "decode response body") {
			t.Fatalf("Get() error = %v", err)
		}
	})
}

func TestClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client, err := NewClient(server.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = client.Get(ctx, "/items", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Get() error = %v, want context cancellation", err)
	}
}

func TestNewClientRejectsInvalidBaseURL(t *testing.T) {
	for _, baseURL := range []string{"", "example.com", "ftp://example.com"} {
		t.Run(baseURL, func(t *testing.T) {
			if _, err := NewClient(baseURL, nil, nil); err == nil {
				t.Fatalf("NewClient(%q) error = nil", baseURL)
			}
		})
	}
}
