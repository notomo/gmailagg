
resource "google_service_account" "cloud_run_job" {
  account_id   = "gmailagg-job"
  display_name = "gmailagg cloud_run job service account"
}

resource "google_project_iam_member" "clound_run_job_object_viewer" {
  role    = "roles/storage.objectViewer"
  member  = "serviceAccount:${google_service_account.cloud_run_job.email}"
  project = var.project_id
}

resource "google_cloud_run_v2_job" "app" {
  name     = "gmailagg-app"
  location = var.region
  project  = var.project_id

  template {
    template {
      containers {
        image = "${var.region}-docker.pkg.dev/${var.project_id}/gmailagg-app/app"

        env {
          name = "GMAILAGG_GMAIL_CREDENTIALS"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.gmail_oauth2_client_credentials.secret_id
              version = "1"
            }
          }
        }

        env {
          name = "INFLUXDB_TOKEN"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.influxdb_token.secret_id
              version = "1"
            }
          }
        }

      }
      timeout     = "300s"
      max_retries = 0

      service_account = google_service_account.cloud_run_job.email
    }
  }

}


resource "google_secret_manager_secret" "gmail_oauth2_client_credentials" {
  secret_id = "gmail_oauth2_client_credentials"
  replication {
    automatic = true
  }
  project = var.project_id
}

resource "google_secret_manager_secret_iam_member" "gmail_oauth2_client_credentials_access" {
  secret_id = google_secret_manager_secret.gmail_oauth2_client_credentials.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run_job.email}"
  project   = var.project_id
}

resource "google_secret_manager_secret" "influxdb_token" {
  secret_id = "influxdb_token"
  replication {
    automatic = true
  }
  project = var.project_id
}

resource "google_secret_manager_secret_iam_member" "influxdb_token_access" {
  secret_id = google_secret_manager_secret.influxdb_token.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run_job.email}"
  project   = var.project_id
}
