terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2"
    }
  }
  # Floci (localstack) production backend - state stored locally on Floci VM
  backend "local" {
    path = "terraform.tfstate"
  }

  # Future AWS production backend (uncomment and comment local above):
  # backend "s3" {
  #   bucket                      = "pulumi-state"
  #   key                         = "resume-analyzer/terraform.tfstate"
  #   region                      = "us-east-1"
  #   encrypt                     = true
  #   use_lockfile                = true
  #   skip_credentials_validation = false
  #   skip_metadata_api_check     = false
  #   skip_requesting_account_id  = false
  # }
}