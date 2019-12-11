provider "helm" {
  kubernetes {
    load_config_file = true

    cluster_ca_certificate = file("~/.kube/cluster-ca-cert.pem")
  }
}

locals {
  kubevault_release_name = "vault-operator"
  kubevault_namespace    = "kube-system"
}

locals {
  kubevault_chart_name = "vault-operator"

  # re-implement: https://github.com/kubevault/operator/blob/0.3.0/charts/vault-operator/templates/_helpers.tpl#L9-L20
  # in hcl
  kubevault_release_fullname = length(regexall("\\Q${local.kubevault_chart_name}\\E", local.kubevault_release_name)) != 0 ? trimsuffix(substr(local.kubevault_release_name, 0, 63), "-") : trimsuffix(substr("${local.kubevault_release_name}-${local.kubevault_chart_name}", 0, 63), "-")
}

data "helm_repository" "appscode" {
  name = "appscode"
  url  = "https://charts.appscode.com/stable"
}

resource "tls_private_key" "kubevault_ca" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "kubevault_ca" {
  key_algorithm   = tls_private_key.kubevault_ca.algorithm
  private_key_pem = tls_private_key.kubevault_ca.private_key_pem

  subject {
    common_name = "ca"
  }

  validity_period_hours = 87600
  set_subject_key_id    = false

  is_ca_certificate = true
  allowed_uses = [
    "digital_signature",
    "key_encipherment",
    "cert_signing",
    "server_auth",
    "client_auth"
  ]
}

resource "tls_private_key" "kubevault_server" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "kubevault_server" {
  key_algorithm   = tls_private_key.kubevault_server.algorithm
  private_key_pem = tls_private_key.kubevault_server.private_key_pem

  subject {
    common_name = local.kubevault_release_fullname
  }

  dns_names = [
    "${local.kubevault_release_fullname}.${local.kubevault_namespace}",
    "${local.kubevault_release_fullname}.${local.kubevault_namespace}.svc",
    "${local.kubevault_release_fullname}.${local.kubevault_namespace}.svc.cluster.local"
  ]
}

resource "tls_locally_signed_cert" "kubevault_server" {
  ca_key_algorithm   = tls_self_signed_cert.kubevault_ca.key_algorithm
  ca_private_key_pem = tls_private_key.kubevault_ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.kubevault_ca.cert_pem

  cert_request_pem = tls_cert_request.kubevault_server.cert_request_pem

  validity_period_hours = 87600
  set_subject_key_id    = false

  allowed_uses = [
    "digital_signature",
    "key_encipherment",
    "server_auth",
    "client_auth"
  ]
}

resource "helm_release" "kubevault" {
  name = local.kubevault_release_name

  namespace = local.kubevault_namespace

  # executing locally, from in chart folder
  chart = "../vault-operator"

  # executing from published chart
  #   repository = data.helm_repository.appscode.metadata[0].name
  #   chart = "vault-operator"
  #   version    = "0.2.0"

  set {
    name  = "apiserver.k8s2operatorCerts.generate"
    value = "false"
  }

  set_sensitive {
    name  = "apiserver.ca"
    value = base64encode(file("~/.kube/cluster-ca-cert.pem"))
  }

  set_sensitive {
    name  = "apiserver.k8s2operatorCerts.caCrt"
    value = base64encode(tls_self_signed_cert.kubevault_ca.cert_pem)
  }

  set_sensitive {
    name  = "apiserver.k8s2operatorCerts.serverKey"
    value = base64encode(tls_private_key.kubevault_server.private_key_pem)
  }

  set_sensitive {
    name  = "apiserver.k8s2operatorCerts.serverCrt"
    value = base64encode(tls_locally_signed_cert.kubevault_server.cert_pem)
  }
}