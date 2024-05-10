
[![Go version](https://img.shields.io/github/go-mod/go-version/karl-cardenas-coding/go-lambda-cleanup)](https://golang.org/dl/)

# MyWhooop 

A tool for gathering and retaining your own Whoop data. 



## Overview

MyWhoop was created to help faciliate ownership of your data and allow you to interfact with your own data in different ways than what Whoop may offer. 


## Environment Variables

The following environment variables are available for use with MyWhoop.

| Variable | Description | Required |
|---|----|---|
| `WHOOP_CLIENT_ID` | The client ID for your Whoop application. | Yes |
| `WHOOP_CLIENT_SECRET` | The client secret for your Whoop application. | Yes |
| `WHOOP_CREDENTIALS_FILE` | The file path to the Whoop credentials file that contains a valid Whoop authentication token. Default value is `token.json`. | No | 

### Order Of Precendence

The order of precendence is as follows.

1. CLI flags.
2. Environment Variables
3. Configuration File

