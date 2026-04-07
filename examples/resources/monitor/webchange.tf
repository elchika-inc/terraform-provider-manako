resource "manako_monitor" "pricing_page" {
  name = "Pricing Page Change"
  type = "webchange"

  webchange_config {
    url        = "https://example.com/pricing"
    check_type = "text"
    selector   = ".pricing-table"
  }
}
