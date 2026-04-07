# terraform-provider-manako - AI Development Guide

## Project Overview

Terraform Provider for [Manako](https://manako.dev) — a Japanese-UI all-in-one monitoring SaaS.
Registry address: `registry.terraform.io/elchika-inc/manako`
Module path: `github.com/elchika-inc/terraform-provider-manako`

### Relationship to main monorepo

The Manako backend lives at `elchika-inc/manako` (private). This provider is a thin consumer of the public API (`/api/v1/*`) which requires an `mk_*` API Key. The provider holds zero business logic — it maps HCL config to API requests and API responses back to Terraform state.

### Resources / Data Sources

| Name | Kind | Description |
|---|---|---|
| `manako_monitor` | Resource | CRUD for any of 7 monitor types |
| `manako_monitor` | Data Source | Fetch a single monitor by ID |
| `manako_monitors` | Data Source | List monitors, optionally filtered by type |

---

## Architecture

3-layer design: **client → resource/data_source → provider**

```
main.go
  └─ internal/provider/provider.go     (Provider: Configure, Resources, DataSources)
       ├─ internal/client/client.go     (HTTP client: DoRequest, Monitor CRUD)
       └─ internal/resources/monitor/
            ├─ schema.go                (Terraform schema + AttrType helpers)
            ├─ resource.go             (CRUD + ImportState + expand/flatten wiring)
            ├─ config_expand.go        (HCL model → API camelCase map)
            ├─ config_flatten.go       (API camelCase map → HCL model)
            ├─ data_source.go          (manako_monitor data source)
            └─ data_source_list.go     (manako_monitors data source)
```

### Config expand/flatten pattern

The Manako API stores per-type configuration in a single `config` field (`map[string]interface{}`).
Terraform exposes this via 7 type-specific `SingleNestedBlock` entries (one per monitor type).

- **expand** (`config_expand.go`): reads the correct typed model from the HCL object, converts
  snake_case field names to camelCase API keys, returns `map[string]interface{}`.
- **flatten** (`config_flatten.go`): reads the API `config` map, converts camelCase keys to typed
  Go models with snake_case tfsdk tags. All non-active blocks are set to `types.ObjectNull`.

> API JSON numbers unmarshal as `float64`. The helper functions in `config_flatten.go`
> (`getInt64Value`, `getInt64ListValue`, etc.) handle this conversion.

### Field name mapping

| HCL (snake_case) | API (camelCase) |
|---|---|
| `interval_seconds` | `intervalSeconds` |
| `is_active` | `isActive` (int 0/1) |
| `expected_status` | `expectedStatus` |
| `timeout_ms` | `timeoutMs` |
| `keyword_must_exist` | `keywordMustExist` |
| `grace_seconds` | `graceSeconds` |
| `check_type` | `checkType` |
| `change_mode` | `changeMode` |
| `warn_days` | `warnDays` |
| `service_id` | `serviceId` |

Note: `isActive` in the API response is an **integer** (0 or 1), not a boolean.
`mapMonitorToState` converts it with `m.IsActive != 0`.

---

## File Structure

```
main.go                              Entry point; injects version via ldflags
go.mod / go.sum                      Go modules (go 1.26)
Makefile                             Dev commands (build, test, testacc, lint, install)
.goreleaser.yml                      Release pipeline: builds for linux/darwin/windows x amd64/arm64
.github/workflows/ci.yml             CI: build + vet + test + lint on push/PR to main
.github/workflows/release.yml        Release: GoReleaser triggered by v* tags
examples/                            Example .tf files for registry documentation
templates/                           tfplugindocs templates
internal/
  client/
    client.go                        HTTP client, retry with exponential backoff, Monitor CRUD
    client_test.go                   Unit tests for client helpers
  provider/
    provider.go                      ManakoProvider: Schema, Configure, registers resources/data sources
  resources/monitor/
    schema.go                        MonitorResourceModel, monitorResourceSchema(), block helpers, AttrType maps
    resource.go                      Create/Read/Update/Delete/ImportState + expandConfig + flattenConfigToState
    config_expand.go                 Expand functions + typed config model structs
    config_flatten.go                Flatten functions + low-level map helper functions
    data_source.go                   manako_monitor data source (single monitor by ID)
    data_source_list.go              manako_monitors data source (list with optional type filter)
    resource_test.go                 Acceptance tests for HTTP/TCP/SSL monitor lifecycle
    config_expand_test.go            Unit tests for expand functions
    config_flatten_test.go           Unit tests for flatten functions
```

---

## Key Conventions

### Go

- Module: `github.com/elchika-inc/terraform-provider-manako`
- Go version: 1.26 (see `go.mod` and CI workflows)
- Test files: `*_test.go`, package `<pkg>_test` for acceptance tests, `<pkg>` for unit tests
- Interface compliance asserted at compile time: `var _ resource.Resource = &MonitorResource{}`

### Terraform Plugin Framework patterns

This provider uses **Plugin Framework** (not SDKv2). Always import from:
- `github.com/hashicorp/terraform-plugin-framework/*`
- `github.com/hashicorp/terraform-plugin-framework-validators/*`

Do NOT use `github.com/hashicorp/terraform-plugin-sdk/v2` patterns.

Key patterns in use:

| Pattern | Location |
|---|---|
| `schema.SingleNestedBlock` for config per monitor type | `schema.go` |
| `types.ObjectNull(attrTypes)` to null unused config blocks | `resource.go` `flattenConfigToState` |
| `types.ObjectValueFrom(ctx, attrTypes, model)` to populate active block | `resource.go` |
| `resource.ImportStatePassthroughID` for import | `resource.go` |
| `resp.State.RemoveResource(ctx)` on 404 (drift detection) | `resource.go` Read |
| `planmodifier.String` with `RequiresReplace` on `type` field | `schema.go` |

### Config blocks: exactly one must match `type`

`monitorResourceSchema()` declares all 7 `*_config` blocks in the schema.
`expandConfig` validates at runtime that the block matching the declared `type` is present and non-null.
`flattenConfigToState` always nulls all 7 blocks first, then sets only the active one.

### 2-step create for `is_active=false`

The create API does not accept `isActive`. If `is_active = false` is set in the plan, the provider:
1. Creates the monitor (always starts active)
2. Immediately calls UpdateMonitor with `isActive: false`

Do not skip this pattern. Omitting it causes permanent drift.

### Drift detection

In `Read`, a 404 API error calls `resp.State.RemoveResource(ctx)` to remove the resource from state
without returning an error. This allows `terraform apply` to recreate it.

In `Delete`, a 404 is silently tolerated (already deleted is fine).

### Import

`terraform import manako_monitor.example <ULID>` — passthrough ID only. No extra parsing needed.

### Rate limiting & retries

`client.DoRequest` retries up to 3 times with exponential backoff starting at 500ms.
On 429, it respects the `Retry-After` header if present.

---

## Commands

```bash
make build       # go build -o terraform-provider-manako
make install     # build + install to ~/.terraform.d/plugins/registry.terraform.io/elchika-inc/manako/0.1.0/<OS_ARCH>
make test        # go test ./... -v -count=1  (unit tests only)
make testacc     # TF_ACC=1 go test ./... -v -count=1 -timeout 120m  (acceptance tests)
make lint        # golangci-lint run ./...
make generate    # go generate ./...
```

---

## Testing

### Unit tests (no credentials needed)

Test files: `config_expand_test.go`, `config_flatten_test.go`, `client_test.go`

These verify config mapping correctness in isolation (no real API calls).
Run with `make test`.

### Acceptance tests (require live API)

Test file: `resource_test.go` (package `monitor_test`)

`testAccPreCheck(t)` calls `t.Skip` if `MANAKO_API_KEY` is not set — so `make test` never runs them by accident.

To run acceptance tests:
```bash
export MANAKO_API_KEY=mk_...
make testacc
```

Acceptance tests cover: HTTP monitor (create → import → update), TCP monitor, SSL monitor.

---

## Release Process

1. Tag the commit: `git tag v0.x.y && git push origin v0.x.y`
2. `release.yml` triggers GoReleaser automatically
3. GoReleaser builds binaries for 4 targets (linux/darwin x amd64/arm64), creates SHA256SUMS, signs with GPG
4. GitHub Release is created; Terraform Registry picks it up automatically

**Required GitHub Secrets:**
- `GPG_PRIVATE_KEY` — armored private key for signing
- `PASSPHRASE` — passphrase for the GPG key

The `version` variable in `main.go` is injected at build time via ldflags:
`-X main.version={{.Version}}`

---

## Adding New Resources

Example: adding `notification_channel` in the future.

1. **Add client methods** in `internal/client/client.go`:
   - Add `NotificationChannel` struct with JSON tags (camelCase)
   - Add `Create/Get/Update/Delete/List` methods

2. **Create resource package** `internal/resources/notification_channel/`:
   - `schema.go` — model struct, `notificationChannelResourceSchema()`, AttrType helpers
   - `resource.go` — `NotificationChannelResource` with CRUD + ImportState + interface assertion
   - `config_expand.go` / `config_flatten.go` — if the resource has type-specific sub-configs
   - `resource_test.go` — acceptance tests

3. **Register in provider** `internal/provider/provider.go`:
   - Add `notification_channel.NewNotificationChannelResource` to `Resources()`
   - Add data sources to `DataSources()` as needed

4. Follow all patterns: interface assertion, drift detection on 404, ImportStatePassthroughID.

---

## What NOT to Do

- Do not hardcode the API base URL. Use provider config `base_url` (default: `https://api.manako.dev`).
- Do not use SDKv2 (`terraform-plugin-sdk/v2`) patterns. This is Plugin Framework only.
- Do not skip the 2-step create pattern for fields the create API does not accept (`isActive`).
- Do not assume JSON numbers are `int64`. After `json.Unmarshal` into `interface{}` they are `float64` — use the helper functions in `config_flatten.go`.
- Do not set only the active config block in `flattenConfigToState`. Always null all 7 blocks first to avoid stale state.
- Do not return an error from `Read` on 404. Call `resp.State.RemoveResource(ctx)` instead.
- Do not add business logic to the provider. It is a thin API wrapper only.
- Do not generate files and commit them without running `go mod tidy` first.
