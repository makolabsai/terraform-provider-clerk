package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/instancesettings"
)

func TestGetInstanceSettingsClient_Success(t *testing.T) {
	c := NewClerkClient("platform-key")
	if err := c.RegisterBackendClient("app_1", "development", "sk_test_dev"); err != nil {
		t.Fatalf("unexpected error registering backend client: %v", err)
	}

	isClient, err := c.GetInstanceSettingsClient("app_1", "development")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isClient == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestGetInstanceSettingsClient_NotRegistered(t *testing.T) {
	c := NewClerkClient("platform-key")

	_, err := c.GetInstanceSettingsClient("app_unknown", "development")
	if err == nil {
		t.Fatal("expected error for unregistered backend client")
	}
}

func TestUpdateInstanceSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/v1/instance" {
			t.Errorf("expected /v1/instance, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk_test_dev" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if body["hibp"] != true {
			t.Errorf("expected hibp=true, got %v", body["hibp"])
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newBackendTestClient(t, server, "app_1", "development", "sk_test_dev")

	hibp := true
	err := c.UpdateInstanceSettings(context.Background(), "app_1", "development", &instancesettings.UpdateParams{
		HIBP: &hibp,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateInstanceSettings_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"errors":[{"message":"invalid value","long_message":"HIBP value is invalid","code":"form_param_value_invalid"}]}`))
	}))
	defer server.Close()

	c := newBackendTestClient(t, server, "app_1", "development", "sk_test_dev")

	hibp := true
	err := c.UpdateInstanceSettings(context.Background(), "app_1", "development", &instancesettings.UpdateParams{
		HIBP: &hibp,
	})
	if err == nil {
		t.Fatal("expected error for 422 response")
	}
}

func TestUpdateInstanceRestrictions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/v1/instance/restrictions" {
			t.Errorf("expected /v1/instance/restrictions, got %s", r.URL.Path)
		}

		resp := map[string]any{
			"object":                          "instance_restrictions",
			"allowlist":                       true,
			"blocklist":                       false,
			"block_email_subaddresses":        false,
			"block_disposable_email_domains":  false,
			"ignore_dots_for_gmail_addresses": false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newBackendTestClient(t, server, "app_1", "development", "sk_test_dev")

	allowlist := true
	result, err := c.UpdateInstanceRestrictions(context.Background(), "app_1", "development", &instancesettings.UpdateRestrictionsParams{
		Allowlist: &allowlist,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowlist {
		t.Error("expected allowlist=true")
	}
}

func TestUpdateOrganizationSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/v1/instance/organization_settings" {
			t.Errorf("expected /v1/instance/organization_settings, got %s", r.URL.Path)
		}

		resp := map[string]any{
			"object":                   "organization_settings",
			"enabled":                  true,
			"max_allowed_memberships":  10,
			"max_allowed_roles":        5,
			"max_allowed_permissions":  20,
			"creator_role":             "org:admin",
			"admin_delete_enabled":     true,
			"domains_enabled":          false,
			"domains_enrollment_modes": []string{},
			"domains_default_role":     "org:member",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newBackendTestClient(t, server, "app_1", "development", "sk_test_dev")

	enabled := true
	maxMemberships := int64(10)
	result, err := c.UpdateOrganizationSettings(context.Background(), "app_1", "development", &instancesettings.UpdateOrganizationSettingsParams{
		Enabled:               &enabled,
		MaxAllowedMemberships: &maxMemberships,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Enabled {
		t.Error("expected enabled=true")
	}
	if result.MaxAllowedMemberships != 10 {
		t.Errorf("expected max_allowed_memberships=10, got %d", result.MaxAllowedMemberships)
	}
}

// newBackendTestClient creates a ClerkClient with a registered backend client
// pointing at the test server.
func newBackendTestClient(t *testing.T, server *httptest.Server, appID, environment, secretKey string) *ClerkClient {
	t.Helper()
	c := NewClerkClient("platform-key")

	// Register the backend client with the secret key.
	config := &clerk.ClientConfig{}
	config.Key = clerk.String(secretKey)
	// Override the backend URL to point at our test server.
	config.URL = clerk.String(server.URL + "/v1/")

	c.mu.Lock()
	c.backendClients[backendClientKey(appID, environment)] = config
	c.mu.Unlock()

	return c
}
