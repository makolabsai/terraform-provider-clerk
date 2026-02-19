package client

import (
	"context"
	"fmt"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/organization"
)

// CreateOrganization creates an organization in the specified application/environment.
func (c *ClerkClient) CreateOrganization(ctx context.Context, appID, environment string, params *organization.CreateParams) (*clerk.Organization, error) {
	config, err := c.GetBackendConfig(appID, environment)
	if err != nil {
		return nil, fmt.Errorf("resolving backend client for %s/%s: %w", appID, environment, err)
	}

	orgClient := organization.NewClient(config)
	return orgClient.Create(ctx, params)
}

// GetOrganization fetches an organization by ID or slug.
func (c *ClerkClient) GetOrganization(ctx context.Context, appID, environment, idOrSlug string) (*clerk.Organization, error) {
	config, err := c.GetBackendConfig(appID, environment)
	if err != nil {
		return nil, fmt.Errorf("resolving backend client for %s/%s: %w", appID, environment, err)
	}

	orgClient := organization.NewClient(config)
	return orgClient.Get(ctx, idOrSlug)
}

// UpdateOrganization updates an organization by ID.
func (c *ClerkClient) UpdateOrganization(ctx context.Context, appID, environment, id string, params *organization.UpdateParams) (*clerk.Organization, error) {
	config, err := c.GetBackendConfig(appID, environment)
	if err != nil {
		return nil, fmt.Errorf("resolving backend client for %s/%s: %w", appID, environment, err)
	}

	orgClient := organization.NewClient(config)
	return orgClient.Update(ctx, id, params)
}

// DeleteOrganization deletes an organization by ID.
func (c *ClerkClient) DeleteOrganization(ctx context.Context, appID, environment, id string) (*clerk.DeletedResource, error) {
	config, err := c.GetBackendConfig(appID, environment)
	if err != nil {
		return nil, fmt.Errorf("resolving backend client for %s/%s: %w", appID, environment, err)
	}

	orgClient := organization.NewClient(config)
	return orgClient.Delete(ctx, id)
}
