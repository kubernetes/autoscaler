terraform {
  required_providers {
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = "~> 3.0"
    }
  }
}

variable "autoscaler_username" {
  type = string
  default = null
}

variable "autoscaler_password" {
  type      = string
  sensitive = true
  default   = null
}

variable "autoscaler_token" {
  type      = string
  sensitive = true
  default   = null
}

variable "cluster_zone" {
  type    = string
  default = "de-fra1"
}
