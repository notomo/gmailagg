resource "google_service_account" "influxdb_instance" {
  account_id   = var.project_id
  display_name = "gmailagg compute engine influxdb instance service account"
}

resource "google_compute_network" "influxdb_instance" {
  name                    = var.project_id
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "influxdb_instance" {
  name          = "${var.project_id}-subnet"
  network       = google_compute_network.influxdb_instance.id
  region        = var.region
  ip_cidr_range = "192.168.0.0/20"
}

resource "google_project_iam_member" "influxdb_instance_log_writer" {
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.influxdb_instance.email}"
  project = var.project_id
}

resource "google_project_iam_member" "influxdb_instance_metrics_writer" {
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.influxdb_instance.email}"
  project = var.project_id
}

module "influxdb_container" {
  source  = "terraform-google-modules/container-vm/google"
  version = "~> 3.1"

  cos_image_name = "cos-beta-109-17800-0-13"

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
      {
        name  = "DOCKER_INFLUXDB_INIT_RETENTION"
        value = var.influxdb_retention
      },
      {
        name  = "DOCKER_INFLUXDB_INIT_ADMIN_TOKEN"
        value = var.influxdb_admin_token
      },
      {
        name  = "INFLUXD_SESSION_RENEW_DISABLED"
        value = "true"
      },
      {
        name  = "INFLUXD_SESSION_LENGTH"
        value = "1440"
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

resource "google_compute_disk" "influxdb_data" {
  project = var.project_id
  name    = "${var.project_id}-influxdb-data-disk"
  type    = "pd-standard"
  zone    = var.zone
  size    = 20
}

resource "google_compute_instance" "influxdb" {
  name         = "${var.project_id}-influxdb"
  machine_type = var.influxdb_machine_type

  allow_stopping_for_update = true
  can_ip_forward            = true

  boot_disk {
    initialize_params {
      image = module.influxdb_container.source_image
      size  = 10
      type  = "pd-standard"
    }
  }

  attached_disk {
    source      = google_compute_disk.influxdb_data.self_link
    device_name = "data-disk-0"
    mode        = "READ_WRITE"
  }

  network_interface {
    subnetwork = google_compute_subnetwork.influxdb_instance.id
    access_config {}
  }

  metadata = {
    gce-container-declaration = module.influxdb_container.metadata_value
    block-project-ssh-keys    = true
    google-logging-enabled    = true
    google-monitoring-enabled = true
    user-data                 = sensitive(data.template_file.cloud_init.rendered)
  }

  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  service_account {
    email  = google_service_account.influxdb_instance.email
    scopes = ["cloud-platform"]
  }
}

resource "tailscale_tailnet_key" "influxdb_instance_onetime_auth" {
  ephemeral     = true
  preauthorized = true
  reusable      = false
}

data "template_file" "cloud_init" {
  template = file("cloud-init.yaml")

  vars = {
    tailscale_auth_key = tailscale_tailnet_key.influxdb_instance_onetime_auth.key
  }
}
