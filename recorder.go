package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"gocv.io/x/gocv"
)

// runCameraLoop поддерживает подключение к RTSP-камере и переподключается при ошибках.
func (a *App) runCameraLoop(ctx context.Context, cam CameraConfig) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := a.runCamera(ctx, cam); err != nil {
			logWarn("camera %s stopped: %v", cam.ID, err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(a.cfg.Recording.Reconnect.Duration):
			logInfo("reconnecting camera: %s", cam.ID)
		}
	}
}

// runCamera читает кадры, ищет движение и пишет видео.
func (a *App) runCamera(ctx context.Context, cam CameraConfig) error {
	logInfo("opening camera %s: %s", cam.ID, cam.RTSP)
	webcam, err := gocv.OpenVideoCapture(cam.RTSP)
	if err != nil {
		return err
	}
	defer webcam.Close()

	frame := gocv.NewMat()
	defer frame.Close()
	resized := gocv.NewMat()
	defer resized.Close()
	gray := gocv.NewMat()
	defer gray.Close()
	prevGray := gocv.NewMat()
	defer prevGray.Close()
	delta := gocv.NewMat()
	defer delta.Close()
	thresh := gocv.NewMat()
	defer thresh.Close()

	motionDetected := false
	lastMotionTime := time.Now()
	var writer *gocv.VideoWriter
	var currentLocalPath string

	for {
		select {
		case <-ctx.Done():
			if writer != nil {
				writer.Close()
			}
			return nil
		default:
		}

		if !webcam.Read(&frame) || frame.Empty() {
			return errEmptyFrame
		}

		gocv.Resize(frame, &resized, imagePoint(a.cfg.Recording.FrameWidth, a.cfg.Recording.FrameHeight), 0, 0, gocv.InterpolationLinear)
		gocv.CvtColor(resized, &gray, gocv.ColorBGRToGray)
		if prevGray.Empty() {
			gray.CopyTo(&prevGray)
			continue
		}

		gocv.AbsDiff(gray, prevGray, &delta)
		gocv.Threshold(delta, &thresh, float32(a.cfg.Motion.Threshold), 255, gocv.ThresholdBinary)
		contours := gocv.FindContours(thresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)

		if hasMotion(contours, float64(a.cfg.Motion.MinContourArea)) {
			if !motionDetected {
				currentLocalPath, writer, err = a.startRecording(cam)
				if err != nil {
					contours.Close()
					return err
				}
				logInfo("camera %s started recording: %s", cam.ID, currentLocalPath)
			}
			motionDetected = true
			lastMotionTime = time.Now()
		}

		if motionDetected && writer != nil {
			writer.Write(resized)
		}

		if motionDetected && time.Since(lastMotionTime) > a.cfg.Recording.PostRecording.Duration {
			logInfo("camera %s stopping recording: %s", cam.ID, currentLocalPath)
			writer.Close()
			writer = nil
			motionDetected = false
			a.afterRecording(cam, currentLocalPath)
		}

		contours.Close()
		gray.CopyTo(&prevGray)
	}
}

// startRecording создаёт файл видеозаписи.
func (a *App) startRecording(cam CameraConfig) (string, *gocv.VideoWriter, error) {
	timestamp := time.Now()
	year := timestamp.Format("2006")
	month := timestamp.Format("01")
	day := timestamp.Format("02")
	dir := cameraDir(a.cfg.Storage.BaseDir, cam, year, month, day)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", nil, err
	}
	filename := timestamp.Format("15-04-05") + ".mp4"
	localPath := filepath.Join(dir, filename)
	writer, err := gocv.VideoWriterFile(localPath, a.cfg.Recording.Codec, a.cfg.Recording.FPS, a.cfg.Recording.FrameWidth, a.cfg.Recording.FrameHeight, true)
	if err != nil {
		return "", nil, err
	}
	return localPath, writer, nil
}

// afterRecording создаёт trigger-файл и запускает WebDAV upload.
func (a *App) afterRecording(cam CameraConfig, localPath string) {
	if a.cfg.NotificationFile.Enabled {
		if err := writeNotificationFile(a.cfg.NotificationFile.Path, localPath); err != nil {
			logWarn("unable to write notification file: %v", err)
		}
	}
	if a.webdav != nil {
		a.uploadWG.Add(1)
		go func() {
			defer a.uploadWG.Done()
			if err := a.webdav.Upload(localPath, a.remotePath(cam, localPath)); err != nil {
				logWarn("webdav upload failed: %v", err)
			}
		}()
	}
}

// remotePath строит путь для WebDAV.
func (a *App) remotePath(cam CameraConfig, localPath string) string {
	filename := filepath.Base(localPath)
	now := time.Now()
	return filepath.ToSlash(filepath.Join(a.cfg.WebDAV.RemoteBaseDir, safeName(cam.NameOrID()), now.Format("2006"), now.Format("01"), now.Format("02"), filename))
}

// writeNotificationFile записывает путь к новому видео для внешней системы.
func writeNotificationFile(path string, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(value), 0644)
}
