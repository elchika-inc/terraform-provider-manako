resource "manako_monitor" "ssl_cert" {
  name = "SSL Certificate"
  type = "ssl"

  ssl_config {
    hostname  = "example.com"
    warn_days = [30, 14, 7, 1]
  }
}
