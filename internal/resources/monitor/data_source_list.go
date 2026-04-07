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
var _ datasource.DataSource = &MonitorsDataSource{}

// MonitorsDataSource implements the manako_monitors data source.
type MonitorsDataSource struct {
	client *client.Client
}

// MonitorsDataSourceModel is the Terraform state model for the manako_monitors data source.
type MonitorsDataSourceModel struct {
	Type     types.String  `tfsdk:"type"`
	Monitors []MonitorListItem `tfsdk:"monitors"`
}

// MonitorListItem is a single monitor entry in the manako_monitors list.
type MonitorListItem struct {
	ID              types.String `tfsdk:"id"`
	Type            types.String `tfsdk:"type"`
	Name            types.String `tfsdk:"name"`
	Status          types.String `tfsdk:"status"`
	IsActive        types.Bool   `tfsdk:"is_active"`
	IntervalSeconds types.Int64  `tfsdk:"interval_seconds"`
}

// NewMonitorsDataSource returns a constructor function for the MonitorsDataSource.
func NewMonitorsDataSource() datasource.DataSource {
	return &MonitorsDataSource{}
}

func (d *MonitorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitors"
}

func (d *MonitorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Lists Manako monitors, optionally filtered by type.",
		Attributes: map[string]dsschema.Attribute{
			"type": dsschema.StringAttribute{
				Description: "Optional monitor type filter (http, tcp, ping, heartbeat, webchange, ssl, domain).",
				Optional:    true,
			},
			"monitors": dsschema.ListNestedAttribute{
				Description: "List of monitors.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "Unique identifier of the monitor.",
							Computed:    true,
						},
						"type": dsschema.StringAttribute{
							Description: "Monitor type.",
							Computed:    true,
						},
						"name": dsschema.StringAttribute{
							Description: "Display name of the monitor.",
							Computed:    true,
						},
						"status": dsschema.StringAttribute{
							Description: "Current status of the monitor.",
							Computed:    true,
						},
						"is_active": dsschema.BoolAttribute{
							Description: "Whether the monitor is active.",
							Computed:    true,
						},
						"interval_seconds": dsschema.Int64Attribute{
							Description: "Check interval in seconds.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *MonitorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MonitorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state MonitorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing monitors data source")

	monitors, err := d.client.ListMonitors()
	if err != nil {
		resp.Diagnostics.AddError("Error listing monitors", err.Error())
		return
	}

	// Apply optional type filter.
	filterType := ""
	if !state.Type.IsNull() && !state.Type.IsUnknown() {
		filterType = state.Type.ValueString()
	}

	items := make([]MonitorListItem, 0, len(monitors))
	for _, m := range monitors {
		if filterType != "" && m.Type != filterType {
			continue
		}
		items = append(items, MonitorListItem{
			ID:              types.StringValue(m.ID),
			Type:            types.StringValue(m.Type),
			Name:            types.StringValue(m.Name),
			Status:          types.StringValue(m.Status),
			IsActive:        types.BoolValue(m.IsActive != 0),
			IntervalSeconds: types.Int64Value(m.IntervalSeconds),
		})
	}

	state.Monitors = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
