package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/instancesettings"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/makolabsai/terraform-provider-clerk/internal/client"
)

var (
	_ resource.Resource                = (*EnvironmentResource)(nil)
	_ resource.ResourceWithImportState = (*EnvironmentResource)(nil)
)

// EnvironmentResource configures a Clerk instance's settings via the Backend API.
// The instance itself is auto-created by Clerk; this resource manages its configuration.
type EnvironmentResource struct {
	client *client.ClerkClient
}

// EnvironmentResourceModel describes the Terraform resource data model.
type EnvironmentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	Environment   types.String `tfsdk:"environment"`

	// Instance settings (PATCH /instance)
	TestMode                    types.Bool   `tfsdk:"test_mode"`
	HIBP                        types.Bool   `tfsdk:"hibp"`
	EnhancedEmailDeliverability types.Bool   `tfsdk:"enhanced_email_deliverability"`
	SupportEmail                types.String `tfsdk:"support_email"`
	ClerkJSVersion              types.String `tfsdk:"clerk_js_version"`
	URLBasedSessionSyncing      types.Bool   `tfsdk:"url_based_session_syncing"`
	DevelopmentOrigin           types.String `tfsdk:"development_origin"`

	// Restrictions (PATCH /instance/restrictions)
	Restrictions types.Object `tfsdk:"restrictions"`

	// Organization settings (PATCH /instance/organization_settings)
	OrganizationSettings types.Object `tfsdk:"organization_settings"`
}

// RestrictionsModel maps the restrictions block.
type RestrictionsModel struct {
	Allowlist                   types.Bool `tfsdk:"allowlist"`
	Blocklist                   types.Bool `tfsdk:"blocklist"`
	BlockEmailSubaddresses      types.Bool `tfsdk:"block_email_subaddresses"`
	BlockDisposableEmailDomains types.Bool `tfsdk:"block_disposable_email_domains"`
	IgnoreDotsForGmailAddresses types.Bool `tfsdk:"ignore_dots_for_gmail_addresses"`
}

// OrganizationSettingsModel maps the organization_settings block.
type OrganizationSettingsModel struct {
	Enabled                types.Bool   `tfsdk:"enabled"`
	MaxAllowedMemberships  types.Int64  `tfsdk:"max_allowed_memberships"`
	CreatorRoleID          types.String `tfsdk:"creator_role_id"`
	AdminDeleteEnabled     types.Bool   `tfsdk:"admin_delete_enabled"`
	DomainsEnabled         types.Bool   `tfsdk:"domains_enabled"`
	DomainsEnrollmentModes types.List   `tfsdk:"domains_enrollment_modes"`
	DomainsDefaultRoleID   types.String `tfsdk:"domains_default_role_id"`
}

var restrictionsAttrTypes = map[string]attr.Type{
	"allowlist":                       types.BoolType,
	"blocklist":                       types.BoolType,
	"block_email_subaddresses":        types.BoolType,
	"block_disposable_email_domains":  types.BoolType,
	"ignore_dots_for_gmail_addresses": types.BoolType,
}

var orgSettingsAttrTypes = map[string]attr.Type{
	"enabled":                  types.BoolType,
	"max_allowed_memberships":  types.Int64Type,
	"creator_role_id":          types.StringType,
	"admin_delete_enabled":     types.BoolType,
	"domains_enabled":          types.BoolType,
	"domains_enrollment_modes": types.ListType{ElemType: types.StringType},
	"domains_default_role_id":  types.StringType,
}

func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

