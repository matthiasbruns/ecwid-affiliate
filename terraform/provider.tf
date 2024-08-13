locals {
  keyName = "ecwid-affiliate-link-${var.stage}-tfstate"
}

terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
  backend "s3" {}
}

provider "aws" {
  region = "eu-central-1"

  default_tags {
    tags = {
      Environment          = var.stage
      Project              = "ecwid-lexoffice"
      Repository           = "https://github.com/matthiasbruns/ecwid-lexoffice"
      TerraformManaged     = "true"
      TerraformStateBucket = local.keyName
    }
  }
}
