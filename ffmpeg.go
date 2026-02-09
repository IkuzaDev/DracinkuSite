package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	ffmpegWindowsURL = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"
	ffmpegLinuxURL   = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz"
	ffmpegMacOSURL   = "https://evermeet.cx/ffmpeg/getrelease/ffmpeg/zip"
)

type SubtitleInfo struct {
	URL         string `json:"url"`
	Language    string `json:"language"`
	DisplayName string `json:"display_name"`
}

type FFmpegManager struct {
	binPath string
	appDir  string
}

func NewFFmpegManager() *FFmpegManager {
	appDir, _ := os.UserHomeDir()
	appDir = filepath.Join(appDir, "DracinkuSite")

	return &FFmpegManager{
		appDir: appDir,
	}
}

// GetFFmpegPath returns the path to ffmpeg binary
func (fm *FFmpegManager) GetFFmpegPath() string {
	if fm.binPath != "" {
		return fm.binPath
	}

	// Determine binary name based on OS
	binaryName := "ffmpeg"
	if runtime.GOOS == "windows" {
		binaryName = "ffmpeg.exe"
	}

	// Check in app directory first
	localPath := filepath.Join(fm.appDir, "ffmpeg", binaryName)
	if _, err := os.Stat(localPath); err == nil {
		fm.binPath = localPath
		return fm.binPath
	}

	// Check in PATH
	path, err := exec.LookPath("ffmpeg")
	if err == nil {
		fm.binPath = path
		return fm.binPath
	}

	return ""
}

// IsAvailable checks if FFmpeg is available
func (fm *FFmpegManager) IsAvailable() bool {
	return fm.GetFFmpegPath() != ""
}

// EnsureFFmpeg downloads FFmpeg if not available (cross-platform)
func (fm *FFmpegManager) EnsureFFmpeg(ctx context.Context, progressFn func(float64)) error {
	if fm.IsAvailable() {
		return nil
	}

	return fm.downloadFFmpeg(ctx, progressFn)
}

// getFFmpegURL returns the download URL for the current platform
func (fm *FFmpegManager) getFFmpegURL() string {
	switch runtime.GOOS {
	case "windows":
		return ffmpegWindowsURL
	case "linux":
		return ffmpegLinuxURL
	case "darwin":
		return ffmpegMacOSURL
	default:
		return ""
	}
}

func (fm *FFmpegManager) downloadFFmpeg(ctx context.Context, progressFn func(float64)) error {
	url := fm.getFFmpegURL()
	if url == "" {
		return fmt.Errorf("FFmpeg auto-download not supported on %s, please install ffmpeg manually", runtime.GOOS)
	}

	// Determine file extension based on URL
	ext := ".zip"
	if strings.Contains(url, ".tar") {
		ext = ".tar.xz"
	}

	// Create temp file for download
	tmpFile, err := os.CreateTemp("", "ffmpeg-*"+ext)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Download FFmpeg
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download ffmpeg: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download ffmpeg: status %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength
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
			_, writeErr := tmpFile.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			if progressFn != nil && totalSize > 0 {
				progressFn(float64(downloaded) / float64(totalSize) * 50) // 50% for download
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	tmpFile.Close()

	// Extract FFmpeg based on platform
	if runtime.GOOS == "linux" {
		return fm.extractTarXZ(tmpFile.Name(), progressFn)
	}
	return fm.extractZip(tmpFile.Name(), progressFn)
}

func (fm *FFmpegManager) extractZip(zipPath string, progressFn func(float64)) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	ffmpegDir := filepath.Join(fm.appDir, "ffmpeg")
	if err := os.MkdirAll(ffmpegDir, 0755); err != nil {
		return err
	}

	binaryName := "ffmpeg"
	if runtime.GOOS == "windows" {
		binaryName = "ffmpeg.exe"
	}

	for _, f := range r.File {
		// We only need the ffmpeg binary
		if !strings.HasSuffix(f.Name, binaryName) {
			continue
		}

		destPath := filepath.Join(ffmpegDir, binaryName)

		outFile, err := os.Create(destPath)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return err
		}

		// Make executable on Unix
		if runtime.GOOS != "windows" {
			os.Chmod(destPath, 0755)
		}

		fm.binPath = destPath
		break
	}

	if progressFn != nil {
		progressFn(100)
	}

	if fm.binPath == "" {
		return fmt.Errorf("ffmpeg binary not found in archive")
	}

	return nil
}

