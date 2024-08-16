# Configuration Reference

A MyWhoop configuration file is used to configure MyWhoop's behavior and enable advanced features. By default, MyWhoop searches for a YAML file located at `~/.mywhoop.yaml`. You can specify a different configuration file using the `--config` flag. The configuration file is divided into the following sections:


> [!IMPORTANT]
> You can learn more about supported environment variables in the [Environment Variables](./environment_variables.md) section.



## Credentials

The credentials section of the configuration file is used to configure where the MyWhoop authentication token is stored or where to find it. The following fields are available for configuration:

| Field | Description | Required | Default |
|---|----|---|---|
| `credentialsFile` | The file path to the Whoop credentials file that contains a valid Whoop authentication token. By default, MyWhoop looks for a token file in the local directory. | No | `token.json` |

```yaml
credentials:
    credentialsFile: "/opt/mywhoop/token.json"
```

## Debug

The debug section of the configuration file is used to enable debug logging for MyWhoop. The following fields are available for configuration:

| Field | Description | Required | Default |
|---|----|---|---|
| `debug` | Enable debug logging. Allowed values are: `INFO`, `DEBUG`, `WARN`, and `ERROR`. | No | `INFO` |


```yaml
debug: "info"
```

## Export

The export section of the configuration file is used to configure the data export feature of MyWhoop. The export feature allows you to export your Whoop data to a remote location such as an S3 bucket. The following fields are available for configuration:

| Field | Description | Required | Default |
|---|----|---|---|
| `method` | The export method to use. Allowed values are `file` and `s3`. | Yes | |
| `fileExport` | The file export configuration. Required if `method` is `file`. | No | |
| `s3Export` | The S3 export configuration. Required if `method` is `s3`. | No | |


### File Export

Local file export accepts the following fields.

| Field | Description | Required | Default |
|---|----|---|---|
| `fileName` | The name of the file to export. | Yes | `user` |
| `filePath` | The path to save the file. By default, a data folder is created in the immediate folder. | Yes | `data/` | 
| `fileType` | The file type to save the file as. Allowed values are `json`, and `xlsx`. | Yes | `json` |
| `fileNamePrefix` | The prefix to add to the file name. In server mode, the data is automatically inserted as a prefix. | No |`""` |
| `serverMode` | Ensures the file name is unique and contains a timestamp. Ensures behaviors match server mode. | No | `false` |

```yaml
export:
 method: file
 fileExport:
    fileName: "user"
    filePath: "app/"
    fileType: "json"
    fileNamePrefix: ""
```

### S3 Export

You can export your Whoop data to an AWS S3 bucket using the S3 export method. By default, AWS credentials are honored from the environment variables or the default profile in the [order of precedence of the AWS SDK](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials). The S3 export method accepts the following fields:

| Field | Description | Required | Default |
|---|----|---|---|
| `bucket` | The S3 bucket to export the data to. | Yes | |
| `region` | The AWS region to use. | Yes | |
| `profile` | The AWS profile to use. | No | |
| `fileConfig` | The file configuration for the S3 export. Refer to the [file export configuration](#configuration-reference) for more information. | Yes | |


```yaml
export:
  method: s3
  awsS3:
    region: "us-east-1"
    bucket: "my-example-bucket"
    fileConfig:
      fileName: ""
      fileType: "json"
      fileNamePrefix: ""
      serverMode: true
```

## Notification

The notification section of the configuration file is used to configure the notification feature of MyWhoop. By default, notifications are sent to stdout. The notification feature allows you to use a different service to receive notifications when MyWhoop completes a task. The following fields are available for configuration:

| Field | Description | Required | Default |
|---|----|---|---|
|`method` | The notification method to use. Allowed values are `ntfy`, or `""`.  | Yes | `""`|
| `ntfy` | The ntfy notification configuration. Required if `method` is `ntfy`. | No | |

### Ntfy

You can use the open-source service, [Ntfy](https://docs.ntfy.sh/), to receive notifications when MyWhoop completes a task or an error is encountered. The Ntfy configuration block accepts the following fields:

| Field | Description | Required | Default |
|---|----|---|---|
| `events` | The events to receive notifications for. Allowed values are `""`, `all`, `success`, and `error`. | Yes | `all` |
| `serverEndpoint` | The Ntfy server endpoint to send notifications to. | Yes | `""` |
| `subscriptionID` | The subscription ID to use for the Ntfy subscription. | Yes | `""` |
| `userName` | The username for the Ntfy subscription. Required if user name and password is used. | No | `""` | 

```yaml
notification:
  method: "ntfy"
  ntfy:
    serverEndpoint: "https://example.my.ntfy.com"
    subscriptionID: "mywhoop_custom_notifications"
    events: "all"
```

> [!IMPORTANT]
>  Use the environment variables `NOTIFICATION_NTFY_AUTH_TOKEN` or `NOTIFICATION_NTFY_PASSWORD` to provide the Ntfy authentication credentials. 



## Server

The server section of the configuration file is used to configure the server feature of MyWhoop. The server feature allows you to start MyWhoop as a server that queries the Whoop API every 24 hours. The following fields are available for configuration:

| Field | Description | Required | Default |
|---|----|---|---|
| `enabled` | Enable the server feature. | No | `false` |
| `crontab` | The crontab schedule to use for the server. By default, the server is configured to download your Whoop data daily at 1pm (13:00). Your local time zone is used. Different Operating System implement local time zone differently. Refer to the Go [time.Location](https://pkg.go.dev/time#Local) for additional details on expected behavior. | No | `0 13 * * *` |
| `jwtRefreshDuration` | The duration to refresh the Whoop API JWT token provided. Default is 45 minutes. This value must be greater than 0 and less than 59 minutes.| No | `45` |


```yaml
server:
  enabled: true
  crontab: "*/55 * * * *" 
  jwtRefreshDuration: 45
```


## Example Configuration File


The following is an example configuration file that configures MyWhoop to export data to an S3 bucket, enable the server feature, and send notifications using the Ntfy service.

```yaml
export:
  method: s3
  awsS3:
    region: "us-east-1"
    bucket: "42acg-primary-data-whoop-bucket"
    fileConfig:
      fileName: ""
      fileType: "json"
      fileNamePrefix: ""
      serverMode: true
server:
  enabled: true
  firstRunDownload: false
notification:
  method: "ntfy"
  ntfy:
    serverEndpoint: "https://ntfy.self-hosted.example"
    subscriptionID: "mywhoop_alerts_at_home"
    events: "all"
debug: info
```


The following is an example configuration file configuring MyWhoop to export data to a local file and enabling debug logging.

```yaml
export:
  method: file
  fileExport:
    fileName: "user"
    filePath: "/opt/mywhoop/data/"
    fileType: "json"
    fileNamePrefix: ""
credentials:
  credentialsFile: "/opt/mywhoop/token.json"
debug: debug
```