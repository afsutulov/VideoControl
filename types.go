package main

import (
	"encoding/json"
	"time"
)

const (
	AppName    = "VideoControl"
	AppVersion = "1.0.0"

	defaultConfigPath = "/etc/videocontrol.conf"
)

// Config описывает все настройки приложения.
type Config struct {
	Logging          LoggingConfig      `json:"logging"`
	Storage          StorageConfig      `json:"storage"`
	Motion           MotionConfig       `json:"motion"`
	Recording        RecordingConfig    `json:"recording"`
	NotificationFile NotificationConfig `json:"notification_file"`
	WebDAV           WebDAVConfig       `json:"webdav"`
	Cameras          []CameraConfig     `json:"cameras"`
}

// LoggingConfig управляет журналированием приложения.
type LoggingConfig struct {
	Level     string `json:"level"`
	File      string `json:"file"`
	ToConsole bool   `json:"to_console"`
}

// StorageConfig задаёт локальную директорию для видеозаписей.
type StorageConfig struct {
	BaseDir string `json:"base_dir"`
}

// MotionConfig задаёт параметры детектора движения.
type MotionConfig struct {
	MinContourArea int `json:"min_contour_area"`
	Threshold      int `json:"threshold"`
}

// RecordingConfig задаёт параметры видеозаписи.
type RecordingConfig struct {
	FrameWidth    int      `json:"frame_width"`
	FrameHeight   int      `json:"frame_height"`
	FPS           float64  `json:"fps"`
	Codec         string   `json:"codec"`
	PostRecording Duration `json:"post_recording"`
	Reconnect     Duration `json:"reconnect_delay"`
}

// NotificationConfig создаёт файл-триггер для внешних систем, например SmartHome.
type NotificationConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

// WebDAVConfig управляет загрузкой записей в облако через WebDAV.
type WebDAVConfig struct {
	Enabled       bool     `json:"enabled"`
	URL           string   `json:"url"`
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	RemoteBaseDir string   `json:"remote_base_dir"`
	Timeout       Duration `json:"timeout"`
}

// CameraConfig описывает одну RTSP-камеру.
type CameraConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RTSP        string `json:"rtsp"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// Duration позволяет задавать длительность строкой: "5s", "2m", "180s".
type Duration struct {
	time.Duration
}

// UnmarshalJSON читает duration из JSON-строки или числа секунд.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		v, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		d.Duration = v
		return nil
	}
	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	d.Duration = time.Duration(n) * time.Second
	return nil
}

// MarshalJSON сохраняет duration в человекочитаемом виде.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}
