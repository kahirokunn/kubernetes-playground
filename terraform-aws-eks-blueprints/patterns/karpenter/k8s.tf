provider "kubectl" {
  apply_retry_count      = 15
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
  load_config_file       = false
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
  }
}

resource "kubectl_manifest" "karpenter-default-node" {
  yaml_body = templatefile("${path.module}/karpenter.yaml", {
    cluster_name = module.eks.cluster_name
  })

  depends_on = [helm_release.karpenter]
}
