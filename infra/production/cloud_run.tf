
resource "google_service_account" "job" {
  account_id   = "gmailagg-job"
  display_name = "gmailagg cloud_run job service account"
}

resource "google_project_iam_member" "job_storage_object_viewer" {
  role    = "roles/storage.objectViewer"
  member  = "serviceAccount:${google_service_account.job.email}"
  project = var.project_id
  condition {
    title      = "limit_to_token_bucket"
    expression = <<-EOT
      resource.name == "projects/_/buckets/${google_storage_bucket.gmailagg_oauth.name}" ||
      resource.name.startsWith("projects/_/buckets/${google_storage_bucket.gmailagg_oauth.name}/objects/")
    EOT
  }
}

resource "google_cloud_run_v2_job" "job" {
  name     = "gmailagg-job"
  location = var.region
  project  = var.project_id

  template {
    template {
      containers {
        image = "${var.region}-docker.pkg.dev/${var.project_id}/gmailagg-app/job"

        env {
          name  = "GMAILAGG_CONFIG"
          value = sensitive("dummy")
        }

        env {
          name = "GMAILAGG_GMAIL_CREDENTIALS"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.gmail_oauth_client_credentials.secret_id
              version = "latest"
            }
          }
        }

        env {
          name = "INFLUXDB_TOKEN"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.influxdb_token.secret_id
              version = "latest"
            }
          }
        }

        env {
          name = "TAILSCALE_AUTHKEY"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.tailscale_reusable_auth_key.secret_id
              version = "latest"
            }
          }
        }

      }
      timeout     = "300s"
      max_retries = 0

      service_account = google_service_account.job.email
    }
  }

  lifecycle {
    ignore_changes = [
      template[0].template[0].containers[0].env[0].value,
      annotations,
      client,
      template[0].annotations
    ]
  }

}

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

resource "google_secret_manager_secret_iam_member" "gmail_oauth_client_credentials_access" {
  secret_id = google_secret_manager_secret.gmail_oauth_client_credentials.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.job.email}"
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

resource "google_secret_manager_secret_iam_member" "influxdb_token_access" {
  secret_id = google_secret_manager_secret.influxdb_token.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.job.email}"
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

resource "google_secret_manager_secret_iam_member" "tailscale_reusable_auth_key_access" {
  secret_id = google_secret_manager_secret.tailscale_reusable_auth_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.job.email}"
  project   = var.project_id
}

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

resource "google_cloud_scheduler_job" "job" {
  name             = "gmailagg-job"
  description      = "gmailagg job scheduler"
  schedule         = "0 0 */3 * *"
  time_zone        = "Asia/Tokyo"
  attempt_deadline = "330s"
  region           = var.region
  project          = var.project_id

  retry_config {
    retry_count = 0
  }

  http_target {
    http_method = "POST"
    uri         = "https://${google_cloud_run_v2_job.job.location}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_number}/jobs/${google_cloud_run_v2_job.job.name}:run"

    oauth_token {
      service_account_email = google_service_account.cloud_run_invoker.email
    }
  }
}
