package main

import (
	"embed"
	"net/http"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// LocalFileHandler serves local files for video playback
type LocalFileHandler struct{}

func (h *LocalFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only handle /localfile/ requests
	if !strings.HasPrefix(r.URL.Path, "/localfile/") {
		http.NotFound(w, r)
		return
	}

	// Extract the file path from URL
	// Format: /localfile/C:/path/to/file.mp4
	filePath := strings.TrimPrefix(r.URL.Path, "/localfile/")

	// Decode URL encoding
	filePath = strings.ReplaceAll(filePath, "%20", " ")
	filePath = strings.ReplaceAll(filePath, "%3A", ":")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Serve the file with proper headers for video streaming
	http.ServeFile(w, r, filePath)
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "Dracinku Downloader App",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: &LocalFileHandler{},
		},
		BackgroundColour: &options.RGBA{R: 18, G: 18, B: 18, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			Theme:                windows.Dark,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
