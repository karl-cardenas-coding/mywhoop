
[![Go version](https://img.shields.io/github/go-mod/go-version/karl-cardenas-coding/go-lambda-cleanup)](https://golang.org/dl/)

# MyWhoop 

A tool for gathering and retaining your own Whoop data. 

<p align="center">
  <img src="/static/images/logo.webp" alt="drawing" width="600"/>
</p>

## Overview

MyWhoop is a tool intended to help you take ownership of your Whoop data. You can use MyWhoop to interfact with your own data in different ways than what Whoop may offer or intend.  MyWhoop is designed to be a simple and easy to use tool that can be run on your own machine or server. It supports the following features:

- 🔐 **Login**: A simple interface to login into Whoop and save your authentication token locally. The token is required for interacting with the Whoop API.
- 🗄️ **Server**: Automatically download your Whoop data daily and save it to a local file or export it to a remote location.
- 📬 **Notifications**: Receive notifications when new data is available or when an error occurs.
- 💾 **Data Export**: Export your Whoop data to a remote location such as an S3 bucket.

## Get Started 🚀

Please check out the [Getting Started](/docs/get-started.md) guide to get started with MyWhoop.


```shell
export WHOOP_CLIENT_ID=<your client id>
export WHOOP_CLIENT_SECRET=<your client
docker run -p 8080:8080 \
-e WHOOP_CLIENT_ID=$WHOOP_CLIENT_ID -e WHOOP_CLIENT_SECRET=$WHOOP_CLIENT_SECRET \
ghcr.io/karl-cardenas-coding/mywhoop:v1.0.0 login -n
```




## Environment Variables

The following environment variables are available for use with MyWhoop.

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

### Order Of Precendence

The order of precendence is as follows.

1. CLI flags.
2. Environment Variables
3. Configuration File

