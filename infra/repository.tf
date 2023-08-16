resource "google_artifact_registry_repository" "gmailagg_app" {
  location      = var.region
  repository_id = "gmailagg-app"
  project       = var.project_id
  format        = "DOCKER"
}
