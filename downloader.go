package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type DownloadStatus string

const (
	StatusPending          DownloadStatus = "pending"
	StatusDownloading      DownloadStatus = "downloading"
	StatusConverting       DownloadStatus = "converting"
	StatusCompleted        DownloadStatus = "completed"
	StatusFailed           DownloadStatus = "failed"
	StatusCancelled        DownloadStatus = "cancelled"
	StatusInstallingFFmpeg DownloadStatus = "installing_ffmpeg"
)

type DownloadProgress struct {
	ID              string         `json:"id"`
	Filename        string         `json:"filename"`
	Status          DownloadStatus `json:"status"`
	StatusText      string         `json:"statusText"`
	Progress        float64        `json:"progress"`
	TotalBytes      int64          `json:"totalBytes"`
	DownloadedBytes int64          `json:"downloadedBytes"`
	Error           string         `json:"error,omitempty"`
	OutputPath      string         `json:"outputPath"`
}

type DownloadTask struct {
	ID        string
	VideoURL  string
	Subtitles []SubtitleInfo
	EmbedSubs bool
	Filename  string
	OutputDir string
	Cancel    context.CancelFunc
}

type DownloadManager struct {
	downloads     map[string]*DownloadProgress
	tasks         map[string]*DownloadTask
	mu            sync.RWMutex
	client        *http.Client
	ffmpegManager *FFmpegManager
}

func NewDownloadManager() *DownloadManager {
	return &DownloadManager{
		downloads: make(map[string]*DownloadProgress),
		tasks:     make(map[string]*DownloadTask),
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
		ffmpegManager: NewFFmpegManager(),
	}
}

func (dm *DownloadManager) StartDownload(videoURL string, subtitles []SubtitleInfo, embedSubs bool, filename, outputDir string) (string, error) {
	id := uuid.New().String()

	ctx, cancel := context.WithCancel(context.Background())

	task := &DownloadTask{
		ID:        id,
		VideoURL:  videoURL,
		Subtitles: subtitles,
		EmbedSubs: embedSubs,
		Filename:  filename,
		OutputDir: outputDir,
		Cancel:    cancel,
	}

	progress := &DownloadProgress{
		ID:       id,
		Filename: filename,
		Status:   StatusPending,
		Progress: 0,
	}

	dm.mu.Lock()
	dm.downloads[id] = progress
	dm.tasks[id] = task
	dm.mu.Unlock()

	go dm.processDownload(ctx, task)

	return id, nil
}

