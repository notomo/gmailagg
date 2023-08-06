terraform {
  required_version = "~> 1.5.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.76.0"
    }
    tailscale = {
      source  = "tailscale/tailscale"
      version = "~> 0.13.6"
    }
  }
  backend "gcs" {
    bucket = "gmailagg-tfstate"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

provider "tailscale" {
  # needs environment variables
  # TAILSCALE_TAILNET
  # TAILSCALE_API_KEY
}
