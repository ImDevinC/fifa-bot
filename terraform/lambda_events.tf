data "aws_kms_secrets" "secrets" {
  secret {
    name    = "slack-webhook-url"
    payload = var.encrypted_slack_url
  }
  secret {
    name    = "sentry-dsn-matches"
    payload = var.encrypted_sentry_dsn_matches
  }
  secret {
    name    = "sentry-dsn-events"
    payload = var.encrypted_sentry_dsn_events
  }
  secret {
    name    = "slack-webhook-test-url"
    payload = var.encrypted_test_slack_url
  }
}

resource "aws_lambda_function" "events" {
  filename         = "../bin/events.zip"
  function_name    = "fifa-events-watcher"
  role             = aws_iam_role.events.arn
  handler          = "bin/events"
  source_code_hash = filebase64sha256("../bin/events.zip")
  runtime          = "go1.x"
  timeout          = 30
  environment {
    variables = {
      SENTRY_DSN = data.aws_kms_secrets.secrets.plaintext["sentry-dsn-events"]
      QUEUE_URL  = aws_sqs_queue.events.url
      TABLE_NAME = aws_dynamodb_table.fifa_bot.id
      # SLACK_WEBHOOK_URL = data.aws_kms_secrets.secrets.plaintext["slack-webhook-url"]
      SLACK_WEBHOOK_URL = data.aws_kms_secrets.secrets.plaintext["slack-webhook-test-url"]
      LOG_LEVEL         = "INFO"
    }
  }
}

resource "aws_iam_role" "events" {
  name               = "fifa-events-watcher"
  assume_role_policy = data.aws_iam_policy_document.events_assume.json
}

data "aws_iam_policy_document" "events_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "events" {
  name   = "fifa-events-watcher"
  policy = data.aws_iam_policy_document.events.json
}

data "aws_iam_policy_document" "events" {
  statement {
    effect = "Allow"
    actions = [
      "sqs:SendMessage",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes"
    ]
    resources = [
      aws_sqs_queue.events.arn
    ]
  }
}

resource "aws_iam_role_policy_attachment" "events" {
  role       = aws_iam_role.events.name
  policy_arn = aws_iam_policy.events.arn
}

resource "aws_iam_role_policy_attachment" "events_basic_execution" {
  role       = aws_iam_role.events.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_permission" "event_sqs" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.events.function_name
  principal     = "sqs.amazonaws.com"
  source_arn    = aws_sqs_queue.events.arn
}

resource "aws_lambda_event_source_mapping" "events" {
  event_source_arn = aws_sqs_queue.events.arn
  function_name    = aws_lambda_function.events.arn
}
