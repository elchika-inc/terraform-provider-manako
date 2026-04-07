package monitor

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Helper functions for safely extracting typed values from API response maps.
// API JSON numbers are always float64 after json.Unmarshal.
// ---------------------------------------------------------------------------

// getStringValue returns a types.String from the map, or StringNull if absent.
func getStringValue(m map[string]interface{}, key string) types.String {
	v, ok := m[key]
	if !ok || v == nil {
		return types.StringNull()
	}
	if s, ok := v.(string); ok {
		return types.StringValue(s)
	}
	return types.StringNull()
}

// getInt64Value returns a types.Int64 from the map, or Int64Null if absent.
// Handles both float64 (from JSON) and int64 raw values.
func getInt64Value(m map[string]interface{}, key string) types.Int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return types.Int64Null()
	}
	switch n := v.(type) {
	case float64:
		return types.Int64Value(int64(n))
	case int64:
		return types.Int64Value(n)
	case int:
		return types.Int64Value(int64(n))
	}
	return types.Int64Null()
}

// getBoolValue returns a types.Bool from the map, or BoolNull if absent.
func getBoolValue(m map[string]interface{}, key string) types.Bool {
	v, ok := m[key]
	if !ok || v == nil {
		return types.BoolNull()
	}
	if b, ok := v.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}

// getMapStringValue returns a types.Map of strings, or MapNull if absent.
func getMapStringValue(m map[string]interface{}, key string) types.Map {
	v, ok := m[key]
	if !ok || v == nil {
		return types.MapNull(types.StringType)
	}
	raw, ok := v.(map[string]interface{})
	if !ok {
		return types.MapNull(types.StringType)
	}
	elems := make(map[string]attr.Value, len(raw))
	for k, val := range raw {
		if s, ok := val.(string); ok {
			elems[k] = types.StringValue(s)
		}
	}
	return types.MapValueMust(types.StringType, elems)
}

// getInt64ListValue returns a types.List of Int64, or ListNull if absent.
// The raw value is expected to be []interface{} with float64 elements (from JSON).
func getInt64ListValue(m map[string]interface{}, key string) types.List {
	v, ok := m[key]
	if !ok || v == nil {
		return types.ListNull(types.Int64Type)
	}
	raw, ok := v.([]interface{})
	if !ok {
		return types.ListNull(types.Int64Type)
	}
	elems := make([]attr.Value, 0, len(raw))
	for _, item := range raw {
		switch n := item.(type) {
		case float64:
			elems = append(elems, types.Int64Value(int64(n)))
		case int64:
			elems = append(elems, types.Int64Value(n))
		case int:
			elems = append(elems, types.Int64Value(int64(n)))
		}
	}
	return types.ListValueMust(types.Int64Type, elems)
}

// ---------------------------------------------------------------------------
// Flatten functions: API config map (camelCase) → HCL typed model
// ---------------------------------------------------------------------------

// flattenHTTPConfig converts an API config map into an HTTPConfigModel.
func flattenHTTPConfig(m map[string]interface{}) HTTPConfigModel {
	return HTTPConfigModel{
		URL:              getStringValue(m, "url"),
		Method:           getStringValue(m, "method"),
		ExpectedStatus:   getInt64Value(m, "expectedStatus"),
		TimeoutMs:        getInt64Value(m, "timeoutMs"),
		Headers:          getMapStringValue(m, "headers"),
		Keyword:          getStringValue(m, "keyword"),
		KeywordMustExist: getBoolValue(m, "keywordMustExist"),
	}
}

// flattenTCPConfig converts an API config map into a TCPConfigModel.
func flattenTCPConfig(m map[string]interface{}) TCPConfigModel {
	return TCPConfigModel{
		Hostname:  getStringValue(m, "hostname"),
		Port:      getInt64Value(m, "port"),
		TimeoutMs: getInt64Value(m, "timeoutMs"),
	}
}

// flattenPingConfig converts an API config map into a PingConfigModel.
func flattenPingConfig(m map[string]interface{}) PingConfigModel {
	return PingConfigModel{
		Hostname:  getStringValue(m, "hostname"),
		TimeoutMs: getInt64Value(m, "timeoutMs"),
		Port:      getInt64Value(m, "port"),
	}
}

// flattenHeartbeatConfig converts an API config map into a HeartbeatConfigModel.
func flattenHeartbeatConfig(m map[string]interface{}) HeartbeatConfigModel {
	return HeartbeatConfigModel{
		GraceSeconds: getInt64Value(m, "graceSeconds"),
	}
}

// flattenWebChangeConfig converts an API config map into a WebChangeConfigModel.
func flattenWebChangeConfig(m map[string]interface{}) WebChangeConfigModel {
	return WebChangeConfigModel{
		URL:        getStringValue(m, "url"),
		CheckType:  getStringValue(m, "checkType"),
		Selector:   getStringValue(m, "selector"),
		ChangeMode: getStringValue(m, "changeMode"),
	}
}

// flattenSSLConfig converts an API config map into an SSLConfigModel.
func flattenSSLConfig(m map[string]interface{}) SSLConfigModel {
	return SSLConfigModel{
		Hostname: getStringValue(m, "hostname"),
		WarnDays: getInt64ListValue(m, "warnDays"),
	}
}

// flattenDomainConfig converts an API config map into a DomainConfigModel.
func flattenDomainConfig(m map[string]interface{}) DomainConfigModel {
	return DomainConfigModel{
		Domain:   getStringValue(m, "domain"),
		WarnDays: getInt64ListValue(m, "warnDays"),
	}
}