func (r *EnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configures a Clerk instance's settings (development or production). " +
			"The instance is auto-created by Clerk when the application is created; this resource manages its configuration. " +
			"Note: Authentication strategies (email/password/OAuth/MFA) are only configurable via the Clerk Dashboard.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier: {application_id}/{environment}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The Clerk application ID this environment belongs to.",
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

			// Instance settings (PATCH /instance)
			"test_mode": schema.BoolAttribute{
				Description: "Whether test mode is enabled. Defaults to true for development instances.",
				Optional:    true,
				Computed:    true,
			},
			"hibp": schema.BoolAttribute{
				Description: "Whether Have I Been Pwned password checking is enabled.",
				Optional:    true,
				Computed:    true,
			},
			"enhanced_email_deliverability": schema.BoolAttribute{
				Description: "Whether Clerk sends OTP emails via shared domain (Postmark) in production.",
				Optional:    true,
				Computed:    true,
			},
			"support_email": schema.StringAttribute{
				Description: "Contact email displayed to users needing support.",
				Optional:    true,
				Computed:    true,
			},
			"clerk_js_version": schema.StringAttribute{
				Description: "Specific Clerk.js version for hosted account pages. Empty string removes pinned version.",
				Optional:    true,
				Computed:    true,
			},
			"url_based_session_syncing": schema.BoolAttribute{
				Description: "Whether URL-based session syncing is enabled (replaces third-party cookies in dev).",
				Optional:    true,
				Computed:    true,
			},
			"development_origin": schema.StringAttribute{
				Description: "Origin URL for development instances to fix third-party cookie issues.",
				Optional:    true,
				Computed:    true,
			},

			// Restrictions (PATCH /instance/restrictions)
			"restrictions": schema.SingleNestedAttribute{
				Description: "Instance restriction settings for email validation and access control.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"allowlist": schema.BoolAttribute{
						Description: "Whether the allowlist is enabled.",
						Optional:    true,
						Computed:    true,
					},
					"blocklist": schema.BoolAttribute{
						Description: "Whether the blocklist is enabled.",
						Optional:    true,
						Computed:    true,
					},
					"block_email_subaddresses": schema.BoolAttribute{
						Description: "Whether email subaddresses (user+tag@domain.com) are blocked.",
						Optional:    true,
						Computed:    true,
					},
					"block_disposable_email_domains": schema.BoolAttribute{
						Description: "Whether disposable email domains are blocked.",
						Optional:    true,
						Computed:    true,
					},
					"ignore_dots_for_gmail_addresses": schema.BoolAttribute{
						Description: "Whether dots are ignored in Gmail addresses for uniqueness checks.",
						Optional:    true,
						Computed:    true,
					},
				},
			},

			// Organization settings (PATCH /instance/organization_settings)
			"organization_settings": schema.SingleNestedAttribute{
				Description: "Organization feature settings for the instance.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether organizations are enabled.",
						Optional:    true,
						Computed:    true,
					},
					"max_allowed_memberships": schema.Int64Attribute{
						Description: "Maximum number of memberships per organization.",
						Optional:    true,
						Computed:    true,
					},
					"creator_role_id": schema.StringAttribute{
						Description: "Role ID assigned to organization creators.",
						Optional:    true,
						Computed:    true,
					},
					"admin_delete_enabled": schema.BoolAttribute{
						Description: "Whether organization admins can delete the organization.",
						Optional:    true,
						Computed:    true,
					},
					"domains_enabled": schema.BoolAttribute{
						Description: "Whether organization domains are enabled.",
						Optional:    true,
						Computed:    true,
					},
					"domains_enrollment_modes": schema.ListAttribute{
						Description: "Enrollment modes for organization domains.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"domains_default_role_id": schema.StringAttribute{
						Description: "Default role ID for domain-enrolled members.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
}

func (r *EnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := plan.ApplicationID.ValueString()
	env := plan.Environment.ValueString()
	plan.ID = types.StringValue(appID + "/" + env)

	// Apply all settings to the instance.
	r.applySettings(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The Clerk Backend API does not provide GET endpoints for instance settings.
	// We preserve the current state as-is. Drift from dashboard changes won't be detected.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applySettings(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	env := state.Environment.ValueString()

	// Reset instance settings to defaults.
	defaultTrue := true
	defaultFalse := false
	emptyStr := ""

	err := r.client.UpdateInstanceSettings(ctx, appID, env, &instancesettings.UpdateParams{
		TestMode:                    &defaultFalse,
		HIBP:                        &defaultTrue,
		EnhancedEmailDeliverability: &defaultTrue,
		SupportEmail:                &emptyStr,
		ClerkJSVersion:              &emptyStr,
		URLBasedSessionSyncing:      &defaultFalse,
		DevelopmentOrigin:           &emptyStr,
	})
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Failed to reset instance settings",
			fmt.Sprintf("Could not reset instance settings for %s/%s: %s. The instance still exists in Clerk.", appID, env, err.Error()),
		)
	}

	// Reset restrictions to defaults.
	_, err = r.client.UpdateInstanceRestrictions(ctx, appID, env, &instancesettings.UpdateRestrictionsParams{
		Allowlist:                   &defaultFalse,
		Blocklist:                   &defaultFalse,
		BlockEmailSubaddresses:      &defaultFalse,
		BlockDisposableEmailDomains: &defaultFalse,
		IgnoreDotsForGmailAddresses: &defaultFalse,
	})
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Failed to reset instance restrictions",
			fmt.Sprintf("Could not reset restrictions for %s/%s: %s. The instance still exists in Clerk.", appID, env, err.Error()),
		)
	}

	// Reset organization settings to defaults.
	_, err = r.client.UpdateOrganizationSettings(ctx, appID, env, &instancesettings.UpdateOrganizationSettingsParams{
		Enabled:            &defaultFalse,
		AdminDeleteEnabled: &defaultFalse,
		DomainsEnabled:     &defaultFalse,
	})
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Failed to reset organization settings",
			fmt.Sprintf("Could not reset organization settings for %s/%s: %s. The instance still exists in Clerk.", appID, env, err.Error()),
		)
	}
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: {application_id}/{environment}, got: %q", req.ID),
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

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment"), parts[1])...)
}

