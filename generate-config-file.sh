#!/bin/sh

set -euo pipefail

# log_level is one of trace debug info warn error fatal panic
LOG_LEVEL="${LOG_LEVEL:-info}"

cat >memo.toml <<EOF
log_level = "${LOG_LEVEL}"

[slack]
enabled = true
bot_token = "${SLACK_BOT_TOKEN}"
app_token = "${SLACK_APP_TOKEN}"

[grafana]
api_key = "${GRAFANA_API_KEY}"
api_url = "${GRAFANA_API_URL}"
EOF
