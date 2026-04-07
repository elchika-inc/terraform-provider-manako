package monitor_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/elchika-inc/terraform-provider-manako/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"manako": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("MANAKO_API_KEY") == "" {
		t.Skip("MANAKO_API_KEY not set, skipping acceptance test")
	}
}

func TestAccMonitorResource_HTTP(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_HTTP("tf-acc-http-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("manako_monitor.test", "type", "http"),
					resource.TestCheckResourceAttr("manako_monitor.test", "name", "tf-acc-http-test"),
					resource.TestCheckResourceAttrSet("manako_monitor.test", "id"),
					resource.TestCheckResourceAttrSet("manako_monitor.test", "status"),
					resource.TestCheckResourceAttrSet("manako_monitor.test", "created_at"),
				),
			},
			{
				ResourceName:      "manako_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitorConfig_HTTP("tf-acc-http-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("manako_monitor.test", "name", "tf-acc-http-updated"),
				),
			},
		},
	})
}

func TestAccMonitorResource_TCP(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_TCP(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("manako_monitor.test", "type", "tcp"),
					resource.TestCheckResourceAttrSet("manako_monitor.test", "id"),
				),
			},
		},
	})
}

func TestAccMonitorResource_SSL(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_SSL(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("manako_monitor.test", "type", "ssl"),
					resource.TestCheckResourceAttrSet("manako_monitor.test", "id"),
				),
			},
		},
	})
}

func testAccMonitorConfig_HTTP(name string) string {
	return fmt.Sprintf(`
resource "manako_monitor" "test" {
  name             = %[1]q
  type             = "http"
  interval_seconds = 300

  http_config {
    url             = "https://httpstat.us/200"
    method          = "GET"
    expected_status = 200
    timeout_ms      = 10000
  }
}
`, name)
}

func testAccMonitorConfig_TCP() string {
	return `
resource "manako_monitor" "test" {
  name = "tf-acc-tcp-test"
  type = "tcp"

  tcp_config {
    hostname   = "google.com"
    port       = 443
    timeout_ms = 10000
  }
}
`
}

func testAccMonitorConfig_SSL() string {
	return `
resource "manako_monitor" "test" {
  name = "tf-acc-ssl-test"
  type = "ssl"

  ssl_config {
    hostname  = "google.com"
    warn_days = [30, 14, 7]
  }
}
`
}
