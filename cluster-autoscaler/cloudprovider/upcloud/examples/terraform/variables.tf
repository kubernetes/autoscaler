terraform {
  required_providers {
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = "~> 5.0"
    }
  }
}

variable "autoscaler_username" {
  type = string
}

variable "autoscaler_password" {
  type      = string
  sensitive = true
}

variable "cluster_zone" {
  type    = string
  default = "de-fra1"
}

variable "api_access_cidr_range" {
  description = "Network CIDR range that is allowed to connect to API server from outside of the cluster."
  type        = string
}
