# Configuration Reference

The MyWhoop configuration file is used to configure the behavior of MyWhoop and enable advanced features. The configuration file is a YAML file that is located at `~/.mywhoop.yaml` by default. The configuration file can be overridden by using the `--config` flag when issuing MyWhoop commands.


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
| `fileType` | The file type to save the file as. Allowed values are `json`. | Yes | `json` |
| `fileNamePrefix` | The prefix to add to the file name. In server mode, the data is automatically inserted as a prefix. | No | |
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

S3 export accepts the following fields.

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
## Credentials

The credentials section of the configuration file is used to configure where the Whoop authentication token is stored or where to find it. The following fields are available for configuration:

| Field | Description | Required | Default |
|---|----|---|---|
| `credentialsFile` | The file path to the Whoop credentials file that contains a valid Whoop authentication token. By default, MyWhoop looks for a token file in the local directory. | No | `token.json` |

```yaml
credentials:
    credentialsFile: "/opt/mywhoop/token.json"
```