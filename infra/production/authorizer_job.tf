resource "google_service_account" "authorizer_job" {
  account_id   = "gmailagg-authorizer-job"
  display_name = "gmailagg cloud_run authorizer job service account"
}

resource "google_project_iam_member" "authorizer_job_storage_object_user" {
  role    = "roles/storage.objectUser"
  member  = "serviceAccount:${google_service_account.authorizer_job.email}"
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

resource "google_cloud_run_v2_job" "authorizer_job" {
  name     = "gmailagg-authorizer-job"
  location = var.region
  project  = var.project_id

  template {
    template {
      containers {
        image = "${var.region}-docker.pkg.dev/${var.project_id}/gmailagg-app/job"

        args = ["auth", "--timeout=2m", "--port=9999"]

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
          name = "TAILSCALE_AUTHKEY"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.tailscale_reusable_auth_key.secret_id
              version = "latest"
            }
          }
        }

        env {
          name = "GMAILAGG_SLACK_WEBHOOK_URL"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.gmailagg_slack_webhook_url.secret_id
              version = "latest"
            }
          }
        }

      }
      timeout     = "300s"
      max_retries = 0

      service_account = google_service_account.authorizer_job.email
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
