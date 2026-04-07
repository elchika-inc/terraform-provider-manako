# Terraform Provider for Manako

Manage [Manako](https://manako.dev) monitoring resources with Terraform.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.26 (for building from source)

## Usage

```hcl
terraform {
  required_providers {
    manako = {
      source  = "elchika-inc/manako"
      version = "~> 0.1"
    }
  }
}

provider "manako" {
  # Set MANAKO_API_KEY environment variable, or:
  # api_key = var.manako_api_key
}

resource "manako_monitor" "api_health" {
  name             = "API Health Check"
  type             = "http"
  interval_seconds = 60

  http_config {
    url             = "https://api.example.com/health"
    expected_status = 200
  }
}
```

## Authentication

Get your API key from the [Manako Dashboard](https://app.manako.dev) under Settings > API Keys.

Set it via environment variable:

```bash
export MANAKO_API_KEY="mk_your_api_key_here"
```

Or in the provider block (not recommended for version control):

```hcl
provider "manako" {
  api_key = var.manako_api_key
}
```

## Resources

- `manako_monitor` — Manages a monitor (HTTP, TCP, Ping, Heartbeat, WebChange, SSL, Domain)

## Data Sources

- `manako_monitor` — Fetch a single monitor by ID
- `manako_monitors` — List monitors, optionally filtered by type

## Development

```bash
make build      # Build the provider
make test       # Run unit tests
make testacc    # Run acceptance tests (requires MANAKO_API_KEY)
make lint       # Run linter
```

## License

[MIT](LICENSE)
