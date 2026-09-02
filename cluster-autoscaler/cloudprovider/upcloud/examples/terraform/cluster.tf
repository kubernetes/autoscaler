
provider "upcloud" {}

resource "upcloud_router" "autoscaler" {
  name = "autoscaler-demo"
  lifecycle {
    ignore_changes = [static_route]
  }
}

resource "upcloud_gateway" "autoscaler" {
  name     = "autoscaler-demo"
  zone     = var.cluster_zone
  features = ["nat"]
  router {
    id = upcloud_router.autoscaler.id
  }
}

resource "upcloud_network" "autoscaler" {
  name = "autoscaler-demo"
  zone = var.cluster_zone
  ip_network {
    address            = "10.10.0.0/24"
    dhcp               = true
    family             = "IPv4"
    dhcp_default_route = true
  }
  router = upcloud_router.autoscaler.id
}

resource "upcloud_kubernetes_cluster" "autoscaler" {
  name                    = "autoscaler-demo"
  network                 = upcloud_network.autoscaler.id
  zone                    = var.cluster_zone
  private_node_groups     = true
  control_plane_ip_filter = ["0.0.0.0/0"]
}

resource "upcloud_kubernetes_node_group" "autoscaler" {
  cluster    = resource.upcloud_kubernetes_cluster.autoscaler.id
  node_count = 3
  name       = "group-1"
  plan       = "1xCPU-2GB"
  lifecycle {
    ignore_changes = [node_count]
  }
}

data "upcloud_kubernetes_cluster" "autoscaler" {
  id = resource.upcloud_kubernetes_cluster.autoscaler.id
}