func (fm *FFmpegManager) extractTarXZ(tarPath string, progressFn func(float64)) error {
	// For Linux tar.xz files, we use xz command since Go doesn't have native xz support
	// First decompress xz to tar
	tarFile := strings.TrimSuffix(tarPath, ".xz")

	cmd := exec.Command("xz", "-d", "-k", tarPath)
	if err := cmd.Run(); err != nil {
		// Try using 7z if xz not available
		cmd = exec.Command("7z", "x", "-y", "-o"+filepath.Dir(tarPath), tarPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to decompress xz: please install xz-utils or 7z")
		}
	}
	defer os.Remove(tarFile)

	// Open tar file
	f, err := os.Open(tarFile)
	if err != nil {
		// Try .tar directly
		f, err = os.Open(strings.TrimSuffix(tarPath, ".tar.xz") + ".tar")
		if err != nil {
			return fmt.Errorf("failed to open tar: %w", err)
		}
	}
	defer f.Close()

	tr := tar.NewReader(f)

	ffmpegDir := filepath.Join(fm.appDir, "ffmpeg")
	if err := os.MkdirAll(ffmpegDir, 0755); err != nil {
		return err
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look for ffmpeg binary
		if strings.HasSuffix(header.Name, "/ffmpeg") && header.Typeflag == tar.TypeReg {
			destPath := filepath.Join(ffmpegDir, "ffmpeg")
			outFile, err := os.Create(destPath)
			if err != nil {
				return err
			}

			_, err = io.Copy(outFile, tr)
			outFile.Close()
			if err != nil {
				return err
			}

			os.Chmod(destPath, 0755)
			fm.binPath = destPath
			break
		}
	}

	if progressFn != nil {
		progressFn(100)
	}

	if fm.binPath == "" {
		return fmt.Errorf("ffmpeg not found in tar archive")
	}

	return nil
}

// IsM3U8 checks if URL is an M3U8 stream
func (fm *FFmpegManager) IsM3U8(url string) bool {
	lower := strings.ToLower(url)
	return strings.Contains(lower, ".m3u8") || strings.Contains(lower, "/hls/")
}

