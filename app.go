package main

import (
	"context"
	"sync"
)

// App содержит конфигурацию и общие сервисы приложения.
type App struct {
	cfg      Config
	webdav   *WebDAVClient
	cameraWG sync.WaitGroup
	uploadWG sync.WaitGroup
}

// NewApp создаёт приложение.
func NewApp(cfg Config) *App {
	app := &App{cfg: cfg}
	if cfg.WebDAV.Enabled {
		app.webdav = NewWebDAVClient(cfg.WebDAV)
	}
	return app
}

// Run запускает обработчики всех включённых камер.
func (a *App) Run(ctx context.Context) error {
	for _, cam := range a.cfg.Cameras {
		if !cam.Enabled {
			logInfo("camera disabled: %s", cam.ID)
			continue
		}
		cam := cam
		a.cameraWG.Add(1)
		go func() {
			defer a.cameraWG.Done()
			a.runCameraLoop(ctx, cam)
		}()
	}
	<-ctx.Done()
	logInfo("shutdown signal received")
	a.cameraWG.Wait()
	a.uploadWG.Wait()
	return nil
}
