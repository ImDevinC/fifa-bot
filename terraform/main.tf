provider "aws" {
  region = "us-west-2"
}

terraform {
  backend "s3" {
    bucket = "imdevinc-tf-storage"
    region = "us-west-1"
    key    = "fifa-bot"
  }
}
