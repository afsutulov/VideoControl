# VideoControl

[Русский](README.RU.md)

VideoControl is a lightweight RTSP motion-recording service written in Go. It is designed as a companion project for small SmartHome installations and can be used as a minimal alternative to heavy video surveillance stacks.

It watches one or more RTSP cameras, records short MP4 clips when motion is detected, stores videos locally, optionally writes a trigger file for another service, and can optionally upload recordings to Yandex Disk or any other WebDAV-compatible cloud storage.

## Features

- Multiple RTSP cameras.
- Motion detection using OpenCV / GoCV.
- Separate folder per camera.
- Local MP4 recording.
- Optional trigger file, for example `/share/video.save`.
- Optional WebDAV upload.
- JSON configuration file.
- Configurable logging.
- `--run` service mode.
- Systemd unit example.

## Project status

This project is intentionally small. It is not a full NVR like Frigate, ZoneMinder, or Home Assistant. It is useful when you only need motion-triggered video clips and simple integration with your own SmartHome/MQTT automation.

## Requirements

- Linux server.
- Go 1.22+.
- OpenCV installed.
- GoCV installed and correctly linked to OpenCV.
- RTSP-capable camera.

GoCV installation depends on your OS. See the official GoCV documentation before building.

## Build

```bash
git clone https://github.com/afsutulov/VideoControl.git
cd VideoControl
go mod tidy
go build -o VideoControl .
```

## Configuration

Copy the example configuration:

```bash
sudo mkdir -p /etc
sudo cp configs/videocontrol.conf.example.json /etc/videocontrol.conf
sudo nano /etc/videocontrol.conf
```

Main file:

```bash
/etc/videocontrol.conf
```

You may also use a custom path:

```bash
./VideoControl --run --config ./videocontrol.conf
```

## Run

Without `--run`, VideoControl only prints a short help message:

```bash
./VideoControl
```

Run as a service process:

```bash
./VideoControl --run --config /etc/videocontrol.conf
```

Show version:

```bash
./VideoControl --version
```

## Install as systemd service

```bash
sudo cp VideoControl /usr/local/bin/VideoControl
sudo cp configs/videocontrol.conf.example.json /etc/videocontrol.conf
sudo cp systemd/videocontrol.service /etc/systemd/system/videocontrol.service

sudo systemctl daemon-reload
sudo systemctl enable videocontrol
sudo systemctl start videocontrol
sudo journalctl -u videocontrol -f
```

Or use the helper script:

```bash
sudo ./scripts/install-systemd.sh
```

## WebDAV / Yandex Disk

WebDAV upload is optional.

To enable upload to Yandex Disk:

```json
"webdav": {
  "enabled": true,
  "url": "https://webdav.yandex.ru",
  "username": "your-login@yandex.ru",
  "password": "your-app-password",
  "remote_base_dir": "Camera",
  "timeout": "180s"
}
```

Use an app password if your account requires it.

## Directory layout

For camera `Front Door`, the local path will look like:

```text
/share/Front Door/2026/05/17/12-30-45.mp4
```

If WebDAV is enabled, the remote path will look like:

```text
Camera/Front Door/2026/05/17/12-30-45.mp4
```

## Integration with SmartHome (https://github.com/afsutulov/SmartHome)

If `notification_file.enabled=true`, VideoControl writes the local video path into:

```text
/share/video.save
```

A separate SmartHome service can watch this file and send the video link or local path to Telegram.

## License

MIT
