# VideoControl configuration

Default config path:

```bash
/etc/videocontrol.conf
```

Custom config path:

```bash
VideoControl --run --config ./videocontrol.conf
```

The format is JSON. A commented JSONC example is available in:

```text
configs/videocontrol.conf.example.jsonc
```

## logging

```json
"logging": {
  "level": "info",
  "file": "/var/log/videocontrol.log",
  "to_console": false
}
```

- `level`: `debug`, `info`, `warn`, `error`.
- `file`: log file path. Empty means console only.
- `to_console`: also print logs to stdout.

## storage

```json
"storage": {
  "base_dir": "/share"
}
```

Base directory for local recordings.

Each camera gets a separate subdirectory.

## motion

```json
"motion": {
  "min_contour_area": 500,
  "threshold": 25
}
```

- `min_contour_area`: minimum movement area. Increase this to ignore small noise.
- `threshold`: binary threshold for frame difference.

## recording

```json
"recording": {
  "frame_width": 640,
  "frame_height": 360,
  "fps": 10,
  "codec": "mp4v",
  "post_recording": "5s",
  "reconnect_delay": "5s"
}
```

- `frame_width`, `frame_height`: output video size.
- `fps`: output video FPS.
- `codec`: OpenCV codec, usually `mp4v`.
- `post_recording`: how long to continue recording after motion stops.
- `reconnect_delay`: wait before reconnecting to a failed RTSP stream.

## notification_file

```json
"notification_file": {
  "enabled": true,
  "path": "/share/video.save"
}
```

Writes the path of the last saved video to a file.

Useful for integration with SmartHome.

## webdav

```json
"webdav": {
  "enabled": false,
  "url": "https://webdav.yandex.ru",
  "username": "user@example.com",
  "password": "CHANGE_ME",
  "remote_base_dir": "Camera",
  "timeout": "180s"
}
```

Set `enabled=false` to disable cloud upload.

Works with Yandex Disk WebDAV and other WebDAV-compatible services.

## cameras

```json
"cameras": [
  {
    "id": "front_door",
    "name": "Front Door",
    "enabled": true,
    "rtsp": "rtsp://login:password@192.168.1.10:554/media/video2",
    "description": "Main entrance camera"
  }
]
```

- `id`: unique camera ID.
- `name`: folder name and display name.
- `enabled`: whether the camera is active.
- `rtsp`: RTSP stream URL.
- `description`: optional note.
