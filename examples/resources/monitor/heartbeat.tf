resource "manako_monitor" "cron_job" {
  name             = "Nightly Backup"
  type             = "heartbeat"
  interval_seconds = 3600

  heartbeat_config {
    grace_seconds = 600
  }
}