// applySettings pushes all configured settings to the Clerk Backend API.
func (r *EnvironmentResource) applySettings(ctx context.Context, plan *EnvironmentResourceModel, diags *diag.Diagnostics) {
	appID := plan.ApplicationID.ValueString()
	env := plan.Environment.ValueString()

	// 1. Apply instance settings.
	r.applyInstanceSettings(ctx, appID, env, plan, diags)
	if diags.HasError() {
		return
	}

	// 2. Apply restrictions.
	r.applyRestrictions(ctx, appID, env, plan, diags)
	if diags.HasError() {
		return
	}

	// 3. Apply organization settings.
	r.applyOrganizationSettings(ctx, appID, env, plan, diags)
}

func (r *EnvironmentResource) applyInstanceSettings(ctx context.Context, appID, env string, plan *EnvironmentResourceModel, diags *diag.Diagnostics) {
	params := &instancesettings.UpdateParams{}
	hasChanges := false

	if !plan.TestMode.IsNull() && !plan.TestMode.IsUnknown() {
		v := plan.TestMode.ValueBool()
		params.TestMode = &v
		hasChanges = true
	}
	if !plan.HIBP.IsNull() && !plan.HIBP.IsUnknown() {
		v := plan.HIBP.ValueBool()
		params.HIBP = &v
		hasChanges = true
	}
	if !plan.EnhancedEmailDeliverability.IsNull() && !plan.EnhancedEmailDeliverability.IsUnknown() {
		v := plan.EnhancedEmailDeliverability.ValueBool()
		params.EnhancedEmailDeliverability = &v
		hasChanges = true
	}
	if !plan.SupportEmail.IsNull() && !plan.SupportEmail.IsUnknown() {
		v := plan.SupportEmail.ValueString()
		params.SupportEmail = &v
		hasChanges = true
	}
	if !plan.ClerkJSVersion.IsNull() && !plan.ClerkJSVersion.IsUnknown() {
		v := plan.ClerkJSVersion.ValueString()
		params.ClerkJSVersion = &v
		hasChanges = true
	}
	if !plan.URLBasedSessionSyncing.IsNull() && !plan.URLBasedSessionSyncing.IsUnknown() {
		v := plan.URLBasedSessionSyncing.ValueBool()
		params.URLBasedSessionSyncing = &v
		hasChanges = true
	}
	if !plan.DevelopmentOrigin.IsNull() && !plan.DevelopmentOrigin.IsUnknown() {
		v := plan.DevelopmentOrigin.ValueString()
		params.DevelopmentOrigin = &v
		hasChanges = true
	}

	if !hasChanges {
		return
	}

	err := r.client.UpdateInstanceSettings(ctx, appID, env, params)
	if err != nil {
		diags.AddError("Error updating instance settings", err.Error())
	}
}

