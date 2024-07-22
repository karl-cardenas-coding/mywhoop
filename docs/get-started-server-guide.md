# Setup Server Mode

MyWhoop can be setup to automatically download the last 24 hours of data. This feature is useful if you want to download your Whoop data daily and backup the data locally or to a remote location. 


## Prerequisites

Before you begin, ensure you have the following items completed:

- Review and complete all the steps in the [Getting Started](/docs/get-started.md) guide.

- Remote access to a Linux machine or server.

- Internet access to download the MyWhoop binary.

- Elevated permissions to configure systemd services.

- `wget`, and `unzip` installed on your Linux machine.

- A Linux machine with systemd and systemctl installed. 

> [!NOTE]
> You can use other operating systems and system services tooling to ensure MyWhoop starts up automatically. In this guide, Ubuntu with systemd is used.


## Steps


1.  Open a terminal session and log into your Linux machine. 


2. Download the MyWhoop binary to your Linux machine. You can download the binary from the [MyWhoop releases page](https://github.com/karl-cardenas-coding/mywhoop/releases).  The following command will download the latest release of MyWhoop for Linux x86_64.

    ```bash
    wget https://github.com/karl-cardenas-coding/mywhoop/releases/latest/download/mywhoop_darwin_x86_64.zip --output-document mywhoop.zip
    ```
> [!NOTE]
> <details><summary>Why not use Docker 🐳</summary><br>
>
>
>   Monitoring and managing Docker containers is not as trivial as using a   binary. If you are interested in using the MyWhoop Docker container with >  systemd, check out the [Running Docker Containers with Systemd](https://blog.container-solutions.com/running-docker-containers-with-systemd) to get an idea of how to use Docker containers with systemd. 
> </details>

3. Unzip the MyWhoop binary and move it to the `/usr/local/bin/` directory. 

    ```bash
    unzip mywhoop.zip && rm mywhoop.zip \
    && chmod +x mywhoop && sudo mv mywhoop /usr/local/bin/
    ```

4. Create a new directory to store the MyWhoop token and data.

    ```bash
    mkdir -p ~/mywhoop
    ```

5. Create a new MyWhoop configuration in your $HOME directory. The following command will create a new configuration file in your $HOME directory. 

> [!TIP]
> To learn more about the MyWhoop configuration file, refer to the [Configuration Reference](./docs/configuration_reference.md) section.

```bash
cat <<EOF > ~/.mywhoop.yaml
export:
  method: file
  fileExport:
      fileName: ""
      filePath: ""
      fileType: "json"
      fileNamePrefix: ""
      serverMode: true
server:
  enabled: true
  crontab: "* 13 * * *"
  jwtRefreshDuration: 45
debug: info
```

6. Create a new systemd service file to start MyWhoop automatically. The following command will create a new systemd service file in the `/etc/systemd/system/` directory. 

> [!NOTE]
> The working directory is set to the $`/home/ubuntu/mywhoop` directory. You can change the working directory to any directory where you want to store the MyWhoop token and data. In this guide the `/home/ubuntu/mywhoop` directory is used. The same applies to the user and group settings. Change the user and group settings to the appropriate user and group on your system.


```bash
sudo cat <<EOF > /etc/systemd/system/mywhoop.service
[Unit]
Description=MyWhoop
Documentation="https://github.com/karl-cardenas-coding/mywhoop"
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/mywhoop server --config /home/ubuntu/.mywhoop.yaml
Restart=on-failure
User=ubuntu
Group=ubuntu
WorkingDirectory=/home/ubuntu/mywhoop
Environment="WHOOP_CLIENT_ID=*************"
Environment="WHOOP_CLIENT_SECRET==*************"
Environment="WHOOP_CREDENTIALS_FILE=/home/ubuntu/mywhoop/token.json

[Install]
WantedBy=multi-user.target
EOF
```

7. Update the systemd service file with your Whoop client ID and client secret. Replace `*************` with your Whoop client ID and client secret.


8. The next step is to authenticate with the Whoop API and save the authentication token in `/home/ubuntu/mywhoop/token.json`. If your system has a GUI and web browser, you can use the `mywhoop login` command to authenticate with the Whoop API. 

    ```bash
    mywhoop login --credentials /home/ubuntu/mywhoop/token.json
    ```

    If your system does not have a GUI and web browser, use a different machine to authenticate with the Whoop API and save the token to a file or use the clipboard to copy the token. Transfer the file or token content to the machine where your want to use MyWhoop in server mode. Copy the token to the `/home/ubuntu/mywhoop/token.json` file.

9. Start the MyWhoop service and enable it to start automatically on boot. 

    ```bash
    sudo systemctl eanble mywhoop.service
    ```

10. Start the MyWhoop service.

    ```bash
    sudo systemctl start mywhoop.service
    ```

11. Verify the MyWhoop service is up and available. MyWhoop automatically refreshes token upon startup and every 45 minutes after that.

    ```bash
    sudo systemctl status mywhoop.service
    ```

    ```shell
    ● mywhoop.service - MyWhoop
     Loaded: loaded (/etc/systemd/system/mywhoop.service; enabled; vendor preset: enabled)
     Active: active (running) since Sun 2024-07-21 11:18:28 MST; 1 day 5h ago
       Docs: https://github.com/karl-cardenas-coding/mywhoop
    Main PID: 34988 (mywhoop)
        Tasks: 10 (limit: 18783)
        Memory: 5.6M
            CPU: 4.229s
        CGroup: /system.slice/mywhoop.service
                └─34988 /usr/local/bin/mywhoop server

    Jul 22 13:33:28 beelink-s12-pro mywhoop[34988]: time="2024/07/22 13:33:28" level=INFO msg="Refreshing auth token token"
    Jul 22 13:33:28 beelink-s12-pro mywhoop[34988]: time="2024/07/22 13:33:28" level=INFO msg="New token generated:" ECaL=....
    ```

Depending on what time you started the service, MyWhoop will download the last 24 hours of data at the specified time in the `server.crontab`. By default, this value is to `1 pm | 13:00`. You can change the crontab to any time you want MyWhoop to download the data.


13. At 1 pm, MyWhoop will download the last 24 hours of data. You can verify the data is downloaded by checking the `/home/ubuntu/mywhoop/data/` directory. The data will be saved in a file name containing today's date. 

    ```bash
    ls -l /home/ubuntu/mywhoop/data/
    ```


14. If you want to modify the MyWhoop configuration file, you can do so by editing the `~/.mywhoop.yaml` file. After making changes, restart the MyWhoop service to apply the changes.

    ```bash
    sudo systemctl restart mywhoop.service
    ```

🎊 You have successfully setup MyWhoop in server mode. MyWhoop will automatically download the last 24 hours of data daily at the specified time. You can configure MyWhoop to save the data locally or to a remote location, such as AWS S3. For more advanced configurations, refer to the [Configuration Reference](./docs/configuration_reference.md) section.