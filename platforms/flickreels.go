package platforms

import (
	"dramawave/models"
)

type FlickReels struct {
	client *WrapperClient
}

func NewFlickReels(language string) *FlickReels {
	return &FlickReels{
		client: NewWrapperClient("flickreels", language),
	}
}

func (f *FlickReels) SetLanguage(language string) {
	f.client.SetLanguage(language)
}

func (f *FlickReels) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetTabList()
}

func (f *FlickReels) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetTabData(tabType, tabKey, positionIndex)
}

func (f *FlickReels) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetTabFeed(pageInfo)
}

func (f *FlickReels) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetSeriesDetail(id, quality)
}

func (f *FlickReels) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.SearchSeries(keyword, pageNo, pageSize)
}
