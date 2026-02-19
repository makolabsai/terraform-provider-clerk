package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/makolabsai/terraform-provider-clerk/internal/client"
)

var (
	_ datasource.DataSource = (*ApplicationDataSource)(nil)
)

// ApplicationDataSource reads a Clerk application via the Platform API.
type ApplicationDataSource struct {
	client *client.ClerkClient
}

// ApplicationDataSourceModel describes the Terraform data source model.
type ApplicationDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	DevInstanceID      types.String `tfsdk:"dev_instance_id"`
	DevSecretKey       types.String `tfsdk:"dev_secret_key"`
	DevPublishableKey  types.String `tfsdk:"dev_publishable_key"`
	ProdInstanceID     types.String `tfsdk:"prod_instance_id"`
	ProdSecretKey      types.String `tfsdk:"prod_secret_key"`
	ProdPublishableKey types.String `tfsdk:"prod_publishable_key"`
}

func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

func (d *ApplicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing Clerk application by its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the Clerk application.",
				Required:    true,
			},
			"dev_instance_id": schema.StringAttribute{
				Description: "The instance ID for the development environment.",
				Computed:    true,
			},
			"dev_secret_key": schema.StringAttribute{
				Description: "The secret key for the development instance.",
				Computed:    true,
				Sensitive:   true,
			},
			"dev_publishable_key": schema.StringAttribute{
				Description: "The publishable key for the development instance.",
				Computed:    true,
			},
			"prod_instance_id": schema.StringAttribute{
				Description: "The instance ID for the production environment.",
				Computed:    true,
			},
			"prod_secret_key": schema.StringAttribute{
				Description: "The secret key for the production instance.",
				Computed:    true,
				Sensitive:   true,
			},
			"prod_publishable_key": schema.StringAttribute{
				Description: "The publishable key for the production instance.",
				Computed:    true,
			},
		},
	}
}

func (d *ApplicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clerkClient, ok := req.ProviderData.(*client.ClerkClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ClerkClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = clerkClient
}

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	application, err := d.client.GetApplication(ctx, data.ID.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Clerk application", err.Error())
		return
	}

	for _, inst := range application.Instances {
		switch inst.EnvironmentType {
		case "development":
			data.DevInstanceID = types.StringValue(inst.InstanceID)
			data.DevPublishableKey = types.StringValue(inst.PublishableKey)
			if inst.SecretKey != "" {
				data.DevSecretKey = types.StringValue(inst.SecretKey)
			}
		case "production":
			data.ProdInstanceID = types.StringValue(inst.InstanceID)
			data.ProdPublishableKey = types.StringValue(inst.PublishableKey)
			if inst.SecretKey != "" {
				data.ProdSecretKey = types.StringValue(inst.SecretKey)
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
