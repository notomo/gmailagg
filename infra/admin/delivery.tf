module "gh_oidc" {
  source  = "terraform-google-modules/github-actions-runners/google//modules/gh-oidc"
  version = "~> 3.1"

  project_id  = var.project_id
  pool_id     = "gmailagg-pool"
  provider_id = "github-actions"
  sa_mapping = {
    "gmailagg-delivery" = {
      sa_name   = "projects/${var.project_id}/serviceAccounts/gmailagg-delivery@${var.project_id}.iam.gserviceaccount.com"
      attribute = "attribute.repository/notomo/gmailagg"
    }
  }
  attribute_condition   = <<-EOT
    assertion.repository=='notomo/gmailagg'
  EOT
  provider_display_name = "github-actions"
}

resource "google_service_account" "delivery" {
  account_id   = "gmailagg-delivery"
  display_name = "gmailagg delivery automation"
}

resource "google_project_iam_member" "delivery_artifactregistry_writer" {
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.delivery.email}"
  project = var.project_id
}

resource "google_project_iam_member" "delivery_storage_object_user" {
  role    = "roles/storage.objectUser"
  member  = "serviceAccount:${google_service_account.delivery.email}"
  project = var.project_id
  condition {
    title      = "limit_to_tfstate_bucket"
    expression = <<-EOT
      resource.name == "projects/_/buckets/gmailagg-tfstate" ||
      resource.name.startsWith("projects/_/buckets/gmailagg-tfstate/objects/")
    EOT
  }
}

resource "google_project_iam_custom_role" "terraform" {
  role_id = "terraform"
  title   = "terraform automation role"
  permissions = [
    "resourcemanager.projects.getIamPolicy",
    "compute.networks.get",
    "compute.subnetworks.get",
    "compute.instances.get",
    "compute.disks.get",
    "storage.buckets.get",
    "secretmanager.secrets.get",
    "secretmanager.secrets.getIamPolicy",
    "iam.serviceAccounts.get",
    "run.jobs.get",
    "cloudscheduler.jobs.get",
    "monitoring.alertPolicies.get",
  ]
}

resource "google_project_iam_member" "terraform" {
  role    = google_project_iam_custom_role.terraform.id
  member  = "serviceAccount:${google_service_account.delivery.email}"
  project = var.project_id
}
