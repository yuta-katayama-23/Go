terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.27"
    }
  }

  required_version = ">= 0.14.9"

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

variable "s3_bucket_name" {
  type    = string
  default = "terraform-build-artifact-go-lambda"
}

variable "s3_key_name" {
  type    = string
  default = "main.zip"
}

variable "s3_key_name_base64sha256" {
  type    = string
  default = "main.zip.base64sha256"
}

variable "backlog_api_key" {
  type = string
}

variable "backlog_domain" {
  type = string
}

variable "slack_api_token" {
  type = string
}

resource "aws_lambda_function" "go-lambda" {
  function_name    = "terraform-go-lambda"
  runtime          = "go1.x"
  handler          = "main"
  description      = "Lambda function create by aws terraform."
  s3_bucket        = var.s3_bucket_name
  s3_key           = var.s3_key_name
  source_code_hash = data.aws_s3_bucket_object.build_artifact.body
  role             = aws_iam_role.iam_for_lambda.arn

  environment {
    variables = {
      "BACKLOG_API_KEY"      = var.backlog_api_key
      "BACKLOG_DOMEIN"       = var.backlog_domain
      "BACKLOG_ISSUE_STATUS" = "未対応"
      "BACKLOG_STATUS_ID"    = "3"
      "SLACK_API_TOKEN"      = var.slack_api_token
      "SLACK_CHANNEL"        = "#lambda実行"
    }
  }
}

resource "aws_iam_role" "iam_for_lambda" {
  name               = "service-role-for-terraform-go-lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_role_doc.json
  path               = "/service-role/"
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "lambda_execute_role_policy" {
  name   = "lambda_execute"
  role   = aws_iam_role.iam_for_lambda.id
  policy = data.aws_iam_policy_document.codecommit_policy.json
}

resource "aws_lambda_permission" "allow_cloudwatchevents" {
  statement_id  = "terraform-go-lambda-GoLambdaFunctionCodeCommitPushPermission"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.go-lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.codecommit_push.arn
}

resource "aws_cloudwatch_event_rule" "codecommit_push" {
  name          = "terraform-go-lambda-GoLambdaFunctionCodeCommitPush"
  description   = "when developer push to codecommit, trigger lambda function."
  event_pattern = <<EOF
{
        "source" : [
                "aws.codecommit"
        ],
        "detail" : {
            "eventSource" : [
                "codecommit.amazonaws.com"
            ]
        },
        "detail-type" : [
                "CodeCommit Repository State Change"
        ]
}
EOF
}

resource "aws_cloudwatch_event_target" "go_lambda_function_codecommit_push" {
  rule = aws_cloudwatch_event_rule.codecommit_push.name
  arn  = aws_lambda_function.go-lambda.arn
}

data "aws_iam_policy_document" "assume_role_doc" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "codecommit_policy" {
  statement {
    sid    = "CodeCommit"
    effect = "Allow"

    actions = [
      "codecommit:GetCommit"
    ]

    resources = [
      "arn:aws:codecommit:ap-northeast-1:${data.aws_caller_identity.current.account_id}:*",
    ]
  }
}

data "aws_s3_bucket_object" "build_artifact" {
  bucket = var.s3_bucket_name
  key    = var.s3_key_name_base64sha256
}

data "aws_caller_identity" "current" {}

output "go_lambda_function_arn" {
  value = aws_lambda_function.go-lambda.arn
}

output "implicit_ima_role_created_for_go_function" {
  value = aws_iam_role.iam_for_lambda.arn
}

output "cloud_watch_events_for_go_function" {
  value = aws_cloudwatch_event_rule.codecommit_push.arn
}
