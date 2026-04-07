package monitor

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Config model types — one per monitor type, matching the schema AttrTypes.
// ---------------------------------------------------------------------------

// HTTPConfigModel holds the Terraform state for an http_config block.
type HTTPConfigModel struct {
	URL              types.String `tfsdk:"url"`
	Method           types.String `tfsdk:"method"`
	ExpectedStatus   types.Int64  `tfsdk:"expected_status"`
	TimeoutMs        types.Int64  `tfsdk:"timeout_ms"`
	Headers          types.Map    `tfsdk:"headers"`
	Keyword          types.String `tfsdk:"keyword"`
	KeywordMustExist types.Bool   `tfsdk:"keyword_must_exist"`
}

// TCPConfigModel holds the Terraform state for a tcp_config block.
type TCPConfigModel struct {
	Hostname  types.String `tfsdk:"hostname"`
	Port      types.Int64  `tfsdk:"port"`
	TimeoutMs types.Int64  `tfsdk:"timeout_ms"`
}

// PingConfigModel holds the Terraform state for a ping_config block.
type PingConfigModel struct {
	Hostname  types.String `tfsdk:"hostname"`
	TimeoutMs types.Int64  `tfsdk:"timeout_ms"`
	Port      types.Int64  `tfsdk:"port"`
}

// HeartbeatConfigModel holds the Terraform state for a heartbeat_config block.
type HeartbeatConfigModel struct {
	GraceSeconds types.Int64 `tfsdk:"grace_seconds"`
}

// WebChangeConfigModel holds the Terraform state for a webchange_config block.
type WebChangeConfigModel struct {
	URL        types.String `tfsdk:"url"`
	CheckType  types.String `tfsdk:"check_type"`
	Selector   types.String `tfsdk:"selector"`
	ChangeMode types.String `tfsdk:"change_mode"`
}

// SSLConfigModel holds the Terraform state for an ssl_config block.
type SSLConfigModel struct {
	Hostname types.String `tfsdk:"hostname"`
	WarnDays types.List   `tfsdk:"warn_days"`
}

// DomainConfigModel holds the Terraform state for a domain_config block.
type DomainConfigModel struct {
	Domain   types.String `tfsdk:"domain"`
	WarnDays types.List   `tfsdk:"warn_days"`
}

// ---------------------------------------------------------------------------
// Expand functions: HCL typed model → API config map (camelCase keys)
// ---------------------------------------------------------------------------

// expandHTTPConfig converts an HTTPConfigModel into the API request map.
func expandHTTPConfig(m HTTPConfigModel) map[string]interface{} {
	result := map[string]interface{}{
		"url":            m.URL.ValueString(),
		"method":         m.Method.ValueString(),
		"expectedStatus": m.ExpectedStatus.ValueInt64(),
		"timeoutMs":      m.TimeoutMs.ValueInt64(),
	}

	if !m.Headers.IsNull() && !m.Headers.IsUnknown() {
		hdrs := make(map[string]string, len(m.Headers.Elements()))
		for k, v := range m.Headers.Elements() {
			if sv, ok := v.(types.String); ok {
				hdrs[k] = sv.ValueString()
			}
		}
		result["headers"] = hdrs
	}

	if !m.Keyword.IsNull() && !m.Keyword.IsUnknown() {
		result["keyword"] = m.Keyword.ValueString()
	}

	if !m.KeywordMustExist.IsNull() && !m.KeywordMustExist.IsUnknown() {
		result["keywordMustExist"] = m.KeywordMustExist.ValueBool()
	}

	return result
}

// expandTCPConfig converts a TCPConfigModel into the API request map.
func expandTCPConfig(m TCPConfigModel) map[string]interface{} {
	return map[string]interface{}{
		"hostname":  m.Hostname.ValueString(),
		"port":      m.Port.ValueInt64(),
		"timeoutMs": m.TimeoutMs.ValueInt64(),
	}
}

// expandPingConfig converts a PingConfigModel into the API request map.
func expandPingConfig(m PingConfigModel) map[string]interface{} {
	return map[string]interface{}{
		"hostname":  m.Hostname.ValueString(),
		"timeoutMs": m.TimeoutMs.ValueInt64(),
		"port":      m.Port.ValueInt64(),
	}
}

// expandHeartbeatConfig converts a HeartbeatConfigModel into the API request map.
func expandHeartbeatConfig(m HeartbeatConfigModel) map[string]interface{} {
	return map[string]interface{}{
		"graceSeconds": m.GraceSeconds.ValueInt64(),
	}
}

// expandWebChangeConfig converts a WebChangeConfigModel into the API request map.
func expandWebChangeConfig(m WebChangeConfigModel) map[string]interface{} {
	result := map[string]interface{}{
		"url":        m.URL.ValueString(),
		"checkType":  m.CheckType.ValueString(),
		"changeMode": m.ChangeMode.ValueString(),
	}

	if !m.Selector.IsNull() && !m.Selector.IsUnknown() {
		result["selector"] = m.Selector.ValueString()
	}

	return result
}

// expandSSLConfig converts an SSLConfigModel into the API request map.
func expandSSLConfig(m SSLConfigModel) map[string]interface{} {
	result := map[string]interface{}{
		"hostname": m.Hostname.ValueString(),
	}

	if !m.WarnDays.IsNull() && !m.WarnDays.IsUnknown() {
		elems := m.WarnDays.Elements()
		days := make([]int64, 0, len(elems))
		for _, e := range elems {
			if iv, ok := e.(types.Int64); ok {
				days = append(days, iv.ValueInt64())
			}
		}
		result["warnDays"] = days
	}

	return result
}

// expandDomainConfig converts a DomainConfigModel into the API request map.
func expandDomainConfig(m DomainConfigModel) map[string]interface{} {
	result := map[string]interface{}{
		"domain": m.Domain.ValueString(),
	}

	if !m.WarnDays.IsNull() && !m.WarnDays.IsUnknown() {
		elems := m.WarnDays.Elements()
		days := make([]int64, 0, len(elems))
		for _, e := range elems {
			if iv, ok := e.(types.Int64); ok {
				days = append(days, iv.ValueInt64())
			}
		}
		result["warnDays"] = days
	}

	return result
}
