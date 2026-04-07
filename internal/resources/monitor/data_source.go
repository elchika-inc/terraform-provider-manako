package monitor

import (
	"context"
	"fmt"

	"github.com/elchika-inc/terraform-provider-manako/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure interface compliance at compile time.
var _ datasource.DataSource = &MonitorDataSource{}

// MonitorDataSource implements the manako_monitor data source.
type MonitorDataSource struct {
	client *client.Client
}

// NewMonitorDataSource returns a constructor function for the MonitorDataSource.
func NewMonitorDataSource() datasource.DataSource {
	return &MonitorDataSource{}
}

func (d *MonitorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (d *MonitorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Fetches a single Manako monitor by ID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Description: "Unique identifier of the monitor.",
				Required:    true,
			},
			"type": dsschema.StringAttribute{
				Description: "Monitor type.",
				Computed:    true,
			},
			"name": dsschema.StringAttribute{
				Description: "Display name of the monitor.",
				Computed:    true,
			},
			"interval_seconds": dsschema.Int64Attribute{
				Description: "Check interval in seconds.",
				Computed:    true,
			},
			"is_active": dsschema.BoolAttribute{
				Description: "Whether the monitor is active.",
				Computed:    true,
			},
			"service_id": dsschema.StringAttribute{
				Description: "Optional service group ID this monitor belongs to.",
				Computed:    true,
			},
			"status": dsschema.StringAttribute{
				Description: "Current status of the monitor.",
				Computed:    true,
			},
			"last_checked_at": dsschema.StringAttribute{
				Description: "Timestamp of the last check.",
				Computed:    true,
			},
			"last_error_message": dsschema.StringAttribute{
				Description: "Last error message if the monitor is down.",
				Computed:    true,
			},
			"maintenance_until": dsschema.StringAttribute{
				Description: "Timestamp until which the monitor is in maintenance mode.",
				Computed:    true,
			},
			"auto_resolve_maintenance": dsschema.BoolAttribute{
				Description: "Whether maintenance mode is automatically resolved.",
				Computed:    true,
			},
			"created_at": dsschema.StringAttribute{
				Description: "Timestamp when the monitor was created.",
				Computed:    true,
			},
			"updated_at": dsschema.StringAttribute{
				Description: "Timestamp when the monitor was last updated.",
				Computed:    true,
			},
			"http_config": dsschema.SingleNestedAttribute{
				Description: "Configuration for HTTP monitors.",
				Computed:    true,
				Attributes: map[string]dsschema.Attribute{
					"url":                dsschema.StringAttribute{Computed: true},
					"method":             dsschema.StringAttribute{Computed: true},
					"expected_status":    dsschema.Int64Attribute{Computed: true},
					"timeout_ms":         dsschema.Int64Attribute{Computed: true},
					"headers":            dsschema.MapAttribute{Computed: true, ElementType: types.StringType},
					"keyword":            dsschema.StringAttribute{Computed: true},
					"keyword_must_exist": dsschema.BoolAttribute{Computed: true},
				},
			},
			"tcp_config": dsschema.SingleNestedAttribute{
				Description: "Configuration for TCP monitors.",
				Computed:    true,
				Attributes: map[string]dsschema.Attribute{
					"hostname":   dsschema.StringAttribute{Computed: true},
					"port":       dsschema.Int64Attribute{Computed: true},
					"timeout_ms": dsschema.Int64Attribute{Computed: true},
				},
			},
			"ping_config": dsschema.SingleNestedAttribute{
				Description: "Configuration for Ping monitors.",
				Computed:    true,
				Attributes: map[string]dsschema.Attribute{
					"hostname":   dsschema.StringAttribute{Computed: true},
					"timeout_ms": dsschema.Int64Attribute{Computed: true},
					"port":       dsschema.Int64Attribute{Computed: true},
				},
			},
			"heartbeat_config": dsschema.SingleNestedAttribute{
				Description: "Configuration for Heartbeat monitors.",
				Computed:    true,
				Attributes: map[string]dsschema.Attribute{
					"grace_seconds": dsschema.Int64Attribute{Computed: true},
				},
			},
			"webchange_config": dsschema.SingleNestedAttribute{
				Description: "Configuration for WebChange monitors.",
				Computed:    true,
				Attributes: map[string]dsschema.Attribute{
					"url":         dsschema.StringAttribute{Computed: true},
					"check_type":  dsschema.StringAttribute{Computed: true},
					"selector":    dsschema.StringAttribute{Computed: true},
					"change_mode": dsschema.StringAttribute{Computed: true},
				},
			},
			"ssl_config": dsschema.SingleNestedAttribute{
				Description: "Configuration for SSL certificate monitors.",
				Computed:    true,
				Attributes: map[string]dsschema.Attribute{
					"hostname":  dsschema.StringAttribute{Computed: true},
					"warn_days": dsschema.ListAttribute{Computed: true, ElementType: types.Int64Type},
				},
			},
			"domain_config": dsschema.SingleNestedAttribute{
				Description: "Configuration for domain expiry monitors.",
				Computed:    true,
				Attributes: map[string]dsschema.Attribute{
					"domain":    dsschema.StringAttribute{Computed: true},
					"warn_days": dsschema.ListAttribute{Computed: true, ElementType: types.Int64Type},
				},
			},
		},
	}
}

func (d *MonitorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = c
}

func (d *MonitorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, "Reading monitor data source", map[string]interface{}{"id": id})

	m, err := d.client.GetMonitor(id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading monitor", err.Error())
		return
	}

	diags := mapMonitorToState(ctx, m, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
