package platforms

import (
	"dramawave/models"
)

type FreeReels struct {
	client *WrapperClient
}

func NewFreeReels(language string) *FreeReels {
	return &FreeReels{
		client: NewWrapperClient("freereels", language),
	}
}

func (f *FreeReels) SetLanguage(language string) {
	f.client.SetLanguage(language)
}

func (f *FreeReels) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetTabList()
}

func (f *FreeReels) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetTabData(tabType, tabKey, positionIndex)
}

func (f *FreeReels) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetTabFeed(pageInfo)
}

func (f *FreeReels) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetSeriesDetail(id, quality)
}

func (f *FreeReels) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.SearchSeries(keyword, pageNo, pageSize)
}
