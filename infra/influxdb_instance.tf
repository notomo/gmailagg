resource "google_service_account" "default" {
  account_id   = var.project_id
  display_name = "gmailagg compute engine instance service account"
}

resource "google_project_iam_member" "instance-log-writer" {
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.default.email}"
  project = var.project_id
}

resource "google_project_iam_member" "instance-metrics-writer" {
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.default.email}"
  project = var.project_id
}

module "influxdb-container" {
  source  = "terraform-google-modules/container-vm/google"
  version = "~> 3.1"

  container = {
    image = "marketplace.gcr.io/google/influxdb2:2.7"

    env = [
      {
        name  = "DOCKER_INFLUXDB_INIT_MODE"
        value = "setup"
      },
      {
        name  = "DOCKER_INFLUXDB_INIT_USERNAME"
        value = var.influxdb_user_name
      },
      {
        name  = "DOCKER_INFLUXDB_INIT_PASSWORD"
        value = var.influxdb_password
      },
      {
        name  = "DOCKER_INFLUXDB_INIT_ORG"
        value = var.influxdb_org
      },
      {
        name  = "DOCKER_INFLUXDB_INIT_BUCKET"
        value = var.influxdb_bucket
      },
    ]

    volumeMounts = [
      {
        mountPath = "/var/lib/influxdb2"
        name      = "data-disk-0"
        readOnly  = false
      },
    ]
  }
  volumes = [
    {
      name = "data-disk-0"

      gcePersistentDisk = {
        pdName = "data-disk-0"
        fsType = "ext4"
      }
    },
  ]
}

resource "google_compute_disk" "influxdb-data-disk" {
  project = var.project_id
  name    = "${var.project_id}-influxdb-data-disk"
  type    = "pd-standard"
  zone    = var.zone
  size    = 20
}

resource "google_compute_instance" "default" {
  name         = "${var.project_id}-instance"
  machine_type = var.machine_type

  can_ip_forward = true

  boot_disk {
    initialize_params {
      image = module.influxdb-container.source_image
      size  = 10
      type  = "pd-standard"
    }
  }

  attached_disk {
    source      = google_compute_disk.influxdb-data-disk.self_link
    device_name = "data-disk-0"
    mode        = "READ_WRITE"
  }

  network_interface {
    network = "default"
    access_config {}
  }

  metadata = {
    gce-container-declaration = module.influxdb-container.metadata_value
    google-logging-enabled    = true
    google-monitoring-enabled = true
    user-data                 = data.template_file.cloud-init.rendered
  }

  service_account {
    email = google_service_account.default.email
    scopes = [
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/trace.append",
    ]
  }
}
