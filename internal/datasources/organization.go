package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/makolabsai/terraform-provider-clerk/internal/client"
)

var (
	_ datasource.DataSource = (*OrganizationDataSource)(nil)
)

// OrganizationDataSource reads a Clerk organization via the Backend API.
type OrganizationDataSource struct {
	client *client.ClerkClient
}

// OrganizationDataSourceModel describes the Terraform data source model.
type OrganizationDataSourceModel struct {
	ApplicationID         types.String `tfsdk:"application_id"`
	Environment           types.String `tfsdk:"environment"`
	ID                    types.String `tfsdk:"id"`
	Slug                  types.String `tfsdk:"slug"`
	Name                  types.String `tfsdk:"name"`
	MaxAllowedMemberships types.Int64  `tfsdk:"max_allowed_memberships"`
	AdminDeleteEnabled    types.Bool   `tfsdk:"admin_delete_enabled"`
	CreatedAt             types.Int64  `tfsdk:"created_at"`
}

func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
}

func (d *OrganizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *OrganizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing Clerk organization by ID or slug.",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Description: "The Clerk application ID the organization belongs to.",
				Required:    true,
			},
			"environment": schema.StringAttribute{
				Description: "The environment type: \"development\" or \"production\".",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("development", "production"),
				},
			},
			"id": schema.StringAttribute{
				Description: "The organization ID to look up. Exactly one of id or slug must be specified.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("slug")),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The organization slug to look up. Exactly one of id or slug must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the organization.",
				Computed:    true,
			},
			"max_allowed_memberships": schema.Int64Attribute{
				Description: "Maximum number of memberships allowed in the organization.",
				Computed:    true,
			},
			"admin_delete_enabled": schema.BoolAttribute{
				Description: "Whether organization admins can delete the organization.",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "Unix timestamp of when the organization was created.",
				Computed:    true,
			},
		},
	}
}

func (d *OrganizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.ApplicationID.ValueString()
	env := data.Environment.ValueString()

	// The Clerk SDK Get method accepts either an ID or slug.
	lookupKey := data.ID.ValueString()
	if lookupKey == "" {
		lookupKey = data.Slug.ValueString()
	}

	org, err := d.client.GetOrganization(ctx, appID, env, lookupKey)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Clerk organization", err.Error())
		return
	}

	data.ID = types.StringValue(org.ID)
	data.Name = types.StringValue(org.Name)
	data.Slug = types.StringValue(org.Slug)
	data.MaxAllowedMemberships = types.Int64Value(org.MaxAllowedMemberships)
	data.AdminDeleteEnabled = types.BoolValue(org.AdminDeleteEnabled)
	data.CreatedAt = types.Int64Value(org.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
