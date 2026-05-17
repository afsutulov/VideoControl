package main

import (
	"net/http"
	"os"
	"path"
	"strings"
)

// WebDAVClient загружает записи на любой WebDAV-совместимый ресурс.
type WebDAVClient struct {
	cfg    WebDAVConfig
	client *http.Client
}

// NewWebDAVClient создаёт WebDAV-клиент.
func NewWebDAVClient(cfg WebDAVConfig) *WebDAVClient {
	return &WebDAVClient{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout.Duration},
	}
}

// Upload создаёт удалённые каталоги и загружает файл.
func (w *WebDAVClient) Upload(localPath string, remotePath string) error {
	remotePath = strings.TrimLeft(path.Clean("/"+filepathToSlash(remotePath)), "/")
	remoteDir := path.Dir(remotePath)
	if err := w.CreateDir(remoteDir); err != nil {
		return err
	}

	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	url := strings.TrimRight(w.cfg.URL, "/") + "/" + remotePath
	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return err
	}
	req.SetBasicAuth(w.cfg.Username, w.cfg.Password)

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return statusError("webdav upload", resp.Status)
	}
	logInfo("uploaded to WebDAV: %s", remotePath)
	return nil
}

// CreateDir рекурсивно создаёт директории на WebDAV.
func (w *WebDAVClient) CreateDir(remoteDir string) error {
	parts := strings.Split(filepathToSlash(remoteDir), "/")
	current := ""
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		current += "/" + part
		url := strings.TrimRight(w.cfg.URL, "/") + current
		req, err := http.NewRequest("MKCOL", url, nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(w.cfg.Username, w.cfg.Password)
		resp, err := w.client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusOK {
			return statusError("webdav mkdir", resp.Status)
		}
	}
	return nil
}

// filepathToSlash приводит путь к WebDAV slash-формату.
func filepathToSlash(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}

// statusError возвращает ошибку HTTP-операции.
func statusError(op string, status string) error {
	return &httpStatusError{op: op, status: status}
}

type httpStatusError struct {
	op     string
	status string
}

func (e *httpStatusError) Error() string {
	return e.op + ": " + e.status
}
