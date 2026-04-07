package monitor

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MonitorResourceModel is the Terraform state model for a manako_monitor resource.
type MonitorResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Type                   types.String `tfsdk:"type"`
	Name                   types.String `tfsdk:"name"`
	IntervalSeconds        types.Int64  `tfsdk:"interval_seconds"`
	IsActive               types.Bool   `tfsdk:"is_active"`
	ServiceID              types.String `tfsdk:"service_id"`
	Status                 types.String `tfsdk:"status"`
	LastCheckedAt          types.String `tfsdk:"last_checked_at"`
	LastErrorMessage       types.String `tfsdk:"last_error_message"`
	MaintenanceUntil       types.String `tfsdk:"maintenance_until"`
	AutoResolveMaintenance types.Bool   `tfsdk:"auto_resolve_maintenance"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	// Type-specific config blocks (exactly one must be set)
	HTTPConfig      types.Object `tfsdk:"http_config"`
	TCPConfig       types.Object `tfsdk:"tcp_config"`
	PingConfig      types.Object `tfsdk:"ping_config"`
	HeartbeatConfig types.Object `tfsdk:"heartbeat_config"`
	WebChangeConfig types.Object `tfsdk:"webchange_config"`
	SSLConfig       types.Object `tfsdk:"ssl_config"`
	DomainConfig    types.Object `tfsdk:"domain_config"`
}

// monitorResourceSchema returns the full Terraform schema for the manako_monitor resource.
func monitorResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a Manako monitor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the monitor.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Monitor type. One of: http, tcp, ping, heartbeat, webchange, ssl, domain.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("http", "tcp", "ping", "heartbeat", "webchange", "ssl", "domain"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Display name of the monitor.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(100),
				},
			},
			"interval_seconds": schema.Int64Attribute{
				Description: "Check interval in seconds. Minimum 60, maximum 86400. Defaults to 300.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(300),
				Validators: []validator.Int64{
					int64validator.Between(60, 86400),
				},
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the monitor is active. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"service_id": schema.StringAttribute{
				Description: "Optional service group ID this monitor belongs to.",
				Optional:    true,
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the monitor (e.g. up, down, unknown).",
				Computed:    true,
			},
			"last_checked_at": schema.StringAttribute{
				Description: "Timestamp of the last check.",
				Computed:    true,
			},
			"last_error_message": schema.StringAttribute{
				Description: "Last error message if the monitor is down.",
				Computed:    true,
			},
			"maintenance_until": schema.StringAttribute{
				Description: "Timestamp until which the monitor is in maintenance mode.",
				Computed:    true,
			},
			"auto_resolve_maintenance": schema.BoolAttribute{
				Description: "Whether maintenance mode is automatically resolved.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the monitor was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the monitor was last updated.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"http_config":      httpConfigBlock(),
			"tcp_config":       tcpConfigBlock(),
			"ping_config":      pingConfigBlock(),
			"heartbeat_config": heartbeatConfigBlock(),
			"webchange_config": webchangeConfigBlock(),
			"ssl_config":       sslConfigBlock(),
			"domain_config":    domainConfigBlock(),
		},
	}
}

// ---------------------------------------------------------------------------
// Config block schema definitions
// ---------------------------------------------------------------------------

func httpConfigBlock() schema.Block {
	return schema.SingleNestedBlock{
		Description: "Configuration for HTTP monitors.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "URL to monitor.",
				Required:    true,
			},
			"method": schema.StringAttribute{
				Description: "HTTP method. One of: GET, HEAD, POST. Defaults to GET.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("GET"),
				Validators: []validator.String{
					stringvalidator.OneOf("GET", "HEAD", "POST"),
				},
			},
			"expected_status": schema.Int64Attribute{
				Description: "Expected HTTP status code. Defaults to 200.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(200),
				Validators: []validator.Int64{
					int64validator.Between(100, 599),
				},
			},
			"timeout_ms": schema.Int64Attribute{
				Description: "Timeout in milliseconds. Defaults to 10000.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10000),
				Validators: []validator.Int64{
					int64validator.Between(1000, 30000),
				},
			},
			"headers": schema.MapAttribute{
				Description: "Additional HTTP headers to send.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"keyword": schema.StringAttribute{
				Description: "Keyword to search for in the response body.",
				Optional:    true,
			},
			"keyword_must_exist": schema.BoolAttribute{
				Description: "Whether the keyword must exist (true) or must not exist (false) in the response.",
				Optional:    true,
			},
		},
	}
}

func tcpConfigBlock() schema.Block {
	return schema.SingleNestedBlock{
		Description: "Configuration for TCP monitors.",
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description: "Hostname or IP address to connect to.",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "TCP port number (1-65535).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"timeout_ms": schema.Int64Attribute{
				Description: "Timeout in milliseconds. Defaults to 10000.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10000),
				Validators: []validator.Int64{
					int64validator.Between(1000, 30000),
				},
			},
		},
	}
}

func pingConfigBlock() schema.Block {
	return schema.SingleNestedBlock{
		Description: "Configuration for Ping monitors.",
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description: "Hostname or IP address to ping.",
				Required:    true,
			},
			"timeout_ms": schema.Int64Attribute{
				Description: "Timeout in milliseconds. Defaults to 10000.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10000),
				Validators: []validator.Int64{
					int64validator.Between(1000, 30000),
				},
			},
			"port": schema.Int64Attribute{
				Description: "Port number (1-65535). Defaults to 443.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(443),
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
		},
	}
}

func heartbeatConfigBlock() schema.Block {
	return schema.SingleNestedBlock{
		Description: "Configuration for Heartbeat monitors.",
		Attributes: map[string]schema.Attribute{
			"grace_seconds": schema.Int64Attribute{
				Description: "Grace period in seconds before alerting. Defaults to 300.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(300),
				Validators: []validator.Int64{
					int64validator.Between(60, 86400),
				},
			},
		},
	}
}

func webchangeConfigBlock() schema.Block {
	return schema.SingleNestedBlock{
		Description: "Configuration for WebChange monitors.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "URL to monitor for changes.",
				Required:    true,
			},
			"check_type": schema.StringAttribute{
				Description: "Check type. One of: text, screenshot, both.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("text", "screenshot", "both"),
				},
			},
			"selector": schema.StringAttribute{
				Description: "CSS selector to scope the change detection.",
				Optional:    true,
			},
			"change_mode": schema.StringAttribute{
				Description: "Change detection mode. One of: tamper, track. Defaults to tamper.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("tamper"),
				Validators: []validator.String{
					stringvalidator.OneOf("tamper", "track"),
				},
			},
		},
	}
}

func sslConfigBlock() schema.Block {
	return schema.SingleNestedBlock{
		Description: "Configuration for SSL certificate monitors.",
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description: "Hostname to check the SSL certificate for.",
				Required:    true,
			},
			"warn_days": schema.ListAttribute{
				Description: "List of days-before-expiry thresholds for warnings.",
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(0),
				},
			},
		},
	}
}

func domainConfigBlock() schema.Block {
	return schema.SingleNestedBlock{
		Description: "Configuration for domain expiry monitors.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "Domain name to monitor.",
				Required:    true,
			},
			"warn_days": schema.ListAttribute{
				Description: "List of days-before-expiry thresholds for warnings.",
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(0),
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// AttrType helper functions — used when converting between API types and
// Terraform object values via types.ObjectValue / types.ObjectNull.
// ---------------------------------------------------------------------------

func httpConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"url":                types.StringType,
		"method":             types.StringType,
		"expected_status":    types.Int64Type,
		"timeout_ms":         types.Int64Type,
		"headers":            types.MapType{ElemType: types.StringType},
		"keyword":            types.StringType,
		"keyword_must_exist": types.BoolType,
	}
}

func tcpConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"hostname":   types.StringType,
		"port":       types.Int64Type,
		"timeout_ms": types.Int64Type,
	}
}

func pingConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"hostname":   types.StringType,
		"timeout_ms": types.Int64Type,
		"port":       types.Int64Type,
	}
}

func heartbeatConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"grace_seconds": types.Int64Type,
	}
}

func webchangeConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"url":         types.StringType,
		"check_type":  types.StringType,
		"selector":    types.StringType,
		"change_mode": types.StringType,
	}
}

func sslConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"hostname":  types.StringType,
		"warn_days": types.ListType{ElemType: types.Int64Type},
	}
}

func domainConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"domain":    types.StringType,
		"warn_days": types.ListType{ElemType: types.Int64Type},
	}
}

