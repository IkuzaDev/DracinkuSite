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
	flickshort      *platforms.FlickShort
	freereels       *platforms.FreeReels
	goodshort        *platforms.GoodShort
	melolo          *platforms.Melolo
	netshort        *platforms.NetShort
	shortmax        *platforms.ShortMax
	reelbuzz        *platforms.ReelBuzz
	dramamax        *platforms.DramaMax
	anime           *platforms.Anime
	movie           *platforms.Movie
	tv              *platforms.TV
	radreels        *platforms.RadReels
	hishort         *platforms.HiShort
	dramawave       *platforms.DramaWave
	jav             *platforms.Jav
	drakor          *platforms.Drakor
	animev2         *platforms.AnimeV2
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
		flickshort:      platforms.NewFlickShort("in"),
		freereels:       platforms.NewFreeReels("id-ID"),
		goodshort:        platforms.NewGoodShort("in"),
		melolo:          platforms.NewMelolo("id"),
		netshort:        platforms.NewNetShort("id_ID"),
		shortmax:        platforms.NewShortMax("id"),
		reelbuzz:        platforms.NewReelBuzz("in"),
		dramamax:        platforms.NewDramaMax("id"),
		anime:           platforms.NewAnime("id"),
		movie:           platforms.NewMovie("id"),
		tv:              platforms.NewTV("id"),
		radreels:        platforms.NewRadReels("in"),
		hishort:         platforms.NewHiShort("id"),
		dramawave:       platforms.NewDramaWave("id_ID"),
		jav:             platforms.NewJav("id"),
		drakor:          platforms.NewDrakor("id"),
		animev2:         platforms.NewAnimeV2("id"),
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
		{ID: "flickshort", Name: "FlickShort"},
		{ID: "goodshort", Name: "GoodShort"},
		{ID: "melolo", Name: "Melolo"},
		{ID: "netshort", Name: "NetShort"},
		{ID: "shortmax", Name: "ShortMax"},
		{ID: "reelbuzz", Name: "ReelBuzz"},
		{ID: "dramamax", Name: "DramaMax"},
		{ID: "anime", Name: "Anime"},
		{ID: "movie", Name: "Movie"},
		{ID: "tv", Name: "TV"},
		{ID: "radreels", Name: "RadReels"},
		{ID: "hishort", Name: "HiShort"},
		{ID: "dramawave", Name: "DramaWave"},
		{ID: "bjav", Name: "Jav"},
		{ID: "drakor", Name: "Drakor"},
		{ID: "animev2", Name: "AnimeV2"},
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
	case "flickshort":
		list, err = a.flickshort.GetTabList(language)
	case "goodshort":
		list, err = a.goodshort.GetTabList(language)
	case "melolo":
		list, err = a.melolo.GetTabList(language)
	case "netshort":
		list, err = a.netshort.GetTabList(language)
	case "shortmax":
		list, err = a.shortmax.GetTabList(language)
	case "reelbuzz":
		list, err = a.reelbuzz.GetTabList(language)
	case "dramamax":
		list, err = a.dramamax.GetTabList(language)
	case "anime":
		list, err = a.anime.GetTabList(language)
	case "movie":
		list, err = a.movie.GetTabList(language)
	case "tv":
		list, err = a.tv.GetTabList(language)
	case "radreels":
		list, err = a.radreels.GetTabList(language)
	case "hishort":
		list, err = a.hishort.GetTabList(language)
	case "dramawave":
		list, err = a.dramawave.GetTabList(language)
	case "bjav":
		list, err = a.jav.GetTabList(language)
	case "drakor":
		list, err = a.drakor.GetTabList(language)
	case "animev2":
		list, err = a.animev2.GetTabList(language)
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
	case "flickshort":
		result, err = a.flickshort.GetSeriesDetail(sid, quality, language)
	case "goodshort":
		result, err = a.goodshort.GetSeriesDetail(sid, quality, language)
	case "melolo":
		result, err = a.melolo.GetSeriesDetail(sid, quality, language)
	case "netshort":
		result, err = a.netshort.GetSeriesDetail(sid, language)
	case "shortmax":
		result, err = a.shortmax.GetSeriesDetail(sid, quality, language)
	case "reelbuzz":
		result, err = a.reelbuzz.GetSeriesDetail(sid, quality, language)
	case "dramamax":
		result, err = a.dramamax.GetSeriesDetail(sid, quality, language)
	case "anime":
		result, err = a.anime.GetSeriesDetail(sid, quality, language)
	case "movie":
		result, err = a.movie.GetSeriesDetail(sid, quality, language)
	case "tv":
		result, err = a.tv.GetSeriesDetail(sid, quality, language)
	case "radreels":
		result, err = a.radreels.GetSeriesDetail(sid, quality, language)
	case "hishort":
		result, err = a.hishort.GetSeriesDetail(sid, quality, language)
	case "dramawave":
		result, err = a.dramawave.GetSeriesDetail(sid, quality, language)
	case "bjav":
		result, err = a.jav.GetSeriesDetail(sid, quality, language)
	case "drakor":
		result, err = a.drakor.GetSeriesDetail(sid, quality, language)
	case "animev2":
		result, err = a.animev2.GetSeriesDetail(sid, quality, language)
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

