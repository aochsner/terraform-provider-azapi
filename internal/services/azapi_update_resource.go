package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/locks"
	"github.com/Azure/terraform-provider-azapi/internal/services/defaults"
	"github.com/Azure/terraform-provider-azapi/internal/services/dynamic"
	"github.com/Azure/terraform-provider-azapi/internal/services/myplanmodifier"
	"github.com/Azure/terraform-provider-azapi/internal/services/myvalidator"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/Azure/terraform-provider-azapi/utils"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type AzapiUpdateResourceModel struct {
	ID                    types.String   `tfsdk:"id"`
	Name                  types.String   `tfsdk:"name"`
	ParentID              types.String   `tfsdk:"parent_id"`
	ResourceID            types.String   `tfsdk:"resource_id"`
	Type                  types.String   `tfsdk:"type"`
	Body                  types.String   `tfsdk:"body"`
	Payload               types.Dynamic  `tfsdk:"payload"`
	IgnoreCasing          types.Bool     `tfsdk:"ignore_casing"`
	IgnoreBodyChanges     types.List     `tfsdk:"ignore_body_changes"`
	IgnoreMissingProperty types.Bool     `tfsdk:"ignore_missing_property"`
	ResponseExportValues  types.List     `tfsdk:"response_export_values"`
	Locks                 types.List     `tfsdk:"locks"`
	Output                types.String   `tfsdk:"output"`
	OutputPayload         types.Dynamic  `tfsdk:"output_payload"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}

type AzapiUpdateResource struct {
	ProviderData *clients.Client
}

var _ resource.Resource = &AzapiUpdateResource{}
var _ resource.ResourceWithConfigure = &AzapiUpdateResource{}
var _ resource.ResourceWithValidateConfig = &AzapiUpdateResource{}
var _ resource.ResourceWithModifyPlan = &AzapiUpdateResource{}

func (r *AzapiUpdateResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*clients.Client); ok {
		r.ProviderData = v
	}
}
func (r *AzapiUpdateResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_update_resource"
}

func (r *AzapiUpdateResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					myvalidator.StringIsNotEmpty(),
				},
			},

			"parent_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					myvalidator.StringIsResourceID(),
				},
			},

			"resource_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					myvalidator.StringIsResourceID(),
				},
			},

			"type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					myvalidator.StringIsResourceType(),
				},
			},

			"body": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  defaults.StringDefault("{}"),
				Validators: []validator.String{
					myvalidator.StringIsJSON(),
				},
				PlanModifiers: []planmodifier.String{
					myplanmodifier.UseStateWhen(func(a, b types.String) bool {
						return utils.NormalizeJson(a.ValueString()) == utils.NormalizeJson(b.ValueString())
					}),
				},
				DeprecationMessage: "This feature is deprecated and will be removed in a major release. Please use the `payload` argument to specify the body of the resource.",
			},

			"payload": schema.DynamicAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Dynamic{
					myplanmodifier.DynamicUseStateWhen(dynamic.SemanticallyEqual),
				},
			},

			"ignore_body_changes": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(myvalidator.StringIsNotEmpty()),
				},
				DeprecationMessage: "This feature is deprecated and will be removed in a major release. Please use the `lifecycle.ignore_changes` argument to specify the fields in `payload` to ignore.",
			},

			"ignore_casing": schema.BoolAttribute{
				Optional:           true,
				Computed:           true,
				Default:            defaults.BoolDefault(false),
				DeprecationMessage: "This feature is deprecated and will be removed in a major release. Please use the `lifecycle.ignore_changes` argument to specify the fields in `payload` to ignore.",
			},

			"ignore_missing_property": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  defaults.BoolDefault(true),
			},

			"response_export_values": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(myvalidator.StringIsNotEmpty()),
				},
			},

			"locks": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(myvalidator.StringIsNotEmpty()),
				},
			},

			"output": schema.StringAttribute{
				Computed:           true,
				DeprecationMessage: "This feature is deprecated and will be removed in a major release. Please use the `payload` argument to specify the body of the resource.",
			},

			"output_payload": schema.DynamicAttribute{
				Computed: true,
			},
		},

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

func (r *AzapiUpdateResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config *AzapiUpdateResourceModel
	if response.Diagnostics.Append(request.Config.Get(ctx, &config)...); response.Diagnostics.HasError() {
		return
	}

	if config == nil {
		return
	}

	// can't specify both body and payload
	if !config.Body.IsNull() && !config.Payload.IsNull() {
		response.Diagnostics.AddError("Invalid config", "can't specify both body and payload")
	}

	if config.Name.IsNull() && !config.ParentID.IsNull() {
		response.Diagnostics.AddError("Invalid configuration", `The argument "name" is required when the argument "parent_id" is set`)
	}
	if !config.Name.IsNull() && config.ParentID.IsNull() {
		response.Diagnostics.AddError("Invalid configuration", `The argument "parent_id" is required when the argument "name" is set`)
	}
	if config.Name.IsNull() && config.ResourceID.IsNull() {
		response.Diagnostics.AddError("Invalid configuration", `One of the arguments "name" or "resource_id" must be set`)
	}
	if !config.Name.IsNull() && !config.ResourceID.IsNull() {
		response.Diagnostics.AddError("Invalid configuration", `Only one of the arguments "name" or "resource_id" can be set`)
	}
	if response.Diagnostics.HasError() {
		return
	}

	if name := config.Name.ValueString(); name != "" {
		parentId := config.ParentID.ValueString()
		resourceType := config.Type.ValueString()

		// verify parent_id when it's known
		if parentId != "" {
			_, err := parse.NewResourceID(name, parentId, resourceType)
			if err != nil {
				response.Diagnostics.AddError("Invalid configuration", err.Error())
				return
			}
		}
	}
}

func (r *AzapiUpdateResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	var config, state, plan *AzapiUpdateResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// destroy doesn't need to modify plan
	if config == nil {
		return
	}

	if state == nil || !plan.ResponseExportValues.Equal(state.ResponseExportValues) || utils.NormalizeJson(plan.Body.ValueString()) != utils.NormalizeJson(state.Body.ValueString()) || !plan.Payload.Equal(state.Payload) {
		plan.Output = types.StringUnknown()
		plan.OutputPayload = basetypes.NewDynamicUnknown()
		response.Diagnostics.Append(response.Plan.Set(ctx, plan)...)
	}
}

func (r *AzapiUpdateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	r.CreateUpdate(ctx, request.Plan, &response.State, &response.Diagnostics)
}

func (r *AzapiUpdateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	r.CreateUpdate(ctx, request.Plan, &response.State, &response.Diagnostics)
}

func (r *AzapiUpdateResource) CreateUpdate(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	var model AzapiUpdateResourceModel
	if diagnostics.Append(plan.Get(ctx, &model)...); diagnostics.HasError() {
		return
	}

	createUpdateTimeout, diags := model.Timeouts.Create(ctx, 30*time.Minute)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createUpdateTimeout)
	defer cancel()

	var id parse.ResourceId
	if name := model.Name.ValueString(); len(name) != 0 {
		buildId, err := parse.NewResourceID(model.Name.ValueString(), model.ParentID.ValueString(), model.Type.ValueString())
		if err != nil {
			diagnostics.AddError("Invalid configuration", err.Error())
			return
		}
		id = buildId
	} else {
		buildId, err := parse.ResourceIDWithResourceType(model.ResourceID.ValueString(), model.Type.ValueString())
		if err != nil {
			diagnostics.AddError("Invalid configuration", err.Error())
			return
		}
		id = buildId
	}

	client := r.ProviderData.ResourceClient
	existing, err := client.Get(ctx, id.AzureResourceId, id.ApiVersion)
	if err != nil {
		diagnostics.AddError("Failed to retrieve resource", fmt.Errorf("checking for presence of existing %s: %+v", id, err).Error())
		return
	}
	if utils.GetId(existing) == nil {
		diagnostics.AddError("Failed to retrieve resource", fmt.Errorf("update target does not exist %s", id).Error())
		return
	}

	var requestBody interface{}
	switch {
	case !model.Payload.IsNull():
		out, err := expandPayload(model.Payload)
		if err != nil {
			diagnostics.AddError("Invalid payload", err.Error())
			return
		}
		requestBody = out
	case !model.Body.IsNull():
		bodyValueString := model.Body.ValueString()
		err := json.Unmarshal([]byte(bodyValueString), &requestBody)
		if err != nil {
			diagnostics.AddError("Invalid JSON string", fmt.Sprintf(`The argument "body" is invalid: value: %s, err: %+v`, model.Body.ValueString(), err))
			return
		}
	default:
		requestBody = map[string]interface{}{}
	}

	requestBody = utils.MergeObject(existing, requestBody)
	if ignoreChanges := AsStringList(model.IgnoreBodyChanges); len(ignoreChanges) != 0 {
		out, err := overrideWithPaths(requestBody, existing, ignoreChanges)
		if err != nil {
			diagnostics.AddError("Invalid configuration", fmt.Sprintf(`The argument "ignore_body_changes" is invalid, value: %q, error: %s`, model.IgnoreBodyChanges.String(), err.Error()))
			return
		}
		requestBody = out
	}

	if id.ResourceDef != nil {
		requestBody = (*id.ResourceDef).GetWriteOnly(utils.NormalizeObject(requestBody))
	}

	for _, id := range AsStringList(model.Locks) {
		locks.ByID(id)
		defer locks.UnlockByID(id)
	}

	responseBody, err := client.CreateOrUpdate(ctx, id.AzureResourceId, id.ApiVersion, requestBody)
	if err != nil {
		diagnostics.AddError("Failed to update resource", fmt.Errorf("updating %q: %+v", id, err).Error())
		return
	}

	model.ID = basetypes.NewStringValue(id.ID())
	model.Name = basetypes.NewStringValue(id.Name)
	model.ParentID = basetypes.NewStringValue(id.ParentId)
	model.ResourceID = basetypes.NewStringValue(id.AzureResourceId)
	model.Output = basetypes.NewStringValue(flattenOutput(responseBody, AsStringList(model.ResponseExportValues)))
	model.OutputPayload = types.DynamicValue(flattenOutputPayload(responseBody, AsStringList(model.ResponseExportValues)))

	diagnostics.Append(state.Set(ctx, model)...)
}

func (r *AzapiUpdateResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var model AzapiUpdateResourceModel
	if response.Diagnostics.Append(request.State.Get(ctx, &model)...); response.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := model.Timeouts.Read(ctx, 5*time.Minute)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	id, err := parse.ResourceIDWithResourceType(model.ID.ValueString(), model.Type.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid resource id", err.Error())
		return
	}

	client := r.ProviderData.ResourceClient
	responseBody, err := client.Get(ctx, id.AzureResourceId, id.ApiVersion)
	if err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			tflog.Info(ctx, fmt.Sprintf("[INFO] Error reading %q - removing from state", id.ID()))
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Failed to retrieve resource", fmt.Errorf("reading %q: %+v", id, err).Error())
		return
	}

	state := model
	state.ID = basetypes.NewStringValue(id.ID())
	state.Name = basetypes.NewStringValue(id.Name)
	state.ParentID = basetypes.NewStringValue(id.ParentId)
	state.ResourceID = basetypes.NewStringValue(id.AzureResourceId)
	state.Type = basetypes.NewStringValue(fmt.Sprintf("%s@%s", id.AzureResourceType, id.ApiVersion))

	var requestBody map[string]interface{}
	var useBody bool
	switch {
	case !model.Payload.IsNull():
		useBody = false
		out, err := expandPayload(model.Payload)
		if err != nil {
			response.Diagnostics.AddError("Invalid payload", err.Error())
			return
		}
		requestBody = out
	case !model.Body.IsNull():
		useBody = true
		bodyValueString := model.Body.ValueString()
		err := json.Unmarshal([]byte(bodyValueString), &requestBody)
		if err != nil {
			response.Diagnostics.AddError("Invalid JSON string", fmt.Sprintf(`The argument "body" is invalid: value: %s, err: %+v`, model.Body.ValueString(), err))
			return
		}
	default:
		requestBody = map[string]interface{}{}
	}

	if out, err := overrideWithPaths(responseBody, requestBody, AsStringList(model.IgnoreBodyChanges)); err == nil {
		responseBody = out
	} else {
		response.Diagnostics.AddError("Invalid configuration", fmt.Sprintf(`The argument "ignore_body_changes" is invalid, value: %q, error: %s`, model.IgnoreBodyChanges.String(), err.Error()))
		return
	}

	option := utils.UpdateJsonOption{
		IgnoreCasing:          model.IgnoreCasing.ValueBool(),
		IgnoreMissingProperty: model.IgnoreMissingProperty.ValueBool(),
	}
	body := utils.UpdateObject(requestBody, responseBody, option)

	data, err := json.Marshal(body)
	if err != nil {
		response.Diagnostics.AddError("Invalid body", err.Error())
		return
	}
	if useBody {
		state.Body = types.StringValue(string(data))
	} else {
		payload, err := dynamic.FromJSON(data, model.Payload.UnderlyingValue().Type(ctx))
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to parse payload: %s", err.Error()))
			payload, err = dynamic.FromJSONImplied(data)
			if err != nil {
				response.Diagnostics.AddError("Invalid payload", err.Error())
				return
			}
		}
		state.Payload = payload
	}

	state.Output = basetypes.NewStringValue(flattenOutput(responseBody, AsStringList(model.ResponseExportValues)))
	state.OutputPayload = types.DynamicValue(flattenOutputPayload(responseBody, AsStringList(model.ResponseExportValues)))

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *AzapiUpdateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {

}
