package main

// Updated by Antigravity to fix platform signatures

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dramawave/models"
	"dramawave/platforms"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx             context.Context
	dramabox        *platforms.Dramabox
	dramadash       *platforms.DramaDash
	flickreels      *platforms.FlickReels
	freereels       *platforms.FreeReels
	gooshort        *platforms.GooShort
	melolo          *platforms.Melolo
	netshort        *platforms.NetShort
	shortmax        *platforms.ShortMax
	downloadManager *DownloadManager
	library         *Library
	seriesCache     *SeriesCache
	ffmpegManager   *FFmpegManager
	configManager   *ConfigManager
}

func NewApp() *App {
	configManager := NewConfigManager()

	// Set global license key from config
	platforms.SetGlobalLicenseKey(configManager.GetLicenseKey())

	return &App{
		dramabox:        platforms.NewDramabox("id"),
		dramadash:       platforms.NewDramaDash("id"),
		flickreels:      platforms.NewFlickReels("id-ID"),
		freereels:       platforms.NewFreeReels("id-ID"),
		gooshort:        platforms.NewGooShort("in"),
		melolo:          platforms.NewMelolo("id"),
		netshort:        platforms.NewNetShort("id_ID"),
		shortmax:        platforms.NewShortMax("id"),
		downloadManager: NewDownloadManager(),
		library:         NewLibrary(),
		seriesCache:     NewSeriesCache(),
		ffmpegManager:   NewFFmpegManager(),
		configManager:   configManager,
	}
}

// GetLicenseKey returns the current license key
func (a *App) GetLicenseKey() string {
	isValid := a.VerifyLicenseKey(a.configManager.GetLicenseKey())
	if !isValid {
		return ""
	}
	return a.configManager.GetLicenseKey()
}

// SetLicenseKey saves and updates the license key
func (a *App) SetLicenseKey(key string) error {
	if key != "" {
		isValid := a.VerifyLicenseKey(key)
		if !isValid {
			return fmt.Errorf("invalid license key")
		}
	}
	err := a.configManager.SetLicenseKey(key)
	if err == nil {
		platforms.SetGlobalLicenseKey(key)
	}
	return err
}

func (a *App) VerifyLicenseKey(key string) bool {
	if key == "" {
		return false
	}

	platforms.SetGlobalLicenseKey(key)
	_, err := a.dramabox.GetTabList("id")
	if err != nil {
		if strings.Contains(err.Error(), "LICENSE_ERROR") {
			return false
		}
		fmt.Printf("[App] Verify License warning: %v\n", err)
	}

	return true
}

// CheckLicense validates the current license key with API call
func (a *App) CheckLicense() bool {
	key := a.GetLicenseKey()
	if key == "" {
		return false
	}

	_, err := a.dramabox.GetTabList("id")
	if err != nil {
		if strings.Contains(err.Error(), "LICENSE_ERROR") {
			fmt.Printf("[App] License invalid or expired: %v. Revoking.\n", err)
			a.SetLicenseKey("") // Clear key
			return false
		}
		fmt.Printf("[App] Verify License warning: %v\n", err)
	}

	return true
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	go func() {
		runtime.EventsEmit(a.ctx, "ffmpeg_status", "checking")
		err := a.ffmpegManager.EnsureFFmpeg(a.ctx, func(progress float64) {
			runtime.EventsEmit(a.ctx, "ffmpeg_status", "downloading")
			runtime.EventsEmit(a.ctx, "ffmpeg_progress", progress)
		})

		if err != nil {
			fmt.Printf("FFmpeg auto-download: %v\n", err)
			runtime.EventsEmit(a.ctx, "ffmpeg_status", "error")
		} else {
			runtime.EventsEmit(a.ctx, "ffmpeg_status", "ready")
		}
	}()
}

type PlatformInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (a *App) GetPlatformList() []PlatformInfo {
	return []PlatformInfo{
		{ID: "dramabox", Name: "DramaBox"},
		{ID: "dramadash", Name: "DramaDash"},
		{ID: "flickreels", Name: "FlickReels"},
		{ID: "freereels", Name: "FreeReels"},
		{ID: "gooshort", Name: "GooShort"},
		{ID: "melolo", Name: "Melolo"},
		{ID: "netshort", Name: "NetShort"},
		{ID: "shortmax", Name: "ShortMax"},
	}
}

func (a *App) checkLicenseError(err error) {
	if err != nil && strings.Contains(err.Error(), "LICENSE_ERROR") {
		fmt.Printf("[App] Realtime License check failed: %v. Revoking.\n", err)
		a.SetLicenseKey("")
		runtime.EventsEmit(a.ctx, "force_logout", true)
	}
}

func (a *App) GetTabList(platform, language string) ([]models.TabItem, error) {
	var list []models.TabItem
	var err error

	switch platform {
	case "dramabox":
		list, err = a.dramabox.GetTabList(language)
	case "dramadash":
		list, err = a.dramadash.GetTabList(language)
	case "flickreels":
		list, err = a.flickreels.GetTabList(language)
	case "freereels":
		list, err = a.freereels.GetTabList(language)
	case "gooshort":
		list, err = a.gooshort.GetTabList(language)
	case "melolo":
		list, err = a.melolo.GetTabList(language)
	case "netshort":
		list, err = a.netshort.GetTabList(language)
	case "shortmax":
		list, err = a.shortmax.GetTabList(language)
	default:
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}

	if err != nil {
		a.checkLicenseError(err)
		return nil, err
	}
	return list, nil
}

func (a *App) GetSeriesDetail(platform string, seriesID interface{}, quality int, language string) (*models.SeriesDetailResponse, error) {
	// Convert seriesID to string
	sid := fmt.Sprintf("%v", seriesID)

	// Create cache key
	cacheKey := fmt.Sprintf("%s:%s:%d:%s", platform, sid, quality, language)

	// Check cache first
	if cached, found := a.seriesCache.Get(cacheKey); found {
		return cached, nil
	}

	var result *models.SeriesDetailResponse
	var err error

	switch platform {
	case "dramabox":
		result, err = a.dramabox.GetSeriesDetail(sid, quality, language)
	case "dramadash":
		result, err = a.dramadash.GetSeriesDetail(sid, quality, language)
	case "flickreels":
		result, err = a.flickreels.GetSeriesDetail(sid, quality, language)
	case "freereels":
		result, err = a.freereels.GetSeriesDetail(sid, quality, language)
	case "gooshort":
		result, err = a.gooshort.GetSeriesDetail(sid, quality, language)
	case "melolo":
		result, err = a.melolo.GetSeriesDetail(sid, quality, language)
	case "netshort":
		result, err = a.netshort.GetSeriesDetail(sid, language)
	case "shortmax":
		result, err = a.shortmax.GetSeriesDetail(sid, quality, language)
	default:
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}

	if err != nil {
		a.checkLicenseError(err)
		return nil, err
	}

	// Cache successful results
	if result != nil {
		a.seriesCache.Set(cacheKey, result)
	}

	return result, err
}

func (a *App) SearchSeries(platform, keyword string, language string) (*models.SearchResponse, error) {
	var result *models.SearchResponse
	var err error

	switch platform {
	case "dramabox":
		result, err = a.dramabox.SearchSeries(keyword, 1, 20, language)
	case "dramadash":
		result, err = a.dramadash.SearchSeries(keyword, 1, 20, language)
	case "flickreels":
		result, err = a.flickreels.SearchSeries(keyword, 1, 20, language)
	case "freereels":
		result, err = a.freereels.SearchSeries(keyword, 1, 20, language)
	case "gooshort":
		result, err = a.gooshort.SearchSeries(keyword, 1, 20, language)
	case "melolo":
		result, err = a.melolo.SearchSeries(keyword, 1, 20, language)
	case "netshort":
		result, err = a.netshort.SearchSeries(keyword, language)
	case "shortmax":
		result, err = a.shortmax.SearchSeries(keyword, 1, 20, language)
	default:
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}

	if err != nil {
		a.checkLicenseError(err)
		return nil, err
	}
	return result, nil
}

