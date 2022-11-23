resource "aws_cloudwatch_event_rule" "watcher_schedule" {
  name                = "start-watcher"
  schedule_expression = "rate(1 minute)"
}

resource "aws_cloudwatch_event_target" "watcher_schedule" {
  rule = aws_cloudwatch_event_rule.watcher_schedule.name
  arn  = aws_lambda_function.watcher.arn
}
