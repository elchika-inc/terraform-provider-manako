package monitor

import (
	"context"
	"fmt"

	"github.com/elchika-inc/terraform-provider-manako/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure interface compliance at compile time.
var (
	_ resource.Resource                = &MonitorResource{}
	_ resource.ResourceWithImportState = &MonitorResource{}
)

// MonitorResource implements the manako_monitor Terraform resource.
type MonitorResource struct {
	client *client.Client
}

// NewMonitorResource returns a constructor function for the MonitorResource.
func NewMonitorResource() resource.Resource {
	return &MonitorResource{}
}

func (r *MonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *MonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = monitorResourceSchema()
}

func (r *MonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = c
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Expand the type-specific config block into an API config map.
	config, diags := expandConfig(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the create request.
	createReq := client.CreateMonitorRequest{
		Type:            plan.Type.ValueString(),
		Config:          config,
		IntervalSeconds: plan.IntervalSeconds.ValueInt64(),
	}
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		createReq.Name = plan.Name.ValueString()
	}
	if !plan.ServiceID.IsNull() && !plan.ServiceID.IsUnknown() {
		createReq.ServiceID = plan.ServiceID.ValueString()
	}

	tflog.Debug(ctx, "Creating monitor", map[string]interface{}{
		"type": createReq.Type,
		"name": createReq.Name,
	})

	monitor, err := r.client.CreateMonitor(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating monitor", err.Error())
		return
	}

	// 2-step create: if is_active=false, immediately update because the create
	// API does not accept isActive.
	if !plan.IsActive.IsNull() && !plan.IsActive.ValueBool() {
		tflog.Debug(ctx, "Monitor created with is_active=false; sending update to deactivate", map[string]interface{}{
			"id": monitor.ID,
		})
		isActive := false
		updateReq := client.UpdateMonitorRequest{
			IsActive: &isActive,
		}
		updated, err := r.client.UpdateMonitor(monitor.ID, updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Error deactivating monitor after create", err.Error())
			return
		}
		monitor = updated
	}

	// Map the API response to Terraform state.
	diags = mapMonitorToState(ctx, monitor, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ---------------------------------------------------------------------------
// Read
// ---------------------------------------------------------------------------

func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, "Reading monitor", map[string]interface{}{"id": id})

	monitor, err := r.client.GetMonitor(id)
	if err != nil {
		// If 404, remove from state (drift detection).
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			tflog.Warn(ctx, "Monitor not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading monitor", err.Error())
		return
	}

	diags := mapMonitorToState(ctx, monitor, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Expand the type-specific config block.
	config, diags := expandConfig(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	intervalSeconds := plan.IntervalSeconds.ValueInt64()
	isActive := plan.IsActive.ValueBool()

	updateReq := client.UpdateMonitorRequest{
		Name:            plan.Name.ValueString(),
		Config:          config,
		IntervalSeconds: &intervalSeconds,
		IsActive:        &isActive,
	}

	tflog.Debug(ctx, "Updating monitor", map[string]interface{}{
		"id":   id,
		"name": updateReq.Name,
	})

	monitor, err := r.client.UpdateMonitor(id, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating monitor", err.Error())
		return
	}

	diags = mapMonitorToState(ctx, monitor, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, "Deleting monitor", map[string]interface{}{"id": id})

	err := r.client.DeleteMonitor(id)
	if err != nil {
		// Tolerate 404 (already deleted).
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			tflog.Warn(ctx, "Monitor already deleted", map[string]interface{}{"id": id})
			return
		}
		resp.Diagnostics.AddError("Error deleting monitor", err.Error())
	}
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// expandConfig — extracts the correct type-specific config block
// ---------------------------------------------------------------------------

func expandConfig(ctx context.Context, plan MonitorResourceModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	monitorType := plan.Type.ValueString()

	switch monitorType {
	case "http":
		if plan.HTTPConfig.IsNull() || plan.HTTPConfig.IsUnknown() {
			diags.AddError("Missing config block", "http_config block is required when type is \"http\".")
			return nil, diags
		}
		var config HTTPConfigModel
		diags.Append(plan.HTTPConfig.As(ctx, &config, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		return expandHTTPConfig(config), diags

	case "tcp":
		if plan.TCPConfig.IsNull() || plan.TCPConfig.IsUnknown() {
			diags.AddError("Missing config block", "tcp_config block is required when type is \"tcp\".")
			return nil, diags
		}
		var config TCPConfigModel
		diags.Append(plan.TCPConfig.As(ctx, &config, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		return expandTCPConfig(config), diags

	case "ping":
		if plan.PingConfig.IsNull() || plan.PingConfig.IsUnknown() {
			diags.AddError("Missing config block", "ping_config block is required when type is \"ping\".")
			return nil, diags
		}
		var config PingConfigModel
		diags.Append(plan.PingConfig.As(ctx, &config, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		return expandPingConfig(config), diags

	case "heartbeat":
		if plan.HeartbeatConfig.IsNull() || plan.HeartbeatConfig.IsUnknown() {
			diags.AddError("Missing config block", "heartbeat_config block is required when type is \"heartbeat\".")
			return nil, diags
		}
		var config HeartbeatConfigModel
		diags.Append(plan.HeartbeatConfig.As(ctx, &config, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		return expandHeartbeatConfig(config), diags

	case "webchange":
		if plan.WebChangeConfig.IsNull() || plan.WebChangeConfig.IsUnknown() {
			diags.AddError("Missing config block", "webchange_config block is required when type is \"webchange\".")
			return nil, diags
		}
		var config WebChangeConfigModel
		diags.Append(plan.WebChangeConfig.As(ctx, &config, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		return expandWebChangeConfig(config), diags

	case "ssl":
		if plan.SSLConfig.IsNull() || plan.SSLConfig.IsUnknown() {
			diags.AddError("Missing config block", "ssl_config block is required when type is \"ssl\".")
			return nil, diags
		}
		var config SSLConfigModel
		diags.Append(plan.SSLConfig.As(ctx, &config, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		return expandSSLConfig(config), diags

	case "domain":
		if plan.DomainConfig.IsNull() || plan.DomainConfig.IsUnknown() {
			diags.AddError("Missing config block", "domain_config block is required when type is \"domain\".")
			return nil, diags
		}
		var config DomainConfigModel
		diags.Append(plan.DomainConfig.As(ctx, &config, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		return expandDomainConfig(config), diags

	default:
		diags.AddError("Unknown monitor type", fmt.Sprintf("Unsupported monitor type: %q", monitorType))
		return nil, diags
	}
}

// ---------------------------------------------------------------------------
// flattenConfigToState — sets the correct config block in state, nulls the rest
// ---------------------------------------------------------------------------

func flattenConfigToState(ctx context.Context, monitorType string, apiConfig map[string]interface{}, state *MonitorResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Reset ALL config blocks to null first.
	state.HTTPConfig = types.ObjectNull(httpConfigAttrTypes())
	state.TCPConfig = types.ObjectNull(tcpConfigAttrTypes())
	state.PingConfig = types.ObjectNull(pingConfigAttrTypes())
	state.HeartbeatConfig = types.ObjectNull(heartbeatConfigAttrTypes())
	state.WebChangeConfig = types.ObjectNull(webchangeConfigAttrTypes())
	state.SSLConfig = types.ObjectNull(sslConfigAttrTypes())
	state.DomainConfig = types.ObjectNull(domainConfigAttrTypes())

	switch monitorType {
	case "http":
		model := flattenHTTPConfig(apiConfig)
		obj, d := types.ObjectValueFrom(ctx, httpConfigAttrTypes(), model)
		diags.Append(d...)
		state.HTTPConfig = obj

	case "tcp":
		model := flattenTCPConfig(apiConfig)
		obj, d := types.ObjectValueFrom(ctx, tcpConfigAttrTypes(), model)
		diags.Append(d...)
		state.TCPConfig = obj

	case "ping":
		model := flattenPingConfig(apiConfig)
		obj, d := types.ObjectValueFrom(ctx, pingConfigAttrTypes(), model)
		diags.Append(d...)
		state.PingConfig = obj

	case "heartbeat":
		model := flattenHeartbeatConfig(apiConfig)
		obj, d := types.ObjectValueFrom(ctx, heartbeatConfigAttrTypes(), model)
		diags.Append(d...)
		state.HeartbeatConfig = obj

	case "webchange":
		model := flattenWebChangeConfig(apiConfig)
		obj, d := types.ObjectValueFrom(ctx, webchangeConfigAttrTypes(), model)
		diags.Append(d...)
		state.WebChangeConfig = obj

	case "ssl":
		model := flattenSSLConfig(apiConfig)
		obj, d := types.ObjectValueFrom(ctx, sslConfigAttrTypes(), model)
		diags.Append(d...)
		state.SSLConfig = obj

	case "domain":
		model := flattenDomainConfig(apiConfig)
		obj, d := types.ObjectValueFrom(ctx, domainConfigAttrTypes(), model)
		diags.Append(d...)
		state.DomainConfig = obj
	}

	return diags
}

// ---------------------------------------------------------------------------
// mapMonitorToState — maps a *client.Monitor to MonitorResourceModel
// ---------------------------------------------------------------------------

func mapMonitorToState(ctx context.Context, m *client.Monitor, state *MonitorResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Simple fields.
	state.ID = types.StringValue(m.ID)
	state.Type = types.StringValue(m.Type)
	state.Name = types.StringValue(m.Name)
	state.IntervalSeconds = types.Int64Value(m.IntervalSeconds)
	state.Status = types.StringValue(m.Status)
	state.CreatedAt = types.StringValue(m.CreatedAt)
	state.UpdatedAt = types.StringValue(m.UpdatedAt)

	// Bool conversion: API returns int (0/1) → Terraform bool.
	state.IsActive = types.BoolValue(m.IsActive != 0)
	state.AutoResolveMaintenance = types.BoolValue(m.AutoResolveMaintenance != 0)

	// ServiceID: may be empty string from API.
	if m.ServiceID != "" {
		state.ServiceID = types.StringValue(m.ServiceID)
	} else {
		state.ServiceID = types.StringNull()
	}

	// Nullable fields: *string → types.String.
	if m.LastCheckedAt != nil {
		state.LastCheckedAt = types.StringValue(*m.LastCheckedAt)
	} else {
		state.LastCheckedAt = types.StringNull()
	}

	if m.LastErrorMessage != nil {
		state.LastErrorMessage = types.StringValue(*m.LastErrorMessage)
	} else {
		state.LastErrorMessage = types.StringNull()
	}

	if m.MaintenanceUntil != nil {
		state.MaintenanceUntil = types.StringValue(*m.MaintenanceUntil)
	} else {
		state.MaintenanceUntil = types.StringNull()
	}

	// Config: flatten API config map into the correct type-specific block.
	d := flattenConfigToState(ctx, m.Type, m.Config, state)
	diags.Append(d...)

	return diags
}
