resource "aws_lambda_function" "watcher" {
  filename         = "../bin/matches.zip"
  function_name    = "fifa-match-watcher"
  role             = aws_iam_role.watcher.arn
  handler          = "bin/matches"
  source_code_hash = filebase64sha256("../bin/matches.zip")
  runtime          = "go1.x"
  timeout          = 30
  environment {
    variables = {
      SENTRY_DSN        = data.aws_kms_secrets.secrets.plaintext["sentry-dsn-matches"]
      QUEUE_URL         = aws_sqs_queue.events.url
      TABLE_NAME        = aws_dynamodb_table.fifa_bot.id
      LOG_LEVEL         = "INFO"
      WATCH_COMPETITION = "17"
    }
  }
}

resource "aws_iam_role" "watcher" {
  name               = "fifa-match-watcher"
  assume_role_policy = data.aws_iam_policy_document.watcher_assume.json
}

data "aws_iam_policy_document" "watcher_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "watcher" {
  name   = "fifa-match-watcher"
  policy = data.aws_iam_policy_document.watcher.json
}

data "aws_iam_policy_document" "watcher" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem"
    ]
    resources = [
      aws_dynamodb_table.fifa_bot.arn
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "sqs:SendMessage",
    ]
    resources = [
      aws_sqs_queue.events.arn
    ]
  }
}

resource "aws_iam_role_policy_attachment" "watcher" {
  role       = aws_iam_role.watcher.name
  policy_arn = aws_iam_policy.watcher.arn
}

resource "aws_iam_role_policy_attachment" "watcher_basic_execution" {
  role       = aws_iam_role.watcher.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_permission" "watcher_timer" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.watcher.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.watcher_schedule.arn
}
