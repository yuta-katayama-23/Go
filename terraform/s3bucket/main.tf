terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.27"
    }
  }

  required_version = ">= 1.0.8"

  backend "remote" {
    organization = "yuta_katayama"

    workspaces {
      name = "terraform-go-lambda-cicd"
    }
  }
}

provider "aws" {
  region = "ap-northeast-1"
}

variable "user_name" {}

resource "aws_s3_bucket" "s3_bucket" {
  bucket = "terraform-build-artifact-go-lambda"
}

resource "aws_s3_bucket_policy" "policy" {
  bucket = aws_s3_bucket.s3_bucket.id
  policy = data.aws_iam_policy_document.policy_doc.json
}

resource "aws_s3_bucket_object" "build_artifact" {
  bucket = aws_s3_bucket.s3_bucket.id
  key    = "main.zip"
  source = "${path.module}/main.zip"
  etag   = filemd5("${path.module}/main.zip")
}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "policy_doc" {
  statement {
    sid = "bucket-policy"

    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:user/${var.user_name}"]
    }

    actions = [
      "s3:GetObject"
    ]

    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/*"
    ]
  }
}