func (dm *DownloadManager) processDownload(ctx context.Context, task *DownloadTask) {
	outputPath := filepath.Join(task.OutputDir, task.Filename+".mp4")

	// Check if it's an M3U8 stream
	isM3U8 := dm.ffmpegManager.IsM3U8(task.VideoURL)

	if isM3U8 || task.EmbedSubs {
		// Ensure FFmpeg is available
		if !dm.ffmpegManager.IsAvailable() {
			dm.updateStatusText(task.ID, StatusInstallingFFmpeg, "Installing FFmpeg...")
			err := dm.ffmpegManager.EnsureFFmpeg(ctx, func(progress float64) {
				dm.updateProgress(task.ID, progress*0.1) // 10% for FFmpeg install
			})
			if err != nil {
				dm.setError(task.ID, fmt.Sprintf("FFmpeg installation failed: %v", err))
				return
			}
		}

		// Use hardsub if subtitle is provided
		var selectedSub *SubtitleInfo
		if task.EmbedSubs && len(task.Subtitles) > 0 {
			selectedSub = &task.Subtitles[0] // Use first subtitle for hardsub
		}

		progressFn := func(status string, progress float64) {
			dm.updateStatusText(task.ID, StatusConverting, status)
			dm.updateProgress(task.ID, progress)
		}

		var err error
		if selectedSub != nil {
			dm.updateStatusText(task.ID, StatusConverting, "Converting with hardsub...")
			err = dm.ffmpegManager.ConvertWithHardsub(ctx, task.VideoURL, outputPath, selectedSub, progressFn)
		} else {
			dm.updateStatusText(task.ID, StatusDownloading, "Converting video...")
			err = dm.ffmpegManager.ConvertM3U8NoSub(ctx, task.VideoURL, outputPath, progressFn)
		}

		if err != nil {
			if ctx.Err() == context.Canceled {
				dm.updateStatus(task.ID, StatusCancelled, 0)
			} else {
				dm.setError(task.ID, err.Error())
			}
			return
		}
	} else {
		// Direct download for MP4 files
		dm.updateStatusText(task.ID, StatusDownloading, "Downloading...")

		err := dm.downloadFile(ctx, task.ID, task.VideoURL, outputPath)
		if err != nil {
			if ctx.Err() == context.Canceled {
				dm.updateStatus(task.ID, StatusCancelled, 0)
			} else {
				dm.setError(task.ID, err.Error())
			}
			return
		}

		// Download subtitles separately if not embedding
		if len(task.Subtitles) > 0 && !task.EmbedSubs {
			for i, sub := range task.Subtitles {
				if sub.URL == "" {
					continue
				}
				ext := ".srt"
				lang := sub.Language
				if lang == "" {
					lang = fmt.Sprintf("sub%d", i+1)
				}
				subtitlePath := filepath.Join(task.OutputDir, fmt.Sprintf("%s.%s%s", task.Filename, lang, ext))
				dm.downloadFile(ctx, task.ID, sub.URL, subtitlePath)
			}
		}
	}

	dm.mu.Lock()
	dm.downloads[task.ID].Status = StatusCompleted
	dm.downloads[task.ID].StatusText = "Completed"
	dm.downloads[task.ID].Progress = 100
	dm.downloads[task.ID].OutputPath = outputPath
	dm.mu.Unlock()
}

func (dm *DownloadManager) downloadFile(ctx context.Context, id, url, outputPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := dm.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	totalSize := resp.ContentLength
	dm.mu.Lock()
	dm.downloads[id].TotalBytes = totalSize
	dm.mu.Unlock()

	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := file.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			var progress float64
			if totalSize > 0 {
				progress = float64(downloaded) / float64(totalSize) * 100
			}

			dm.mu.Lock()
			dm.downloads[id].DownloadedBytes = downloaded
			dm.downloads[id].Progress = progress
			dm.mu.Unlock()
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (dm *DownloadManager) updateStatus(id string, status DownloadStatus, progress float64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if p, ok := dm.downloads[id]; ok {
		p.Status = status
		p.Progress = progress
	}
}

func (dm *DownloadManager) updateStatusText(id string, status DownloadStatus, text string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if p, ok := dm.downloads[id]; ok {
		p.Status = status
		p.StatusText = text
	}
}

func (dm *DownloadManager) updateProgress(id string, progress float64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if p, ok := dm.downloads[id]; ok {
		p.Progress = progress
	}
}

func (dm *DownloadManager) setError(id string, errMsg string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if p, ok := dm.downloads[id]; ok {
		p.Status = StatusFailed
		p.StatusText = "Failed"
		p.Error = errMsg
	}
}

func (dm *DownloadManager) GetProgress(id string) *DownloadProgress {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.downloads[id]
}

func (dm *DownloadManager) GetAllProgress() []*DownloadProgress {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	result := make([]*DownloadProgress, 0, len(dm.downloads))
	for _, p := range dm.downloads {
		result = append(result, p)
	}
	return result
}

func (dm *DownloadManager) CancelDownload(id string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if task, ok := dm.tasks[id]; ok {
		task.Cancel()
		return nil
	}
	return fmt.Errorf("download not found: %s", id)
}

// CheckFFmpeg checks if FFmpeg is available
func (dm *DownloadManager) CheckFFmpeg() bool {
	return dm.ffmpegManager.IsAvailable()
}

// InstallFFmpeg downloads and installs FFmpeg
func (dm *DownloadManager) InstallFFmpeg(ctx context.Context) error {
	return dm.ffmpegManager.EnsureFFmpeg(ctx, nil)
}
