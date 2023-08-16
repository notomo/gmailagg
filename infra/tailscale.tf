resource "tailscale_tailnet_key" "influxdb_instance_onetime_auth" {
  ephemeral     = true
  preauthorized = true
  reusable      = false
  expiry        = 600
}

data "template_file" "cloud_init" {
  template = file("cloud-init.yaml")

  vars = {
    tailscale_auth_key = tailscale_tailnet_key.influxdb_instance_onetime_auth.key
  }
}