func (r *EnvironmentResource) applyRestrictions(ctx context.Context, appID, env string, plan *EnvironmentResourceModel, diags *diag.Diagnostics) {
	if plan.Restrictions.IsNull() || plan.Restrictions.IsUnknown() {
		return
	}

	var restrictions RestrictionsModel
	diags.Append(plan.Restrictions.As(ctx, &restrictions, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}

	params := &instancesettings.UpdateRestrictionsParams{}
	if !restrictions.Allowlist.IsNull() && !restrictions.Allowlist.IsUnknown() {
		v := restrictions.Allowlist.ValueBool()
		params.Allowlist = &v
	}
	if !restrictions.Blocklist.IsNull() && !restrictions.Blocklist.IsUnknown() {
		v := restrictions.Blocklist.ValueBool()
		params.Blocklist = &v
	}
	if !restrictions.BlockEmailSubaddresses.IsNull() && !restrictions.BlockEmailSubaddresses.IsUnknown() {
		v := restrictions.BlockEmailSubaddresses.ValueBool()
		params.BlockEmailSubaddresses = &v
	}
	if !restrictions.BlockDisposableEmailDomains.IsNull() && !restrictions.BlockDisposableEmailDomains.IsUnknown() {
		v := restrictions.BlockDisposableEmailDomains.ValueBool()
		params.BlockDisposableEmailDomains = &v
	}
	if !restrictions.IgnoreDotsForGmailAddresses.IsNull() && !restrictions.IgnoreDotsForGmailAddresses.IsUnknown() {
		v := restrictions.IgnoreDotsForGmailAddresses.ValueBool()
		params.IgnoreDotsForGmailAddresses = &v
	}

	result, err := r.client.UpdateInstanceRestrictions(ctx, appID, env, params)
	if err != nil {
		diags.AddError("Error updating instance restrictions", err.Error())
		return
	}

	// Update state from the API response.
	restrictionsObj, d := types.ObjectValueFrom(ctx, restrictionsAttrTypes, &RestrictionsModel{
		Allowlist:                   types.BoolValue(result.Allowlist),
		Blocklist:                   types.BoolValue(result.Blocklist),
		BlockEmailSubaddresses:      types.BoolValue(result.BlockEmailSubaddresses),
		BlockDisposableEmailDomains: types.BoolValue(result.BlockDisposableEmailDomains),
		IgnoreDotsForGmailAddresses: types.BoolValue(result.IgnoreDotsForGmailAddresses),
	})
	diags.Append(d...)
	plan.Restrictions = restrictionsObj
}

func (r *EnvironmentResource) applyOrganizationSettings(ctx context.Context, appID, env string, plan *EnvironmentResourceModel, diags *diag.Diagnostics) {
	if plan.OrganizationSettings.IsNull() || plan.OrganizationSettings.IsUnknown() {
		return
	}

	var orgSettings OrganizationSettingsModel
	diags.Append(plan.OrganizationSettings.As(ctx, &orgSettings, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}

	params := &instancesettings.UpdateOrganizationSettingsParams{}
	if !orgSettings.Enabled.IsNull() && !orgSettings.Enabled.IsUnknown() {
		v := orgSettings.Enabled.ValueBool()
		params.Enabled = &v
	}
	if !orgSettings.MaxAllowedMemberships.IsNull() && !orgSettings.MaxAllowedMemberships.IsUnknown() {
		v := orgSettings.MaxAllowedMemberships.ValueInt64()
		params.MaxAllowedMemberships = &v
	}
	if !orgSettings.CreatorRoleID.IsNull() && !orgSettings.CreatorRoleID.IsUnknown() {
		v := orgSettings.CreatorRoleID.ValueString()
		params.CreatorRoleID = &v
	}
	if !orgSettings.AdminDeleteEnabled.IsNull() && !orgSettings.AdminDeleteEnabled.IsUnknown() {
		v := orgSettings.AdminDeleteEnabled.ValueBool()
		params.AdminDeleteEnabled = &v
	}
	if !orgSettings.DomainsEnabled.IsNull() && !orgSettings.DomainsEnabled.IsUnknown() {
		v := orgSettings.DomainsEnabled.ValueBool()
		params.DomainsEnabled = &v
	}
	if !orgSettings.DomainsEnrollmentModes.IsNull() && !orgSettings.DomainsEnrollmentModes.IsUnknown() {
		var modes []string
		diags.Append(orgSettings.DomainsEnrollmentModes.ElementsAs(ctx, &modes, false)...)
		if diags.HasError() {
			return
		}
		params.DomainsEnrollmentModes = &modes
	}
	if !orgSettings.DomainsDefaultRoleID.IsNull() && !orgSettings.DomainsDefaultRoleID.IsUnknown() {
		v := orgSettings.DomainsDefaultRoleID.ValueString()
		params.DomainsDefaultRoleID = &v
	}

	result, err := r.client.UpdateOrganizationSettings(ctx, appID, env, params)
	if err != nil {
		diags.AddError("Error updating organization settings", err.Error())
		return
	}

	// Update state from the API response.
	enrollmentModes, d := types.ListValueFrom(ctx, types.StringType, result.DomainsEnrollmentModes)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	// The API returns role keys (e.g. "org:admin") in creator_role / domains_default_role,
	// but the params accept role IDs via creator_role_id / domains_default_role_id.
	// To avoid state drift, preserve the user's configured values for these fields
	// rather than overwriting with the differently-formatted API response.
	creatorRoleID := orgSettings.CreatorRoleID
	domainsDefaultRoleID := orgSettings.DomainsDefaultRoleID

	orgObj, d := types.ObjectValueFrom(ctx, orgSettingsAttrTypes, &OrganizationSettingsModel{
		Enabled:                types.BoolValue(result.Enabled),
		MaxAllowedMemberships:  types.Int64Value(result.MaxAllowedMemberships),
		CreatorRoleID:          creatorRoleID,
		AdminDeleteEnabled:     types.BoolValue(result.AdminDeleteEnabled),
		DomainsEnabled:         types.BoolValue(result.DomainsEnabled),
		DomainsEnrollmentModes: enrollmentModes,
		DomainsDefaultRoleID:   domainsDefaultRoleID,
	})
	diags.Append(d...)
	plan.OrganizationSettings = orgObj
}