// ConvertM3U8 converts M3U8 to MP4 with optional subtitles
func (fm *FFmpegManager) ConvertM3U8(ctx context.Context, inputURL, outputPath string, subtitles []SubtitleInfo, progressFn func(float64)) error {
	ffmpeg := fm.GetFFmpegPath()
	if ffmpeg == "" {
		return fmt.Errorf("ffmpeg not available")
	}

	args := []string{
		"-y",           // Overwrite output
		"-i", inputURL, // Input M3U8
	}

	// Download and add subtitles
	subtitleFiles := []string{}
	for i, sub := range subtitles {
		if sub.URL == "" {
			continue
		}

		// Download subtitle to temp file
		subPath := filepath.Join(os.TempDir(), fmt.Sprintf("sub_%d.srt", i))
		if err := fm.downloadSubtitle(ctx, sub.URL, subPath); err != nil {
			continue // Skip failed subtitles
		}
		subtitleFiles = append(subtitleFiles, subPath)
		defer os.Remove(subPath)

		args = append(args, "-i", subPath)
	}

	// Map streams
	args = append(args, "-map", "0:v", "-map", "0:a?")

	// Map subtitles
	for i := range subtitleFiles {
		args = append(args, "-map", fmt.Sprintf("%d:s", i+1))
	}

	// Codec settings
	args = append(args,
		"-c:v", "copy", // Copy video stream
		"-c:a", "aac", // Convert audio to AAC
		"-c:s", "mov_text", // Subtitle codec for MP4
	)

	// Add subtitle metadata
	for i, sub := range subtitles {
		if i >= len(subtitleFiles) {
			break
		}
		lang := sub.Language
		if lang == "" {
			lang = "und"
		}
		args = append(args,
			fmt.Sprintf("-metadata:s:s:%d", i),
			fmt.Sprintf("language=%s", lang),
		)
		if sub.DisplayName != "" {
			args = append(args,
				fmt.Sprintf("-metadata:s:s:%d", i),
				fmt.Sprintf("title=%s", sub.DisplayName),
			)
		}
	}

	args = append(args, outputPath)

	cmd := exec.CommandContext(ctx, ffmpeg, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	return nil
}

func (fm *FFmpegManager) downloadSubtitle(ctx context.Context, url, outputPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download subtitle: status %d", resp.StatusCode)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}

// ConvertDirect copies a direct video file (no processing)
func (fm *FFmpegManager) ConvertDirect(ctx context.Context, inputURL, outputPath string) error {
	ffmpeg := fm.GetFFmpegPath()
	if ffmpeg == "" {
		return fmt.Errorf("ffmpeg not available")
	}

	args := []string{
		"-y",
		"-i", inputURL,
		"-c", "copy",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, ffmpeg, args...)
	return cmd.Run()
}

// ConvertWithHardsub converts M3U8 to MP4 with burned-in subtitle (hardsub)
func (fm *FFmpegManager) ConvertWithHardsub(ctx context.Context, inputURL, outputPath string, subtitle *SubtitleInfo, progressFn func(status string, progress float64)) error {
	ffmpeg := fm.GetFFmpegPath()
	if ffmpeg == "" {
		return fmt.Errorf("ffmpeg not available")
	}

	// Detect CPU cores - gunakan 1/2 cores untuk low-end device
	numThreads := runtime.NumCPU()
	if numThreads > 4 {
		numThreads = numThreads / 2 // High-end: pakai setengah
	}
	// Low-end (1-2 cores): pakai semua

	args := []string{
		"-y",
		"-hwaccel", "auto",
		"-threads", fmt.Sprintf("%d", numThreads),
		"-progress", "pipe:1",
		"-i", inputURL,
	}

	var subtitlePath string

	if subtitle != nil && subtitle.URL != "" {
		if progressFn != nil {
			progressFn("Downloading subtitle...", 5)
		}

		subtitlePath = filepath.Join(os.TempDir(), "hardsub.srt")
		if err := fm.downloadSubtitle(ctx, subtitle.URL, subtitlePath); err != nil {
			return fmt.Errorf("failed to download subtitle: %w", err)
		}
		defer os.Remove(subtitlePath)

		escapedPath := strings.ReplaceAll(subtitlePath, "\\", "/")
		escapedPath = strings.ReplaceAll(escapedPath, ":", "\\:")

		// OPTIMASI 1: Gunakan preset ultrafast + tune zerolatency
		args = append(args,
			"-vf", fmt.Sprintf("subtitles='%s':force_style='FontSize=16,PrimaryColour=&HFFFFFF&,OutlineColour=&H000000&,Outline=2'", escapedPath),
			"-c:v", "libx264",
			"-preset", "ultrafast", // PERUBAHAN: fast -> ultrafast (5-10x lebih cepat)
			"-tune", "zerolatency", // TAMBAHAN: optimasi untuk speed
			"-crf", "28", // PERUBAHAN: 23->28 (lebih cepat, sedikit turun quality)
			"-maxrate", "2M", // TAMBAHAN: batasi bitrate
			"-bufsize", "4M", // TAMBAHAN: buffer size
		)
	} else {
		// No subtitle - stream copy (instant, no re-encoding)
		args = append(args, "-c:v", "copy")
	}

	// OPTIMASI 2: Copy audio langsung (jangan re-encode)
	args = append(args,
		"-c:a", "copy", // PERUBAHAN: aac -> copy (JAUH lebih cepat)
		"-movflags", "+faststart",
		outputPath,
	)

	cmd := exec.CommandContext(ctx, ffmpeg, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// OPTIMASI 3: Capture stderr untuk error handling
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	if progressFn != nil {
		progressFn("Converting video...", 10)
	}

	// Parse progress
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine untuk progress
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		var progress float64 = 10

		for scanner.Scan() {
			line := scanner.Text()

			// Parse out_time_ms untuk progress lebih akurat
			if strings.HasPrefix(line, "out_time_ms=") {
				timeStr := strings.TrimPrefix(line, "out_time_ms=")
				var timeMs int64
				if _, err := fmt.Sscanf(timeStr, "%d", &timeMs); err == nil {
					// Estimasi: 1800000000 microseconds = 30 menit
					estimatedProgress := float64(timeMs) / 1800000000.0 * 85
					if estimatedProgress > 85 {
						estimatedProgress = 85
					}
					if estimatedProgress > progress {
						progress = estimatedProgress
						if progressFn != nil {
							progressFn("Converting video...", 10+progress)
						}
					}
				}
			}
		}
	}()

	// Goroutine untuk error logging
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			// Log errors jika perlu
			_ = scanner.Text()
		}
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w", err)
	}

	if progressFn != nil {
		progressFn("Complete", 100)
	}

	return nil
}

// ConvertM3U8NoSub converts M3U8 to MP4 without subtitles
func (fm *FFmpegManager) ConvertM3U8NoSub(ctx context.Context, inputURL, outputPath string, progressFn func(status string, progress float64)) error {
	ffmpeg := fm.GetFFmpegPath()
	if ffmpeg == "" {
		return fmt.Errorf("ffmpeg not available")
	}

	args := []string{
		"-y",
		"-i", inputURL,
		"-c:v", "copy",
		"-c:a", "aac",
		"-movflags", "+faststart",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, ffmpeg, args...)

	if progressFn != nil {
		progressFn("Converting video...", 10)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	if progressFn != nil {
		progressFn("Complete", 100)
	}

	return nil
}