func (a *App) SelectDownloadFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Download Folder",
	})
}

func (a *App) StartDownload(videoURL string, subtitles []SubtitleInfo, embedSubs bool, filename, outputDir string) (string, error) {
	if outputDir == "" {
		home, _ := os.UserHomeDir()
		outputDir = filepath.Join(home, "Downloads", "Dracinku")
	}
	os.MkdirAll(outputDir, 0755)
	return a.downloadManager.StartDownload(videoURL, subtitles, embedSubs, filename, outputDir)
}

func (a *App) GetDownloadProgress(id string) *DownloadProgress {
	return a.downloadManager.GetProgress(id)
}

func (a *App) GetAllDownloads() []*DownloadProgress {
	return a.downloadManager.GetAllProgress()
}

func (a *App) GetTabData(platform string, tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	switch platform {
	case "dramabox":
		return a.dramabox.GetTabData(tabType, tabKey, positionIndex, language)
	case "dramadash":
		return a.dramadash.GetTabData(tabType, tabKey, positionIndex, language)
	case "flickreels":
		return a.flickreels.GetTabData(tabType, tabKey, positionIndex, language)
	case "freereels":
		return a.freereels.GetTabData(tabType, tabKey, positionIndex, language)
	case "gooshort":
		return a.gooshort.GetTabData(tabType, tabKey, positionIndex, language)
	case "melolo":
		return a.melolo.GetTabData(tabType, tabKey, positionIndex, language)
	case "netshort":
		return a.netshort.GetTabData(tabType, tabKey, positionIndex, language)
	case "shortmax":
		return a.shortmax.GetTabData(tabType, tabKey, positionIndex, language)
	default:
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}
}

func (a *App) GetTabFeed(platform string, pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	var result *models.TabFeedResponse
	var err error

	switch platform {
	case "dramabox":
		result, err = a.dramabox.GetTabFeed(pageInfo, language)
	case "dramadash":
		result, err = a.dramadash.GetTabFeed(pageInfo, language)
	case "flickreels":
		result, err = a.flickreels.GetTabFeed(pageInfo, language)
	case "freereels":
		result, err = a.freereels.GetTabFeed(pageInfo, language)
	case "gooshort":
		result, err = a.gooshort.GetTabFeed(pageInfo, language)
	case "melolo":
		result, err = a.melolo.GetTabFeed(pageInfo, language)
	case "netshort":
		result, err = a.netshort.GetTabFeed(pageInfo, language)
	case "shortmax":
		result, err = a.shortmax.GetTabFeed(pageInfo, language)
	default:
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}

	if err != nil {
		a.checkLicenseError(err)
		return nil, err
	}
	return result, nil
}

func (a *App) CancelDownload(id string) error {
	return a.downloadManager.CancelDownload(id)
}

// Library API methods

func (a *App) GetLibraryItems() []*LibraryItem {
	return a.library.GetAllItems()
}

func (a *App) GetLibraryItem(id string) *LibraryItem {
	return a.library.GetItem(id)
}

func (a *App) GetLibraryByDrama(dramaID string) []*LibraryItem {
	return a.library.GetItemsByDrama(dramaID)
}

func (a *App) GetRecentlyWatched() []*LibraryItem {
	return a.library.GetRecentlyWatched(10)
}

func (a *App) RemoveFromLibrary(id string, deleteFile bool) error {
	return a.library.RemoveItem(id, deleteFile)
}

func (a *App) UpdateWatchProgress(id string, progress float64) error {
	return a.library.UpdateWatchProgress(id, progress)
}

func (a *App) AddToLibrary(item *LibraryItem) error {
	return a.library.AddItem(item)
}

func (a *App) GetDownloadDir() string {
	return a.library.GetDownloadDir()
}
