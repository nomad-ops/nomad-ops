# Server configuraiton

This document describes all the available environment variable settings used in the `main.go` file of the project.

## General Settings

- **TRACE**
    - Description: Enables trace logging.
    - Default: `FALSE`
    - Example: `TRACE=TRUE`

## Slack Notifier Settings

- **SLACK_WEBHOOK_URL**
    - Description: The webhook URL for Slack notifications.
    - Default: `""`
    - Example: `SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...`

- **SLACK_BASE_URL**
    - Description: The base URL for Slack notifications.
    - Default: `localhost:3000/ui/sources/`
    - Example: `SLACK_BASE_URL=https://example.com/ui/sources/`

- **SLACK_ICON_SUCCESS**
    - Description: The icon for successful Slack notifications.
    - Default: `:check:`
    - Example: `SLACK_ICON_SUCCESS=:white_check_mark:`

- **SLACK_ICON_ERROR**
    - Description: The icon for error Slack notifications.
    - Default: `:check-no:`
    - Example: `SLACK_ICON_ERROR=:x:`

- **SLACK_ENV_INFO_TEXT**
    - Description: The footer text for Slack notifications.
    - Default: `Sent by nomad-ops (dev)`
    - Example: `SLACK_ENV_INFO_TEXT=Sent by nomad-ops (production)`

## Webhook Notifier Settings

- **WEBHOOK_URL**
    - Description: The webhook URL for notifications.
    - Default: `""`
    - Example: `WEBHOOK_URL=https://example.com/webhook`

- **WEBHOOK_TIMEOUT**
    - Description: The timeout for webhook notifications.
    - Default: `10s`
    - Example: `WEBHOOK_TIMEOUT=30s`

- **WEBHOOK_METHOD**
    - Description: The HTTP method for webhook notifications.
    - Default: `""`
    - Example: `WEBHOOK_METHOD=POST`

- **WEBHOOK_INSECURE**
    - Description: Allows insecure connections for webhook notifications.
    - Default: `FALSE`
    - Example: `WEBHOOK_INSECURE=TRUE`

- **WEBHOOK_LOG_TEMPLATE_RESULTS**
    - Description: Logs the template results for webhook notifications in server logs.
    - Default: `FALSE`
    - Example: `WEBHOOK_LOG_TEMPLATE_RESULTS=TRUE`

- **WEBHOOK_FIRE_ON**
    - Description: Specifies when to fire the webhook notifications.
    - Default: `success`
    - Example: `WEBHOOK_FIRE_ON=success,error`

- **WEBHOOK_AUTH_HEADER_NAME**
    - Description: The name of the authorization header for webhook notifications.
    - Default: `""`
    - Example: `WEBHOOK_AUTH_HEADER_NAME=Authorization`

- **WEBHOOK_AUTH_HEADER_VALUE_FILE**
    - Description: The file that contains the authorization header value for webhook notifications.
    - Default: `""`
    - Example: `WEBHOOK_AUTH_HEADER_VALUE_FILE=/secrets/webhook-auth`

- **WEBHOOK_BODY_TEMPLATE_FILE**
    - Description: The file that contains the body template of the webhook request.
    - Default: `""`
    - Example: `WEBHOOK_BODY_TEMPLATE_FILE=/templates/webhook-body.tpl`

- **WEBHOOK_QUERY_TEMPLATE_FILE**
    - Description: The file that contains the query template of the webhook request.
    - Default: `""`
    - Example: `WEBHOOK_QUERY_TEMPLATE_FILE=/templates/webhook-query.tpl`

## PocketBase Settings

- **POCKETBASE_APP_NAME**
    - Description: The name of the PocketBase application.
    - Default: `Nomad-Ops`
    - Example: `POCKETBASE_APP_NAME=MyApp`

- **POCKETBASE_APP_URL**
    - Description: The URL of the PocketBase application.
    - Default: `http://localhost:8090`
    - Example: `POCKETBASE_APP_URL=https://example.com`

- **POCKETBASE_SENDER_NAME**
    - Description: The sender name for PocketBase emails.
    - Default: `Support`
    - Example: `POCKETBASE_SENDER_NAME=Admin`

- **POCKETBASE_SENDER_ADDRESS**
    - Description: The sender address for PocketBase emails.
    - Default: `support@localhost.com`
    - Example: `POCKETBASE_SENDER_ADDRESS=support@example.com`

- **POCKETBASE_HIDE_CONTROLS**
    - Description: Hides controls in the PocketBase UI.
    - Default: `TRUE`
    - Example: `POCKETBASE_HIDE_CONTROLS=FALSE`

- **POCKETBASE_ENABLE_SMTP**
    - Description: Enables SMTP for PocketBase.
    - Default: `FALSE`
    - Example: `POCKETBASE_ENABLE_SMTP=TRUE`

