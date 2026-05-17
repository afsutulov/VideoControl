package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// loadConfig читает JSON-конфигурацию и выставляет значения по умолчанию.
func loadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var cfg Config
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, err
	}

	applyDefaults(&cfg)
	if err := validateConfig(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// applyDefaults заполняет безопасные значения по умолчанию.
func applyDefaults(cfg *Config) {
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.File == "" && !cfg.Logging.ToConsole {
		cfg.Logging.ToConsole = true
	}
	if cfg.Storage.BaseDir == "" {
		cfg.Storage.BaseDir = "/share"
	}
	if cfg.Motion.MinContourArea == 0 {
		cfg.Motion.MinContourArea = 500
	}
	if cfg.Motion.Threshold == 0 {
		cfg.Motion.Threshold = 25
	}
	if cfg.Recording.FrameWidth == 0 {
		cfg.Recording.FrameWidth = 640
	}
	if cfg.Recording.FrameHeight == 0 {
		cfg.Recording.FrameHeight = 360
	}
	if cfg.Recording.FPS == 0 {
		cfg.Recording.FPS = 10
	}
	if cfg.Recording.Codec == "" {
		cfg.Recording.Codec = "mp4v"
	}
	if cfg.Recording.PostRecording.Duration == 0 {
		cfg.Recording.PostRecording.Duration = 5 * time.Second
	}
	if cfg.Recording.Reconnect.Duration == 0 {
		cfg.Recording.Reconnect.Duration = 5 * time.Second
	}
	if cfg.NotificationFile.Path == "" {
		cfg.NotificationFile.Path = "/share/video.save"
	}
	if cfg.WebDAV.Timeout.Duration == 0 {
		cfg.WebDAV.Timeout.Duration = 180 * time.Second
	}
	if cfg.WebDAV.RemoteBaseDir == "" {
		cfg.WebDAV.RemoteBaseDir = "Камера"
	}
}

// validateConfig проверяет обязательные поля.
func validateConfig(cfg Config) error {
	if len(cfg.Cameras) == 0 {
		return errors.New("at least one camera must be configured")
	}
	ids := map[string]bool{}
	for _, cam := range cfg.Cameras {
		if !cam.Enabled {
			continue
		}
		if cam.ID == "" {
			return errors.New("camera.id is required")
		}
		if ids[cam.ID] {
			return errors.New("duplicate camera.id: " + cam.ID)
		}
		ids[cam.ID] = true
		if strings.TrimSpace(cam.RTSP) == "" {
			return errors.New("camera.rtsp is required for " + cam.ID)
		}
	}
	if len(ids) == 0 {
		return errors.New("no enabled cameras configured")
	}
	if cfg.WebDAV.Enabled {
		if cfg.WebDAV.URL == "" {
			return errors.New("webdav.url is required when webdav.enabled=true")
		}
		if cfg.WebDAV.Username == "" || cfg.WebDAV.Password == "" {
			return errors.New("webdav.username and webdav.password are required when webdav.enabled=true")
		}
	}
	return nil
}

// cameraDir возвращает локальную папку камеры за конкретную дату.
func cameraDir(baseDir string, cam CameraConfig, year string, month string, day string) string {
	return filepath.Join(baseDir, safeName(cam.NameOrID()), year, month, day)
}

// NameOrID возвращает имя камеры, а если его нет — ID.
func (c CameraConfig) NameOrID() string {
	if strings.TrimSpace(c.Name) != "" {
		return c.Name
	}
	return c.ID
}

// safeName делает имя безопасным для файловой системы.
func safeName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	if s == "" {
		return "camera"
	}
	return s
}
