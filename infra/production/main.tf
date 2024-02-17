terraform {
  required_version = "~> 1.5.0"
  required_providers {
    template = {
      source  = "hashicorp/template"
      version = "~> 2.2.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.16.0"
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
