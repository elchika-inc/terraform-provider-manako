package monitor

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFlattenHTTPConfig(t *testing.T) {
	m := map[string]interface{}{
		"url":              "https://example.com",
		"method":           "GET",
		"expectedStatus":   float64(200),
		"timeoutMs":        float64(10000),
		"headers":          map[string]interface{}{"X-Custom": "value"},
		"keyword":          "OK",
		"keywordMustExist": true,
	}

	got := flattenHTTPConfig(m)

	if got.URL.ValueString() != "https://example.com" {
		t.Errorf("URL = %v, want https://example.com", got.URL.ValueString())
	}
	if got.Method.ValueString() != "GET" {
		t.Errorf("Method = %v, want GET", got.Method.ValueString())
	}
	if got.ExpectedStatus.ValueInt64() != 200 {
		t.Errorf("ExpectedStatus = %v, want 200", got.ExpectedStatus.ValueInt64())
	}
	if got.TimeoutMs.ValueInt64() != 10000 {
		t.Errorf("TimeoutMs = %v, want 10000", got.TimeoutMs.ValueInt64())
	}
	if got.Keyword.ValueString() != "OK" {
		t.Errorf("Keyword = %v, want OK", got.Keyword.ValueString())
	}
	if got.KeywordMustExist.ValueBool() != true {
		t.Errorf("KeywordMustExist = %v, want true", got.KeywordMustExist.ValueBool())
	}
	// Check headers map
	hdrs := got.Headers.Elements()
	if len(hdrs) != 1 {
		t.Errorf("Headers length = %v, want 1", len(hdrs))
	}
}

func TestFlattenHTTPConfigMissingOptionals(t *testing.T) {
	m := map[string]interface{}{
		"url":            "https://example.com",
		"method":         "POST",
		"expectedStatus": float64(201),
		"timeoutMs":      float64(5000),
	}

	got := flattenHTTPConfig(m)

	if !got.Keyword.IsNull() {
		t.Errorf("Keyword should be null when missing from API response")
	}
	if !got.KeywordMustExist.IsNull() {
		t.Errorf("KeywordMustExist should be null when missing")
	}
	if !got.Headers.IsNull() {
		t.Errorf("Headers should be null when missing")
	}
}

func TestFlattenTCPConfig(t *testing.T) {
	m := map[string]interface{}{
		"hostname":  "example.com",
		"port":      float64(443),
		"timeoutMs": float64(5000),
	}

	got := flattenTCPConfig(m)

	if got.Hostname.ValueString() != "example.com" {
		t.Errorf("Hostname = %v, want example.com", got.Hostname.ValueString())
	}
	if got.Port.ValueInt64() != 443 {
		t.Errorf("Port = %v, want 443", got.Port.ValueInt64())
	}
	if got.TimeoutMs.ValueInt64() != 5000 {
		t.Errorf("TimeoutMs = %v, want 5000", got.TimeoutMs.ValueInt64())
	}
}

func TestFlattenPingConfig(t *testing.T) {
	m := map[string]interface{}{
		"hostname":  "8.8.8.8",
		"timeoutMs": float64(3000),
		"port":      float64(80),
	}

	got := flattenPingConfig(m)

	if got.Hostname.ValueString() != "8.8.8.8" {
		t.Errorf("Hostname = %v, want 8.8.8.8", got.Hostname.ValueString())
	}
	if got.TimeoutMs.ValueInt64() != 3000 {
		t.Errorf("TimeoutMs = %v, want 3000", got.TimeoutMs.ValueInt64())
	}
	if got.Port.ValueInt64() != 80 {
		t.Errorf("Port = %v, want 80", got.Port.ValueInt64())
	}
}

func TestFlattenHeartbeatConfig(t *testing.T) {
	m := map[string]interface{}{
		"graceSeconds": float64(300),
	}

	got := flattenHeartbeatConfig(m)

	if got.GraceSeconds.ValueInt64() != 300 {
		t.Errorf("GraceSeconds = %v, want 300", got.GraceSeconds.ValueInt64())
	}
}

func TestFlattenWebChangeConfig(t *testing.T) {
	m := map[string]interface{}{
		"url":        "https://example.com",
		"checkType":  "text",
		"selector":   ".main",
		"changeMode": "tamper",
	}

	got := flattenWebChangeConfig(m)

	if got.URL.ValueString() != "https://example.com" {
		t.Errorf("URL = %v, want https://example.com", got.URL.ValueString())
	}
	if got.CheckType.ValueString() != "text" {
		t.Errorf("CheckType = %v, want text", got.CheckType.ValueString())
	}
	if got.Selector.ValueString() != ".main" {
		t.Errorf("Selector = %v, want .main", got.Selector.ValueString())
	}
	if got.ChangeMode.ValueString() != "tamper" {
		t.Errorf("ChangeMode = %v, want tamper", got.ChangeMode.ValueString())
	}
}

func TestFlattenWebChangeConfigNullSelector(t *testing.T) {
	m := map[string]interface{}{
		"url":        "https://example.com",
		"checkType":  "screenshot",
		"changeMode": "track",
	}

	got := flattenWebChangeConfig(m)

	if !got.Selector.IsNull() {
		t.Errorf("Selector should be null when missing from API response")
	}
}

func TestFlattenSSLConfig(t *testing.T) {
	m := map[string]interface{}{
		"hostname": "example.com",
		"warnDays": []interface{}{float64(30), float64(14)},
	}

	got := flattenSSLConfig(m)

	if got.Hostname.ValueString() != "example.com" {
		t.Errorf("Hostname = %v, want example.com", got.Hostname.ValueString())
	}
	elems := got.WarnDays.Elements()
	if len(elems) != 2 {
		t.Fatalf("WarnDays length = %v, want 2", len(elems))
	}
	if elems[0].(types.Int64).ValueInt64() != 30 {
		t.Errorf("WarnDays[0] = %v, want 30", elems[0])
	}
	if elems[1].(types.Int64).ValueInt64() != 14 {
		t.Errorf("WarnDays[1] = %v, want 14", elems[1])
	}
}

func TestFlattenSSLConfigNullWarnDays(t *testing.T) {
	m := map[string]interface{}{
		"hostname": "example.com",
	}

	got := flattenSSLConfig(m)

	if !got.WarnDays.IsNull() {
		t.Errorf("WarnDays should be null when missing from API response")
	}
}

func TestFlattenDomainConfig(t *testing.T) {
	m := map[string]interface{}{
		"domain":   "example.com",
		"warnDays": []interface{}{float64(60), float64(30)},
	}

	got := flattenDomainConfig(m)

	if got.Domain.ValueString() != "example.com" {
		t.Errorf("Domain = %v, want example.com", got.Domain.ValueString())
	}
	elems := got.WarnDays.Elements()
	if len(elems) != 2 {
		t.Fatalf("WarnDays length = %v, want 2", len(elems))
	}
	if elems[0].(types.Int64).ValueInt64() != 60 {
		t.Errorf("WarnDays[0] = %v, want 60", elems[0])
	}
	if elems[1].(types.Int64).ValueInt64() != 30 {
		t.Errorf("WarnDays[1] = %v, want 30", elems[1])
	}
}

func TestFlattenDomainConfigNullWarnDays(t *testing.T) {
	m := map[string]interface{}{
		"domain": "example.com",
	}

	got := flattenDomainConfig(m)

	if !got.WarnDays.IsNull() {
		t.Errorf("WarnDays should be null when missing")
	}
}