func (a *App) SearchSeries(platform, keyword string, page, limit int, language string) (*models.SearchResponse, error) {
	var result *models.SearchResponse
	var err error

	// Default values if 0
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	switch platform {
	case "dramabox":
		result, err = a.dramabox.SearchSeries(keyword, page, limit, language)
	case "dramadash":
		result, err = a.dramadash.SearchSeries(keyword, page, limit, language)
	case "flickreels":
		result, err = a.flickreels.SearchSeries(keyword, page, limit, language)
	case "freereels":
		result, err = a.freereels.SearchSeries(keyword, page, limit, language)
	case "flickshort":
		result, err = a.flickshort.SearchSeries(keyword, page, limit, language)
	case "goodshort":
		result, err = a.goodshort.SearchSeries(keyword, page, limit, language)
	case "melolo":
		result, err = a.melolo.SearchSeries(keyword, page, limit, language)
	case "netshort":
		result, err = a.netshort.SearchSeries(keyword, page, limit, language)
	case "shortmax":
		result, err = a.shortmax.SearchSeries(keyword, page, limit, language)
	case "reelbuzz":
		result, err = a.reelbuzz.SearchSeries(keyword, page, limit, language)
	case "dramamax":
		result, err = a.dramamax.SearchSeries(keyword, page, limit, language)
	case "anime":
		result, err = a.anime.SearchSeries(keyword, page, limit, language)
	case "movie":
		result, err = a.movie.SearchSeries(keyword, page, limit, language)
	case "tv":
		result, err = a.tv.SearchSeries(keyword, page, limit, language)
	case "radreels":
		result, err = a.radreels.SearchSeries(keyword, page, limit, language)
	case "hishort":
		result, err = a.hishort.SearchSeries(keyword, page, limit, language)
	case "dramawave":
		result, err = a.dramawave.SearchSeries(keyword, page, limit, language)
	case "bjav":
		result, err = a.jav.SearchSeries(keyword, page, limit, language)
	case "drakor":
		result, err = a.drakor.SearchSeries(keyword, page, limit, language)
	case "animev2":
		result, err = a.animev2.SearchSeries(keyword, page, limit, language)
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
	case "flickshort":
		return a.flickshort.GetTabData(tabType, tabKey, positionIndex, language)
	case "goodshort":
		return a.goodshort.GetTabData(tabType, tabKey, positionIndex, language)
	case "melolo":
		return a.melolo.GetTabData(tabType, tabKey, positionIndex, language)
	case "netshort":
		return a.netshort.GetTabData(tabType, tabKey, positionIndex, language)
	case "shortmax":
		return a.shortmax.GetTabData(tabType, tabKey, positionIndex, language)
	case "reelbuzz":
		return a.reelbuzz.GetTabData(tabType, tabKey, positionIndex, language)
	case "dramamax":
		return a.dramamax.GetTabData(tabType, tabKey, positionIndex, language)
	case "anime":
		return a.anime.GetTabData(tabType, tabKey, positionIndex, language)
	case "movie":
		return a.movie.GetTabData(tabType, tabKey, positionIndex, language)
	case "tv":
		return a.tv.GetTabData(tabType, tabKey, positionIndex, language)
	case "radreels":
		return a.radreels.GetTabData(tabType, tabKey, positionIndex, language)
	case "hishort":
		return a.hishort.GetTabData(tabType, tabKey, positionIndex, language)
	case "dramawave":
		return a.dramawave.GetTabData(tabType, tabKey, positionIndex, language)
	case "bjav":
		return a.jav.GetTabData(tabType, tabKey, positionIndex, language)
	case "drakor":
		return a.drakor.GetTabData(tabType, tabKey, positionIndex, language)
	case "animev2":
		return a.animev2.GetTabData(tabType, tabKey, positionIndex, language)
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
	case "flickshort":
		result, err = a.flickshort.GetTabFeed(pageInfo, language)
	case "goodshort":
		result, err = a.goodshort.GetTabFeed(pageInfo, language)
	case "melolo":
		result, err = a.melolo.GetTabFeed(pageInfo, language)
	case "netshort":
		result, err = a.netshort.GetTabFeed(pageInfo, language)
	case "shortmax":
		result, err = a.shortmax.GetTabFeed(pageInfo, language)
	case "reelbuzz":
		result, err = a.reelbuzz.GetTabFeed(pageInfo, language)
	case "dramamax":
		result, err = a.dramamax.GetTabFeed(pageInfo, language)
	case "anime":
		result, err = a.anime.GetTabFeed(pageInfo, language)
	case "movie":
		result, err = a.movie.GetTabFeed(pageInfo, language)
	case "tv":
		result, err = a.tv.GetTabFeed(pageInfo, language)
	case "radreels":
		result, err = a.radreels.GetTabFeed(pageInfo, language)
	case "hishort":
		result, err = a.hishort.GetTabFeed(pageInfo, language)
	case "dramawave":
		result, err = a.dramawave.GetTabFeed(pageInfo, language)
	case "bjav":
		result, err = a.jav.GetTabFeed(pageInfo, language)
	case "drakor":
		result, err = a.drakor.GetTabFeed(pageInfo, language)
	case "animev2":
		result, err = a.animev2.GetTabFeed(pageInfo, language)
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
