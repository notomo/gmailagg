variable "project_id" {
  type    = string
  default = "gmailagg"
}

variable "region" {
  type    = string
  default = "us-west1"
}

variable "zone" {
  type    = string
  default = "us-west1-b"
}

variable "machine_type" {
  type    = string
  default = "e2-micro"
}

variable "influxdb_user_name" {
  type    = string
  default = "admin"
}

variable "influxdb_password" {
  type    = string
  default = "example-password"
}

variable "influxdb_org" {
  type    = string
  default = "example-org"
}

variable "influxdb_bucket" {
  type    = string
  default = "gmailagg"
}
