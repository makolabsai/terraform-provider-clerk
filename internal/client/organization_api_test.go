package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2/organization"
)

func TestCreateOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/organizations" {
			t.Errorf("expected /v1/organizations, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk_test_dev" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		resp := map[string]any{
			"object":                  "organization",
			"id":                      "org_test123",
			"name":                    "Acme Corp",
			"slug":                    "acme-corp",
			"max_allowed_memberships": 50,
			"admin_delete_enabled":    false,
			"created_at":              1700000000000,
			"updated_at":              1700000000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newBackendTestClient(t, server, "app_1", "development", "sk_test_dev")

	name := "Acme Corp"
	maxMembers := int64(50)
	result, err := c.CreateOrganization(context.Background(), "app_1", "development", &organization.CreateParams{
		Name:                  &name,
		MaxAllowedMemberships: &maxMembers,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "org_test123" {
		t.Errorf("expected org_test123, got %s", result.ID)
	}
	if result.Name != "Acme Corp" {
		t.Errorf("expected Acme Corp, got %s", result.Name)
	}
}

func TestGetOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/organizations/org_test123" {
			t.Errorf("expected /v1/organizations/org_test123, got %s", r.URL.Path)
		}

		resp := map[string]any{
			"object":                  "organization",
			"id":                      "org_test123",
			"name":                    "Acme Corp",
			"slug":                    "acme-corp",
			"max_allowed_memberships": 50,
			"admin_delete_enabled":    false,
			"created_at":              1700000000000,
			"updated_at":              1700000000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newBackendTestClient(t, server, "app_1", "development", "sk_test_dev")

	result, err := c.GetOrganization(context.Background(), "app_1", "development", "org_test123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Slug != "acme-corp" {
		t.Errorf("expected acme-corp, got %s", result.Slug)
	}
}

func TestUpdateOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/v1/organizations/org_test123" {
			t.Errorf("expected /v1/organizations/org_test123, got %s", r.URL.Path)
		}

		resp := map[string]any{
			"object":                  "organization",
			"id":                      "org_test123",
			"name":                    "Acme Corp Updated",
			"slug":                    "acme-corp",
			"max_allowed_memberships": 100,
			"admin_delete_enabled":    true,
			"created_at":              1700000000000,
			"updated_at":              1700001000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newBackendTestClient(t, server, "app_1", "development", "sk_test_dev")

	name := "Acme Corp Updated"
	result, err := c.UpdateOrganization(context.Background(), "app_1", "development", "org_test123", &organization.UpdateParams{
		Name: &name,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Acme Corp Updated" {
		t.Errorf("expected Acme Corp Updated, got %s", result.Name)
	}
}

func TestDeleteOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/organizations/org_test123" {
			t.Errorf("expected /v1/organizations/org_test123, got %s", r.URL.Path)
		}

		resp := map[string]any{
			"object":  "organization",
			"id":      "org_test123",
			"deleted": true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newBackendTestClient(t, server, "app_1", "development", "sk_test_dev")

	result, err := c.DeleteOrganization(context.Background(), "app_1", "development", "org_test123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Deleted {
		t.Error("expected deleted=true")
	}
}

func TestCreateOrganization_NotRegistered(t *testing.T) {
	c := NewClerkClient("platform-key")

	name := "Test"
	_, err := c.CreateOrganization(context.Background(), "app_unknown", "development", &organization.CreateParams{
		Name: &name,
	})
	if err == nil {
		t.Fatal("expected error for unregistered backend client")
	}
}

func TestGetOrganization_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{
				{
					"code":    "resource_not_found",
					"message": "Organization not found",
				},
			},
		})
	}))
	defer server.Close()

	c := newBackendTestClient(t, server, "app_1", "development", "sk_test_dev")

	_, err := c.GetOrganization(context.Background(), "app_1", "development", "org_nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}
