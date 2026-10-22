package dxapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRelationEdgeExistsPaginates(t *testing.T) {
	var cursors []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/catalog.relationEdges.list" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("entity_identifier") != "src" || q.Get("relation_identifier") != "rel" || q.Get("direction") != "outgoing" || q.Get("limit") != "100" {
			t.Fatalf("unexpected query %s", r.URL.RawQuery)
		}
		cursors = append(cursors, q.Get("cursor"))

		switch q.Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"ok":true,"entity_relations":[{"identifier":"a","name":"A"}],"response_metadata":{"next_cursor":"page2"}}`))
		case "page2":
			_, _ = w.Write([]byte(`{"ok":true,"entity_relations":[{"identifier":"b","name":"B"}],"response_metadata":{"next_cursor":null}}`))
		default:
			t.Fatalf("unexpected cursor %q", q.Get("cursor"))
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "test")

	exists, err := client.RelationEdgeExists(context.Background(), "rel", "src", "b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !exists {
		t.Fatal("expected edge to exist")
	}
	if len(cursors) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(cursors))
	}

	cursors = nil
	exists, err = client.RelationEdgeExists(context.Background(), "rel", "src", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if exists {
		t.Fatal("expected edge not to exist")
	}
	if len(cursors) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(cursors))
	}
}

func TestRelationEdgeExistsNotFound(t *testing.T) {
	for _, code := range []string{"entity_not_found", "relation_not_found"} {
		t.Run(code, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"ok":false,"error":"` + code + `"}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "token", "test")
			exists, err := client.RelationEdgeExists(context.Background(), "rel", "src", "tgt")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if exists {
				t.Fatal("expected edge not to exist")
			}
		})
	}
}

func TestRelationEdgeExistsOtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_auth"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "test")
	_, err := client.RelationEdgeExists(context.Background(), "rel", "src", "tgt")

	var apiErr *RelationEdgeAPIError
	if !errors.As(err, &apiErr) || apiErr.Code != "invalid_auth" {
		t.Fatalf("expected invalid_auth API error, got %v", err)
	}
}

func TestAddRelationEdge(t *testing.T) {
	var got map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/catalog.relationEdges.add" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decoding body: %s", err)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "test")
	if err := client.AddRelationEdge(context.Background(), "rel", "src", "tgt"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	want := map[string]string{
		"relation_identifier":      "rel",
		"source_entity_identifier": "src",
		"target_entity_identifier": "tgt",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

func TestAddRelationEdgeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_cardinality"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "test")
	err := client.AddRelationEdge(context.Background(), "rel", "src", "tgt")

	var apiErr *RelationEdgeAPIError
	if !errors.As(err, &apiErr) || apiErr.Code != "invalid_cardinality" {
		t.Fatalf("expected invalid_cardinality API error, got %v", err)
	}
}

func TestRemoveRelationEdgeTreatsNotFoundAsSuccess(t *testing.T) {
	for _, code := range []string{"edge_not_found", "entity_not_found", "relation_not_found"} {
		t.Run(code, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/catalog.relationEdges.remove" {
					t.Fatalf("unexpected path %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"ok":false,"error":"` + code + `"}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "token", "test")
			if err := client.RemoveRelationEdge(context.Background(), "rel", "src", "tgt"); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestRemoveRelationEdgeOtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_entity_type"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "test")
	if err := client.RemoveRelationEdge(context.Background(), "rel", "src", "tgt"); err == nil {
		t.Fatal("expected error")
	}
}
