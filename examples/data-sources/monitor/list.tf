data "manako_monitors" "http_only" {
  type = "http"
}

output "http_monitor_count" {
  value = length(data.manako_monitors.http_only.monitors)
}
