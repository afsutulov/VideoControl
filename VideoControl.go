package main

import (
    "fmt"
    "os"
    "time"
    "path/filepath"
    "net/http"
    "strings"
    "sync"

    "gocv.io/x/gocv"
)

const (
    RTSP_URL       = "rtsp://login:password@192.168.1.6:554/media/video2"
    FRAME_WIDTH    = 640
    FRAME_HEIGHT   = 360
    MIN_CONTOUR    = 500
    POST_RECORDING = 5 * time.Second
    WEBDAV_URL     = "https://webdav.yandex.ru"
    WEBDAV_USER    = "e-mail"
    WEBDAV_PASS    = "password"
    WEBDAV_TIMEOUT = 180 * time.Second
)

var wg sync.WaitGroup

func createWebDAVDir(remoteDir string) error {
    dirs := strings.Split(remoteDir, "/")
    path := ""
    for _, dir := range dirs {
	if dir == "" {
	    continue
	}
	path += "/" + dir
	req, err := http.NewRequest("MKCOL", WEBDAV_URL+path, nil)
	if err != nil {
	    return err
	}
	req.SetBasicAuth(WEBDAV_USER, WEBDAV_PASS)

	client := &http.Client{Timeout: WEBDAV_TIMEOUT}
	resp, err := client.Do(req)
	if err != nil {
	    return err
	}
	resp.Body.Close()
    }
    return nil
}

func uploadToWebDAV(localPath, remotePath string) {
    defer wg.Done()
    file, err := os.Open(localPath)
    if err != nil {
	fmt.Println("Error opening file for upload:", err)
	return
    }
    defer file.Close()

    req, err := http.NewRequest("PUT", WEBDAV_URL+"/"+remotePath, file)
    if err != nil {
	fmt.Println("Error creating upload request:", err)
	return
    }
    req.SetBasicAuth(WEBDAV_USER, WEBDAV_PASS)

    client := &http.Client{Timeout: WEBDAV_TIMEOUT}
    resp, err := client.Do(req)
    if err != nil {
	fmt.Println("Error uploading file:", err)
	return
    }
    resp.Body.Close()
    fmt.Println("Successfully uploaded to WebDAV:", remotePath)
}

func main() {
    webcam, err := gocv.OpenVideoCapture(RTSP_URL)
    if err != nil {
	fmt.Println("Error opening video stream:", err)
	return
    }
    defer webcam.Close()

    frame := gocv.NewMat()
    defer frame.Close()

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

    filename := ""

    for {
	if !webcam.Read(&frame) || frame.Empty() {
	    fmt.Println("Error reading frame or empty frame received")
	    break
	}

	gocv.CvtColor(frame, &gray, gocv.ColorBGRToGray)
	if prevGray.Empty() {
	    gray.CopyTo(&prevGray)
	    continue
	}

	gocv.AbsDiff(gray, prevGray, &delta)
	gocv.Threshold(delta, &thresh, 25, 255, gocv.ThresholdBinary)
	contours := gocv.FindContours(thresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)

	if contours.Size() > 0 {
	    for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		if gocv.ContourArea(c) > MIN_CONTOUR {
		    if !motionDetected {
			timestamp := time.Now()
			year := timestamp.Format("2006")
			month := timestamp.Format("01")
			day := timestamp.Format("02")
			dir := filepath.Join("/share", "Камера", year, month, day)
			os.MkdirAll(dir, os.ModePerm)
			filename = timestamp.Format("15-04-05") + ".mp4"
			localPath := filepath.Join(dir, filename)
			writer, err = gocv.VideoWriterFile(localPath, "mp4v", 10, FRAME_WIDTH, FRAME_HEIGHT, true)
			if err != nil {
			    fmt.Println("Error creating video writer:", err)
			    return
			}
			fmt.Println("Started recording:", localPath)
		    }
		    motionDetected = true
		    lastMotionTime = time.Now()
		    break
		}
	    }
	}

	if motionDetected {
	    writer.Write(frame)
	}

	if motionDetected && time.Since(lastMotionTime) > POST_RECORDING {
	    fmt.Println("Stopping recording")
	    writer.Close()
	    motionDetected = false

	    timestamp := lastMotionTime.UTC()
	    year := timestamp.Format("2006")
	    month := timestamp.Format("01")
	    day := timestamp.Format("02")
	    remoteDir := filepath.Join("Камера", year, month, day)
	    remotePath := filepath.Join(remoteDir, filename)
	    localPath := filepath.Join("/share", "Камера", year, month, day, filename)

	    err := createWebDAVDir(remoteDir)
	    if err != nil {
		fmt.Println("Error creating WebDAV directory:", err)
	    }

	    file, err := os.Create("/share/video.save")
	    if err != nil {
		fmt.Println("Unable to create file /share/video.save:", err)
	    }
	    defer file.Close()
	    file.WriteString(localPath)

	    wg.Add(1)
	    go uploadToWebDAV(localPath, remotePath)
	}

	contours.Close()
	gray.CopyTo(&prevGray)
    }

    wg.Wait()
}