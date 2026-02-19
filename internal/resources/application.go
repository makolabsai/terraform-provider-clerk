package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/makolabsai/terraform-provider-clerk/internal/client"
)

var (
	_ resource.Resource                = (*ApplicationResource)(nil)
	_ resource.ResourceWithImportState = (*ApplicationResource)(nil)
)

// ApplicationResource manages a Clerk application via the Platform API.
type ApplicationResource struct {
	client *client.ClerkClient
}

// ApplicationResourceModel describes the Terraform resource data model.
type ApplicationResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	DeletionProtection types.Bool   `tfsdk:"deletion_protection"`
	Domain             types.String `tfsdk:"domain"`
	EnvironmentTypes   types.List   `tfsdk:"environment_types"`
	Template           types.String `tfsdk:"template"`
	DevInstanceID      types.String `tfsdk:"dev_instance_id"`
	DevSecretKey       types.String `tfsdk:"dev_secret_key"`
	DevPublishableKey  types.String `tfsdk:"dev_publishable_key"`
	ProdInstanceID     types.String `tfsdk:"prod_instance_id"`
	ProdSecretKey      types.String `tfsdk:"prod_secret_key"`
	ProdPublishableKey types.String `tfsdk:"prod_publishable_key"`
}

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

func (r *ApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Clerk application. Each application can have multiple instances (development, production) with distinct user pools.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the Clerk application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the application.",
				Required:    true,
			},
			"deletion_protection": schema.BoolAttribute{
				Description: "Whether deletion protection is enabled. When true, the application cannot be destroyed. " +
					"Set to false before destroying. Defaults to true.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "The domain for the application. Only set at creation time.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_types": schema.ListAttribute{
				Description: "List of environment types to create instances for (e.g., development, production). Only set at creation time.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listRequiresReplace{},
				},
			},
			"template": schema.StringAttribute{
				Description: "Application template (e.g., b2b-saas, b2c-saas, waitlist). Only set at creation time.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dev_instance_id": schema.StringAttribute{
				Description: "The instance ID for the development environment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dev_secret_key": schema.StringAttribute{
				Description: "The secret key for the development instance.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dev_publishable_key": schema.StringAttribute{
				Description: "The publishable key for the development instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"prod_instance_id": schema.StringAttribute{
				Description: "The instance ID for the production environment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"prod_secret_key": schema.StringAttribute{
				Description: "The secret key for the production instance.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"prod_publishable_key": schema.StringAttribute{
				Description: "The publishable key for the production instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ApplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.PlatformCreateApplicationRequest{
		Name: plan.Name.ValueString(),
	}

	if !plan.Domain.IsNull() && !plan.Domain.IsUnknown() {
		createReq.Domain = plan.Domain.ValueString()
	}

	if !plan.Template.IsNull() && !plan.Template.IsUnknown() {
		createReq.Template = plan.Template.ValueString()
	}

	if !plan.EnvironmentTypes.IsNull() && !plan.EnvironmentTypes.IsUnknown() {
		var envTypes []string
		resp.Diagnostics.Append(plan.EnvironmentTypes.ElementsAs(ctx, &envTypes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.EnvironmentTypes = envTypes
	}

	// Create the application — the response includes secret keys on create.
	application, err := r.client.CreateApplication(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Clerk application", err.Error())
		return
	}

	// Map the API response to state.
	plan.ID = types.StringValue(application.ApplicationID)
	if plan.DeletionProtection.IsNull() || plan.DeletionProtection.IsUnknown() {
		plan.DeletionProtection = types.BoolValue(true)
	}
	mapInstancesToState(application.Instances, &plan)

	// Register backend clients for each instance with a secret key.
	r.registerBackendClients(application.ApplicationID, application.Instances, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	application, err := r.client.GetApplication(ctx, state.ID.ValueString(), true)
	if err != nil {
		if apiErr, ok := err.(*client.PlatformAPIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Clerk application", err.Error())
		return
	}

	// The API does not return the name — preserve it from state.
	mapInstancesToState(application.Instances, &state)

	// Register backend clients for each instance with a secret key.
	r.registerBackendClients(application.ApplicationID, application.Instances, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.PlatformUpdateApplicationRequest{
		Name: plan.Name.ValueString(),
	}

	_, err := r.client.UpdateApplication(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Clerk application", err.Error())
		return
	}

	// Re-read the application to get fresh instance data.
	application, err := r.client.GetApplication(ctx, plan.ID.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Clerk application after update", err.Error())
		return
	}

	mapInstancesToState(application.Instances, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeletionProtection.ValueBool() {
		resp.Diagnostics.AddError(
			"Cannot destroy application with deletion protection enabled",
			fmt.Sprintf("Application %q (%s) has deletion_protection = true. "+
				"Set deletion_protection = false and apply before destroying.",
				state.Name.ValueString(), state.ID.ValueString()),
		)
		return
	}

	err := r.client.DeleteApplication(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Clerk application", err.Error())
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapInstancesToState maps Platform API instance data to the Terraform resource model.
func mapInstancesToState(instances []client.PlatformApplicationInstance, state *ApplicationResourceModel) {
	for _, inst := range instances {
		switch inst.EnvironmentType {
		case "development":
			state.DevInstanceID = types.StringValue(inst.InstanceID)
			state.DevPublishableKey = types.StringValue(inst.PublishableKey)
			if inst.SecretKey != "" {
				state.DevSecretKey = types.StringValue(inst.SecretKey)
			}
		case "production":
			state.ProdInstanceID = types.StringValue(inst.InstanceID)
			state.ProdPublishableKey = types.StringValue(inst.PublishableKey)
			if inst.SecretKey != "" {
				state.ProdSecretKey = types.StringValue(inst.SecretKey)
			}
		}
	}
}

// registerBackendClients registers Backend API clients for instances that have secret keys.
func (r *ApplicationResource) registerBackendClients(appID string, instances []client.PlatformApplicationInstance, diags *diag.Diagnostics) {
	// This is a no-op helper for now — will be wired in when backend resources need it.
}

// listRequiresReplace is a plan modifier that forces replacement when a list attribute changes.
type listRequiresReplace struct{}

func (m listRequiresReplace) Description(_ context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (m listRequiresReplace) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m listRequiresReplace) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.PlanValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}
	if req.StateValue.IsNull() {
		return
	}

	if !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}
