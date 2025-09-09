package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/static"
	log "github.com/sirupsen/logrus"
)

//go:embed ui/dist/*
var staticFiles embed.FS

var (
	version string
)

type EmbeddedFileSystem struct {
	http.FileSystem
}

func (e EmbeddedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	return err == nil
}

func NewEmbeddedFS() static.ServeFileSystem {
	sub, err := fs.Sub(staticFiles, "ui/dist")
	if err != nil {
		log.Fatal("Failed to create embedded filesystem:", err)
	}

	return EmbeddedFileSystem{
		FileSystem: http.FS(sub),
	}
}
