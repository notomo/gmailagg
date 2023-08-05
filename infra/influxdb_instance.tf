resource "google_service_account" "default" {
  account_id   = "gmailagg"
  display_name = "gmailagg compute engine instance service account"
}

resource "google_project_iam_member" "gmailagg-instance-log-writer" {
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.default.email}"
  project = "gmailagg"
}

resource "google_project_iam_member" "gmailagg-instance-metrics-writer" {
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.default.email}"
  project = "gmailagg"
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
        value = "admin"
      },
      {
        name  = "DOCKER_INFLUXDB_INIT_PASSWORD"
        value = "example-password"
      },
      {
        name  = "DOCKER_INFLUXDB_INIT_ORG"
        value = "example-org"
      },
      {
        name  = "DOCKER_INFLUXDB_INIT_BUCKET"
        value = "gmailagg"
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
  project = "gmailagg"
  name    = "gmailagg-influxdb-data-disk"
  type    = "pd-standard"
  zone    = "us-west1-b"
  size    = 20
}

resource "google_compute_instance" "default" {
  name         = "gmailagg-instance"
  machine_type = "e2-micro"

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
