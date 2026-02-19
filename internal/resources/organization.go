package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/organization"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/makolabsai/terraform-provider-clerk/internal/client"
)

var (
	_ resource.Resource                = (*OrganizationResource)(nil)
	_ resource.ResourceWithImportState = (*OrganizationResource)(nil)
)

// OrganizationResource manages a Clerk organization via the Backend API.
type OrganizationResource struct {
	client *client.ClerkClient
}

// OrganizationResourceModel describes the Terraform resource data model.
type OrganizationResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	ApplicationID         types.String `tfsdk:"application_id"`
	Environment           types.String `tfsdk:"environment"`
	Name                  types.String `tfsdk:"name"`
	Slug                  types.String `tfsdk:"slug"`
	MaxAllowedMemberships types.Int64  `tfsdk:"max_allowed_memberships"`
	AdminDeleteEnabled    types.Bool   `tfsdk:"admin_delete_enabled"`
	CreatedAt             types.Int64  `tfsdk:"created_at"`
	UpdatedAt             types.Int64  `tfsdk:"updated_at"`
}

func NewOrganizationResource() resource.Resource {
	return &OrganizationResource{}
}

func (r *OrganizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *OrganizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Clerk organization within a specific application environment. " +
			"Organizations represent tenants or teams that group users together.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the Clerk organization.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The Clerk application ID this organization belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment": schema.StringAttribute{
				Description: "The environment type: \"development\" or \"production\".",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("development", "production"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the organization.",
				Required:    true,
			},
			"slug": schema.StringAttribute{
				Description: "URL-friendly identifier for the organization. Auto-generated from name if not provided.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"max_allowed_memberships": schema.Int64Attribute{
				Description: "Maximum number of memberships allowed in the organization. 0 means unlimited.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"admin_delete_enabled": schema.BoolAttribute{
				Description: "Whether organization admins can delete the organization.",
				Optional:    true,
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "Unix timestamp of when the organization was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.Int64Attribute{
				Description: "Unix timestamp of when the organization was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *OrganizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clerkClient, ok := req.ProviderData.(*client.ClerkClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.ClerkClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = clerkClient
}

func (r *OrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OrganizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	params := &organization.CreateParams{
		Name: &name,
	}

	if !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		slug := plan.Slug.ValueString()
		params.Slug = &slug
	}
	if !plan.MaxAllowedMemberships.IsNull() && !plan.MaxAllowedMemberships.IsUnknown() {
		v := plan.MaxAllowedMemberships.ValueInt64()
		params.MaxAllowedMemberships = &v
	}

	appID := plan.ApplicationID.ValueString()
	env := plan.Environment.ValueString()

	org, err := r.client.CreateOrganization(ctx, appID, env, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Clerk organization", err.Error())
		return
	}

	mapOrganizationToState(org, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OrganizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	env := state.Environment.ValueString()

	org, err := r.client.GetOrganization(ctx, appID, env, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*clerk.APIErrorResponse); ok && apiErr.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Clerk organization", err.Error())
		return
	}

	mapOrganizationToState(org, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrganizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &organization.UpdateParams{}

	name := plan.Name.ValueString()
	params.Name = &name

	if !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		slug := plan.Slug.ValueString()
		params.Slug = &slug
	}
	if !plan.MaxAllowedMemberships.IsNull() && !plan.MaxAllowedMemberships.IsUnknown() {
		v := plan.MaxAllowedMemberships.ValueInt64()
		params.MaxAllowedMemberships = &v
	}
	if !plan.AdminDeleteEnabled.IsNull() && !plan.AdminDeleteEnabled.IsUnknown() {
		v := plan.AdminDeleteEnabled.ValueBool()
		params.AdminDeleteEnabled = &v
	}

	appID := plan.ApplicationID.ValueString()
	env := plan.Environment.ValueString()

	org, err := r.client.UpdateOrganization(ctx, appID, env, plan.ID.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Clerk organization", err.Error())
		return
	}

	mapOrganizationToState(org, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OrganizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	env := state.Environment.ValueString()

	_, err := r.client.DeleteOrganization(ctx, appID, env, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Clerk organization", err.Error())
		return
	}
}

func (r *OrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected format: {application_id}/{environment}/{organization_id}
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: {application_id}/{environment}/{organization_id}, got: %q", req.ID),
		)
		return
	}

	if parts[1] != "development" && parts[1] != "production" {
		resp.Diagnostics.AddError(
			"Invalid Environment",
			fmt.Sprintf("Environment must be \"development\" or \"production\", got: %q", parts[1]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment"), parts[1])...)
}

// mapOrganizationToState maps a Clerk Organization API response to the Terraform model.
func mapOrganizationToState(org *clerk.Organization, state *OrganizationResourceModel) {
	state.ID = types.StringValue(org.ID)
	state.Name = types.StringValue(org.Name)
	state.Slug = types.StringValue(org.Slug)
	state.MaxAllowedMemberships = types.Int64Value(org.MaxAllowedMemberships)
	state.AdminDeleteEnabled = types.BoolValue(org.AdminDeleteEnabled)
	state.CreatedAt = types.Int64Value(org.CreatedAt)
	state.UpdatedAt = types.Int64Value(org.UpdatedAt)
}
