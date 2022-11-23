variable "encrypted_slack_url" {
  type        = string
  description = "KMS Encrypted Slack Webhook URL"
  default     = "AQICAHgsJXuUfVcivglHv1qjCTscwl6NzhvBut8NdXoUXyTZ4gEyPnEVzAsaiF+AdBYY4pDYAAAAszCBsAYJKoZIhvcNAQcGoIGiMIGfAgEAMIGZBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDC2msURpIstFV960cwIBEIBsdah/Yotpy0W5YOfKKiD4Ca5UNQPsbQ2LYT6fWxswGkGssBtVC42AqcKrA4Me2Ay+UmgW5p/cq82mJDIRJLrmY9gTkqCk9vfX1UmypwIpem5l+MY8/NPyug6GCyalcsEpY9zq4SCT5GdkS5Co"
}

variable "encrypted_sentry_dsn_events" {
  type        = string
  description = "KMS Encrypted Sentry DSN for fifa-events"
  default     = "AQICAHgsJXuUfVcivglHv1qjCTscwl6NzhvBut8NdXoUXyTZ4gFaAirWoSrvywTIlRfzOldGAAAAvjCBuwYJKoZIhvcNAQcGoIGtMIGqAgEAMIGkBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDDYWzH/s3oUmzlY9NAIBEIB3rf0n8Fy95D3fOTmjYXCkWiRn8LtxZLW9299cRFcaKs2nn9W8321/BQdJMX0TaM9WgosdTdzlsTmCUbuSR6oSbZjzx6i4onnp/cCrqhM6fMpglQrBnsMsBk2OYPLIzcWN1yhT/pF/jAgMvXzdvG9B+IZ0TpcVufo="
}

variable "encrypted_sentry_dsn_matches" {
  type        = string
  description = "KMS Encrypted Sentry DSN for fifa-matches"
  default     = "AQICAHgsJXuUfVcivglHv1qjCTscwl6NzhvBut8NdXoUXyTZ4gFoU3zFBzi196LmOkomVN5MAAAAvjCBuwYJKoZIhvcNAQcGoIGtMIGqAgEAMIGkBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDJtJtAM2Dy4hYAY+1gIBEIB3XF5ET8S3faV4E+os7/9TI6g7ScsK2EHvDjXobbAX/WLSaZy22A6ooUzteJTm/iE6TTivbDt9prwTH7ur6QSQ0agUoGNbJcFptvKZ3+GlxvZtHsmvToI25wCcyv8y52sCRbTbDIIPtUkJrEr9eIvJejR6a3yfkOE="
}
