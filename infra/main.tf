terraform {
  required_version = "~> 1.5.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.76.0"
    }
  }
  backend "gcs" {
    bucket = "gmailagg-tfstate"
  }
}

provider "google" {
  project = "gmailagg"
  region  = "us-west1"
  zone    = "us-west1-b"
}
