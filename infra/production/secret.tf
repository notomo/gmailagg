resource "google_secret_manager_secret" "gmail_oauth_client_credentials" {
  secret_id = "gmail_oauth_client_credentials"
  replication {
    auto {}
  }
  project = var.project_id
}

resource "google_secret_manager_secret_iam_member" "runner_job_gmail_oauth_client_credentials_access" {
  secret_id = google_secret_manager_secret.gmail_oauth_client_credentials.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.runner_job.email}"
  project   = var.project_id
}

resource "google_secret_manager_secret_iam_member" "authorizer_job_gmail_oauth_client_credentials_access" {
  secret_id = google_secret_manager_secret.gmail_oauth_client_credentials.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.authorizer_job.email}"
  project   = var.project_id
}

resource "google_secret_manager_secret" "influxdb_token" {
  secret_id = "influxdb_token"
  replication {
    auto {}
  }
  project = var.project_id
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
    auto {}
  }
  project = var.project_id
}

resource "google_secret_manager_secret_iam_member" "runner_job_tailscale_reusable_auth_key_access" {
  secret_id = google_secret_manager_secret.tailscale_reusable_auth_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.runner_job.email}"
  project   = var.project_id
}

resource "google_secret_manager_secret_iam_member" "authorizer_job_tailscale_reusable_auth_key_access" {
  secret_id = google_secret_manager_secret.tailscale_reusable_auth_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.authorizer_job.email}"
  project   = var.project_id
}

resource "google_secret_manager_secret" "gmailagg_slack_webhook_url" {
  secret_id = "gmailagg_slack_webhook_url"
  replication {
    auto {}
  }
  project = var.project_id
}

resource "google_secret_manager_secret_iam_member" "authorizer_job_gmailagg_slack_webhook_url_access" {
  secret_id = google_secret_manager_secret.gmailagg_slack_webhook_url.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.authorizer_job.email}"
  project   = var.project_id
}
