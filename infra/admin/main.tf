terraform {
  required_version = "~> 1.9.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.42.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.42.0"
    }
  }
  backend "gcs" {
    bucket = "gmailagg-tfstate"
    prefix = "admin"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}
