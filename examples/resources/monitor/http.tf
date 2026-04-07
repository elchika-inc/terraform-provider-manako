resource "manako_monitor" "api_health" {
  name             = "API Health Check"
  type             = "http"
  interval_seconds = 60

  http_config {
    url             = "https://api.example.com/health"
    method          = "GET"
    expected_status = 200
    keyword         = "ok"
    keyword_must_exist = true
    timeout_ms      = 5000
  }
}
