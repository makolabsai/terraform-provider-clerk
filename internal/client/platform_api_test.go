package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/platform/applications" {
			t.Errorf("expected /v1/platform/applications, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}

		var req PlatformCreateApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if req.Name != "test-app" {
			t.Errorf("expected name test-app, got %s", req.Name)
		}

		resp := PlatformApplicationResponse{
			ApplicationID: "app_123",
			Instances: []PlatformApplicationInstance{
				{
					InstanceID:      "ins_dev",
					EnvironmentType: "development",
					PublishableKey:  "pk_test_dev",
					SecretKey:       "sk_test_dev",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server, "test-key")

	result, err := c.CreateApplication(context.Background(), PlatformCreateApplicationRequest{
		Name: "test-app",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ApplicationID != "app_123" {
		t.Errorf("expected app_123, got %s", result.ApplicationID)
	}
	if len(result.Instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(result.Instances))
	}
	if result.Instances[0].SecretKey != "sk_test_dev" {
		t.Errorf("expected sk_test_dev, got %s", result.Instances[0].SecretKey)
	}
}

func TestGetApplication_WithSecretKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("include_secret_keys") != "true" {
			t.Error("expected include_secret_keys=true query param")
		}

		resp := PlatformApplicationResponse{
			ApplicationID: "app_123",
			Instances: []PlatformApplicationInstance{
				{InstanceID: "ins_dev", EnvironmentType: "development", PublishableKey: "pk_dev", SecretKey: "sk_dev"},
				{InstanceID: "ins_prod", EnvironmentType: "production", PublishableKey: "pk_prod", SecretKey: "sk_prod"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server, "test-key")

	result, err := c.GetApplication(context.Background(), "app_123", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(result.Instances))
	}
}

func TestGetApplication_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	c := newTestClient(server, "test-key")

	_, err := c.GetApplication(context.Background(), "app_missing", false)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	apiErr, ok := err.(*PlatformAPIError)
	if !ok {
		t.Fatalf("expected *PlatformAPIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestDeleteApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		resp := PlatformDeletedObjectResponse{Deleted: true, Object: "application", ID: "app_123"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server, "test-key")

	err := c.DeleteApplication(context.Background(), "app_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		var req PlatformUpdateApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if req.Name != "updated-name" {
			t.Errorf("expected updated-name, got %s", req.Name)
		}

		resp := PlatformApplicationResponse{ApplicationID: "app_123"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server, "test-key")

	result, err := c.UpdateApplication(context.Background(), "app_123", PlatformUpdateApplicationRequest{Name: "updated-name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ApplicationID != "app_123" {
		t.Errorf("expected app_123, got %s", result.ApplicationID)
	}
}

// newTestClient creates a ClerkClient that points at the given test server.
func newTestClient(server *httptest.Server, apiKey string) *ClerkClient {
	c := NewClerkClient(apiKey)
	// Override the base URL by replacing the platformRequest method's target.
	// Since platformRequest uses the package-level const, we need a different approach:
	// use a custom HTTP client that rewrites URLs.
	c.PlatformHTTPClient = &http.Client{
		Transport: &rewriteTransport{
			base:    http.DefaultTransport,
			baseURL: server.URL,
		},
	}
	return c
}

// rewriteTransport rewrites the request URL to point at the test server.
type rewriteTransport struct {
	base    http.RoundTripper
	baseURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.baseURL[len("http://"):]
	return t.base.RoundTrip(req)
}
