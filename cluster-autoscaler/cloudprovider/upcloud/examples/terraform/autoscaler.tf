provider "kubernetes" {
  client_certificate     = data.upcloud_kubernetes_cluster.autoscaler.client_certificate
  client_key             = data.upcloud_kubernetes_cluster.autoscaler.client_key
  cluster_ca_certificate = data.upcloud_kubernetes_cluster.autoscaler.cluster_ca_certificate
  host                   = data.upcloud_kubernetes_cluster.autoscaler.host
}

resource "kubernetes_secret" "autoscaler" {
  metadata {
    name      = "upcloud-autoscaler"
    namespace = "kube-system"
  }
  data = {
    token    = var.autoscaler_token != null ? var.autoscaler_token : ""
    username = var.autoscaler_username != null ? var.autoscaler_username : ""
    password = var.autoscaler_password != null ? var.autoscaler_password : ""
  }
}

resource "kubernetes_deployment" "autoscaler" {
  depends_on = [kubernetes_role_binding.autoscaler]
  metadata {
    name      = "cluster-autoscaler"
    namespace = "kube-system"
    labels = {
      app = "cluster-autoscaler"
    }
  }
  spec {
    replicas = 1
    selector {
      match_labels = {
        app = "cluster-autoscaler"
      }
    }
    template {
      metadata {
        labels = {
          app = "cluster-autoscaler"
        }
      }
      spec {
        service_account_name = "cluster-autoscaler"
        priority_class_name  = "system-cluster-critical"
        container {
          image = "ghcr.io/upcloudltd/autoscaler:v1.29.5"
          name  = "cluster-autoscaler"
          resources {
            limits = {
              cpu    = "100m"
              memory = "300Mi"
            }
            requests = {
              cpu    = "100m"
              memory = "300Mi"
            }
          }
          command = [
            "/cluster-autoscaler",
            "--cloud-provider=upcloud",
            "--stderrthreshold=info",
            "--scale-down-enabled=true",
            "-v=4"
          ]
          env {
            name  = "UPCLOUD_CLUSTER_ID"
            value = resource.upcloud_kubernetes_cluster.autoscaler.id
          }
          env {
            name = "UPCLOUD_TOKEN"
            value_from {
              secret_key_ref {
                key  = "token"
                name = "upcloud-autoscaler"
              }
            }
          }
          env {
            name = "UPCLOUD_USERNAME"
            value_from {
              secret_key_ref {
                key  = "username"
                name = "upcloud-autoscaler"
              }
            }
          }
          env {
            name = "UPCLOUD_PASSWORD"
            value_from {
              secret_key_ref {
                key  = "password"
                name = "upcloud-autoscaler"
              }
            }
          }
          volume_mount {
            name       = "ssl-certs"
            mount_path = "/etc/ssl/certs/ca-certificates.crt"
            read_only  = true
          }
        }
        volume {
          name = "ssl-certs"
          host_path {
            path = "/etc/ssl/certs/ca-certificates.crt"
          }
        }
      }
    }
  }
}
