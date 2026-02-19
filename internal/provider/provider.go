package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/makolabsai/terraform-provider-clerk/internal/client"
	"github.com/makolabsai/terraform-provider-clerk/internal/datasources"
	"github.com/makolabsai/terraform-provider-clerk/internal/resources"
)

var _ provider.Provider = (*ClerkProvider)(nil)

// ClerkProvider implements the Terraform provider for Clerk.
type ClerkProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance tests.
	version string
}

// ClerkProviderModel describes the provider configuration data model.
type ClerkProviderModel struct {
	PlatformAPIKey types.String `tfsdk:"platform_api_key"`
}

// New returns a function that creates a new instance of the Clerk provider.
// This is the constructor used by main.go and test helpers.
func New() func() provider.Provider {
	return func() provider.Provider {
		return &ClerkProvider{
			version: "dev",
		}
	}
}

// NewWithVersion returns a provider constructor with a specific version string.
func NewWithVersion(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ClerkProvider{
			version: version,
		}
	}
}

func (p *ClerkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "clerk"
	resp.Version = p.version
}

func (p *ClerkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Clerk provider enables managing Clerk authentication platform resources via Terraform.",
		Attributes: map[string]schema.Attribute{
			"platform_api_key": schema.StringAttribute{
				Description: "The Clerk Platform API key for workspace-level operations. " +
					"Can also be set via the CLERK_PLATFORM_API_KEY environment variable.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *ClerkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ClerkProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve the Platform API key: config takes precedence over env var.
	platformAPIKey := os.Getenv("CLERK_PLATFORM_API_KEY")
	if !data.PlatformAPIKey.IsNull() && !data.PlatformAPIKey.IsUnknown() {
		platformAPIKey = data.PlatformAPIKey.ValueString()
	}

	if platformAPIKey == "" {
		resp.Diagnostics.AddError(
			"Missing Platform API Key",
			"The Clerk Platform API key must be set in the provider configuration "+
				"block (platform_api_key) or via the CLERK_PLATFORM_API_KEY environment variable.",
		)
		return
	}

	clerkClient := client.NewClerkClient(platformAPIKey)

	resp.DataSourceData = clerkClient
	resp.ResourceData = clerkClient
}

func (p *ClerkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewApplicationResource,
	}
}

func (p *ClerkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewApplicationDataSource,
	}
}
