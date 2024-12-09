# Please manually set up VPC Peering if necessary
module "mng" {
  source   = "../../patterns/karpenter"
  name     = "mng"
  vpc_cidr = "10.10.0.0/16"
}

module "worker1" {
  source   = "../../patterns/karpenter"
  name     = "worker1"
  vpc_cidr = "10.11.0.0/16"
}

module "worker2" {
  source   = "../../patterns/karpenter"
  name     = "worker2"
  vpc_cidr = "10.12.0.0/16"
}
