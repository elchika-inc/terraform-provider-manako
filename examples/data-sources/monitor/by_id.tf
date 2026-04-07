data "manako_monitor" "existing" {
  id = "01HXYZ..."
}

output "monitor_status" {
  value = data.manako_monitor.existing.status
}
