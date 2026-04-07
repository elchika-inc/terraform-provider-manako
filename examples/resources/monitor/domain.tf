resource "manako_monitor" "domain_expiry" {
  name = "Domain Expiry"
  type = "domain"

  domain_config {
    domain    = "example.com"
    warn_days = [30, 14, 7]
  }
}
