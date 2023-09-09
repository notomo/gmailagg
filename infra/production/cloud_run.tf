resource "google_service_account" "cloud_run_invoker" {
  account_id   = "cloud-run-invoker"
  display_name = "cloud run invoker"
  project      = var.project_id
}

resource "google_project_iam_member" "cloud_run_invoker" {
  role    = "roles/run.invoker"
  member  = "serviceAccount:${google_service_account.cloud_run_invoker.email}"
  project = var.project_id
}
