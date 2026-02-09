package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LibraryItem represents a downloaded video in the library
type LibraryItem struct {
	ID            string     `json:"id"`
	DramaID       string     `json:"dramaId"`
	DramaName     string     `json:"dramaName"`
	DramaCover    string     `json:"dramaCover"`
	EpisodeIndex  int        `json:"episodeIndex"`
	EpisodeName   string     `json:"episodeName"`
	Platform      string     `json:"platform"`
	FilePath      string     `json:"filePath"`
	FileSize      int64      `json:"fileSize"`
	Duration      float64    `json:"duration"`
	HasSubtitles  bool       `json:"hasSubtitles"`
	DownloadedAt  time.Time  `json:"downloadedAt"`
	LastPlayedAt  *time.Time `json:"lastPlayedAt,omitempty"`
	WatchProgress float64    `json:"watchProgress"` // percentage 0-100
}

// Library manages the offline video collection
type Library struct {
	items   map[string]*LibraryItem
	mu      sync.RWMutex
	dataDir string
}

func NewLibrary() *Library {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".dramawave")
	os.MkdirAll(dataDir, 0755)

	lib := &Library{
		items:   make(map[string]*LibraryItem),
		dataDir: dataDir,
	}
	lib.load()
	return lib
}

func (l *Library) getLibraryPath() string {
	return filepath.Join(l.dataDir, "library.json")
}

func (l *Library) load() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := os.ReadFile(l.getLibraryPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var items []*LibraryItem
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	for _, item := range items {
		l.items[item.ID] = item
	}
	return nil
}

func (l *Library) save() error {
	items := make([]*LibraryItem, 0, len(l.items))
	for _, item := range l.items {
		items = append(items, item)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(l.getLibraryPath(), data, 0644)
}

// AddItem adds a new item to the library
func (l *Library) AddItem(item *LibraryItem) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check for duplicates and remove them
	for id, existing := range l.items {
		if existing.DramaID == item.DramaID && existing.EpisodeIndex == item.EpisodeIndex && existing.Platform == item.Platform {
			delete(l.items, id)
		}
	}

	item.DownloadedAt = time.Now()
	l.items[item.ID] = item
	return l.save()
}

// RemoveItem removes an item from the library (optionally delete file)
func (l *Library) RemoveItem(id string, deleteFile bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.items[id]
	if !ok {
		return nil
	}

	if deleteFile && item.FilePath != "" {
		os.Remove(item.FilePath)
	}

	delete(l.items, id)
	return l.save()
}

// GetItem returns a specific library item
func (l *Library) GetItem(id string) *LibraryItem {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.items[id]
}

// GetAllItems returns all library items
func (l *Library) GetAllItems() []*LibraryItem {
	l.mu.RLock()
	defer l.mu.RUnlock()

	items := make([]*LibraryItem, 0, len(l.items))
	for _, item := range l.items {
		// Check if file still exists
		if _, err := os.Stat(item.FilePath); err == nil {
			items = append(items, item)
		}
	}
	return items
}

// GetItemsByDrama returns all episodes for a specific drama
func (l *Library) GetItemsByDrama(dramaID string) []*LibraryItem {
	l.mu.RLock()
	defer l.mu.RUnlock()

	items := make([]*LibraryItem, 0)
	for _, item := range l.items {
		if item.DramaID == dramaID {
			if _, err := os.Stat(item.FilePath); err == nil {
				items = append(items, item)
			}
		}
	}
	return items
}

// UpdateWatchProgress updates the watch progress for an item
func (l *Library) UpdateWatchProgress(id string, progress float64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.items[id]
	if !ok {
		return nil
	}

	now := time.Now()
	item.WatchProgress = progress
	item.LastPlayedAt = &now
	return l.save()
}

// GetRecentlyWatched returns recently watched items
func (l *Library) GetRecentlyWatched(limit int) []*LibraryItem {
	l.mu.RLock()
	defer l.mu.RUnlock()

	items := make([]*LibraryItem, 0)
	for _, item := range l.items {
		if item.LastPlayedAt != nil {
			if _, err := os.Stat(item.FilePath); err == nil {
				items = append(items, item)
			}
		}
	}

	// Sort by last played (newest first) - simple bubble sort for small lists
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].LastPlayedAt.After(*items[i].LastPlayedAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

// GetDownloadDir returns the default download directory
func (l *Library) GetDownloadDir() string {
	home, _ := os.UserHomeDir()
	downloadDir := filepath.Join(home, "Downloads", "Dracinku")
	os.MkdirAll(downloadDir, 0755)
	return downloadDir
}
