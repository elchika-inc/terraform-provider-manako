resource "manako_monitor" "db_check" {
  name = "PostgreSQL"
  type = "tcp"

  tcp_config {
    hostname   = "db.example.com"
    port       = 5432
    timeout_ms = 10000
  }
}
