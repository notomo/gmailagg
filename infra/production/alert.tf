resource "google_monitoring_alert_policy" "job_error" {
  display_name = "job_error"

  combiner = "OR"

  notification_channels = [
    "projects/gmailagg/notificationChannels/4926407567727485434",
    "projects/gmailagg/notificationChannels/438750772524549281",
  ]

  project = var.project_id

  alert_strategy {
    auto_close = "604800s"

    notification_rate_limit {
      period = "300s"
    }
  }

  conditions {
    display_name = "Log match condition"
    condition_matched_log {
      filter = <<-EOT
        resource.type = "cloud_run_job"
        resource.labels.job_name = "${google_cloud_run_v2_job.runner_job.name}"
        resource.labels.location = "${var.region}"
        jsonPayload.level = "ERROR"
      EOT
      label_extractors = {
        "error" = "EXTRACT(jsonPayload.\"msg\")"
      }
    }
  }

  timeouts {}
}
