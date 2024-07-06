# Environment Variables


MyWhoop uses environment variables to configure the application. The environment variables are used to configure the Whoop API client and other integrations such as notifications. 

### Order Of Precendence

The order of precendence is as follows.

1. CLI flags.
2. Environment Variables
3. Configuration File


## MyWhoop Variables

The following environment variables are used to configure MyWhoop. The two required variables are `WHOOP_CLIENT_ID` and `WHOOP_CLIENT_SECRET`.


| Variable | Description | Required |
|---|----|---|
| `WHOOP_CLIENT_ID` | The client ID for your Whoop application. | Yes |
| `WHOOP_CLIENT_SECRET` | The client secret for your Whoop application. | Yes |
| `WHOOP_CREDENTIALS_FILE` | The file path to the Whoop credentials file that contains a valid Whoop authentication token. Default value is `token.json`. | No | 


### Notification  Variables

Depending on the notification service you use, you may need to provide additional environment variables.

| Variable | Description | Required |
|---|----|---|
| `NOTIFICATION_NTFY_AUTH_TOKEN`| The token for the [Ntfy](https://docs.ntfy.sh/) service. Required if the ntfy subscription requires a token. | No |
| `NOTIFICATION_NTFY_PASSWORD` | The password for the ntfy subscription if username/password authentication is used. Required if the ntfy subscription requires a username and password. | No |

