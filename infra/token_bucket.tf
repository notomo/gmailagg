
resource "google_storage_bucket" "token" {
  name                     = "gmailagg-token"
  storage_class            = "STANDARD"
  public_access_prevention = "enforced"
  location                 = var.region
  project                  = var.project_id
}
