resource "kubernetes_service_account" "autoscaler" {
  metadata {
    labels = {
      "k8s-addon" = "cluster-autoscaler.addons.k8s.io"
      "k8s-app"   = "cluster-autoscaler"
    }
    name      = "cluster-autoscaler"
    namespace = "kube-system"
  }
}

resource "kubernetes_cluster_role" "autoscaler" {
  metadata {
    labels = {
      "k8s-addon" = "cluster-autoscaler.addons.k8s.io"
      "k8s-app"   = "cluster-autoscaler"
    }
    name = "cluster-autoscaler"
  }
  rule {
    api_groups = [
      "",
    ]
    resources = [
      "events",
      "endpoints",
    ]
    verbs = [
      "create",
      "patch",
    ]
  }
  rule {
    api_groups = [
      "",
    ]
    resources = [
      "pods/eviction",
    ]
    verbs = [
      "create",
    ]
  }
  rule {
    api_groups = [
      "",
    ]
    resources = [
      "pods/status",
    ]
    verbs = [
      "update",
    ]
  }
  rule {
    api_groups = [
      "",
    ]
    resource_names = [
      "cluster-autoscaler",
    ]
    resources = [
      "endpoints",
    ]
    verbs = [
      "get",
      "update",
    ]
  }
  rule {
    api_groups = [
      "",
    ]
    resources = [
      "namespaces",
    ]
    verbs = [
      "watch",
      "list",
      "get",
    ]
  }
  rule {
    api_groups = [
      "",
    ]
    resources = [
      "nodes",
    ]
    verbs = [
      "watch",
      "list",
      "get",
      "update",
    ]
  }
  rule {
    api_groups = [
      "",
    ]
    resources = [
      "namespaces",
      "pods",
      "services",
      "replicationcontrollers",
      "persistentvolumeclaims",
      "persistentvolumes",
    ]
    verbs = [
      "watch",
      "list",
      "get",
    ]
  }
  rule {
    api_groups = [
      "extensions",
    ]
    resources = [
      "replicasets",
      "daemonsets",
    ]
    verbs = [
      "watch",
      "list",
      "get",
    ]
  }
  rule {
    api_groups = [
      "policy",
    ]
    resources = [
      "poddisruptionbudgets",
    ]
    verbs = [
      "watch",
      "list",
    ]
  }
  rule {
    api_groups = [
      "apps",
    ]
    resources = [
      "statefulsets",
      "replicasets",
      "daemonsets",
    ]
    verbs = [
      "watch",
      "list",
      "get",
    ]
  }
  rule {
    api_groups = [
      "storage.k8s.io",
    ]
    resources = [
      "storageclasses",
      "csinodes",
      "csistoragecapacities",
      "csidrivers",
    ]
    verbs = [
      "watch",
      "list",
      "get",
    ]
  }
  rule {
    api_groups = [
      "batch",
      "extensions",
    ]
    resources = [
      "jobs",
    ]
    verbs = [
      "get",
      "list",
      "watch",
      "patch",
    ]
  }
  rule {
    api_groups = [
      "coordination.k8s.io",
    ]
    resources = [
      "leases",
    ]
    verbs = [
      "create",
    ]
  }
  rule {
    api_groups = [
      "coordination.k8s.io",
    ]
    resource_names = [
      "cluster-autoscaler",
    ]
    resources = [
      "leases",
    ]
    verbs = [
      "get",
      "update",
    ]
  }
}

resource "kubernetes_role" "autoscaler" {
  metadata {
    labels = {
      "k8s-addon" = "cluster-autoscaler.addons.k8s.io"
      "k8s-app"   = "cluster-autoscaler"
    }
    name      = "cluster-autoscaler"
    namespace = "kube-system"
  }
  rule {
    api_groups = [
      "",
    ]
    resources = [
      "configmaps",
    ]
    verbs = [
      "create",
      "list",
      "watch",
    ]
  }
  rule {
    api_groups = [
      "",
    ]
    resource_names = [
      "cluster-autoscaler-status",
      "cluster-autoscaler-priority-expander",
    ]
    resources = [
      "configmaps",
    ]
    verbs = [
      "delete",
      "get",
      "update",
      "watch",
    ]
  }
}

resource "kubernetes_cluster_role_binding" "autoscaler" {
  metadata {
    labels = {
      "k8s-addon" = "cluster-autoscaler.addons.k8s.io"
      "k8s-app"   = "cluster-autoscaler"
    }
    name = "cluster-autoscaler"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-autoscaler"
  }
  subject {
    kind      = "ServiceAccount"
    name      = "cluster-autoscaler"
    namespace = "kube-system"
  }
}

resource "kubernetes_role_binding" "autoscaler" {
  metadata {
    labels = {
      "k8s-addon" = "cluster-autoscaler.addons.k8s.io"
      "k8s-app"   = "cluster-autoscaler"
    }
    name      = "cluster-autoscaler"
    namespace = "kube-system"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = "cluster-autoscaler"
  }
  subject {
    kind      = "ServiceAccount"
    name      = "cluster-autoscaler"
    namespace = "kube-system"
  }
}
