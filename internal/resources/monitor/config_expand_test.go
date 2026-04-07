package monitor

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandHTTPConfig(t *testing.T) {
	headers, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"X-Custom": types.StringValue("value"),
	})
	m := HTTPConfigModel{
		URL:              types.StringValue("https://example.com"),
		Method:           types.StringValue("GET"),
		ExpectedStatus:   types.Int64Value(200),
		TimeoutMs:        types.Int64Value(10000),
		Headers:          headers,
		Keyword:          types.StringValue("OK"),
		KeywordMustExist: types.BoolValue(true),
	}

	got := expandHTTPConfig(m)

	if got["url"] != "https://example.com" {
		t.Errorf("url = %v, want https://example.com", got["url"])
	}
	if got["method"] != "GET" {
		t.Errorf("method = %v, want GET", got["method"])
	}
	if got["expectedStatus"] != int64(200) {
		t.Errorf("expectedStatus = %v, want 200", got["expectedStatus"])
	}
	if got["timeoutMs"] != int64(10000) {
		t.Errorf("timeoutMs = %v, want 10000", got["timeoutMs"])
	}
	hdrs, ok := got["headers"].(map[string]string)
	if !ok {
		t.Fatalf("headers is not map[string]string: %T", got["headers"])
	}
	if hdrs["X-Custom"] != "value" {
		t.Errorf("headers[X-Custom] = %v, want value", hdrs["X-Custom"])
	}
	if got["keyword"] != "OK" {
		t.Errorf("keyword = %v, want OK", got["keyword"])
	}
	if got["keywordMustExist"] != true {
		t.Errorf("keywordMustExist = %v, want true", got["keywordMustExist"])
	}
}

func TestExpandHTTPConfigNullOptionals(t *testing.T) {
	headers, _ := types.MapValue(types.StringType, map[string]attr.Value{})
	m := HTTPConfigModel{
		URL:              types.StringValue("https://example.com"),
		Method:           types.StringValue("GET"),
		ExpectedStatus:   types.Int64Value(200),
		TimeoutMs:        types.Int64Value(10000),
		Headers:          types.MapNull(types.StringType),
		Keyword:          types.StringNull(),
		KeywordMustExist: types.BoolNull(),
	}
	_ = headers

	got := expandHTTPConfig(m)

	if _, ok := got["keyword"]; ok {
		t.Errorf("keyword should not be present when null")
	}
	if _, ok := got["keywordMustExist"]; ok {
		t.Errorf("keywordMustExist should not be present when null")
	}
	if _, ok := got["headers"]; ok {
		t.Errorf("headers should not be present when null")
	}
}

func TestExpandTCPConfig(t *testing.T) {
	m := TCPConfigModel{
		Hostname:  types.StringValue("example.com"),
		Port:      types.Int64Value(443),
		TimeoutMs: types.Int64Value(5000),
	}

	got := expandTCPConfig(m)

	if got["hostname"] != "example.com" {
		t.Errorf("hostname = %v, want example.com", got["hostname"])
	}
	if got["port"] != int64(443) {
		t.Errorf("port = %v, want 443", got["port"])
	}
	if got["timeoutMs"] != int64(5000) {
		t.Errorf("timeoutMs = %v, want 5000", got["timeoutMs"])
	}
}

func TestExpandPingConfig(t *testing.T) {
	m := PingConfigModel{
		Hostname:  types.StringValue("8.8.8.8"),
		TimeoutMs: types.Int64Value(3000),
		Port:      types.Int64Value(80),
	}

	got := expandPingConfig(m)

	if got["hostname"] != "8.8.8.8" {
		t.Errorf("hostname = %v, want 8.8.8.8", got["hostname"])
	}
	if got["timeoutMs"] != int64(3000) {
		t.Errorf("timeoutMs = %v, want 3000", got["timeoutMs"])
	}
	if got["port"] != int64(80) {
		t.Errorf("port = %v, want 80", got["port"])
	}
}

func TestExpandHeartbeatConfig(t *testing.T) {
	m := HeartbeatConfigModel{
		GraceSeconds: types.Int64Value(300),
	}

	got := expandHeartbeatConfig(m)

	if got["graceSeconds"] != int64(300) {
		t.Errorf("graceSeconds = %v, want 300", got["graceSeconds"])
	}
}

func TestExpandWebChangeConfig(t *testing.T) {
	m := WebChangeConfigModel{
		URL:        types.StringValue("https://example.com"),
		CheckType:  types.StringValue("text"),
		Selector:   types.StringValue(".main"),
		ChangeMode: types.StringValue("tamper"),
	}

	got := expandWebChangeConfig(m)

	if got["url"] != "https://example.com" {
		t.Errorf("url = %v, want https://example.com", got["url"])
	}
	if got["checkType"] != "text" {
		t.Errorf("checkType = %v, want text", got["checkType"])
	}
	if got["selector"] != ".main" {
		t.Errorf("selector = %v, want .main", got["selector"])
	}
	if got["changeMode"] != "tamper" {
		t.Errorf("changeMode = %v, want tamper", got["changeMode"])
	}
}

func TestExpandWebChangeConfigNullSelector(t *testing.T) {
	m := WebChangeConfigModel{
		URL:        types.StringValue("https://example.com"),
		CheckType:  types.StringValue("text"),
		Selector:   types.StringNull(),
		ChangeMode: types.StringValue("tamper"),
	}

	got := expandWebChangeConfig(m)

	if _, ok := got["selector"]; ok {
		t.Errorf("selector should not be present when null")
	}
}

func TestExpandSSLConfig(t *testing.T) {
	warnDays, _ := types.ListValue(types.Int64Type, []attr.Value{
		types.Int64Value(30),
		types.Int64Value(14),
	})
	m := SSLConfigModel{
		Hostname: types.StringValue("example.com"),
		WarnDays: warnDays,
	}

	got := expandSSLConfig(m)

	if got["hostname"] != "example.com" {
		t.Errorf("hostname = %v, want example.com", got["hostname"])
	}
	days, ok := got["warnDays"].([]int64)
	if !ok {
		t.Fatalf("warnDays is not []int64: %T", got["warnDays"])
	}
	if len(days) != 2 || days[0] != 30 || days[1] != 14 {
		t.Errorf("warnDays = %v, want [30 14]", days)
	}
}

func TestExpandSSLConfigNullWarnDays(t *testing.T) {
	m := SSLConfigModel{
		Hostname: types.StringValue("example.com"),
		WarnDays: types.ListNull(types.Int64Type),
	}

	got := expandSSLConfig(m)

	if _, ok := got["warnDays"]; ok {
		t.Errorf("warnDays should not be present when null")
	}
}

func TestExpandDomainConfig(t *testing.T) {
	warnDays, _ := types.ListValue(types.Int64Type, []attr.Value{
		types.Int64Value(60),
		types.Int64Value(30),
	})
	m := DomainConfigModel{
		Domain:   types.StringValue("example.com"),
		WarnDays: warnDays,
	}

	got := expandDomainConfig(m)

	if got["domain"] != "example.com" {
		t.Errorf("domain = %v, want example.com", got["domain"])
	}
	days, ok := got["warnDays"].([]int64)
	if !ok {
		t.Fatalf("warnDays is not []int64: %T", got["warnDays"])
	}
	if len(days) != 2 || days[0] != 60 || days[1] != 30 {
		t.Errorf("warnDays = %v, want [60 30]", days)
	}
}
