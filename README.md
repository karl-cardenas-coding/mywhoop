
[![Go version](https://img.shields.io/github/go-mod/go-version/karl-cardenas-coding/go-lambda-cleanup)](https://golang.org/dl/)

# MyWhoop 

A tool for gathering and retaining your own Whoop data. 

<p align="center">
  <img src="/static/images/logo.webp" alt="drawing" width="600"/>
</p>

## Overview

MyWhoop is a tool intended to help you take ownership of your Whoop data. You can use MyWhoop to interface with your own data in different ways than what Whoop may offer or intend.  MyWhoop is designed to be a simple and easy to use and designed to be ran on your own machine or server. It supports the following features:

- 🔐 **Login**: A simple interface to login into the Whoop devloper portal and save an authentication token locally. The token is required for interacting with the Whoop API.
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

| Flag | Description | Required | Default |
|---|----|---|---|
| `--config` | The file path to the MyWhoop configuration file. | No | `~/.mywhoop.yaml` |
| `--credentials` | The file path to the Whoop credentials file that contains a valid Whoop authentication token. | No | `token.json` |
| `--debug` | Enable debug logging. | No | False |


> [!IMPORTANT]
> For more information on the MyWhoop configuration file, refer to the [Configuration Reference](./docs/configuration_reference.md) section.



### Dump

The dump command is used to download your Whoop data and save it to a local file. For more advanced configurations, use a Mywhoop configuration file. Refer to the [Configuration Reference](./docs/configuration_reference.md) section for more information.

```bash
mywhoop dump
```


### Login

The login command is used to authenticate with the Whoop API and save the authentication token locally. The command will standup a local HTTP server to handle the OAuth2 handshake with the Whoop API and save the token to a local file.

```bash
mywhoop login
```

| Flag | Description | Required | Default |
|---|----|---|---|
| `--no-auto-open` | By default, the login command will automatically open a browser window to the Whoop login page. Use this flag to disable this behavior. | No | False |
| `--port` | The port to use for the local HTTP server. | No | `8080` |
| `--redirect-url` | The redirect URL to use for the OAuth2 handshake. | No | `http://localhost:8080/redirect`. |






