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

variable "influxdb_machine_type" {
  type    = string
  default = "e2-micro"
}

variable "influxdb_user_name" {
  type    = string
  default = "admin"
}

variable "influxdb_password" {
  type      = string
  default   = "example-password"
  sensitive = true
}

variable "influxdb_org" {
  type    = string
  default = "example-org"
}

variable "influxdb_bucket" {
  type    = string
  default = "gmailagg"
}

variable "influxdb_retention" {
  type    = string
  default = ""
}

variable "influxdb_admin_token" {
  type      = string
  default   = "example-token"
  sensitive = true
}
