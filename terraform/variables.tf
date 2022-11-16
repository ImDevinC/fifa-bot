variable "encrypted_slack_url" {
  type        = string
  description = "KMS Encrypted Slack Webhook URL"
  default     = "AQICAHgsJXuUfVcivglHv1qjCTscwl6NzhvBut8NdXoUXyTZ4gE4gg5CPEm6Iia0eM9ECNlJAAAAszCBsAYJKoZIhvcNAQcGoIGiMIGfAgEAMIGZBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDM4HLQFcf2++ULSBVgIBEIBsbDLTWOnPFIaQIGXf+zchb03IvMpSmW0rpEPkfbnJYPUwK86RbGbsFgivjSFGwjcPjOiioy0PtzwayRh6u509CLqCp5s3mdpx/drxOhJ+hoPei6WygG9dJnjCPtMja36rq+tF79BnhdEmnvQh"
}

variable "encrypted_sentry_dsn" {
  type        = string
  description = "KMS Encrypted Sentry DSN"
  default     = "AQICAHgsJXuUfVcivglHv1qjCTscwl6NzhvBut8NdXoUXyTZ4gFaAirWoSrvywTIlRfzOldGAAAAvjCBuwYJKoZIhvcNAQcGoIGtMIGqAgEAMIGkBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDDYWzH/s3oUmzlY9NAIBEIB3rf0n8Fy95D3fOTmjYXCkWiRn8LtxZLW9299cRFcaKs2nn9W8321/BQdJMX0TaM9WgosdTdzlsTmCUbuSR6oSbZjzx6i4onnp/cCrqhM6fMpglQrBnsMsBk2OYPLIzcWN1yhT/pF/jAgMvXzdvG9B+IZ0TpcVufo="
}
