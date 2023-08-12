resource "tailscale_tailnet_key" "default" {
  ephemeral     = true
  preauthorized = true
  reusable      = false
  expiry        = 600
}

data "template_file" "cloud-init" {
  template = file("cloud-init.yaml")

  vars = {
    tailscale_auth_key = tailscale_tailnet_key.default.key
  }
}
