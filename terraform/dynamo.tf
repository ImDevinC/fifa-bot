resource "aws_dynamodb_table" "fifa_bot" {
  name     = "fifa-bot"
  hash_key = "MatchId"
  attribute {
    name = "MatchId"
    type = "S"
  }
  ttl {
    attribute_name = "Expiration"
    enabled        = true
  }
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1
}
