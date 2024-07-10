
[![Go version](https://img.shields.io/github/go-mod/go-version/karl-cardenas-coding/go-lambda-cleanup)](https://golang.org/dl/)

# MyWhoop 

A tool for gathering and retaining your own Whoop data. 

<p align="center">
  <img src="/static/images/logo.webp" alt="drawing" width="600"/>
</p>

## Overview

MyWhoop is a tool intended to help you take ownership of your Whoop data. You can use MyWhoop to interface with your own data in ways that are different from what Whoop may offer or intend.  MyWhoop is designed to be simple and easy to use and designed to be deployed on your own machine or server. MyWhoop supports the following features:

- 🔐 **Login**: A simple interface to log into the Whoop developer portal and save an authentication token locally. The token is required to interact with the Whoop API.
- 🗄️ **Server**: Automatically download your Whoop data daily and save it to a local file or export it to a remote location.
- 📬 **Notifications**: Receive notifications when new data is available or when an error occurs.
- 💾 **Data Export**: Export your Whoop data to a remote location such as an S3 bucket.

## Get Started 🚀

Please check out the [Getting Started](/docs/get-started.md) guide to get started with MyWhoop.


## Commands

MyWhoop supports the following commands and global flags:

- [Dump](#dump) - Download your Whoop data and save it to a local file.
- [Login](#login) - Authenticate with the Whoop API and save the authentication token locally.
- [Help](#help) - Display help information for MyWhoop.
- [Server](#server) - Automatically download your Whoop data daily and save it to a local file or export it to a remote location.
- [Version](#version) - Display the version of MyWhoop.

#### Global Flags

| Long Flag | Short Flag |Description | Required | Default |
|---|--|---|---|---|
| `--config` | - |The file path to the MyWhoop [configuration file](./docs/configuration_reference.md). | No | `~/.mywhoop.yaml` |
| `--credentials` | - |The file path to the Whoop credentials file that contains a valid Whoop authentication token. | No | `token.json` |
| `--debug` | `-d` | Modify the output logging level. | No | `INFO` |


> [!IMPORTANT]
> For more information on the MyWhoop configuration file, refer to the [Configuration Reference](./docs/configuration_reference.md) section.



### Dump

The dump command downloads **all your Whoop data** and saves it to a local file. For more advanced configurations, use a Mywhoop configuration file. For more information, refer to the [Configuration Reference](./docs/configuration_reference.md) section.

```bash
mywhoop dump
```

| Long Flag | Short Flag | Description | Required | Default |
|---|---|---|---|---|
| `--location` | `-l` |The location to save the Whoop data file. | No | `./data/` |


> [!IMPORTANT]
> MyWhoop has exponential backoff and retries logic built in for the Whoop API. If the API is down or the request fails, MyWhoop will retry the request. Whoop has an [API rate limit of 100 requests per minute](https://developer.whoop.com/docs/developing/rate-limiting). If the rate limit is exceeded, MyWhoop will attempt to retry the request after a delay for up to a maximum of 5 minutes. If the request fails after 5 minutes, the application will exit with an error. If the Whoop API rejects the authentication token, the application will exit with an error.


### Login

The login command is used to authenticate with the Whoop API and save the authentication token locally. The command will set up a local HTTP server to handle the OAuth2 handshake with the Whoop API and save the token to a local file.

```bash
mywhoop login
```

| Long Flag | Short Flag |Description | Required | Default |
|---|--|--|---|---|
| `--no-auto-open` | `-n` |By default, the login command will automatically open a browser window to the Whoop login page. Use this flag to disable this behavior. | No | False |
| `--port` | `-p` | The port to use for the local HTTP server. | No | `8080` |
| `--redirect-url` | `-r` |The redirect URL to use for the OAuth2 handshake. | No | `http://localhost:8080/redirect`. |



## Server

The server command automatically downloads your Whoop data daily. If specified, it saves or exports the data to a local file or a remote location. It is designed to be started as a background process and will automatically download your Whoop data daily. The command will refresh the Whoop authentication token every 55 minutes and update the local token file. The Whoop API is queried precisely every 24 hours from when the server is started.   

Use a MyWhoop configuration file for more advanced configurations. For more information, refer to the [Configuration Reference](./docs/configuration_reference.md) section.


| Long Flag | Short Flag |Description | Required | Default |
|---|--|--|---|---|
| `--first-run-download` | - |Download all the available Whoop data on the first run. | No | False |


```bash
mywhoop server
```


## Version

The version command is used to display the version of MyWhoop. The version command checks for the latest version of MyWhoop and displays the current version. If a new version is available, the command will notify you.


```bash
mywhoop version
```
```
2024/07/06 10:50:29 INFO mywhoop v1.0.0
```

