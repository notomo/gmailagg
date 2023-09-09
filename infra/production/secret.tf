resource "google_secret_manager_secret" "gmail_oauth_client_credentials" {
  secret_id = "gmail_oauth_client_credentials"
  replication {
    automatic = true
  }
  project = var.project_id
}

resource "google_secret_manager_secret_version" "gmail_oauth_client_credentials_value" {
  secret      = google_secret_manager_secret.gmail_oauth_client_credentials.id
  secret_data = "dummy"
  lifecycle {
    ignore_changes = [
      enabled,
      secret_data,
    ]
  }
}

resource "google_secret_manager_secret_iam_member" "runner_job_gmail_oauth_client_credentials_access" {
  secret_id = google_secret_manager_secret.gmail_oauth_client_credentials.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.runner_job.email}"
  project   = var.project_id
}

resource "google_secret_manager_secret" "influxdb_token" {
  secret_id = "influxdb_token"
  replication {
    automatic = true
  }
  project = var.project_id
}

resource "google_secret_manager_secret_version" "influxdb_token_value" {
  secret      = google_secret_manager_secret.influxdb_token.id
  secret_data = "dummy"
  lifecycle {
    ignore_changes = [
      enabled,
      secret_data,
    ]
  }
}

resource "google_secret_manager_secret_iam_member" "runner_job_influxdb_token_access" {
  secret_id = google_secret_manager_secret.influxdb_token.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.runner_job.email}"
  project   = var.project_id
}

resource "google_secret_manager_secret" "tailscale_reusable_auth_key" {
  secret_id = "tailscale_reusable_auth_key"
  replication {
    automatic = true
  }
  project = var.project_id
}

resource "google_secret_manager_secret_version" "tailscale_reusable_auth_key_value" {
  secret      = google_secret_manager_secret.tailscale_reusable_auth_key.id
  secret_data = "dummy"
  lifecycle {
    ignore_changes = [
      enabled,
      secret_data,
    ]
  }
}

resource "google_secret_manager_secret_iam_member" "runner_job_tailscale_reusable_auth_key_access" {
  secret_id = google_secret_manager_secret.tailscale_reusable_auth_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.runner_job.email}"
  project   = var.project_id
}

