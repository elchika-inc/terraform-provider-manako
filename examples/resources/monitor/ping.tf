resource "manako_monitor" "server_ping" {
  name = "Web Server Ping"
  type = "ping"

  ping_config {
    hostname = "web.example.com"
  }
}
