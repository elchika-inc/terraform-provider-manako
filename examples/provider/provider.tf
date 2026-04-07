terraform {
  required_providers {
    manako = {
      source  = "elchika-inc/manako"
      version = "~> 0.1"
    }
  }
}

provider "manako" {
  # Set via MANAKO_API_KEY environment variable, or:
  # api_key = var.manako_api_key
}
