package client

import (
	"context"
	"fmt"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/instancesettings"
)

// GetInstanceSettingsClient returns an instancesettings.Client configured for
// the given application and environment. The secret key is resolved from the
// internal backend client registry.
func (c *ClerkClient) GetInstanceSettingsClient(appID, environment string) (*instancesettings.Client, error) {
	config, err := c.GetBackendConfig(appID, environment)
	if err != nil {
		return nil, fmt.Errorf("resolving backend client for %s/%s: %w", appID, environment, err)
	}
	return instancesettings.NewClient(config), nil
}

// UpdateInstanceSettings updates the general settings of a Clerk instance.
func (c *ClerkClient) UpdateInstanceSettings(ctx context.Context, appID, environment string, params *instancesettings.UpdateParams) error {
	isClient, err := c.GetInstanceSettingsClient(appID, environment)
	if err != nil {
		return err
	}
	return isClient.Update(ctx, params)
}

// UpdateInstanceRestrictions updates the restriction settings of a Clerk instance.
func (c *ClerkClient) UpdateInstanceRestrictions(ctx context.Context, appID, environment string, params *instancesettings.UpdateRestrictionsParams) (*clerk.InstanceRestrictions, error) {
	isClient, err := c.GetInstanceSettingsClient(appID, environment)
	if err != nil {
		return nil, err
	}
	return isClient.UpdateRestrictions(ctx, params)
}

// UpdateOrganizationSettings updates the organization settings of a Clerk instance.
func (c *ClerkClient) UpdateOrganizationSettings(ctx context.Context, appID, environment string, params *instancesettings.UpdateOrganizationSettingsParams) (*clerk.OrganizationSettings, error) {
	isClient, err := c.GetInstanceSettingsClient(appID, environment)
	if err != nil {
		return nil, err
	}
	return isClient.UpdateOrganizationSettings(ctx, params)
}
