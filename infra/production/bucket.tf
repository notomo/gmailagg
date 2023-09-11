
resource "google_storage_bucket" "gmailagg_oauth" {
  name                        = "gmailagg-oauth"
  storage_class               = "STANDARD"
  public_access_prevention    = "enforced"
  location                    = var.region
  project                     = var.project_id
  force_destroy               = true
  uniform_bucket_level_access = true
}

resource "google_storage_bucket" "gmailagg_config" {
  name                        = "gmailagg-config"
  storage_class               = "STANDARD"
  public_access_prevention    = "enforced"
  location                    = var.region
  project                     = var.project_id
  force_destroy               = true
  uniform_bucket_level_access = true
}
