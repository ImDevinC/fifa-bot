# fifa-bot
This repository contains the code to run a bot designed to watch different FIFA matches and post the contents to a Slack channel.

## Usage
This bot has two pieces, a `match-watcher` and an `event-watcher`. Each runs as a separate lambda but has different triggers.

### match-watcher
The `match-watcher` is triggered by a CloudWatch Event Rule that runs on a schedule. This schedule can be adjusted in `terraform/cloudwatch.tf`. It's currently set to 1 minute to make sure we start watching a match as soon as it's available.
When this bot is triggered, it queries the FIFA API for live matches. It then takes this list of live matches and queries dynamoDB to see if the match is already being watched. If it is, the match is skipped. If the match is not being watched, then the match info is added to dynamoDB and a message is sent to SQS for `event-watcher` to pickup.

### event-watcher
The `event-watcher` is triggered by a message on the SQS queue. When the bot receives a message, it performs the following:
1. The events for the match are looked up from the FIFA API
1. The list of events are looped over and compared to the `lastEvent` message attribute that was attached to the SQS message
1. All events are skipped until the existing eventId is found, and then each event after that is processed into a string
  - Some events are skipped, which is controlled in `pkg/fifa/consts.go`. This helps remove non-essential events from spamming Slack.
1. The final list of events are sent to Slack
1. The bot determines if this match is still active by looking at the final event and also looking at the list of live matches.
1. If the match is not over, a new message is put onto the SQS queue with all the relevant match information, and an updated `lastEvent` for querying again.
  - This message has a 1 minute delay to not overutilize the lambda events.

## Installing
The bot is currently designed to run as a lambda on AWS. The following resources are needed:
 - Two lambdas (`match-watcher` & `event-watcher`)
 - A CloudWatch event rule to trigger `match-watcher`
 - An SQS queue for `match-watcher` to send matches to `event-watcher`
 - A DynamoDB table for `match-watcher` to determine which matches are being watched
 - A Slack Webhook URL
 - A KMS key to encrypt/decrypt variables

If you are familiar with terraform, you can use terraform to create all the needed resources.
### Required terraform updates
1. Update the region and bucket info in `terraform/main.tf` to match your accounts information
1. Base64Encode your Slack Webhook URL `echo -n <SLACK_WEBHOOK_URL> | base64`
1. Encrypt your Slack Webhook URL using the following command `aws kms encrypt --key-id <KMS_KEY_ID> --plaintext <BASE64_SLACK_WEBHOOK_URL>`
1. Copy the ciphertextBlob and paste it into the `encrypted_slack_url` in `terraform/variables.tf`
1. Run `make dist` in the root folder to generate the zip file for each lambda
1. In the `terraform` directory, run `terraform apply`