- **POCKETBASE_SMTP_HOST**
    - Description: The SMTP host for PocketBase.
    - Default: `localhost`
    - Example: `POCKETBASE_SMTP_HOST=smtp.example.com`

- **POCKETBASE_SMTP_PORT**
    - Description: The SMTP port for PocketBase.
    - Default: `25`
    - Example: `POCKETBASE_SMTP_PORT=587`

- **POCKETBASE_SMTP_USERNAME**
    - Description: The SMTP username for PocketBase.
    - Default: `""`
    - Example: `POCKETBASE_SMTP_USERNAME=user@example.com`

- **POCKETBASE_SMTP_PASSWORD**
    - Description: The SMTP password for PocketBase.
    - Default: `""`
    - Example: `POCKETBASE_SMTP_PASSWORD=secret`

- **POCKETBASE_SMTP_AUTH_METHOD**
    - Description: The SMTP authentication method for PocketBase.
    - Default: `PLAIN`
    - Example: `POCKETBASE_SMTP_AUTH_METHOD=LOGIN`

- **POCKETBASE_SMTP_TLS**
    - Description: Enables TLS for SMTP in PocketBase.
    - Default: `FALSE`
    - Example: `POCKETBASE_SMTP_TLS=TRUE`

- **POCKETBASE_AUTH_MICROSOFT_ENABLED**
    - Description: Enables Microsoft authentication for PocketBase.
    - Default: `FALSE`
    - Example: `POCKETBASE_AUTH_MICROSOFT_ENABLED=TRUE`

- **POCKETBASE_AUTH_MICROSOFT_CLIENT_ID**
    - Description: The Microsoft client ID for PocketBase authentication.
    - Default: `""`
    - Example: `POCKETBASE_AUTH_MICROSOFT_CLIENT_ID=your-client-id`

- **POCKETBASE_AUTH_MICROSOFT_CLIENT_SECRET**
    - Description: The Microsoft client secret for PocketBase authentication.
    - Default: `""`
    - Example: `POCKETBASE_AUTH_MICROSOFT_CLIENT_SECRET=your-client-secret`

- **POCKETBASE_AUTH_MICROSOFT_AUTH_URL**
    - Description: The Microsoft authentication URL for PocketBase.
    - Default: `""`
    - Example: `POCKETBASE_AUTH_MICROSOFT_AUTH_URL=https://login.microsoftonline.com/...`

- **POCKETBASE_AUTH_MICROSOFT_TOKEN_URL**
    - Description: The Microsoft token URL for PocketBase.
    - Default: `""`
    - Example: `POCKETBASE_AUTH_MICROSOFT_TOKEN_URL=https://login.microsoftonline.com/...`

- **TEAM_NAME_MICROSOFT_PROPERTY**
    - Description: The Microsoft team name property.
    - Default: `""`
    - Example: `TEAM_NAME_MICROSOFT_PROPERTY=team-name`

## Nomad Settings

- **NOMAD_TOKEN_FILE**
    - Description: The file path for the Nomad token.
    - Default: `""`
    - Example: `NOMAD_TOKEN_FILE=/path/to/token`

- **NOMAD_OPS_LOCAL_REPO_DIR**
    - Description: The local repository directory for Nomad Ops.
    - Default: `repos`
    - Example: `NOMAD_OPS_LOCAL_REPO_DIR=/path/to/repos`

## Monitor Settings

- **MONITOR_ADDRESS**
    - Description: The address for the monitor.
    - Default: `:8080`
    - Example: `MONITOR_ADDRESS=:9090`

## Default User Settings

- **DEFAULT_USER_EMAIL**
    - Description: The default email for the initial user.
    - Default: `user@nomad-ops.org`
    - Example: `DEFAULT_USER_EMAIL=admin@example.com`

- **DEFAULT_USER_NAME**
    - Description: The default username for the initial user.
    - Default: `user`
    - Example: `DEFAULT_USER_NAME=admin`

- **DEFAULT_USER_PASSWORD**
    - Description: The default password for the initial user.
    - Default: `simple-nomad-ops`
    - Example: `DEFAULT_USER_PASSWORD=securepassword`

## Default Admin Settings

- **DEFAULT_ADMIN_EMAIL**
    - Description: The default email for the initial admin user.
    - Default: `admin@nomad-ops.org`
    - Example: `DEFAULT_ADMIN_EMAIL=admin@example.com`

- **DEFAULT_ADMIN_PASSWORD**
    - Description: The default password for the initial admin user.
    - Default: `simple-nomad-ops`
    - Example: `DEFAULT_ADMIN_PASSWORD=securepassword`