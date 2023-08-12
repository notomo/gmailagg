resource "google_artifact_registry_repository" "app" {
  location      = var.region
  repository_id = "gmailagg-app"
  project       = var.project_id
  format        = "DOCKER"
}
