resource "google_service_account" "runner_job" {
  account_id   = "gmailagg-runner-job"
  display_name = "gmailagg cloud_run runner job service account"
}

resource "google_project_iam_member" "runner_job_storage_object_viewer" {
  role    = "roles/storage.objectViewer"
  member  = "serviceAccount:${google_service_account.runner_job.email}"
  project = var.project_id
  condition {
    title      = "limit_buckets"
    expression = <<-EOT
      resource.name == "projects/_/buckets/${google_storage_bucket.gmailagg_oauth.name}" ||
      resource.name.startsWith("projects/_/buckets/${google_storage_bucket.gmailagg_oauth.name}/objects/") ||
      resource.name == "projects/_/buckets/${google_storage_bucket.gmailagg_config.name}" ||
      resource.name.startsWith("projects/_/buckets/${google_storage_bucket.gmailagg_config.name}/objects/")
    EOT
  }
}

resource "google_cloud_run_v2_job" "runner_job" {
  name     = "gmailagg-runner-job"
  location = var.region
  project  = var.project_id

  template {
    template {
      containers {
        image = "${var.region}-docker.pkg.dev/${var.project_id}/gmailagg-app/job"

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

      service_account = google_service_account.runner_job.email
    }
  }

  lifecycle {
    ignore_changes = [
      annotations,
      client,
      template[0].annotations
    ]
  }

}

resource "google_cloud_scheduler_job" "runner_job" {
  name             = "gmailagg-runner-job"
  description      = "gmailagg runner job scheduler"
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
    uri         = "https://${google_cloud_run_v2_job.runner_job.location}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_number}/jobs/${google_cloud_run_v2_job.runner_job.name}:run"

    oauth_token {
      service_account_email = google_service_account.cloud_run_invoker.email
    }
  }
}
