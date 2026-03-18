package platforms

import (
	"dramawave/models"
)

type RadReels struct {
	client *WrapperClient
}

func NewRadReels(language string) *RadReels {
	return &RadReels{
		client: NewWrapperClient("radreels", language),
	}
}

func (r *RadReels) SetLanguage(language string) {
	r.client.SetLanguage(language)
}

func (r *RadReels) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		r.SetLanguage(language)
	}
	return r.client.GetTabList()
}

func (r *RadReels) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		r.SetLanguage(language)
	}
	return r.client.GetTabData(tabType, tabKey, positionIndex)
}

func (r *RadReels) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		r.SetLanguage(language)
	}
	return r.client.GetTabFeed(pageInfo)
}

func (r *RadReels) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		r.SetLanguage(language)
	}
	return r.client.GetSeriesDetail(id, quality)
}

func (r *RadReels) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		r.SetLanguage(language)
	}
	return r.client.SearchSeries(keyword, pageNo, pageSize)
}
